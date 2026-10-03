package blnkgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-querystring/query"
)

// CloudProxyBaseURL is the Blnk Cloud Proxy API base. Use it as BaseURL with
// WithInstanceID so Core paths are routed through Cloud:
// https://api.cloud.blnkfinance.com/proxy/{path}?instance_id=YOUR_INSTANCE_ID
const CloudProxyBaseURL = "https://api.cloud.blnkfinance.com/proxy/"

type Client struct {
	// ApiKey is sent as X-Blnk-Key. As in earlier v1 releases, the client keeps
	// the pointer passed to NewClient or SetAPIKey and reads it on each request.
	// Direct edits are not synchronized. While requests may be running, rotate
	// the key with SetAPIKey and a new string instead.
	ApiKey *string
	// BaseURL is the Core or Cloud Proxy origin. As in earlier v1 releases, the
	// client keeps the pointer passed to NewClient or SetBaseURL and reads it on
	// each request. Direct edits are not synchronized. While requests may be
	// running, fail over with SetBaseURL and a new *url.URL instead.
	BaseURL        *url.URL
	instanceID     string
	mu             sync.RWMutex
	options        Options
	client         *http.Client
	Ledger         *LedgerService
	LedgerBalance  *LedgerBalanceService
	Transaction    *TransactionService
	BalanceMonitor *BalanceMonitorService
	Identity       *IdentityService
	Search         *SearchService
	Reconciliation *ReconciliationService
	Metadata       *MetadataService
	Health         *HealthService
	ApiKeys        *ApiKeysService
	Hooks          *HooksService
}

// create a client interface
type ClientInterface interface {
	NewRequest(endpoint, method string, opt interface{}) (*http.Request, error)
	CallWithRetry(req *http.Request, resBody interface{}) (*http.Response, error)
	NewFileUploadRequest(endpoint string, fileParam string, file interface{}, fileName string, fields map[string]string) (*http.Request, error)
}

type service struct {
	client ClientInterface
}

type Options struct {
	RetryCount int
	RetryDelay time.Duration
	Timeout    time.Duration
	Logger     Logger
}

func DefaultOptions() Options {
	return Options{
		RetryCount: 1,
		RetryDelay: defaultRetryDelay,
		Timeout:    time.Second * 10,
		Logger:     NewDefaultLogger(),
	}
}

func NewClient(baseURL *url.URL, apiKey *string, opts ...ClientOption) *Client {
	//if base url is nil or empty, return error
	if baseURL == nil || baseURL.String() == "" {
		panic(errors.New("base url is required"))
	}

	ensureTrailingSlash(baseURL)

	//set default options if not provided
	client := &Client{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		options: DefaultOptions(),
		client: &http.Client{
			Timeout:       10 * time.Second,
			CheckRedirect: refuseCrossHostRedirect,
		},
	}

	//apply options
	for _, opt := range opts {
		opt(client)
		//if options.timeout is set, update the client.client timeout
		if client.options.Timeout != 0 {
			client.client.Timeout = client.options.Timeout
		}
		if client.options.RetryCount == 0 {
			client.options.RetryCount = 1
		}
	}

	//initialize services
	client.Ledger = &LedgerService{client: client}
	client.LedgerBalance = &LedgerBalanceService{client: client}
	client.Transaction = &TransactionService{client: client}
	client.BalanceMonitor = &BalanceMonitorService{client: client}
	client.Identity = &IdentityService{client: client}
	client.Search = &SearchService{client: client}
	client.Reconciliation = &ReconciliationService{client: client}
	client.Metadata = &MetadataService{client: client}
	client.Health = &HealthService{client: client}
	client.ApiKeys = &ApiKeysService{client: client}
	client.Hooks = &HooksService{client: client}

	return client
}

// SetBaseURL replaces the base URL and keeps the pointer, as in earlier v1
// releases. It is safe to call while other goroutines build requests. Pass a
// new *url.URL rather than editing one the client already uses.
func (c *Client) SetBaseURL(baseURL *url.URL) {
	c.mu.Lock()
	c.BaseURL = baseURL
	c.mu.Unlock()
}

// SetAPIKey replaces the key sent as X-Blnk-Key and keeps the pointer, like
// the ApiKey field. It is safe to call while other goroutines build requests.
// Pass a new string rather than editing one the client already uses.
func (c *Client) SetAPIKey(apiKey *string) {
	c.mu.Lock()
	c.ApiKey = apiKey
	c.mu.Unlock()
}

// SetBaseURLAndAPIKey replaces the base URL and API key together, so no
// request can pair the new host with the old key or the old host with the new
// key. Use it for failover to another Core instance that has its own key.
func (c *Client) SetBaseURLAndAPIKey(baseURL *url.URL, apiKey *string) {
	c.mu.Lock()
	c.BaseURL = baseURL
	c.ApiKey = apiKey
	c.mu.Unlock()
}

// requestConfig is one request's view of the client configuration, read under
// a single lock so the host and key always come from the same moment.
type requestConfig struct {
	base   *url.URL
	apiKey string
	hasKey bool
}

func (c *Client) snapshotConfig() (requestConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.BaseURL == nil || c.BaseURL.String() == "" {
		return requestConfig{}, errors.New("base url is required")
	}
	// Copy so URL resolution cannot edit the caller's BaseURL, and so one
	// request validates and uses the same host.
	cfg := requestConfig{base: cloneURL(c.BaseURL)}
	if c.ApiKey != nil {
		cfg.apiKey = *c.ApiKey
		cfg.hasKey = true
	}
	return cfg, nil
}

func ensureTrailingSlash(u *url.URL) {
	if u == nil || u.String() == "" {
		return
	}
	if u.String()[len(u.String())-1:] != "/" {
		u.Path += "/"
	}
}

// applyInstanceID adds instance_id to u. Proxy requests must carry an
// instance_... ID; base is the configured base URL the request was built from.
func (c *Client) applyInstanceID(u, base *url.URL) error {
	id := c.instanceID
	if isProxyBase(base) || isCloudProxyURL(u) {
		if id == "" {
			return errors.New("instance_id is required for Cloud Proxy requests")
		}
		if !strings.HasPrefix(id, "instance_") {
			return errors.New("Cloud instance_id must start with instance_")
		}
	}
	if id == "" {
		return nil
	}
	q := u.Query()
	q.Set("instance_id", id)
	u.RawQuery = q.Encode()
	return nil
}

// requestURL reads the configuration once and resolves endpoint against it, so
// the returned URL and key belong to the same snapshot.
func (c *Client) requestURL(endpoint string) (*url.URL, requestConfig, error) {
	cfg, err := c.snapshotConfig()
	if err != nil {
		return nil, requestConfig{}, err
	}
	u, err := resolveEndpoint(cfg.base, endpoint)
	if err != nil {
		return nil, requestConfig{}, err
	}
	return u, cfg, nil
}

// resolveEndpoint joins endpoint onto base using URL resolution so a missing
// trailing slash on SetBaseURL, or a leading slash on the endpoint, does not
// produce a malformed path. Absolute, scheme-relative, userinfo, and
// base-escaping endpoints are rejected so the client cannot send its API key
// off-host. base is a private copy and may be edited.
func resolveEndpoint(base *url.URL, endpoint string) (*url.URL, error) {
	if strings.Contains(endpoint, "\\") {
		return nil, errors.New("endpoint must be a relative path")
	}
	ref, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if ref.IsAbs() || ref.Scheme != "" || ref.Host != "" || ref.User != nil || ref.Opaque != "" {
		return nil, errors.New("endpoint must be a relative path")
	}
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	if strings.HasPrefix(ref.Path, "/") {
		ref.Path = strings.TrimPrefix(ref.Path, "/")
	}
	ref.Path = trimRedundantProxyPrefix(base.Path, ref.Path)
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != base.Scheme || resolved.Host != base.Host {
		return nil, errors.New("endpoint must be a relative path")
	}
	if !staysWithinBase(base.Path, resolved.Path) {
		return nil, errors.New("endpoint escapes the configured base URL")
	}
	return resolved, nil
}

// refuseCrossHostRedirect only follows redirects that keep the same scheme and
// host, so X-Blnk-Key never goes to another host or downgrades from https to
// http. Go's default client strips Authorization and Cookie on a cross-host
// redirect and leaves custom headers in place.
func refuseCrossHostRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if len(via) == 0 || req.URL == nil || via[len(via)-1].URL == nil {
		return nil
	}
	prev := via[len(via)-1].URL
	if !strings.EqualFold(req.URL.Scheme, prev.Scheme) {
		return errors.New("refusing cross-host redirect: scheme changed")
	}
	if !strings.EqualFold(req.URL.Host, prev.Host) {
		return errors.New("refusing cross-host redirect")
	}
	return nil
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	copied := *u
	if u.User != nil {
		user := *u.User
		copied.User = &user
	}
	return &copied
}

// isProxyBase reports whether the configured base URL points at a Cloud Proxy
// mount (its path ends in /proxy). This is host-independent so enterprise
// Cloud deployments on their own domain are validated the same way.
func isProxyBase(base *url.URL) bool {
	if base == nil {
		return false
	}
	baseDir := strings.TrimSuffix(base.Path, "/")
	return baseDir == "/proxy" || strings.HasSuffix(baseDir, "/proxy")
}

// isCloudProxyURL reports whether a resolved request URL targets the hosted
// Cloud Proxy, for callers whose base URL is the Cloud origin and whose
// endpoint carries the /proxy prefix.
func isCloudProxyURL(u *url.URL) bool {
	if u == nil {
		return false
	}
	nameMatch := strings.EqualFold(u.Hostname(), "api.cloud.blnkfinance.com")
	path := u.Path
	return nameMatch && (path == "/proxy" || strings.HasPrefix(path, "/proxy/"))
}

func trimRedundantProxyPrefix(basePath, refPath string) string {
	baseDir := strings.TrimSuffix(basePath, "/")
	if baseDir != "/proxy" && !strings.HasSuffix(baseDir, "/proxy") {
		return refPath
	}
	if refPath == "proxy" {
		return ""
	}
	return strings.TrimPrefix(refPath, "proxy/")
}

func staysWithinBase(basePath, resolvedPath string) bool {
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	if resolvedPath == strings.TrimSuffix(basePath, "/") {
		return true
	}
	return strings.HasPrefix(resolvedPath, basePath)
}

func (c *Client) NewRequest(endpoint, method string, opt interface{}) (*http.Request, error) {
	//creates and returns a new HTTP request
	//endpoint is the API endpoint
	//method is the HTTP method
	//opt is the request body
	//returns the request and an error if any

	u, cfg, err := c.requestURL(endpoint)
	if err != nil {
		return nil, err
	}

	// For GET, merge opt into any query already on the endpoint. Keys from opt
	// win when both set the same key.
	if method == http.MethodGet && opt != nil {
		values, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for key, vals := range values {
			q[key] = vals
		}
		u.RawQuery = q.Encode()
	}

	if err := c.applyInstanceID(u, cfg.base); err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if method != http.MethodGet && opt != nil {
		var err error
		bodyBytes, err = json.Marshal(opt)
		if err != nil {
			return nil, err
		}
	}

	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	if len(bodyBytes) > 0 {
		req.ContentLength = int64(len(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	//if c has api key, add it to the header
	if cfg.hasKey {
		req.Header.Add("X-Blnk-Key", cfg.apiKey)
	}
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (c *Client) CallWithRetry(req *http.Request, resBody interface{}) (*http.Response, error) {
	maxAttempts := normalizeRetryCount(c.options.RetryCount)
	baseDelay := normalizeRetryDelay(c.options.RetryDelay)
	canRetry := maxAttempts > 1 && isRetryableHTTPMethod(req.Method)

	var lastResp *http.Response
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 && canRetry {
			delay := retryDelayForAttempt(attempt-1, baseDelay)
			c.options.Logger.Info(fmt.Sprintf("Retrying request (attempt %d/%d) after %v", attempt, maxAttempts, delay))
			time.Sleep(delay)
			if err := resetRequestBody(req); err != nil {
				return lastResp, err
			}
		}

		resp, err := c.client.Do(req)
		if err != nil {
			if resp != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			lastErr = err
			c.options.Logger.Info(err.Error())
			if canRetry && attempt < maxAttempts && isRetryableNetworkError(err) {
				continue
			}
			return lastResp, err
		}

		if canRetry && isRetryableHTTPStatus(resp.StatusCode) && attempt < maxAttempts {
			c.options.Logger.Error(fmt.Sprintf("Request failed with status %d; retrying.", resp.StatusCode))
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastResp = resp
			continue
		}

		decodeErr := c.DecodeResponse(resp, resBody)
		resp.Body.Close()
		if decodeErr != nil {
			return resp, decodeErr
		}
		return resp, nil
	}

	if lastErr != nil {
		return lastResp, lastErr
	}
	return lastResp, errors.New("request failed after maximum retry attempts")
}

// decode response, this function will take in a response, and an interface it'll then decode the response body into the interface
// before that it will call checkResponse to check if the response is valid
// the function returns 2 values, the interface and an error if any
// the value passed should be a pointer to a struct
func (c *Client) DecodeResponse(resp *http.Response, v interface{}) error {
	err := c.CheckResponse(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) NewFileUploadRequest(endpoint string, fileParam string, file interface{}, fileName string, fields map[string]string) (*http.Request, error) {
	// Prepare multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file to the form
	var fileReader io.Reader

	switch v := file.(type) {
	case string: // File path
		openedFile, err := os.Open(v)
		if err != nil {
			return nil, err
		}
		defer openedFile.Close()
		fileReader = openedFile
		if fileName == "" {
			fileName = filepath.Base(v)
		}
	case io.Reader: // Read stream
		fileReader = v
		// Default file name
		if fileName == "" {
			fileName = "upload"
		}
	default:
		return nil, fmt.Errorf("unsupported file input type")
	}

	part, err := writer.CreateFormFile(fileParam, fileName)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, fileReader); err != nil {
		return nil, err
	}

	// Add additional form fields
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	u, cfg, err := c.requestURL(endpoint)
	if err != nil {
		return nil, err
	}
	if err := c.applyInstanceID(u, cfg.base); err != nil {
		return nil, err
	}

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, u.String(), io.NopCloser(body))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if cfg.hasKey {
		req.Header.Add("X-Blnk-Key", cfg.apiKey)
	}

	return req, nil
}
