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
	"time"

	"github.com/google/go-querystring/query"
)

type Client struct {
	ApiKey         *string
	BaseURL        *url.URL
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
	Timeout    time.Duration
	Logger     Logger
}

func DefaultOptions() Options {
	return Options{
		RetryCount: 1,
		Timeout:    time.Second * 10,
		Logger:     NewDefaultLogger(),
	}
}

func NewClient(baseURL *url.URL, apiKey *string, opts ...ClientOption) *Client {
	//if base url is nil or empty, return error
	if baseURL == nil || baseURL.String() == "" {
		panic(errors.New("base url is required"))
	}

	//check if base url ends with a "/", if it doesnt append it
	if baseURL.String()[len(baseURL.String())-1:] != "/" {
		baseURL.Path += "/"
	}

	//set default options if not provided
	client := &Client{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		options: DefaultOptions(),
		client:  &http.Client{Timeout: 10 * time.Second},
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

	return client
}

func (c *Client) SetBaseURL(baseURL *url.URL) {
	c.BaseURL = baseURL
}

func (c *Client) NewRequest(endpoint, method string, opt interface{}) (*http.Request, error) {
	//creates and returns a new HTTP request
	//endpoint is the API endpoint
	//method is the HTTP method
	//opt is the request body
	//returns the request and an error if any

	u, err := url.Parse(c.BaseURL.String() + endpoint)
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

	var bodyBuf io.ReadWriter

	if method != http.MethodGet && opt != nil {
		bodyBuf = new(bytes.Buffer)
		err := json.NewEncoder(bodyBuf).Encode(opt)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), bodyBuf)
	if err != nil {
		return nil, err
	}

	//if c has api key, add it to the header
	if c.ApiKey != nil {
		req.Header.Add("X-Blnk-Key", *c.ApiKey)
	}
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

// to:Do implement retry strategies
func (c *Client) CallWithRetry(req *http.Request, resBody interface{}) (*http.Response, error) {
	retryCount := c.options.RetryCount

	var resp *http.Response
	var err error

	for i := 0; i < retryCount; i++ {
		resp, err = c.client.Do(req)
		if err != nil {
			c.options.Logger.Info(err.Error())
			time.Sleep(time.Second * 2)
			continue
		}

		defer resp.Body.Close()
		if resp.StatusCode >= 500 {
			logString := fmt.Sprintf("Request failed with status code %v and Status %v", resp.StatusCode, resp.Status)
			c.options.Logger.Error(logString)
			time.Sleep(time.Second * 2)
			continue
		}

		//check resp
		err = c.DecodeResponse(resp, resBody)
		if err != nil {
			c.options.Logger.Error(err.Error())
			return resp, err
		}

		return resp, nil
	}
	return nil, errors.New("max retry count exceeded")
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

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, c.BaseURL.ResolveReference(&url.URL{Path: endpoint}).String(), io.NopCloser(body))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	//print out the file type
	if c.ApiKey != nil {
		req.Header.Add("X-Blnk-Key", *c.ApiKey)
	}

	return req, nil
}
