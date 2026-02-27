package live

import "io"

type VideoStream struct {
	Name   string
	stream io.ReadCloser
}

func (v *VideoStream) Read(p []byte) (int, error) {
	return v.stream.Read(p)
}

func (v *VideoStream) Close() error {
	return v.stream.Close()
}
