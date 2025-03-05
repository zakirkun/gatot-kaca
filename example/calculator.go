package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// CalculatorTool implements the Tool interface to evaluate basic arithmetic expressions.
type CalculatorTool struct{}

// Name returns the name of the calculator tool.
func (c CalculatorTool) Name() string {
	return "calculator"
}

// Description returns a brief description of the calculator tool.
func (c CalculatorTool) Description() string {
	return "Evaluates basic arithmetic expressions (supports addition in the format 'number+number')."
}

// Execute parses a simple arithmetic expression and returns the result.
// For simplicity, only simple addition expressions (e.g., "2+2") are supported.
func (c CalculatorTool) Execute(ctx context.Context, input string) (string, error) {
	// Remove any spaces and newlines.
	expression := strings.ReplaceAll(input, " ", "")
	expression = strings.TrimSpace(expression)

	// Check if the expression is an addition.
	if strings.Contains(expression, "+") {
		parts := strings.Split(expression, "+")
		if len(parts) != 2 {
			return "", fmt.Errorf("unsupported expression format: %s", input)
		}
		a, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return "", fmt.Errorf("failed parsing left operand: %w", err)
		}
		b, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return "", fmt.Errorf("failed parsing right operand: %w", err)
		}
		result := a + b
		return fmt.Sprintf("%v", result), nil
	}
	return "", fmt.Errorf("unsupported operator or expression: %s", input)
}
