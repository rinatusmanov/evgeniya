// Package client is a JSON-RPC client.
// It invokes request to endpoint by URL and parses response to custom type.
package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/pjrpc/pjrpc/v2"
	"gitlab.com/pjrpc/pjrpc/v2/pjson"
)

// Invoker client interface for using in proto generator.
type Invoker interface {
	Invoke(ctx context.Context, id, method string, params, result interface{}, mods ...Mod) error
}

// Mod is a request modificator.
// Mod calls before Doing request.
type Mod func(req *http.Request)

// ModWithBasicAuth adds basic auth to http request.
func ModWithBasicAuth(username, password string) Mod {
	return func(req *http.Request) {
		req.SetBasicAuth(username, password)
	}
}

// ModWithHeader sets your custom header to http request.
func ModWithHeader(header, value string) Mod {
	return func(req *http.Request) {
		req.Header.Set(header, value)
	}
}

// HTTPDoer interface of http client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client JSON-RPC client.
// Contains url of the endpont, http client and request modificators.
type Client struct {
	URL        string
	HTTPClient HTTPDoer
	Mods       []Mod
	Logger     *log.Logger
}

// New creates new JSON-RPC client with default HTTP client.
func New(endpoint string, mods ...Mod) (*Client, error) {
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint url: %w", err)
	}

	cl := &Client{
		URL:        endpoint,
		HTTPClient: new(http.Client),
		Mods:       mods,
		Logger:     nil,
	}

	return cl, nil
}

// SetLogger sets your io.Writer as error logger of the pjrpc client.
// Also you can create and set your own *log.Logger in Logger field.
func (c *Client) SetLogger(w io.Writer) {
	c.Logger = log.New(w, "[pjrpc-client] ", 0)
}

func (c *Client) logError(place, message string, err error) {
	if c.Logger == nil {
		return
	}

	c.Logger.Printf("(%s): %s: %s\n", place, message, err)
}

func (c *Client) parseResponse(resp *http.Response) (*pjrpc.Response, error) {
	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			c.logError("parseResponse", "failed to close http response body", errClose)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", pjrpc.ErrBadStatusCode, resp.StatusCode)
	}

	ct := resp.Header.Get(pjrpc.ContentTypeHeaderName)
	if !strings.Contains(ct, pjrpc.ContentTypeHeaderValue) {
		return nil, fmt.Errorf("%w: '%s'", pjrpc.ErrWrongContentType, ct)
	}

	jResp := new(pjrpc.Response)
	err := pjson.NewDecoder(resp.Body).Decode(jResp)
	if err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	return jResp, nil
}

// Invoke creates http request with modifacators and do it.
// Takes request id, JSON-RPC method name, sending params and pointer type to unmarshall result.
// Returns inner error or JSON-RPC error.
func (c *Client) Invoke(ctx context.Context, id, method string, params, result interface{}, mods ...Mod) error {
	jsonParams, err := pjson.Marshal(params)
	if err != nil {
		return fmt.Errorf("json.Marshal params: %w", err)
	}

	jsonID, _ := pjson.Marshal(id) // nolint:errcheck // unreal error when encoding regular string.

	jReq := pjrpc.Request{
		JSONRPC: pjrpc.JSONRPCVersion,
		ID:      jsonID,
		Method:  method,
		Params:  jsonParams,
	}

	reqBody := bytes.NewBuffer(nil)
	pjson.NewEncoder(reqBody).Encode(jReq) // nolint:errcheck,gosec // unreal error when encoding regular struct.

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, reqBody)
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}

	for _, mod := range mods {
		mod(req)
	}
	for _, mod := range c.Mods {
		mod(req)
	}

	req.Header.Set(pjrpc.ContentTypeHeaderName, pjrpc.ContentTypeHeaderValue)

	httpResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("do http request: %w", err)
	}

	jResp, err := c.parseResponse(httpResp)
	if err != nil {
		return fmt.Errorf("parseResponse: %w", err)
	}

	if jResp.Error != nil {
		return jResp.Error
	}

	err = pjson.Unmarshal(jResp.Result, result)
	if err != nil {
		return fmt.Errorf("json.Unmarshal JSON-RPC result: %w", err)
	}

	return nil
}
