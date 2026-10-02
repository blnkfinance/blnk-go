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
	ApiKey *string
	// BaseURL is the configured base URL. It stays exported so existing code
	// that reads or assigns client.BaseURL keeps compiling.
	//
	// Requests do not use this pointer directly. NewClient and SetBaseURL store
	// a private copy, and the next request adopts a new assignment the same way.
	// Changing Host, Path, or other fields on this URL does not redirect later
	// requests. Prefer SetBaseURL when replacing the base URL.
	BaseURL        *url.URL
	instanceID     string
	mu             sync.RWMutex
	owned          *url.URL
	published      *url.URL
	apiKeyValue    string
	apiKeySet      bool
	publishedKey   *string
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

	// Own a copy so later mutation of the caller's *url.URL cannot redirect
	// requests (and the API key) to another host. BaseURL is a separate copy
	// for existing callers that read the exported field.
	owned := cloneURL(baseURL)
	ensureTrailingSlash(owned)
	published := cloneURL(owned)

	//set default options if not provided
	client := &Client{
		ApiKey:    nil,
		BaseURL:   published,
		owned:     owned,
		published: published,
		options:   DefaultOptions(),
		client:    &http.Client{Timeout: 10 * time.Second},
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
	client.installAPIKey(apiKey)

	return client
}

func (c *Client) SetBaseURL(baseURL *url.URL) {
	copied := cloneURL(baseURL)
	ensureTrailingSlash(copied)
	published := cloneURL(copied)
	c.mu.Lock()
	c.owned = copied
	c.published = published
	c.BaseURL = published
	c.mu.Unlock()
}

// adoptBaseURLLocked keeps Client.BaseURL assignable without letting in-place
// edits of that URL change where requests go. Caller must hold c.mu.
func (c *Client) adoptBaseURLLocked() error {
	if c.BaseURL != c.published {
		if c.BaseURL == nil || c.BaseURL.String() == "" {
			return errors.New("base url is required")
		}
		owned := cloneURL(c.BaseURL)
		ensureTrailingSlash(owned)
		c.owned = owned
		c.published = cloneURL(owned)
		c.BaseURL = c.published
		return nil
	}
	if c.owned == nil {
		return errors.New("base url is required")
	}
	if c.published != nil && c.published.String() != c.owned.String() {
		c.published = cloneURL(c.owned)
		c.BaseURL = c.published
	}
	return nil
}

func (c *Client) installAPIKey(apiKey *string) {
	if apiKey == nil {
		c.apiKeySet = false
		c.apiKeyValue = ""
		c.publishedKey = nil
		c.ApiKey = nil
		return
	}
	c.apiKeyValue = *apiKey
	c.apiKeySet = true
	published := c.apiKeyValue
	c.publishedKey = &published
	c.ApiKey = c.publishedKey
}

// adoptAPIKeyLocked keeps Client.ApiKey assignable without following later
// writes to the caller's string variable. Caller must hold c.mu.
func (c *Client) adoptAPIKeyLocked() {
	if c.ApiKey != c.publishedKey {
		c.installAPIKey(c.ApiKey)
		return
	}
	if c.publishedKey != nil && *c.publishedKey != c.apiKeyValue {
		published := c.apiKeyValue
		c.publishedKey = &published
		c.ApiKey = c.publishedKey
	}
}

func (c *Client) apiKeyForRequest() (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.adoptAPIKeyLocked()
	if !c.apiKeySet {
		return "", false
	}
	return c.apiKeyValue, true
}

func ensureTrailingSlash(u *url.URL) {
	if u == nil || u.String() == "" {
		return
	}
	if u.String()[len(u.String())-1:] != "/" {
		u.Path += "/"
	}
}

func (c *Client) applyInstanceID(u *url.URL) error {
	id := c.instanceID
	if isCloudProxyURL(u) {
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

func (c *Client) snapshotBase() (*url.URL, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.adoptBaseURLLocked(); err != nil {
		return nil, err
	}
	return cloneURL(c.owned), nil
}

// resolveEndpoint joins endpoint onto BaseURL using URL resolution so a missing
// trailing slash on SetBaseURL, or a leading slash on the endpoint, does not
// produce a malformed path. Absolute, scheme-relative, userinfo, and
// base-escaping endpoints are rejected so the client cannot send its API key
// off-host.
func (c *Client) resolveEndpoint(endpoint string) (*url.URL, error) {
	base, err := c.snapshotBase()
	if err != nil {
		return nil, err
	}
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

func isCloudProxyURL(u *url.URL) bool {
	if u == nil {
		return false
	}
	if !strings.EqualFold(u.Host, "api.cloud.blnkfinance.com") {
		return false
	}
	path := u.Path
	return path == "/proxy" || strings.HasPrefix(path, "/proxy/")
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

	u, err := c.resolveEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	//if method is get and opt is not nil, add query params to the url
	if method == http.MethodGet && opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}

		u.RawQuery = q.Encode()
	}

	if err := c.applyInstanceID(u); err != nil {
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
	if key, ok := c.apiKeyForRequest(); ok {
		req.Header.Add("X-Blnk-Key", key)
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
		fmt.Println("in error", err)
		return nil, err
	}

	u, err := c.resolveEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if err := c.applyInstanceID(u); err != nil {
		return nil, err
	}

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, u.String(), io.NopCloser(body))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	//print out the file type
	if key, ok := c.apiKeyForRequest(); ok {
		req.Header.Add("X-Blnk-Key", key)
	}

	return req, nil
}
