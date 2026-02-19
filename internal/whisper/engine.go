package whisper

import "context"

type TranscriptionRequest struct {
	AudioPath string
	ModelPath string
	Language  string
}

type Engine interface {
	Transcribe(ctx context.Context, req TranscriptionRequest) (string, error)
}
