package llm

import (
	"context"
	"errors"
)

// ModelProvider mendefinisikan penyedia model LLM
type ModelProvider string

const (
	OpenAI    ModelProvider = "openai"
	Anthropic ModelProvider = "anthropic"
	Gemini    ModelProvider = "gemini"
)

// ModelRequest mewakili permintaan ke model LLM
type ModelRequest struct {
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ModelResponse mewakili respons dari model LLM
type ModelResponse struct {
	Text       string                 `json:"text"`
	Usage      Usage                  `json:"usage"`
	ModelName  string                 `json:"model_name"`
	Provider   ModelProvider          `json:"provider"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	FinishType string                 `json:"finish_type,omitempty"`
}

// Usage mencatat penggunaan token
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Model interface {
	// Generate menghasilkan respons dari prompt
	Generate(ctx context.Context, req ModelRequest) (ModelResponse, error)

	// GetProvider mengembalikan penyedia model
	GetProvider() ModelProvider

	// GetModelName mengembalikan nama model
	GetModelName() string

	// Embedding
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
}

// ModelConfig menyimpan konfigurasi untuk model LLM
type ModelConfig struct {
	Provider  ModelProvider          `json:"provider"`
	ModelName string                 `json:"model_name"`
	APIKey    string                 `json:"api_key"`
	BaseURL   string                 `json:"base_url,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// ModelFactory membuat instance Model berdasarkan konfigurasi
func ModelFactory(config ModelConfig) (Model, error) {
	switch config.Provider {
	case OpenAI:
		return NewOpenAIModel(config)
	case Anthropic:
		return NewAnthropicModel(config)
	case Gemini:
		return NewGeminiModel(config)
	default:
		return nil, errors.New("provider tidak didukung")
	}
}
