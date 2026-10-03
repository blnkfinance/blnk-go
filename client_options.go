package blnkgo

import "time"

type ClientOption func(*Client)

func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.options.Logger = logger
	}
}

func WithRetry(count int) ClientOption {
	return func(c *Client) {
		c.options.RetryCount = count
	}
}

func WithRetryDelay(delay time.Duration) ClientOption {
	return func(c *Client) {
		c.options.RetryDelay = delay
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.options.Timeout = timeout
	}
}

// WithInstanceID fixes the Cloud Core instance ID for this client. When
// non-empty, every request includes instance_id as a query parameter. Required
// for Cloud Proxy API calls:
// https://docs.blnkfinance.com/cloud/reference/proxy-api
//
// The ID is immutable after NewClient. Create one client per Core instance.
func WithInstanceID(instanceID string) ClientOption {
	return func(c *Client) {
		c.instanceID = instanceID
	}
}
