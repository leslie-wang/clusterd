package manager

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Client struct {
	*http.Client
	host string
	port uint
}

func NewClient(host string, port uint) *Client {
	return &Client{
		Client: &http.Client{},
		host:   host,
		port:   port,
	}
}

func (c *Client) makeURL(paths ...string) string {
	return fmt.Sprintf("http://%s:%d%s", c.host, c.port, path.Join(paths...))
}

func (c *Client) addQuery(targetURL string, query map[string]string) string {
	ret := fmt.Sprintf("%s?", targetURL)

	vars := []string{}
	for key, value := range query {
		vars = append(vars, fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(value)))
	}

	return ret + strings.Join(vars, "&")
}
