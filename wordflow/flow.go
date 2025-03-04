package wordflow

import (
	"context"
	"fmt"
)

// Flow represents a sequence of workflow nodes executed in order.
type Flow struct {
	Nodes []Node
}

// NewFlow creates a new Flow instance with the provided nodes.
func NewFlow(nodes []Node) *Flow {
	return &Flow{
		Nodes: nodes,
	}
}

// Run executes each node in the flow sequentially.
// The output from one node is passed as input to the next.
func (f *Flow) Run(ctx context.Context, initialInput string) (string, error) {
	currentInput := initialInput
	var err error
	for _, node := range f.Nodes {
		currentInput, err = node.Execute(ctx, currentInput)
		if err != nil {
			return "", err
		}
	}
	return currentInput, nil
}

// RunWithLogging is an enhanced version of Run that logs the output of each node.
func (f *Flow) RunWithLogging(ctx context.Context, initialInput string, logger func(step int, output string)) (string, error) {
	currentInput := initialInput
	var err error
	for i, node := range f.Nodes {
		currentInput, err = node.Execute(ctx, currentInput)
		if err != nil {
			return "", fmt.Errorf("error at step %d: %w", i, err)
		}
		if logger != nil {
			logger(i, currentInput)
		}
	}
	return currentInput, nil
}
