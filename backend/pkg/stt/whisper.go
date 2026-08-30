package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type WhisperClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewWhisperClient(apiKey string) *WhisperClient {
	return &WhisperClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type whisperResponse struct {
	Segments []struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Text  string  `json:"text"`
	} `json:"segments"`
}

func (w *WhisperClient) transcribeSingle(ctx context.Context, audio []byte, speaker string) ([]DialogueSegment, error) {
	if len(audio) == 0 {
		return nil, nil
	}

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("file", "audio.webm")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(audio); err != nil {
		return nil, err
	}

	_ = writer.WriteField("model", "whisper-1")
	_ = writer.WriteField("response_format", "verbose_json")
	_ = writer.WriteField("language", "en")
	_ = writer.WriteField("temperature", "0")
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/transcriptions", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("whisper: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("whisper: status %d: %s", resp.StatusCode, string(body))
	}

	var wResp whisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&wResp); err != nil {
		return nil, fmt.Errorf("whisper: decode response: %w", err)
	}

	var segments []DialogueSegment
	for _, s := range wResp.Segments {
		segments = append(segments, DialogueSegment{
			Speaker:    speaker,
			Text:       s.Text,
			StartSecs:  s.Start,
			EndSecs:    s.End,
			Confidence: 0.95,
		})
	}

	return segments, nil
}

func (w *WhisperClient) Transcribe(ctx context.Context, audioA, audioB []byte, speakerA, speakerB string) (*TranscriptResult, error) {
	segsA, err := w.transcribeSingle(ctx, audioA, speakerA)
	if err != nil {
		return nil, err
	}

	segsB, err := w.transcribeSingle(ctx, audioB, speakerB)
	if err != nil {
		return nil, err
	}

	res := InterleaveDialogue(segsA, segsB)
	res.Provider = "openai-whisper-1"
	return res, nil
}