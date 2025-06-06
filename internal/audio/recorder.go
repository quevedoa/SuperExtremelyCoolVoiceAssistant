package audio

import (
	"github.com/gordonklaus/portaudio"
)

type AudioRecorder interface {
	Start() error
	ReadChunk() ([]int16, error)
	Stop() error
	Close() error
}

type PortAudioRecorder struct {
	stream      *portaudio.Stream
	sampleRate  float64
	channels    int
	chunkFrames int
	buffer      []int16
}

func NewAudioRecorder(sampleRate, channels int) (*PortAudioRecorder, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}

	chunkFrames := sampleRate / 10
	buffer := make([]int16, chunkFrames*channels)

	stream, err := portaudio.OpenDefaultStream(
		channels, 0, float64(sampleRate), len(buffer), &buffer,
	)
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}

	return &PortAudioRecorder{
		stream:      stream,
		sampleRate:  float64(sampleRate),
		channels:    channels,
		chunkFrames: int(chunkFrames),
		buffer:      buffer,
	}, nil
}

func (r *PortAudioRecorder) Start() error {
	return r.stream.Start()
}

func (r *PortAudioRecorder) ReadChunk() ([]int16, error) {
	if err := r.stream.Read(); err != nil {
		return nil, err
	}

	chunk := make([]int16, len(r.buffer))
	copy(chunk, r.buffer)
	return chunk, nil
}

func (r *PortAudioRecorder) Stop() error {
	r.stream.Close()
	return portaudio.Terminate()
}

func (r *PortAudioRecorder) Close() error {
	r.stream.Close()
	return portaudio.Terminate()
}
