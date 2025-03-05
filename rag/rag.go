package rag

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/zakirkun/gatot-kaca/llm"
)

// Document represents a piece of text stored in the knowledge base.
type Document struct {
	ID        string
	Text      string
	Embedding []float64
}

// KnowledgeBase is an in‑memory store for documents. It uses an llm.Client and a designated model
// to generate the real embeddings for documents and queries.
type KnowledgeBase struct {
	Documents []*Document
	Client    *llm.Client
	ModelName string
}

// NewKnowledgeBase creates a new empty knowledge base.
func NewKnowledgeBase(client *llm.Client, modelName string) *KnowledgeBase {
	return &KnowledgeBase{
		Documents: []*Document{},
		Client:    client,
		ModelName: modelName,
	}
}

// AddDocument adds a new document to the knowledge base using an embedding from the llm client.
func (kb *KnowledgeBase) AddDocument(ctx context.Context, id, text string) error {
	embedding, err := kb.Client.Embedding(ctx, kb.ModelName, text)
	if err != nil {
		return fmt.Errorf("failed to compute embedding for document '%s': %w", id, err)
	}

	doc := &Document{
		ID:        id,
		Text:      text,
		Embedding: embedding,
	}
	kb.Documents = append(kb.Documents, doc)
	return nil
}

// cosineSimilarity calculates the cosine similarity between two vectors.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// RetrievalResult holds a document along with its similarity score for a query.
type RetrievalResult struct {
	Doc   *Document
	Score float64
}

// Query returns the top k documents that are most similar to the provided query text.
func (kb *KnowledgeBase) Query(ctx context.Context, query string, k int) ([]RetrievalResult, error) {
	queryEmbedding, err := kb.Client.Embedding(ctx, kb.ModelName, query)
	if err != nil {
		return nil, fmt.Errorf("failed to compute embedding for query: %w", err)
	}

	results := []RetrievalResult{}
	for _, doc := range kb.Documents {
		score := cosineSimilarity(queryEmbedding, doc.Embedding)
		results = append(results, RetrievalResult{
			Doc:   doc,
			Score: score,
		})
	}

	// Sort results by similarity score in descending order.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if k > len(results) {
		k = len(results)
	}
	return results[:k], nil
}

// AugmentPrompt constructs a new prompt by prepending the retrieved documents to the query.
func AugmentPrompt(query string, results []RetrievalResult) string {
	augmented := "The following information might be useful:\n"
	for _, res := range results {
		augmented += "- " + strings.TrimSpace(res.Doc.Text) + "\n"
	}
	augmented += "\nBased on the above, please answer the following question:\n" + query
	return augmented
}
