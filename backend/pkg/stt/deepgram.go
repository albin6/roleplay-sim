package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DeepgramClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewDeepgramClient(apiKey string) *DeepgramClient {
	return &DeepgramClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

type deepgramResponse struct {
	Results struct {
		Channels []struct {
			Alternatives []struct {
				Transcript string  `json:"transcript"`
				Confidence float64 `json:"confidence"`
				Words      []struct {
					Word       string  `json:"word"`
					Start      float64 `json:"start"`
					End        float64 `json:"end"`
					Confidence float64 `json:"confidence"`
				} `json:"words"`
			} `json:"alternatives"`
		} `json:"channels"`
	} `json:"results"`
}

func (d *DeepgramClient) transcribeSingle(ctx context.Context, audio []byte, speaker string) ([]DialogueSegment, error) {
	if len(audio) == 0 {
		return nil, nil
	}

	url := "https://api.deepgram.com/v1/listen?model=nova-2&punctuate=true&utterances=true&words=true&smart_format=true"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(audio))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+d.apiKey)
	req.Header.Set("Content-Type", "audio/webm")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deepgram: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deepgram: status %d: %s", resp.StatusCode, string(body))
	}

	var dgResp deepgramResponse
	if err := json.NewDecoder(resp.Body).Decode(&dgResp); err != nil {
		return nil, fmt.Errorf("deepgram: decode response: %w", err)
	}

	var segments []DialogueSegment
	for _, ch := range dgResp.Results.Channels {
		for _, alt := range ch.Alternatives {
			if len(alt.Words) > 0 {
				first := alt.Words[0]
				last := alt.Words[len(alt.Words)-1]
				segments = append(segments, DialogueSegment{
					Speaker:    speaker,
					Text:       alt.Transcript,
					StartSecs:  first.Start,
					EndSecs:    last.End,
					Confidence: alt.Confidence,
				})
			}
		}
	}

	return segments, nil
}

func (d *DeepgramClient) Transcribe(ctx context.Context, audioA, audioB []byte, speakerA, speakerB string) (*TranscriptResult, error) {
	segsA, err := d.transcribeSingle(ctx, audioA, speakerA)
	if err != nil {
		return nil, err
	}

	segsB, err := d.transcribeSingle(ctx, audioB, speakerB)
	if err != nil {
		return nil, err
	}

	res := InterleaveDialogue(segsA, segsB)
	res.Provider = "deepgram-nova-2"
	return res, nil
}