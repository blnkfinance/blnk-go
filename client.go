package blnkgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
)

type Client struct {
	ApiKey  *string
	BaseURL *url.URL
	Options Options
	client  *http.Client
}

type Options struct {
	RetryCount int
	Timeout    time.Duration
	Logger     Logger
}

func DefaultOptions() Options {
	return Options{
		RetryCount: 3,
		Timeout:    time.Second * 10,
		Logger:     NewDefaultLogger(),
	}
}

func NewClient(baseURL *url.URL, apiKey *string, opts ...ClientOption) *Client {
	//set default options if not provided
	client := &Client{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		Options: DefaultOptions(),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
	//if base url is nil or empty, return error
	if baseURL == nil || baseURL.String() == "" {
		panic(errors.New("base url is required"))
	}

	//apply options
	for _, opt := range opts {
		opt(client)
		//if options.timeout is set, update the client.client timeout
		if client.Options.Timeout != 0 {
			client.client.Timeout = client.Options.Timeout
		}
	}

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
func (c *Client) CallWithRetry(endpoint, method string, opt, resBody interface{}) (*http.Response, error) {
	retryCount := c.Options.RetryCount

	var resp *http.Response

	for i := 0; i < retryCount; i++ {
		req, err := c.NewRequest(endpoint, method, opt)
		if err != nil {
			return nil, err
		}
		//print the request
		resp, err = c.client.Do(req)
		//print the resp
		if err != nil {
			c.Options.Logger.Info(err.Error())
			time.Sleep(time.Second * 2)
			continue
		}

		if resp.StatusCode >= 500 {
			logString := fmt.Sprintf("Request failed with status code %v and Status %v", resp.StatusCode, resp.Status)
			c.Options.Logger.Error(logString)
			time.Sleep(time.Second * 2)
			continue
		}

		//check resp
		err = c.DecodeResponse(resp, resBody)
		if err != nil {
			c.Options.Logger.Error(err.Error())
			return resp, err
		}

		return resp, nil
	}

	defer resp.Body.Close()
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