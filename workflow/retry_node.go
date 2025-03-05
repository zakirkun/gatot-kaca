package workflow

import (
	"context"
	"fmt"
	"time"
)

// RetryNode is a workflow node that wraps another node and attempts to retry its execution a specified number of times upon failure.
type RetryNode struct {
	Node       Node          // The child node to execute.
	MaxRetries int           // Maximum number of retries.
	Delay      time.Duration // Delay between retries.
}

// Execute attempts to execute the wrapped node. If it fails, it retries up to MaxRetries times with Delay between attempts.
func (rn *RetryNode) Execute(ctx context.Context, input string) (string, error) {
	var result string
	var err error
	for attempt := 0; attempt <= rn.MaxRetries; attempt++ {
		result, err = rn.Node.Execute(ctx, input)
		if err == nil {
			return result, nil
		}
		if attempt < rn.MaxRetries {
			time.Sleep(rn.Delay)
		}
	}
	return "", fmt.Errorf("retry node: failed after %d attempts, last error: %w", rn.MaxRetries+1, err)
}
