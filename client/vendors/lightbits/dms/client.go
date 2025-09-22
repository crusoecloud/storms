package dms

import "fmt"

type Client struct{}

func NewClient() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) CloneResource() error {
	return fmt.Errorf("not implemented") //nolint:err113 // wip
}

func (c *Client) GetCloneStatus() error {
	return fmt.Errorf("not implemented") //nolint:err113 // wip
}
