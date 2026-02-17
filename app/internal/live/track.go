package live

import "io"

type OpusStream struct {
	header     []byte
	headerDone bool

	browser chan []byte
}

func (o *OpusStream) Write(b []byte) (int, error) {
	if !o.headerDone {
		o.header = append(o.header, b...)
		o.headerDone = len(o.header) > 80
		// log.Printf("wrote opus header. opus header done = %v\n", o.headerDone)
		return len(b), nil
	}

	if o.browser != nil {
		o.browser <- b
	}
	return len(b), nil
}

func (o *OpusStream) Subscribe() ([]byte, chan []byte) {
	header := append([]byte(nil), o.header...)

	ch := make(chan []byte, 100)

	if o.browser != nil {
		close(o.browser)
	}
	o.browser = ch

	return header, ch
}

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
