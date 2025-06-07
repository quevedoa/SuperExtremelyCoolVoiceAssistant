package main

import (
	"ai-assistant/internal/audio"
	"ai-assistant/internal/openai"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	recorder, err := audio.NewAudioRecorder(sampleRate, numChannels)
	if err != nil {
		log.Fatalf("audio setup failed: %v", err)
	}
	defer recorder.Close()

	runAssistant(ctx, openaiClient, recorder)
}

func runAssistant(
	ctx context.Context,
	openaiClient *openai.Client,
	recorder audio.AudioRecorder,
) {
	fmt.Println("Starting Recording...")
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

	fmt.Println("Calling Whisper for Speech to Text...")
	whisperRes, err := openaiClient.TranscribeRawWAV(ctx, wavFileForUpload)
	if err != nil {
		log.Fatalf("error calling whisper model: %v", err)
	}
	fmt.Printf("Prompt: %s\n", whisperRes.Text)

	fmt.Println("Calling ChatGPT API...")
	chatRes, err := openaiClient.GetLLMResponse(ctx, whisperRes.Text)
	if err != nil {
		log.Fatalf("failed to call chat: %v", err)
	}
	fmt.Printf("ChatGPT Response: %s\n", *chatRes)

	f, _ := os.Create("speech.mp3")
	defer f.Close()

	fmt.Println("Converting text to speech...")
	if err := openaiClient.ConvertTextToSpeech(ctx, *chatRes, f); err != nil {
		log.Fatalf("failed converting text to speech: %v", err)
	}

	cmd := exec.Command("mpg123", "-")
	cmd.Stdin = f
	cmd.Stdout = nil // or os.Stdout if you want logs
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		log.Fatalf("playback failed: %v", err)
	}
}
