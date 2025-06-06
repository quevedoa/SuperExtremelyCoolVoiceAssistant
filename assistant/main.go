package main

import (
	"ai-assistant/internal/audio"
	"ai-assistant/internal/openai"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	goaudio "github.com/go-audio/audio"
	wav "github.com/go-audio/wav"
)

const (
	recordSeconds = 5
	sampleRate    = 16000
	numChannels   = 1
	bitsPerSample = 16
)

func main() {
	openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	recorder, err := audio.NewAudioRecorder(sampleRate, numChannels)
	if err != nil {
		log.Fatalf("audio setup failed: %v", err)
	}
	defer recorder.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	runAssistant(ctx, openaiClient, recorder)
}

func runAssistant(
	ctx context.Context,
	openaiClient *openai.Client,
	recorder audio.AudioRecorder,
) {
	if err := recorder.Start(); err != nil {
		log.Fatalf("could not start recorder: %v", err)
		return
	}
	defer recorder.Stop()

	const totalChunks = 5 * 10

	var allSamples []int16

	for i := 0; i < totalChunks; i++ {
		chunk, err := recorder.ReadChunk()
		if err != nil {
			log.Fatalf("error while reading chunk: %v", err)
			break
		}
		allSamples = append(allSamples, chunk...)
	}
	fmt.Printf("Recorded %d total samples (%d frames)", len(allSamples), len(allSamples))

	wavFileName := "recording.wav"
	wavFile, err := os.Create(wavFileName)
	if err != nil {
		log.Fatalf("error creating WAV file on disk: %v", err)
	}
	defer wavFile.Close()

	const audioFormatPCM = 1
	wavEnc := wav.NewEncoder(
		wavFile,
		sampleRate,
		bitsPerSample,
		numChannels,
		audioFormatPCM,
	)

	intBuf := &goaudio.IntBuffer{
		Format: &goaudio.Format{
			NumChannels: numChannels,
			SampleRate:  sampleRate,
		},
		Data:           make([]int, sampleRate*5),
		SourceBitDepth: sampleRate,
	}
	for i, v := range allSamples {
		intBuf.Data[i] = int(v)
	}

	if err := wavEnc.Write(intBuf); err != nil {
		log.Fatalf("error writing WAV data: %v", err)
	}
	if err := wavEnc.Close(); err != nil {
		log.Fatalf("error closing WAV encoder: %v", err)
	}
	fmt.Printf("Saved WAV to %s\n", wavFileName)

	// Llamamos a OpenAI
	wavFileForUpload, err := os.Open(wavFileName)
	if err != nil {
		log.Fatalf("error opening WAV file: %v", err)
	}
	defer wavFileForUpload.Close()

	whisperRes, err := openaiClient.TranscribeRawWAV(ctx, wavFileForUpload)
	if err != nil {
		log.Fatalf("error calling whisper model: %v", err)
	}

	fmt.Println("TRANSCRIPTION RESULTSSS:")
	fmt.Println(whisperRes.Text)

}
