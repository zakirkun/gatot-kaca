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

// OpenAIModel mengimplementasikan interface Model untuk OpenAI
type OpenAIModel struct {
	apiKey    string
	modelName string
	baseURL   string
}

// NewOpenAIModel membuat instance baru OpenAIModel
func NewOpenAIModel(config ModelConfig) (Model, error) {
	if config.APIKey == "" {
		return nil, errors.New("api key diperlukan untuk OpenAI")
	}

	baseURL := "https://api.openai.com/v1"
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	return &OpenAIModel{
		apiKey:    config.APIKey,
		modelName: config.ModelName,
		baseURL:   baseURL,
	}, nil
}

// OpenAIRequest adalah struktur permintaan untuk API OpenAI
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message merepresentasikan format pesan untuk ChatGPT
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse adalah struktur respons dari API OpenAI
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Choice merepresentasikan pilihan respons dari OpenAI
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Generate mengimplementasikan interface Model.Generate untuk OpenAI
func (m *OpenAIModel) Generate(ctx context.Context, req ModelRequest) (ModelResponse, error) {
	// Konversi ModelRequest ke OpenAIRequest
	openAIReq := OpenAIRequest{
		Model: m.modelName,
		Messages: []Message{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// Serialize request body
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return ModelResponse{}, err
	}

	// Buat HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/chat/completions", m.baseURL),
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return ModelResponse{}, err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiKey))

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
		return ModelResponse{}, fmt.Errorf("error dari OpenAI API: %s", string(respBody))
	}

	// Unmarshal respons
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return ModelResponse{}, err
	}

	// Konversi OpenAIResponse ke ModelResponse
	if len(openAIResp.Choices) == 0 {
		return ModelResponse{}, errors.New("tidak ada respons dari model")
	}

	return ModelResponse{
		Text:       openAIResp.Choices[0].Message.Content,
		ModelName:  m.modelName,
		Provider:   OpenAI,
		FinishType: openAIResp.Choices[0].FinishReason,
		Usage: Usage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
	}, nil
}

// GetProvider mengimplementasikan interface Model.GetProvider
func (m *OpenAIModel) GetProvider() ModelProvider {
	return OpenAI
}

// GetModelName mengimplementasikan interface Model.GetModelName
func (m *OpenAIModel) GetModelName() string {
	return m.modelName
}
