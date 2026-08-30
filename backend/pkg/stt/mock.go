package stt

import (
	"context"
	"fmt"
)

type MockSTT struct{}

func NewMockSTT() *MockSTT {
	return &MockSTT{}
}

func (m *MockSTT) Transcribe(ctx context.Context, audioA, audioB []byte, speakerA, speakerB string) (*TranscriptResult, error) {
	segsA := []DialogueSegment{
		{
			Speaker:   speakerA,
			Text:      fmt.Sprintf("Hi %s, thank you for taking the time to sync with me today. I understand we have major sprint commitments, but I wanted to discuss an urgent leave request for next week.", speakerB),
			StartSecs: 0.5,
			EndSecs:   5.8,
			Confidence: 0.98,
		},
		{
			Speaker:   speakerA,
			Text:      "I need two consecutive days off—specifically Wednesday and Thursday. To make sure the release stays on schedule, I am prepared to work ahead this weekend to finish my pending pull requests.",
			StartSecs: 11.2,
			EndSecs:   18.4,
			Confidence: 0.97,
		},
		{
			Speaker:   speakerA,
			Text:      "I completely understand your concern about the sprint velocity. How about we compromise on one and a half days, and I will hand over my code review checklist to Alex before I sign off?",
			StartSecs: 26.0,
			EndSecs:   34.5,
			Confidence: 0.99,
		},
	}

	segsB := []DialogueSegment{
		{
			Speaker:   speakerB,
			Text:      fmt.Sprintf("Hi %s. Of course. We are only three weeks from the production release, so things are pretty tight right now. What's on your mind?", speakerA),
			StartSecs: 6.2,
			EndSecs:   10.8,
			Confidence: 0.96,
		},
		{
			Speaker:   speakerB,
			Text:      "Taking two full days off in the middle of this sprint is really tough. We have QA testing on Wednesday, and if your feature has bugs, the release will slip. Can you limit it to one day?",
			StartSecs: 19.0,
			EndSecs:   25.4,
			Confidence: 0.98,
		},
		{
			Speaker:   speakerB,
			Text:      "That handover plan and working ahead this weekend shows good initiative. If you get all critical tickets merged by Tuesday 5 PM, I will approve the leave. Deal?",
			StartSecs: 35.0,
			EndSecs:   42.2,
			Confidence: 0.99,
		},
	}

	res := InterleaveDialogue(segsA, segsB)
	res.Provider = "mock-stt"
	return res, nil
}