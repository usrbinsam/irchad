package live

import (
	"fmt"
	"io"
	"os/exec"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
)

type StreamedProcess struct {
	cmd         *exec.Cmd
	Stdout      io.ReadCloser
	publication *lksdk.LocalTrackPublication
}

func NewStreamedProcess(name string, arg ...string) (*StreamedProcess, error) {
	cmd := exec.Command(
		name, arg...,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return &StreamedProcess{
		cmd:    cmd,
		Stdout: stdout,
	}, nil
}

func (p *StreamedProcess) Start() error {
	return p.cmd.Start()
}

func (p *StreamedProcess) Read(b []byte) (int, error) {
	return p.Stdout.Read(b)
}

func (p *StreamedProcess) Stop() error {
	err := p.cmd.Process.Kill()
	if err != nil {
		return err
	}

	return p.cmd.Wait()
}

func (p *StreamedProcess) Close() error {
	_ = p.Stdout.Close()
	return p.Stop()
}

func (p *StreamedProcess) SetMuted(muted bool) {
	if p.publication != nil {
		p.publication.SetMuted(muted)
	}
}

func (p *StreamedProcess) SID() string {
	if p.publication != nil {
		return p.publication.SID()
	}
	return ""
}

func (p *StreamedProcess) Publish(
	room *lksdk.Room,
	mime string,
	duration time.Duration,
	onWriteComplete func(),
	opts *lksdk.TrackPublicationOptions,
) error {
	track, err := lksdk.NewLocalReaderTrack(
		p,
		mime,
		lksdk.ReaderTrackWithFrameDuration(duration),
		lksdk.ReaderTrackWithOnWriteComplete(onWriteComplete),
	)
	if err != nil {
		return err
	}

	pub, err := room.LocalParticipant.PublishTrack(track, opts)
	if err != nil {
		return fmt.Errorf("failed to publish track: %w", err)
	}

	p.publication = pub

	return nil
}
