package eval

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zakirkun/gatot-kaca/llm"
)

// Evaluator is an interface for evaluating LLM outputs.
// It takes the original input prompt and the generated output and returns
// a normalized score between 0 and 1.
type Evaluator interface {
	Evaluate(ctx context.Context, input, output string) (float64, error)
}

// RuleBasedEvaluator scores the output based on the presence of required keywords.
// The score is computed as the ratio of matching keywords to the total number required.
type RuleBasedEvaluator struct {
	RequiredKeywords []string
}

// Evaluate checks if each of the required keywords is present in the output.
// It returns the fraction of keywords matched.
func (r *RuleBasedEvaluator) Evaluate(ctx context.Context, input, output string) (float64, error) {
	if len(r.RequiredKeywords) == 0 {
		return 0, errors.New("no required keywords specified")
	}
	normalizedOutput := strings.ToLower(output)
	var count float64
	for _, kw := range r.RequiredKeywords {
		if strings.Contains(normalizedOutput, strings.ToLower(kw)) {
			count++
		}
	}
	score := count / float64(len(r.RequiredKeywords))
	return score, nil
}

// DummyEvaluator is a simple evaluator that always returns a constant score.
type DummyEvaluator struct{}

// Evaluate always returns a dummy score (e.g. 0.5).
func (d *DummyEvaluator) Evaluate(ctx context.Context, input, output string) (float64, error) {
	return 0.5, nil
}

// EvalFunc defines a function prototype that can be used to implement custom evaluation logic.
type EvalFunc func(ctx context.Context, input, output string) (float64, error)

// CustomEvaluator wraps a custom evaluation function so that it satisfies the Evaluator interface.
type CustomEvaluator struct {
	Eval EvalFunc
}

// Evaluate calls the custom evaluation function.
func (c *CustomEvaluator) Evaluate(ctx context.Context, input, output string) (float64, error) {
	if c.Eval == nil {
		return 0, errors.New("no evaluation function provided")
	}
	return c.Eval(ctx, input, output)
}

// CompositeEvaluator aggregates multiple evaluators and returns the average score.
type CompositeEvaluator struct {
	Evaluators []Evaluator
}

// Evaluate computes each evaluator's score and returns the average.
// If no evaluators are provided, an error is returned.
func (c *CompositeEvaluator) Evaluate(ctx context.Context, input, output string) (float64, error) {
	if len(c.Evaluators) == 0 {
		return 0, errors.New("no evaluators provided")
	}
	var total float64
	for _, evaluator := range c.Evaluators {
		score, err := evaluator.Evaluate(ctx, input, output)
		if err != nil {
			return 0, err
		}
		total += score
	}
	return total / float64(len(c.Evaluators)), nil
}

// ModelGradedEvaluator uses an LLM to grade the output based on a custom prompt.
// It sends the input and output to an LLM and expects a numerical score (0 to 1) in its response.
type ModelGradedEvaluator struct {
	Client           *llm.Client
	ModelName        string
	EvaluationPrompt string // Optional: custom prompt template; if empty, a default prompt is used.
}

// Evaluate sends a request to the LLM to grade the output and parses its numerical response.
func (m *ModelGradedEvaluator) Evaluate(ctx context.Context, input, output string) (float64, error) {
	// Use a default prompt if none is provided.
	prompt := m.EvaluationPrompt
	if prompt == "" {
		prompt = fmt.Sprintf(
			"Evaluate the following output for correctness, completeness, and clarity on a score from 0 to 1.\n\nInput: %s\nOutput: %s\n\nProvide only a numerical score as your response.",
			input, output)
	}
	// Build a model request.
	req := llm.ModelRequest{
		Prompt:      prompt,
		Temperature: 0.0, // Use deterministic output.
		MaxTokens:   10,
	}
	resp, err := m.Client.Generate(ctx, m.ModelName, req)
	if err != nil {
		return 0.0, err
	}
	// Attempt to parse a score from the response text.
	score, err := parseScore(resp.Text)
	if err != nil {
		return 0, fmt.Errorf("failed to parse score: %w", err)
	}
	return score, nil
}

// parseScore attempts to extract a float score from the given text.
func parseScore(text string) (float64, error) {
	trimmed := strings.TrimSpace(text)
	// Use a regex to extract the first floating-point number.
	re := regexp.MustCompile(`\d*\.?\d+`)
	match := re.FindString(trimmed)
	if match == "" {
		return 0, errors.New("no numeric score found in response")
	}
	score, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0, err
	}
	// Ensure that the score is within the expected range.
	if score < 0 || score > 1 {
		return 0, errors.New("score out of range (should be between 0 and 1)")
	}
	return score, nil
}

// WeightedEvaluator pairs an evaluator with a weight.
type WeightedEvaluator struct {
	Evaluator Evaluator
	Weight    float64
}

// WeightedCompositeEvaluator aggregates multiple evaluators with weights,
// returning the weighted average of their scores.
type WeightedCompositeEvaluator struct {
	WeightedEvaluators []WeightedEvaluator
}

// Evaluate computes the weighted average score of all evaluators.
func (w *WeightedCompositeEvaluator) Evaluate(ctx context.Context, input, output string) (float64, error) {
	if len(w.WeightedEvaluators) == 0 {
		return 0, errors.New("no weighted evaluators provided")
	}
	var total, totalWeight float64
	for _, we := range w.WeightedEvaluators {
		score, err := we.Evaluator.Evaluate(ctx, input, output)
		if err != nil {
			return 0, err
		}
		total += score * we.Weight
		totalWeight += we.Weight
	}
	if totalWeight == 0 {
		return 0, errors.New("total weight is zero")
	}
	return total / totalWeight, nil
}
