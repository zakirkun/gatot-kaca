package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ParallelNode is a workflow node that executes multiple child nodes concurrently and merges their outputs.
// The MergeFunc function combines the outputs of individual nodes; if not provided, outputs are joined with newlines.
// The FailFast flag indicates whether to return immediately as soon as one child node fails.
type ParallelNode struct {
	Nodes     []Node
	MergeFunc func([]string) string // Optional merge function.
	FailFast  bool                  // If true, stops execution as soon as a child node returns an error.
}

// Execute runs all child nodes concurrently with the given input and merges their results.
func (pn *ParallelNode) Execute(ctx context.Context, input string) (string, error) {
	if len(pn.Nodes) == 0 {
		return "", fmt.Errorf("parallel node: no nodes provided")
	}

	results := make([]string, len(pn.Nodes))
	errs := make([]error, len(pn.Nodes))
	var wg sync.WaitGroup
	wg.Add(len(pn.Nodes))

	for i, node := range pn.Nodes {
		go func(i int, n Node) {
			defer wg.Done()
			res, err := n.Execute(ctx, input)
			results[i] = res
			errs[i] = err
		}(i, node)
	}

	wg.Wait()

	if pn.FailFast {
		for _, err := range errs {
			if err != nil {
				return "", err
			}
		}
	} else {
		// Log warnings for errors but continue.
		for i, err := range errs {
			if err != nil {
				fmt.Printf("Warning: node %d returned error: %v\n", i, err)
			}
		}
	}

	// Merge the results.
	if pn.MergeFunc != nil {
		return pn.MergeFunc(results), nil
	}

	// Default merge: combine outputs with newline delimiters.
	return strings.Join(results, "\n"), nil
}
