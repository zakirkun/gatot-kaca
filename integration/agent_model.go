package integration

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

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
// It then scans the response for one or more embedded tool commands, executes them via the Agent,
// and replaces those commands with the tool outputs.
func (am *AgentModel) Generate(ctx context.Context, req llm.ModelRequest) (llm.ModelResponse, error) {
	// Generate the initial response from the inner model.
	resp, err := am.InnerModel.Generate(ctx, req)
	if err != nil {
		log.Printf("[AgentModel] Error generating response: %v", err)
		return resp, err
	}

	// Enhance the response by processing all embedded tool commands.
	resp.Text = am.processToolCommands(ctx, resp.Text)
	return resp, nil
}

// processToolCommands scans the provided text for any tool command patterns and replaces them with their outputs.
// It supports multiple commands in a single response.
func (am *AgentModel) processToolCommands(ctx context.Context, text string) string {
	// Define the regex pattern for tool commands:
	// Expected format: "CALL TOOL: <toolName> <toolInput>"
	re := regexp.MustCompile(`(?i)CALL TOOL:\s*(\w+)\s+(.+?)(?:\n|$)`)

	// Replace all matches using a function that calls the corresponding tool.
	enhancedText := re.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			// If parsing of command fails, preserve the original text.
			return match
		}
		toolName := submatches[1]
		toolInput := strings.TrimSpace(submatches[2])

		log.Printf("[AgentModel] Detected tool command: '%s' with input: '%s'", toolName, toolInput)

		// Invoke the tool via the agent.
		toolOutput, err := am.Agent.CallTool(ctx, toolName, toolInput)
		if err != nil {
			log.Printf("[AgentModel] Failed to execute tool '%s': %v", toolName, err)
			// If execution fails, return the original command text.
			return match
		}

		// Format the replacement text to include the tool's output.
		replacement := fmt.Sprintf("Tool Output (%s): %s", toolName, toolOutput)
		return replacement
	})
	return enhancedText
}

// GetProvider returns the underlying model's provider.
func (am *AgentModel) GetProvider() llm.ModelProvider {
	return am.InnerModel.GetProvider()
}

// GetModelName returns the underlying model's name.
func (am *AgentModel) GetModelName() string {
	return am.InnerModel.GetModelName()
}
