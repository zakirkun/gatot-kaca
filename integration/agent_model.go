package integration

import (
	"context"
	"regexp"

	"github.com/zakirkun/gatot-kaca/agent"
	"github.com/zakirkun/gatot-kaca/llm"
)

// AgentModel is an integrated model that wraps an inner LLM model and uses an agent for enhanced processing.
// It checks the generated response for embedded tool commands and, when found, automatically calls the tool.
type AgentModel struct {
	Agent      *agent.Agent // An agent instance that provides tool integration.
	InnerModel llm.Model    // The underlying LLM model (e.g., OpenAI, Anthropic, Gemini, etc.)
}

// NewAgentModel wraps an existing model with agent integration.
func NewAgentModel(agentInstance *agent.Agent, inner llm.Model) *AgentModel {
	return &AgentModel{
		Agent:      agentInstance,
		InnerModel: inner,
	}
}

// Generate processes a ModelRequest by calling the inner model's Generate method.
// After generating the initial response, it inspects the response text for tool commands.
// If a tool command is found in the format "CALL TOOL: <tool_name> <tool_input>",
// it uses the agent to call the tool and appends the tool's output to the response.
func (am *AgentModel) Generate(ctx context.Context, req llm.ModelRequest) (llm.ModelResponse, error) {
	// Get the initial response from the inner model.
	resp, err := am.InnerModel.Generate(ctx, req)
	if err != nil {
		return resp, err
	}

	// Use a regex to check for an embedded tool command.
	// Expected format: "CALL TOOL: <tool_name> <tool_input>"
	re := regexp.MustCompile(`(?i)^CALL TOOL:\s*(\w+)\s+(.+)$`)
	matches := re.FindStringSubmatch(resp.Text)
	if len(matches) == 3 {
		toolName := matches[1]
		toolInput := matches[2]
		// Use the agent to call the tool.
		toolOutput, err := am.Agent.CallTool(ctx, toolName, toolInput)
		if err == nil && toolOutput != "" {
			// Append the tool output to the original response.
			resp.Text += "\nTool Output: " + toolOutput
		}
	}

	return resp, nil
}

// GetProvider returns the underlying model's provider.
func (am *AgentModel) GetProvider() llm.ModelProvider {
	return am.InnerModel.GetProvider()
}

// GetModelName returns the underlying model's name.
func (am *AgentModel) GetModelName() string {
	return am.InnerModel.GetModelName()
}
