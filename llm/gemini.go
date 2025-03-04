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

// GeminiModel mengimplementasikan interface Model untuk Google Gemini
type GeminiModel struct {
	apiKey    string
	modelName string
	baseURL   string
}

// NewGeminiModel membuat instance baru GeminiModel
func NewGeminiModel(config ModelConfig) (Model, error) {
	if config.APIKey == "" {
		return nil, errors.New("api key diperlukan untuk Gemini")
	}

	// Default base URL untuk Gemini API
	baseURL := "https://generativelanguage.googleapis.com/v1"
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	return &GeminiModel{
		apiKey:    config.APIKey,
		modelName: config.ModelName,
		baseURL:   baseURL,
	}, nil
}

// GeminiRequest adalah struktur permintaan untuk API Gemini
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent merepresentasikan konten dalam permintaan Gemini
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart merepresentasikan bagian dari konten Gemini
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig berisi konfigurasi untuk generasi Gemini
type GeminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// GeminiResponse adalah struktur respons dari API Gemini
type GeminiResponse struct {
	Candidates     []GeminiCandidate    `json:"candidates"`
	PromptFeedback GeminiPromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  GeminiUsageMetadata  `json:"usageMetadata,omitempty"`
}

// GeminiCandidate merepresentasikan satu kandidat respons dari Gemini
type GeminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

// GeminiPromptFeedback berisi feedback tentang prompt
type GeminiPromptFeedback struct {
	BlockReason string `json:"blockReason,omitempty"`
}

// GeminiUsageMetadata berisi informasi penggunaan token
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// Generate mengimplementasikan interface Model.Generate untuk Gemini
func (m *GeminiModel) Generate(ctx context.Context, req ModelRequest) (ModelResponse, error) {
	// Konversi ModelRequest ke GeminiRequest
	geminiReq := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: req.Prompt,
					},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			MaxOutputTokens: req.MaxTokens,
			Temperature:     req.Temperature,
			TopP:            req.TopP,
		},
	}

	// Serialize request body
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return ModelResponse{}, err
	}

	// Buat HTTP request
	modelEndpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		m.baseURL, m.modelName, m.apiKey)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		modelEndpoint,
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return ModelResponse{}, err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

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
		return ModelResponse{}, fmt.Errorf("error dari Gemini API: %s", string(respBody))
	}

	// Unmarshal respons
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return ModelResponse{}, err
	}

	// Periksa apakah ada kandidat
	if len(geminiResp.Candidates) == 0 {
		return ModelResponse{}, errors.New("tidak ada respons dari model Gemini")
	}

	// Ekstrak teks dari respons
	var responseText string
	for _, part := range geminiResp.Candidates[0].Content.Parts {
		responseText += part.Text
	}

	// Konversi GeminiResponse ke ModelResponse
	return ModelResponse{
		Text:       responseText,
		ModelName:  m.modelName,
		Provider:   Gemini,
		FinishType: geminiResp.Candidates[0].FinishReason,
		Usage: Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

// GetProvider mengimplementasikan interface Model.GetProvider
func (m *GeminiModel) GetProvider() ModelProvider {
	return Gemini
}

// GetModelName mengimplementasikan interface Model.GetModelName
func (m *GeminiModel) GetModelName() string {
	return m.modelName
}
