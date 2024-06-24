package server

type client struct {
	token    string
	username string
	active   bool
	stream   *stream
}
