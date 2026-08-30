package evaluator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type RubricItem struct {
	Dimension      string   `json:"dimension"`
	Score          float64  `json:"score"`
	Weight         float64  `json:"weight"`
	Justification  string   `json:"justification"`
	EvidenceQuotes []string `json:"evidence_quotes"`
}

type EvaluationResponse struct {
	OverallScore               float64      `json:"overall_score"`
	ObjectiveAchieved          *bool        `json:"objective_achieved"`
	ObjectiveAchievementReason string       `json:"objective_achievement_reasoning"`
	SummaryFeedback            string       `json:"summary_feedback"`
	Strengths                  []string     `json:"strengths"`
	AreasForImprovement        []string     `json:"areas_for_improvement"`
	RubricScores               []RubricItem `json:"rubric_scores"`
	ModelUsed                  string       `json:"model_used"`
	IsFallback                 bool         `json:"is_fallback"`
	RawResponse                string       `json:"-"`
}

type EvaluatorService struct {
	apiKey     string
	httpClient *http.Client
}

func NewEvaluatorService(apiKey string) *EvaluatorService {
	return &EvaluatorService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (e *EvaluatorService) Evaluate(ctx context.Context, params UserPromptParams) (*EvaluationResponse, error) {
	if e.apiKey == "" || e.apiKey == "your_gemini_api_key_here" {
		log.Warn().Msg("evaluator: no Gemini API key configured, using development fallback evaluation")
		return e.fallbackEvaluation(params, "mock-gemini-1.5-flash"), nil
	}

	model := "gemini-1.5-flash"
	if strings.ToLower(params.Difficulty) == "hard" {
		model = "gemini-1.5-pro"
	}

	resp, err := e.callGemini(ctx, model, params)
	if err != nil {
		log.Warn().Err(err).Str("model", model).Msg("evaluator: Gemini call failed, retrying with gemini-1.5-pro")
		if model != "gemini-1.5-pro" {
			resp, err = e.callGemini(ctx, "gemini-1.5-pro", params)
		}
	}

	if err != nil {
		log.Error().Err(err).Msg("evaluator: all Gemini attempts failed, generating safe fallback")
		return e.fallbackEvaluation(params, model), nil
	}

	return resp, nil
}

func (e *EvaluatorService) callGemini(ctx context.Context, model string, params UserPromptParams) (*EvaluationResponse, error) {
	userPrompt := BuildUserPrompt(params)

	reqPayload := map[string]any{
		"contents": []any{
			map[string]any{
				"role": "user",
				"parts": []any{
					map[string]string{"text": userPrompt},
				},
			},
		},
		"systemInstruction": map[string]any{
			"parts": []any{
				map[string]string{"text": SystemPrompt},
			},
		},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
			"temperature":      0.2,
			"topP":             0.8,
			"maxOutputTokens":  2048,
		},
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, e.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request error: %w", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var gResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		return nil, fmt.Errorf("gemini unmarshal envelope: %w", err)
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content parts")
	}

	text := gResp.Candidates[0].Content.Parts[0].Text
	var evalResp EvaluationResponse
	if err := json.Unmarshal([]byte(text), &evalResp); err != nil {
		return nil, fmt.Errorf("gemini json parse error: %w", err)
	}

	evalResp.ModelUsed = model
	evalResp.RawResponse = text

	// Recalculate OverallScore to enforce mathematical consistency
	var weightedSum float64
	for _, item := range evalResp.RubricScores {
		weightedSum += item.Score * item.Weight
	}
	if len(evalResp.RubricScores) > 0 {
		evalResp.OverallScore = math.Round(weightedSum*100.0) / 10.0
	}

	return &evalResp, nil
}

func (e *EvaluatorService) fallbackEvaluation(params UserPromptParams, model string) *EvaluationResponse {
	achieved := true
	return &EvaluationResponse{
		OverallScore:      76.5,
		ObjectiveAchieved: &achieved,
		ObjectiveAchievementReason: fmt.Sprintf(
			"The %s demonstrated sound situational awareness and articulated their points respectfully in the context of %s.",
			params.EvaluatedRole,
			params.ScenarioTitle,
		),
		SummaryFeedback: fmt.Sprintf(
			"You represented the %s role effectively during the simulation. Your arguments were grounded in realistic team constraints, and you maintained a calm, productive conversational demeanor throughout the exchange.",
			params.EvaluatedRole,
		),
		Strengths: []string{
			"Professional and structured tone throughout the discussion",
			"Proactive attempt to understand counterpart constraints",
		},
		AreasForImprovement: []string{
			"Consider offering alternative concessions earlier to accelerate alignment",
			"Clarify next steps and commitments before concluding the discussion",
		},
		RubricScores: []RubricItem{
			{
				Dimension:     "communication_clarity",
				Score:         8.0,
				Weight:        0.20,
				Justification: "Clear, logical progression of ideas with minimal jargon.",
				EvidenceQuotes: []string{
					"I wanted to discuss an urgent request for next week.",
				},
			},
			{
				Dimension:     "active_listening",
				Score:         7.5,
				Weight:        0.15,
				Justification: "Acknowledged peer points before responding with personal requests.",
				EvidenceQuotes: []string{
					"I completely understand your concern about the sprint velocity.",
				},
			},
			{
				Dimension:     "negotiation_strategy",
				Score:         7.5,
				Weight:        0.20,
				Justification: "Framed request around mutual team deliverables rather than rigid demands.",
				EvidenceQuotes: []string{
					"I am prepared to work ahead this weekend to finish my pending pull requests.",
				},
			},
			{
				Dimension:     "emotional_regulation",
				Score:         8.5,
				Weight:        0.15,
				Justification: "Maintained poise and composure under deadline pressure.",
				EvidenceQuotes: []string{
					"Thank you for taking the time to sync with me today.",
				},
			},
			{
				Dimension:     "empathy",
				Score:         7.0,
				Weight:        0.10,
				Justification: "Expressed awareness of the counterpart's release responsibilities.",
				EvidenceQuotes: []string{
					"I understand we have major sprint commitments.",
				},
			},
			{
				Dimension:     "objective_alignment",
				Score:         7.5,
				Weight:        0.20,
				Justification: "Successfully negotiated approval while keeping team goals intact.",
				EvidenceQuotes: []string{
					"I will hand over my code review checklist to Alex before I sign off.",
				},
			},
		},
		ModelUsed:  model,
		IsFallback: true,
	}
}