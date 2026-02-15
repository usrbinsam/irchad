package live

import (
	"fmt"
	"log"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

func NewLiveChat(url string) *LiveChat {
	return &LiveChat{
		url:          "",
		room:         nil,
		roomCallback: nil,
	}
}

type LiveChat struct {
	url          string
	room         *lksdk.Room
	roomCallback *lksdk.RoomCallback
}

func codecSelector() *mediadevices.CodecSelector {
	x264Params, err := x264.NewParams()
	if err != nil {
		panic(err)
	}
	x264Params.BitRate = 500_000 // 500kbps

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}

	return mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
		mediadevices.WithAudioEncoders(&opusParams),
	)
}

func (l *LiveChat) getToken() string {
	// FIXME: get token from livekit-server wrapper
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzExOTg2MjAsImlkZW50aXR5IjoiY2hhZCIsImlzcyI6ImRldmtleSIsIm5hbWUiOiJjaGFkIiwibmJmIjoxNzcxMTEyMjIwLCJzdWIiOiJjaGFkIiwidmlkZW8iOnsicm9vbSI6IiNjaGFkIiwicm9vbUpvaW4iOnRydWV9fQ.kKrUUVUVVe3OCHw_CGL24LHEzbQyNXR42R_JvDyWXz"
}

func (l *LiveChat) Connect(channelName string) error {
	log.Printf("connecting to %s", channelName)
	// XXX: this might need to be a goroutine?
	room, err := lksdk.ConnectToRoomWithToken(
		l.url,
		l.getToken(),
		l.roomCallback,
	)
	if err != nil {
		log.Printf("failed to connect to channel: %s", err.Error())
		return err
	}
	l.room = room
	log.Printf("connected to %s", channelName)
	return nil
}

func (l *LiveChat) Connected() bool {
	if l.room == nil {
		return false
	}

	return l.room.ConnectionState() != lksdk.ConnectionStateConnected
}

func (l *LiveChat) PublishWebcam() error {
	if !l.Connected() {
		return fmt.Errorf("cannot publish webcam: room not connected")
	}

	stream := getMediaDevices()
	videoTracks := stream.GetVideoTracks()

	if len(videoTracks) == 0 {
		return fmt.Errorf("cannot publish webcam: no webcam detected")
	}

	webcam := videoTracks[0]
	_, err := l.room.LocalParticipant.PublishTrack(
		webcam,
		&lksdk.TrackPublicationOptions{
			Name: "webcam",
		},
	)

	log.Printf("publlished webcam track: %+v", webcam)

	return err
}

func getMediaDevices() mediadevices.MediaStream {
	stream, err := mediadevices.GetUserMedia(
		mediadevices.MediaStreamConstraints{
			Video: func(c *mediadevices.MediaTrackConstraints) {
				c.FrameFormat = prop.FrameFormat(frame.FormatI420)
				c.Width = prop.Int(640)
				c.Height = prop.Int(480)
			},
			Audio: func(c *mediadevices.MediaTrackConstraints) {},
			Codec: codecSelector(),
		},
	)
	if err != nil {
		panic(err)
	}

	return stream
}
