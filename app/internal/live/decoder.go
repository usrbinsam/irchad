package live

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const maxVideoLate = 1024

type AudioVideoDecoder struct {
	pipeline      *gst.Pipeline
	audioSource   *app.Source
	videoSource   *app.Source
	sink          *app.Sink
	header        []byte
	pliWriterFunc func()
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
	pipelineStr := fmt.Sprintf(`
		appsrc 
			name=video_src 
			min-latency=0 
			do-timestamp=true 
			format=time 
			is-live=true 
			leaky-type=downstream
			caps=application/x-rtp,media=video,clock-rate=90000,encoding-name=H264,payload=125 !
		rtpjitterbuffer latency=100  drop-on-latency=true do-retransmission=true !
		rtph264depay !
		h264parse !
	  queue max-size-time=500000000 leaky=downstream !
		mux.

	  appsrc 
			leaky-type=downstream 
			name=audio_src 
			min-latency=0 
			do-timestamp=true 
			format=time 
			is-live=true 
			caps=application/x-rtp,media=audio,encoding-name=OPUS,clock-rate=48000,payload=111 !
		rtpjitterbuffer latency=100  drop-on-latency=true do-retransmission=true !
	  rtpopusdepay !
		opusdec !
		audioconvert !
		audiorate !
		audioresample !
		opusenc !
	  queue max-size-time=500000000 leaky=downstream !
	  mux.
 
	  matroskamux name=mux streamable=true !
	  appsink name=sink emit-signals=false sync=false
	`,
		videoPayloadType)

	log.Printf("screenshare decoder pipeline: %s", pipelineStr)

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		return nil, err
	}

	// block the pipeline until we get the first keyframe.
	// prevents locking up the matroksa decoder in the frontend
	// if the first thing it sees is a delta frame.
	muxElem, err := pipeline.GetElementByName("mux")
	if err != nil {
		log.Fatalf("could not find muxer: %w", err)
	}

	videoSinkPad := muxElem.GetStaticPad("video_0")
	if videoSinkPad == nil {
		log.Fatalf("no video sink pad")
	}

	videoSinkPad.AddProbe(gst.PadProbeTypeBuffer, func(pad *gst.Pad, info *gst.PadProbeInfo) gst.PadProbeReturn {
		buffer := info.GetBuffer()
		if !buffer.HasFlags(gst.BufferFlagDeltaUnit) {
			return gst.PadProbeRemove
		}
		return gst.PadProbeDrop
	})

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

	srcPad := videoElem.GetStaticPad("src")
	if srcPad == nil {
		log.Fatalf("video_src has no src pad")
	}

	dec := &AudioVideoDecoder{
		pipeline,
		audioSource,
		videoSource,
		sink,
		nil,
		func() { panic("PLI writer func not set") },
	}
	srcPad.AddProbe(
		gst.PadProbeTypeEventUpstream,
		dec.pliCallback,
	)

	return dec, nil
}

func (d *AudioVideoDecoder) SetPLIWriter(pliWriterFunc func()) {
	d.pliWriterFunc = pliWriterFunc
}

func (d *AudioVideoDecoder) pliCallback(pad *gst.Pad, info *gst.PadProbeInfo) gst.PadProbeReturn {
	event := info.GetEvent()
	if event.Type() != gst.EventTypeCustomUpstream {
		return gst.PadProbeOK
	}

	if event.GetStructure() != nil && event.GetStructure().Name() == "GstForceKeyUnit" {
		log.Println("GStreamer wants a keyframe")
		d.pliWriterFunc()
	}

	return gst.PadProbeOK
}

func (d *AudioVideoDecoder) WriteVideoSample(samp *media.Sample) error {
	buffer := gst.NewBufferFromBytes(samp.Data)
	buffer.SetDuration(gst.ClockTime(samp.Duration))
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

const (
	ScreenShareRegistryKey = "SCREEN_SHARE"
	MicrophoneRegistryKey  = "MICROPHONE"
)

func (l *LiveChat) decodeScreenShare(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	var dec *AudioVideoDecoder
	streamHandler, exists := l.registry.Get(rp.Identity(), ScreenShareRegistryKey)
	log.Printf("decodeScreenShare - %s", track.ID())
	if exists {
		_dec, ok := streamHandler.(*AudioVideoDecoder)
		if !ok {
			log.Fatalf("BUG: expected an AudioVideoDecoder, got %+v", streamHandler)
		}
		dec = _dec
		log.Printf("using existing decoder: %v", streamHandler)
	} else {
		log.Println("making NewAudioVideoDecoder")

		newDec, err := NewAudioVideoDecoder()
		if err != nil {
			log.Printf("error creating AudioVideoDecoder: %s", err.Error())
			return
		}
		dec = newDec
		l.registry.Add(rp.Identity(), ScreenShareRegistryKey, dec)
	}

	if track.Kind() == webrtc.RTPCodecTypeVideo {
		log.Printf("video payload type: %d", uint8(track.PayloadType()))
		dec.SetPLIWriter(func() { rp.WritePLI(track.SSRC()) })

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
		// dec.SetAudioPayloadType(track.PayloadType())
		log.Printf("audio payload type: %d", uint8(track.PayloadType()))
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

	// if d.header == nil {
	// 	sample := d.sink.PullSample()
	// 	if sample != nil {
	// 		if len(d.header) == 0 {
	// 			log.Printf("AudioVideoDecoder: got mkv header")
	// 			d.header = append([]byte(nil), sample.GetBuffer().Bytes()...)
	// 		}
	// 	}
	// }

	if d.pliWriterFunc != nil {
		d.pliWriterFunc()
	}

	// if d.header != nil {
	// 	log.Printf("AudioVideoDecoder: writing cached header")
	// 	w.Write(d.header)
	// }

	for {
		sample := d.sink.PullSample()
		if sample == nil {
			if d.sink.IsEOS() {
				return
			}

			continue
		}
		buffer := sample.GetBuffer()
		// log.Printf("Writing: %d bytes | PTS: %v", buffer.GetSize(), buffer.PresentationTimestamp())
		// log.Printf("Buffer received - IsDelta: %v, Size: %d", isDelta, buffer.GetSize())

		// log.Println("writing buffer")
		_, err := w.Write(buffer.Bytes())
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
				buffer.SetDuration(gst.ClockTime(sample.Duration))
				if src.PushBuffer(buffer) != gst.FlowOK {
					break
				}
			}
		}

		l.registry.Remove(participantID, trackID)
	}()

	return vStream, nil
}

type AudioDecoder struct {
	ID       string
	pipeline *gst.Pipeline
	src      *app.Source
	track    *webrtc.TrackRemote
	volume   *gst.Element
}

func (ad *AudioDecoder) DecodeStream() {
	err := ad.pipeline.SetState(gst.StatePlaying)
	if err != nil {
		return
	}

	packetBuffer := make([]byte, 1500)
	for {
		n, _, err := ad.track.Read(packetBuffer)
		if err != nil {
			break
		}

		buf := gst.NewBufferFromBytes(packetBuffer[:n])
		flow := ad.src.PushBuffer(buf)
		if flow != gst.FlowOK {
			break
		}
	}
}

func (ad *AudioDecoder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// AudioDecoder plays direct to the sound system
	http.NotFound(w, r)
}

func NewAudioDecoder(ID string, track *webrtc.TrackRemote) (*AudioDecoder, error) {
	pipelineStr := `
	  appsrc 
	  	name=src 
			do-timestamp=true 
	    format=time 
		  is-live=true
			min-latency=0
			leaky-type=downstream
	    caps=application/x-rtp,media=audio,encoding-name=OPUS,clock-rate=48000,payload=111 !
	  rtpjitterbuffer latency=100 !
	  rtpopusdepay !
		opusdec !
	  audioconvert !
	  volume name=vol volume=1.0 !
	  audioresample !
		queue max-size-time=100000000 !
		autoaudiosink sync=false
	`
	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		log.Fatalf("NewAudioDecoder: invalid gst pipeline: %s", err.Error())
	}

	sourceElem, err := pipeline.GetElementByName("src")
	if err != nil {
		log.Fatalf("no audio source: %s", err.Error())
	}
	src := app.SrcFromElement(sourceElem)

	volumeElem, err := pipeline.GetElementByName("vol")
	if err != nil {
		log.Fatalf("no volume element: %s", err.Error())
	}

	return &AudioDecoder{
		pipeline: pipeline,
		src:      src,
		track:    track,
		volume:   volumeElem,
	}, nil
}

func (ad *AudioDecoder) SetVolume(vol float64) error {
	if vol < 0.0 || vol >= 2.0 {
		return fmt.Errorf("volume must be between 0 and 200")
	}

	err := ad.volume.SetProperty("volume", vol)
	if err != nil {
		return err
	}

	log.Printf("set user volume to %f", vol)
	return nil
}

func (ad *AudioDecoder) Close() {
	ad.pipeline.SetState(gst.StateNull)
}

func (l *LiveChat) decodeAudioStream(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	participantID := rp.Identity()
	trackID := track.ID()

	log.Printf("decoding audio track from %s: %s %s", rp.Identity(), pub.Name(), pub.Source())
	dec, err := NewAudioDecoder(rp.Identity(), track)
	l.registry.Add(
		participantID,
		MicrophoneRegistryKey,
		dec,
	)
	if err != nil {
		log.Printf("NewAudioDecoder() error: %s", err.Error())
		return
	}
	go func() {
		dec.DecodeStream()
		log.Printf("decoding ended for %s: %s %s", rp.Identity(), pub.Name(), pub.Source())
		l.registry.Remove(participantID, trackID)
	}()
}

func (l *LiveChat) SetParticipantVolume(participantID string, vol float64) error {
	handler, ok := l.registry.Get(participantID, MicrophoneRegistryKey)
	if !ok {
		return fmt.Errorf("participant or track not found")
	}

	decoder := handler.(*AudioDecoder)
	return decoder.SetVolume(vol)
}
