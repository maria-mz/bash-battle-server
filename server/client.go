package server

import "fmt"

type client struct {
	token    string
	username string
	active   bool
	stream   *stream
}

func (client *client) InfoString() string {
	return fmt.Sprintf("%+v", client)
}
