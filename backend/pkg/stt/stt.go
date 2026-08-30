package stt

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// DialogueSegment represents a timestamped utterance by a specific participant.
type DialogueSegment struct {
	Speaker    string  `json:"speaker"`
	SpeakerID  string  `json:"speaker_id"`
	Text       string  `json:"text"`
	StartSecs  float64 `json:"start_secs"`
	EndSecs    float64 `json:"end_secs"`
	Confidence float64 `json:"confidence"`
}

// TranscriptResult holds the full interleaved dialogue and metadata.
type TranscriptResult struct {
	Segments        []DialogueSegment `json:"segments"`
	FullInterleaved string            `json:"full_interleaved"`
	DurationSeconds float64           `json:"duration_seconds"`
	TotalWords      int               `json:"total_words"`
	Provider        string            `json:"provider"`
}

// STTProvider abstracts speech-to-text service implementations.
type STTProvider interface {
	Transcribe(ctx context.Context, audioA, audioB []byte, speakerA, speakerB string) (*TranscriptResult, error)
}

// InterleaveDialogue merges two sets of dialogue segments chronologically.
func InterleaveDialogue(segmentsA, segmentsB []DialogueSegment) *TranscriptResult {
	all := make([]DialogueSegment, 0, len(segmentsA)+len(segmentsB))
	all = append(all, segmentsA...)
	all = append(all, segmentsB...)

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].StartSecs < all[j].StartSecs
	})

	var b strings.Builder
	totalWords := 0
	maxEnd := 0.0

	for _, seg := range all {
		mins := int(seg.StartSecs) / 60
		secs := int(seg.StartSecs) % 60
		cleanText := strings.TrimSpace(seg.Text)
		if cleanText == "" {
			continue
		}

		b.WriteString(fmt.Sprintf("[%02d:%02d] [%s]: %s\n", mins, secs, seg.Speaker, cleanText))
		words := strings.Fields(cleanText)
		totalWords += len(words)

		if seg.EndSecs > maxEnd {
			maxEnd = seg.EndSecs
		}
	}

	return &TranscriptResult{
		Segments:        all,
		FullInterleaved: strings.TrimSpace(b.String()),
		DurationSeconds: maxEnd,
		TotalWords:      totalWords,
	}
}