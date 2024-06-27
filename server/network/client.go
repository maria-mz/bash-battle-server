package network

import (
	"fmt"
)

type Client struct {
	Token    string
	Username string
	Active   bool
	Stream   *Stream
}

func (client *Client) InfoString() string {
	return fmt.Sprintf("%+v", client)
}
