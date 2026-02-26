package live

import (
	"fmt"
	"io"
	"log"
	"net/http"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const maxVideoLate = 1024

type AudioVideoDecoder struct {
	pipeline    *gst.Pipeline
	audioSource *app.Source
	videoSource *app.Source
	sink        *app.Sink
}

func (d *AudioVideoDecoder) WriteVideoRTP(packet *rtp.Packet) error {
	raw, err := packet.Marshal()
	if err != nil {
		return err
	}
	buffer := gst.NewBufferFromBytes(raw)
	flowReturn := d.videoSource.PushBuffer(buffer)
	if flowReturn != gst.FlowOK && flowReturn != gst.FlowFlushing {
		return fmt.Errorf("audio_src push failed with %s", flowReturn)
	}

	return nil
}

func NewAudioVideoDecoder() (*AudioVideoDecoder, error) {
	pipelineStr := `
		appsrc name=video_src do-timestamp=true format=time is-live=true caps=application/x-rtp,media=video,clock-rate=90000,encoding-name=H264,payload=125 !
		rtpjitterbuffer !
		rtph264depay !
		h264parse !
	  queue !
		mux.

	  appsrc name=audio_src do-timestamp=true format=time caps=application/x-rtp,media=audio,encoding-name=OPUS,clock-rate=48000,payload=111 !
	  rtpjitterbuffer !
	  rtpopusdepay !
	  opusparse !
	  queue !
	  mux.
 
	  matroskamux name=mux streamable=true !
	  appsink name=sink emit-signals=false sync=false
	`

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		return nil, err
	}

	audioElem, err := pipeline.GetElementByName("audio_src")
	if err != nil {
		log.Fatalf("BUG: error getting audio_src: %s", err.Error())
	}
	audioSource := app.SrcFromElement(audioElem)

	videoElem, err := pipeline.GetElementByName("video_src")
	if err != nil {
		log.Fatalf("BUG: error getting video_src: %s", err.Error())
	}
	videoSource := app.SrcFromElement(videoElem)

	sinkElem, err := pipeline.GetElementByName("sink")
	if err != nil {
		log.Fatalf("BUG: error getting sink: %s", err.Error())
	}
	sink := app.SinkFromElement(sinkElem)

	err = pipeline.SetState(gst.StatePlaying)
	if err != nil {
		log.Fatalf("failed to set state: %s", err.Error())
	}

	return &AudioVideoDecoder{pipeline, audioSource, videoSource, sink}, nil
}

func (d *AudioVideoDecoder) SetAudioPayloadType(pt webrtc.PayloadType) {
	d.audioSource.SetProperty("payload", pt)
}

func (d *AudioVideoDecoder) SetVideoPayloadType(pt webrtc.PayloadType) {
	d.videoSource.SetProperty("payload", pt)
}

func (d *AudioVideoDecoder) WriteVideoSample(samp *media.Sample) error {
	buffer := gst.NewBufferFromBytes(samp.Data)
	buffer.SetDuration(samp.Duration)
	flowReturn := d.videoSource.PushBuffer(buffer)
	if flowReturn != gst.FlowOK {
		return fmt.Errorf("video_src push failed with %s", flowReturn)
	}

	return nil
}

func (d *AudioVideoDecoder) WriteAudioRTP(packet *rtp.Packet) error {
	raw, err := packet.Marshal()
	if err != nil {
		return err
	}
	buffer := gst.NewBufferFromBytes(raw)
	flowReturn := d.audioSource.PushBuffer(buffer)
	if flowReturn != gst.FlowOK && flowReturn != gst.FlowFlushing {
		return fmt.Errorf("audio_src push failed with %s", flowReturn)
	}

	return nil
}

func (d *AudioVideoDecoder) Close() {
	d.pipeline.SetState(gst.StateNull)
}

const ScreenShareRegistryKey = "SCREEN_SHARE"

func (l *LiveChat) decodeScreenShare(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	var dec *AudioVideoDecoder
	streamHandler, exists := l.registry.Get(rp.Identity(), ScreenShareRegistryKey)
	if exists {
		_dec, ok := streamHandler.(*AudioVideoDecoder)
		if !ok {
			log.Fatalf("BUG: expected an AudioVideoDecoder, got %+v", streamHandler)
		}
		dec = _dec
	} else {
		newDec, err := NewAudioVideoDecoder()
		if err != nil {
			log.Printf("error creating AudioVideoDecoder: %s", err.Error())
			return
		}
		dec = newDec
		l.registry.Add(rp.Identity(), ScreenShareRegistryKey, dec)
	}

	if track.Kind() == webrtc.RTPCodecTypeVideo {
		dec.SetVideoPayloadType(track.PayloadType())

		go func() {
			for {
				packet, _, err := track.ReadRTP()
				if err != nil {
					log.Printf("err reading RTP packet from video stream: %s", err.Error())
					break
				}
				err = dec.WriteVideoRTP(packet)
				if err != nil {
					log.Printf("error writing rtp: %s", err.Error())
				}
			}
			log.Printf("video track ended")
			dec.Close()
			l.registry.Remove(rp.Identity(), ScreenShareRegistryKey)
		}()

		ev := ParticipantTrackPublished{
			Identity:  rp.Identity(),
			TrackID:   ScreenShareRegistryKey,
			Source:    pub.Source().String(),
			Kind:      track.Kind().String(),
			TrackName: pub.Name(),
			SubscribeURL: fmt.Sprintf(
				"http://%s/stream?pid=%s&tid=%s",
				l.decoderServer.Addr(),
				rp.Identity(),
				ScreenShareRegistryKey,
			),
		}
		wapp := application.Get()
		wapp.Event.Emit(
			EventParticipantTrackPublished,
			ev,
		)
	} else if track.Kind() == webrtc.RTPCodecTypeAudio {
		dec.SetAudioPayloadType(track.PayloadType())
		go func() {
			for {
				packet, _, err := track.ReadRTP()
				if err != nil {
					log.Printf("err reading rtp from audio stream: %s", err.Error())
					break
				}

				if err := dec.WriteAudioRTP(packet); err != nil {
					log.Printf("WriteAudioRTP failed: %s", err.Error())
					break
				}
			}
			// no pipeline cleanup for audio break
			log.Printf("audio track ended")
		}()
	} else {
		log.Printf("unknown codec type: %s", track.Kind())
	}
}

func (d *AudioVideoDecoder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("new webm stream opened")
	w.Header().Set("content-type", "video/webm")
	w.Header().Set("connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	flusher := w.(http.Flusher)
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		sample := d.sink.PullSample()
		if sample == nil {
			if d.sink.IsEOS() {
				return
			}

			continue
		}
		_, err := w.Write(sample.GetBuffer().Bytes())
		if err != nil {
			log.Printf("error writing webm to browser: %s", err.Error())
			return
		}
		flusher.Flush()
	}
}

func (l *LiveChat) decodeVideoStream(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) (*VideoStream, error) {
	gst.Init(nil)

	participantID := rp.Identity()
	trackID := track.ID()

	trackCodec := track.Codec().MimeType
	if trackCodec != webrtc.MimeTypeH264 {
		return nil, fmt.Errorf("can only decode h264")
	}

	pipelineStr := "appsrc name=irchad format=time is-live=true ! " +
		"h264parse ! " +
		"avdec_h264 !" +
		"videoconvert ! " +
		"queue ! " +
		"jpegenc quality=85 ! " +
		"appsink name=sink sync=false emit-signals=true drop=true max-buffers=1"

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		return nil, fmt.Errorf("pipeline error: %s", err.Error())
	}

	mjpegReader, mjpegWriter := io.Pipe()
	sinkElem, err := pipeline.GetElementByName("sink")
	if sinkElem == nil {
		return nil, fmt.Errorf("sinkElem == nil")
	}
	if err != nil {
		log.Printf("couldn't find sink element: %s", err.Error())
		return nil, err
	}

	sink := app.SinkFromElement(sinkElem)
	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
			sample := s.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			data := sample.GetBuffer().Bytes()
			fmt.Fprintf(mjpegWriter, "--irchad\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(data))
			_, _ = mjpegWriter.Write(data)
			fmt.Fprintf(mjpegWriter, "\r\n")
			return gst.FlowOK
		},
	})

	err = pipeline.SetState(gst.StatePlaying)
	if err != nil {
		log.Printf("couldn't set pipeline to playing: %s", err.Error())
		return nil, err
	}

	vStream := &VideoStream{
		Name:   pub.Name(),
		stream: mjpegReader,
	}

	l.registry.Add(
		participantID,
		trackID,
		&VideoTrackHandler{stream: vStream},
	)

	go func() {
		defer mjpegWriter.Close()
		defer pipeline.SetState(gst.StateNull)
		srcElem, err := pipeline.GetElementByName("irchad")
		if err != nil {
			log.Printf("failed to get src 'irchad' - %s", err.Error())
			return
		}
		src := app.SrcFromElement(srcElem)
		sb := samplebuilder.New(
			maxVideoLate,
			&codecs.H264Packet{},
			track.Codec().ClockRate,
		)
		for {
			packet, _, err := track.ReadRTP()
			if err != nil {
				break
			}

			sb.Push(packet)

			for {
				sample := sb.Pop()
				if sample == nil {
					break
				}

				buffer := gst.NewBufferFromBytes(sample.Data)
				buffer.SetDuration(sample.Duration)
				if src.PushBuffer(buffer) != gst.FlowOK {
					break
				}
			}
		}

		l.registry.Remove(participantID, trackID)
	}()

	return vStream, nil
}

func (l *LiveChat) decodeAudioStream(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	stream := &OpusStream{}

	participantID := rp.Identity()
	trackID := track.ID()

	log.Printf("decoding audio track from %s: %s %s", rp.Identity(), pub.Name(), pub.Source())
	l.registry.Add(
		participantID,
		trackID,
		&AudioTrackHandler{stream: stream},
	)
	go func() {
		ogg, err := oggwriter.NewWith(stream, 48000, track.Codec().Channels)
		if err != nil {
			log.Printf("ogg writer error: %s", err.Error())
			return
		}
		defer ogg.Close()

		for {
			packet, _, err := track.ReadRTP()
			if err != nil {
				break
			}

			if err := ogg.WriteRTP(packet); err != nil {
				break
			}
		}

		log.Printf("decoding ended for %s: %s %s", rp.Identity(), pub.Name(), pub.Source())
		l.registry.Remove(participantID, trackID)
	}()
}
