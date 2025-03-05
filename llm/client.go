package llm

import (
	"context"
	"errors"
	"sync"
)

// Client adalah klien untuk berinteraksi dengan berbagai model LLM
type Client struct {
	models   map[string]Model
	fallback Model
	mu       sync.RWMutex
}

// NewClient membuat instance baru Client LLM
func NewClient() *Client {
	return &Client{
		models: make(map[string]Model),
	}
}

// AddModel menambahkan model ke client
func (c *Client) AddModel(name string, model Model) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.models[name] = model
}

// SetFallbackModel menetapkan model fallback yang akan digunakan jika model yang diminta tidak ditemukan
func (c *Client) SetFallbackModel(model Model) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fallback = model
}

// GetModel mendapatkan model berdasarkan nama
func (c *Client) GetModel(name string) (Model, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	model, exists := c.models[name]
	if !exists {
		if c.fallback != nil {
			return c.fallback, nil
		}
		return nil, errors.New("model tidak ditemukan dan tidak ada fallback")
	}

	return model, nil
}

// Generate menggunakan model tertentu untuk menghasilkan respons
func (c *Client) Generate(ctx context.Context, modelName string, req ModelRequest) (ModelResponse, error) {
	model, err := c.GetModel(modelName)
	if err != nil {
		return ModelResponse{}, err
	}

	return model.Generate(ctx, req)
}

// Embedding
func (c *Client) Embedding(ctx context.Context, modelName string, text string) ([]float64, error) {
	model, err := c.GetModel(modelName)
	if err != nil {
		return nil, err
	}

	return model.GenerateEmbedding(ctx, text)
}

// ConfigureFromOptions mengonfigurasi client dari opsi
func (c *Client) ConfigureFromOptions(options []ModelConfig) error {
	for _, config := range options {
		model, err := ModelFactory(config)
		if err != nil {
			return err
		}

		modelName := config.ModelName
		c.AddModel(modelName, model)

		// Set model pertama sebagai fallback jika belum ada fallback
		if c.fallback == nil {
			c.SetFallbackModel(model)
		}
	}

	return nil
}

// ListModels mengembalikan daftar nama model yang tersedia
func (c *Client) ListModels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	models := make([]string, 0, len(c.models))
	for name := range c.models {
		models = append(models, name)
	}

	return models
}
