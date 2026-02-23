package live

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"syscall"
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
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	// Try graceful termination first
	_ = p.cmd.Process.Signal(syscall.SIGTERM)

	// Wait for process to exit or kill it after a timeout
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(500 * time.Millisecond):
		log.Printf("ffmpeg did not exit after SIGTERM, killing")
		err := p.cmd.Process.Kill()
		if err != nil {
			return err
		}
		return <-done
	}
}

func (p *StreamedProcess) Close() error {
	log.Printf("StreamedProcess closing: %+v", p)
	if p.Stdout != nil {
		err := p.Stdout.Close()
		if err != nil {
			log.Printf("error closing stdout: %s", err.Error())
		}
	}
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
