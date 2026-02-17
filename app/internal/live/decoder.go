package live

import (
	"log"
	"os/exec"
	"strings"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264writer"
	"github.com/pion/webrtc/v4/pkg/media/h265writer"
	"github.com/pion/webrtc/v4/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
)

func (l *LiveChat) decodeVideoStream(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) (*VideoStream, error) {
	participantID := rp.Identity()
	trackID := track.ID()

	trackCodec := track.Codec().MimeType
	log.Printf("participant %s is publishing %s - %s", participantID, trackCodec, pub.Name())
	var format string
	switch {
	case strings.EqualFold(webrtc.MimeTypeVP8, trackCodec), strings.EqualFold(webrtc.MimeTypeVP9, trackCodec), strings.EqualFold(webrtc.MimeTypeAV1, trackCodec):
		format = "ivf"
	case strings.EqualFold(webrtc.MimeTypeH264, trackCodec):
		format = "h264"
	case strings.EqualFold(webrtc.MimeTypeH265, trackCodec):
		format = "h265"
	default:
		log.Printf("unsuported codec %s", trackCodec)
		return nil, nil
	}

	cmd := exec.Command(
		"ffmpeg",
		"-f", format,
		"-analyzeduration", "5000000",
		"-probesize", "32000000",
		"-i", "pipe:0",
		"-c:v", "mjpeg",
		"-q:v", "5",
		"-f", "mpjpeg",
		"-boundary_tag", "irchad",
		"pipe:1",
	)

	mediaPipeOut, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	mjpegIn, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// cmd.stderr = os.stdout

	if err := cmd.Start(); err != nil {
		log.Printf("error launching ffmpeg: %s", err.Error())
		return nil, err
	}

	vStream := &VideoStream{
		Name:   pub.Name(),
		stream: mjpegIn,
	}

	l.registry.Add(
		participantID,
		trackID,
		&VideoTrackHandler{stream: vStream},
	)

	go func() {
		defer mediaPipeOut.Close()
		var writer media.Writer

		switch format {
		case "ivf":
			writer, err = ivfwriter.NewWith(mediaPipeOut)
			if err != nil {
				log.Printf("error creating ivfwriter: %s", err)
				return
			}
		case "h264":
			writer = h264writer.NewWith(mediaPipeOut)
		case "h265":
			writer = h265writer.NewWith(mediaPipeOut)
		}

		defer writer.Close()

		for {
			packet, _, err := track.ReadRTP()
			if err != nil {
				log.Printf("track.ReadRTP error: %s", err.Error())
				break
			}

			if err := writer.WriteRTP(packet); err != nil {
				log.Printf("writer error: %s", err.Error())
				break
			}
		}

		l.registry.Remove(participantID, trackID)
	}()

	return vStream, nil
}

func (l *LiveChat) decodeAudioStream(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	// audioIn, audioOut := io.Pipe()

	stream := &OpusStream{}

	participantID := rp.Identity()
	trackID := track.ID()

	l.registry.Add(
		participantID,
		trackID,
		&AudioTrackHandler{stream: stream},
	)
	go func() {
		// defer audioOut.Close()

		ogg, err := oggwriter.NewWith(stream, 48000, track.Codec().Channels)
		if err != nil {
			log.Printf("ogg writer error: %s", err.Error())
			return
		}
		defer ogg.Close()

		for {
			// log.Printf("reading packet from track\n")
			packet, _, err := track.ReadRTP()
			if err != nil {
				break
			}

			// log.Printf("writing packet to ogg writer: %s\n", packet.String())
			if err := ogg.WriteRTP(packet); err != nil {
				break
			}

		}

		l.registry.Remove(participantID, trackID)
	}()
}
