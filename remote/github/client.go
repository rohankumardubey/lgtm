package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	pathBranch = "%srepos/%s/%s/branches/%s/protection"

	// protected branch
	pathBranchStatusCheck = "%srepos/%s/%s/branches/%s/protection/required_status_checks"
)

// Client represents the simple HTTP client for the GitHub API.
type Client struct {
	client *http.Client
	base   string // base url
}

// NewClient returns a client at the specified url.
func NewClient(uri string) *Client {
	return &Client{http.DefaultClient, uri}
}

// NewClientToken returns a client at the specified url that
// authenticates all outbound requests with the given token.
func NewClientToken(uri, token string) *Client {
	config := new(oauth2.Config)
	auther := config.Client(context.TODO(), &oauth2.Token{AccessToken: token})
	return &Client{auther, uri}
}

// SetClient sets the default http client. This should be
// used in conjunction with golang.org/x/oauth2 to
// authenticate requests to the server.
func (c *Client) SetClient(client *http.Client) {
	c.client = client
}

// UpdateBranch enables the branch protection for a specific branch.
func (c *Client) UpdateBranch(owner, name, branch string, in *Branch) error {
	uri := fmt.Sprintf(pathBranch, c.base, owner, name, branch)
	return c.put(uri, in, nil)
}

// GetBranchStatusCheck retrives informations about status checks of protected branch from the GitHub API.
func (c *Client) GetBranchStatusCheck(owner, name, branch string) (*RequiredStatusChecks, error) {
	out := new(RequiredStatusChecks)
	uri := fmt.Sprintf(pathBranchStatusCheck, c.base, owner, name, branch)
	err := c.get(uri, out)
	return out, err
}

// PatchBranchStatusCheck update required status checks of protected branchEnabled for GitHub Apps
func (c *Client) PatchBranchStatusCheck(owner, name, branch string, in *RequiredStatusChecks) error {
	uri := fmt.Sprintf(pathBranchStatusCheck, c.base, owner, name, branch)
	return c.patch(uri, in, nil)
}

//
// http request helper functions
//

// helper function for making an http GET request.
func (c *Client) get(rawurl string, out interface{}) error {
	return c.do(rawurl, "GET", nil, out)
}

// helper function for making an http PUT request.
func (c *Client) put(rawurl string, in, out interface{}) error {
	return c.do(rawurl, "PUT", in, out)
}

// helper function for making an http PATCH request.
func (c *Client) patch(rawurl string, in, out interface{}) error {
	return c.do(rawurl, "PATCH", in, out)
}

// helper function to make an http request
func (c *Client) do(rawurl, method string, in, out interface{}) error {
	// executes the http request and returns the body as
	// and io.ReadCloser
	body, err := c.stream(rawurl, method, in, out)
	if err != nil {
		return err
	}
	defer body.Close()

	// if a json response is expected, parse and return
	// the json response.
	if out != nil {
		return json.NewDecoder(body).Decode(out)
	}
	return nil
}

// helper function to stream an http request
func (c *Client) stream(rawurl, method string, in, out interface{}) (io.ReadCloser, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	// if we are posting or putting data, we need to
	// write it to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(in)
		if err != nil {
			return nil, err
		}
	}

	// creates a new http request to bitbucket.
	req, err := http.NewRequest(method, uri.String(), buf)
	if err != nil {
		return nil, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/vnd.github.loki-preview+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > http.StatusPartialContent {
		defer resp.Body.Close()
		out, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf(string(out))
	}
	return resp.Body, nil
}
