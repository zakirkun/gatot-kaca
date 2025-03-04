package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// AnthropicModel mengimplementasikan interface Model untuk Anthropic
type AnthropicModel struct {
	apiKey    string
	modelName string
	baseURL   string
}

// NewAnthropicModel membuat instance baru AnthropicModel
func NewAnthropicModel(config ModelConfig) (Model, error) {
	if config.APIKey == "" {
		return nil, errors.New("api key diperlukan untuk Anthropic")
	}

	baseURL := "https://api.anthropic.com/v1"
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	return &AnthropicModel{
		apiKey:    config.APIKey,
		modelName: config.ModelName,
		baseURL:   baseURL,
	}, nil
}

// AnthropicRequest adalah struktur permintaan untuk API Anthropic
type AnthropicRequest struct {
	Model         string   `json:"model"`
	Prompt        string   `json:"prompt"`
	MaxTokens     int      `json:"max_tokens_to_sample,omitempty"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"top_p,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
}

// AnthropicResponse adalah struktur respons dari API Anthropic
type AnthropicResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
}

// Generate mengimplementasikan interface Model.Generate untuk Anthropic
func (m *AnthropicModel) Generate(ctx context.Context, req ModelRequest) (ModelResponse, error) {
	// Format prompt untuk Anthropic (Claude mengharapkan format tertentu)
	prompt := fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", req.Prompt)

	// Konversi ModelRequest ke AnthropicRequest
	anthropicReq := AnthropicRequest{
		Model:       m.modelName,
		Prompt:      prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// Serialize request body
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return ModelResponse{}, err
	}

	// Buat HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/complete", m.baseURL),
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return ModelResponse{}, err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", m.apiKey)
	httpReq.Header.Set("Anthropic-Version", "2023-06-01")

	// Kirim request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return ModelResponse{}, err
	}
	defer resp.Body.Close()

	// Baca response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ModelResponse{}, err
	}

	// Periksa status code
	if resp.StatusCode != http.StatusOK {
		return ModelResponse{}, fmt.Errorf("error dari Anthropic API: %s", string(respBody))
	}

	// Unmarshal respons
	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return ModelResponse{}, err
	}

	// Hitung token usage secara kasar (karena Anthropic tidak memberikan info token)
	// Ini hanya perkiraan kasar: ~4 karakter per token
	promptChars := len(req.Prompt)
	completionChars := len(anthropicResp.Completion)

	promptTokens := promptChars / 4
	completionTokens := completionChars / 4

	// Konversi AnthropicResponse ke ModelResponse
	return ModelResponse{
		Text:       anthropicResp.Completion,
		ModelName:  m.modelName,
		Provider:   Anthropic,
		FinishType: anthropicResp.StopReason,
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}, nil
}

// GetProvider mengimplementasikan interface Model.GetProvider
func (m *AnthropicModel) GetProvider() ModelProvider {
	return Anthropic
}

// GetModelName mengimplementasikan interface Model.GetModelName
func (m *AnthropicModel) GetModelName() string {
	return m.modelName
}
