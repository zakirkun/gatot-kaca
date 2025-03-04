package wordflow

import (
	"context"
	"errors"
	"math/rand"
	"sync/atomic"
	"time"
)

// BalancingNode is a workflow node that selects one out of multiple nodes based on a balancing algorithm.
// If Weights is provided (its length equals len(Nodes)), weighted random selection is used;
// otherwise, a round-robin algorithm is applied.
type BalancingNode struct {
	Nodes   []Node // Available child nodes.
	Weights []int  // Optional: if provided and len(Weights)==len(Nodes), use weighted random selection.

	rrCounter uint64 // Internal counter for round-robin selection.
}

// init seeds the random number generator.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Execute selects one child node based on the balancing algorithm and then executes it with the input.
func (bn *BalancingNode) Execute(ctx context.Context, input string) (string, error) {
	if len(bn.Nodes) == 0 {
		return "", errors.New("balancing node: no nodes available")
	}

	var selected Node
	if len(bn.Weights) == len(bn.Nodes) {
		// Use weighted random selection.
		total := 0
		for _, w := range bn.Weights {
			total += w
		}
		r := rand.Intn(total)
		for i, w := range bn.Weights {
			if r < w {
				selected = bn.Nodes[i]
				break
			}
			r -= w
		}
		// Fallback to the last node if none selected.
		if selected == nil {
			selected = bn.Nodes[len(bn.Nodes)-1]
		}
	} else {
		// Use round-robin selection.
		idx := int(atomic.AddUint64(&bn.rrCounter, 1)-1) % len(bn.Nodes)
		selected = bn.Nodes[idx]
	}

	return selected.Execute(ctx, input)
}
