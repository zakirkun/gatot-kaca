package wordflow

import (
	"context"
	"errors"
	"log"
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
		if total <= 0 {
			// If total weight is non-positive, fall back to round-robin.
			log.Printf("BalancingNode: total weight %d is non-positive; falling back to round-robin", total)
			idx := int(atomic.AddUint64(&bn.rrCounter, 1)-1) % len(bn.Nodes)
			selected = bn.Nodes[idx]
			log.Printf("BalancingNode (fallback round-robin) selected node at index %d", idx)
		} else {
			r := rand.Intn(total)
			selectedIndex := -1
			for i, w := range bn.Weights {
				if r < w {
					selected = bn.Nodes[i]
					selectedIndex = i
					break
				}
				r -= w
			}
			// Fallback to the last node if none selected.
			if selected == nil {
				selected = bn.Nodes[len(bn.Nodes)-1]
				selectedIndex = len(bn.Nodes) - 1
			}
			log.Printf("BalancingNode (weighted) selected node at index %d", selectedIndex)
		}
	} else {
		// Use round-robin selection.
		idx := int(atomic.AddUint64(&bn.rrCounter, 1)-1) % len(bn.Nodes)
		selected = bn.Nodes[idx]
		log.Printf("BalancingNode (round-robin) selected node at index %d", idx)
	}

	return selected.Execute(ctx, input)
}
