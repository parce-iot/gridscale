package gridscale

import (
	"bytes"
	"net/http"
)

// Client is the gridscale Client class
type Client struct {
	userID    string
	authToken string
	endpoint  string
}

// NewClient creates a new gridscale Client. You have to provide the gridscale
// API user-id UUID, auth token and the API endpoint URL.
func NewClient(userID, authToken, endpoint string) (*Client, error) {
	return &Client{userID, authToken, endpoint}, nil
}

func (c *Client) httpCall(method, path string, body []byte) (*http.Response, error) {
	var req *http.Request
	if len(body) > 0 {
		req, _ = http.NewRequest(method, c.endpoint+path, bytes.NewBuffer(body))
	} else {
		req, _ = http.NewRequest(method, c.endpoint+path, nil)
	}

	req.Header.Add("X-Auth-UserId", c.userID)
	req.Header.Add("X-Auth-Token", c.authToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	hc := &http.Client{}
	return hc.Do(req)
}

func (c *Client) get(path string) (*http.Response, error) {
	return c.httpCall("GET", path, nil)
}

func (c *Client) put(path string, body []byte) (*http.Response, error) {
	return c.httpCall("PUT", path, body)
}

func (c *Client) patch(path string, body []byte) (*http.Response, error) {
	return c.httpCall("PATCH", path, body)
}

func (c *Client) post(path string, body []byte) (*http.Response, error) {
	return c.httpCall("POST", path, body)
}

func (c *Client) delete(path string, body []byte) (*http.Response, error) {
	return c.httpCall("DELETE", path, body)
}
