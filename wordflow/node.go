package wordflow

import (
	"context"

	"github.com/zakirkun/gatot-kaca/agent"
)

// Node defines an interface for a step in the wordflow workflow.
type Node interface {
	Execute(ctx context.Context, input string) (string, error)
}

// LLMNode is a workflow step that uses an agent to generate text based on a prompt.
type LLMNode struct {
	// Agent instance used to communicate with the LLM.
	Agent *agent.Agent
	// Message is a static instruction or prefix for the node.
	Message string
}

// Execute resets the agent’s conversation, sends the prompt, and returns its response.
func (n *LLMNode) Execute(ctx context.Context, input string) (string, error) {
	n.Agent.Reset()
	prompt := n.Message
	if input != "" {
		prompt += "\n" + input
	}
	return n.Agent.Send(ctx, prompt)
}

// ToolNode is a workflow step that calls a registered tool via the agent.
type ToolNode struct {
	// Agent instance used to call the tool.
	Agent *agent.Agent
	// ToolName is the registered name of the tool to call.
	ToolName string
	// Instruction is an optional static instruction to accompany the input.
	Instruction string
}

// Execute resets the agent’s conversation, calls the tool with the provided instruction and input,
// and then returns the tool’s response.
func (n *ToolNode) Execute(ctx context.Context, input string) (string, error) {
	n.Agent.Reset()
	instruct := n.Instruction
	if input != "" {
		instruct += "\n" + input
	}
	return n.Agent.CallTool(ctx, n.ToolName, instruct)
}

// FuncNode is a workflow step that executes a custom function.
type FuncNode struct {
	Process func(ctx context.Context, input string) (string, error)
}

// Execute calls the custom function logic.
func (n *FuncNode) Execute(ctx context.Context, input string) (string, error) {
	return n.Process(ctx, input)
}

// ConditionalNode allows branching based on a condition function.
type ConditionalNode struct {
	// Condition evaluates the input and returns true/false to decide the branch.
	Condition func(input string) bool
	// TrueNode is executed if the condition returns true.
	TrueNode Node
	// FalseNode is executed if the condition returns false (optional).
	FalseNode Node
}

// Execute runs the appropriate node dependent on the condition.
func (n *ConditionalNode) Execute(ctx context.Context, input string) (string, error) {
	if n.Condition(input) {
		return n.TrueNode.Execute(ctx, input)
	} else if n.FalseNode != nil {
		return n.FalseNode.Execute(ctx, input)
	}
	// If no false branch is provided, return the input unchanged.
	return input, nil
}
