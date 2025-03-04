package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zakirkun/gatot-kaca/llm"
)

// LLMConfig menampung konfigurasi untuk semua model LLM
type LLMConfig struct {
	Models  []llm.ModelConfig `json:"models"`
	Default string            `json:"default,omitempty"`
}

// LoadLLMConfig memuat konfigurasi LLM dari file
func LoadLLMConfig(configPath string) (*LLMConfig, error) {
	// Baca file konfigurasi
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file konfigurasi: %w", err)
	}

	// Parse konfigurasi
	var config LLMConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("gagal mem-parse konfigurasi: %w", err)
	}

	// Ganti variabel lingkungan dalam string
	for i := range config.Models {
		if strings.HasPrefix(config.Models[i].APIKey, "${") && strings.HasSuffix(config.Models[i].APIKey, "}") {
			envVar := config.Models[i].APIKey[2 : len(config.Models[i].APIKey)-1]
			config.Models[i].APIKey = os.Getenv(envVar)
		}
	}

	return &config, nil
}

// ConfigureLLMClient mengonfigurasi klien LLM dari konfigurasi
func ConfigureLLMClient(config *LLMConfig) (*llm.Client, error) {
	client := llm.NewClient()

	if err := client.ConfigureFromOptions(config.Models); err != nil {
		return nil, fmt.Errorf("gagal mengonfigurasi klien LLM: %w", err)
	}

	// Set model default jika ditentukan
	if config.Default != "" {
		defaultModel, err := client.GetModel(config.Default)
		if err != nil {
			return nil, fmt.Errorf("model default '%s' tidak ditemukan: %w", config.Default, err)
		}
		client.SetFallbackModel(defaultModel)
	}

	return client, nil
}

// SaveLLMConfig menyimpan konfigurasi LLM ke file
func SaveLLMConfig(config *LLMConfig, configPath string) error {
	// Buat direktori jika belum ada
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori konfigurasi: %w", err)
	}

	// Marshal konfigurasi ke JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal mengonversi konfigurasi ke JSON: %w", err)
	}

	// Tulis ke file
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("gagal menulis file konfigurasi: %w", err)
	}

	return nil
}

// CreateDefaultConfig membuat konfigurasi default
func CreateDefaultConfig() *LLMConfig {
	return &LLMConfig{
		Models: []llm.ModelConfig{
			{
				Provider:  llm.OpenAI,
				ModelName: "gpt-4",
				APIKey:    "${OPENAI_API_KEY}",
			},
			{
				Provider:  llm.Anthropic,
				ModelName: "claude-3-opus-20240229",
				APIKey:    "${ANTHROPIC_API_KEY}",
			},
			{
				Provider:  llm.Gemini,
				ModelName: "gemini-pro",
				APIKey:    "${GEMINI_API_KEY}",
			},
		},
		Default: "gpt-4",
	}
}
