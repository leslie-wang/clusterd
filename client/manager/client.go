package manager

import (
	"fmt"
	"net/http"
	"path"
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
