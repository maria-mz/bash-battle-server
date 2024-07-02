package network

import "fmt"

type ClientMeta struct {
	Active bool
}

type Client struct {
	Token    string
	Username string
	Stream   *Stream

	meta ClientMeta // TODO: make public ? or read only
}

func (client *Client) InfoString() string {
	return fmt.Sprintf("%+v", client)
}
