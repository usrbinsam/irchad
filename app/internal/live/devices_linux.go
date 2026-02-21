package live

func ffmpegWebcamCapture() (*StreamedProcess, error) {
	return NewStreamedProcess(
		"ffmpeg",
		"-f",
		"v4l2",
		"-i",
		"/dev/video0",
		"-c:v",
		"libvpx",
		"-b:v",
		"2M",
		"-deadline",
		"realtime",
		"-f",
		"ivf",
		"pipe:1",
	)
}

func ffmpegMicCapture() (*StreamedProcess, error) {
	return NewStreamedProcess(
		"ffmpeg",
		"-f",
		"pulse",
		"-i",
		"default",
		"-c:a",
		"libopus",
		"-b:a",
		"64k",
		"-vbr",
		"on",
		"-f",
		"opus",
		"pipe:1",
	)
}
