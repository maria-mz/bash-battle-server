package client

import (
	"fmt"

	"github.com/maria-mz/bash-battle-server/server/stream"
)

type Client struct {
	Token    string
	Username string
	Active   bool
	Stream   *stream.Stream
}

func (client *Client) InfoString() string {
	return fmt.Sprintf("%+v", client)
}
