package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/zakirkun/gatot-kaca/agent/tools"
	"github.com/zakirkun/gatot-kaca/llm"
)

// ConversationMessage holds the details of each message in the conversation.
type ConversationMessage struct {
	Role    string
	Content string
	// (Optional) Timestamp string or any other metadata can be added here.
}

// Middleware defines an interface to pre- and post-process conversation messages.
type Middleware interface {
	// ProcessBeforeSend allows a middleware to modify or augment the conversation history before sending.
	ProcessBeforeSend(ctx context.Context, history []ConversationMessage) []ConversationMessage
	// ProcessAfterReceive allows a middleware to post-process the LLM response.
	ProcessAfterReceive(ctx context.Context, response string) string
}

// Agent encapsulates the conversation logic with the LLM-based client
// and now supports calling external tools.
type Agent struct {
	client       *llm.Client
	modelName    string
	history      []ConversationMessage
	Temperature  float64
	MaxTokens    int
	TopP         float64
	tools        *tools.Manager
	systemPrompt string
	middlewares  []Middleware
}

// NewAgent creates a new Agent instance and initializes its tools manager.
func NewAgent(client *llm.Client, modelName string) *Agent {
	return &Agent{
		client:      client,
		modelName:   modelName,
		history:     []ConversationMessage{},
		Temperature: 0.7, // Default value; adjust as needed.
		MaxTokens:   150, // Default value; adjust as needed.
		TopP:        0.9, // Default value; adjust as needed.
		tools:       tools.NewManager(),
	}
}

// SetSystemPrompt sets a system-level instruction that will be prepended to every conversation.
func (a *Agent) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}

// RegisterMiddleware registers a middleware to allow pre- and post-processing of conversation messages.
func (a *Agent) RegisterMiddleware(m Middleware) {
	a.middlewares = append(a.middlewares, m)
}

// AppendMessage adds a new message to the conversation history.
func (a *Agent) AppendMessage(role, content string) {
	a.history = append(a.history, ConversationMessage{
		Role:    role,
		Content: content,
	})
}

// BuildPrompt constructs a prompt from the conversation history,
// including the system prompt (if set) and applying any registered middleware.
func (a *Agent) BuildPrompt(ctx context.Context) string {
	var modHistory []ConversationMessage
	// Prepend the system prompt if present.
	if a.systemPrompt != "" {
		modHistory = append(modHistory, ConversationMessage{Role: "System", Content: a.systemPrompt})
	}

	// Append conversation history.
	modHistory = append(modHistory, a.history...)

	// Allow middleware to process/modify the conversation before sending.
	for _, m := range a.middlewares {
		modHistory = m.ProcessBeforeSend(ctx, modHistory)
	}

	var builder strings.Builder
	for _, msg := range modHistory {
		builder.WriteString(msg.Role + ": " + msg.Content + "\n")
	}
	return builder.String()
}

// Send sends a user message to the agent, retrieves the LLM response, applies middleware,
// processes tool commands and updates the conversation history.
func (a *Agent) Send(ctx context.Context, userInput string) (string, error) {
	// Append the user's message.
	a.AppendMessage("User", userInput)

	// Construct the prompt including system prompt and middleware modifications.
	prompt := a.BuildPrompt(ctx)

	// Create the model request using the agent's default parameters.
	req := llm.ModelRequest{
		Prompt:      prompt,
		Temperature: a.Temperature,
		MaxTokens:   a.MaxTokens,
		TopP:        a.TopP,
	}

	// Get the response from the LLM client.
	res, err := a.client.Generate(ctx, a.modelName, req)
	if err != nil {
		return "", err
	}

	// Allow middleware to post-process the LLM response.
	responseText := res.Text
	for _, m := range a.middlewares {
		responseText = m.ProcessAfterReceive(ctx, responseText)
	}

	// Append the assistant's response to the history.
	a.AppendMessage("Assistant", responseText)

	// Check if the response includes an embedded tool command.
	if toolOutput, err := a.processToolCommand(ctx, responseText); err == nil && toolOutput != "" {
		// Append the tool output automatically.
		a.AppendMessage("Tool Response", toolOutput)
		// Return the combined output (initial response + tool output).
		return fmt.Sprintf("%s\nTool Output: %s", responseText, toolOutput), nil
	}

	return responseText, nil
}

// Reset clears the conversation history in the agent.
func (a *Agent) Reset() {
	a.history = []ConversationMessage{}
}

// RegisterTool registers a new tool with the agent.
func (a *Agent) RegisterTool(tool tools.Tool) {
	a.tools.RegisterTool(tool)
}

// CallTool executes a registered tool by name with the provided input.
// It appends both the tool invocation and its response to the conversation history.
func (a *Agent) CallTool(ctx context.Context, toolName, input string) (string, error) {
	tool, err := a.tools.GetTool(toolName)
	if err != nil {
		return "", err
	}

	// Record the tool invocation.
	a.AppendMessage("Tool Call ("+toolName+")", input)

	// Execute the tool.
	result, err := tool.Execute(ctx, input)
	if err != nil {
		return "", err
	}

	// Record the tool's response.
	a.AppendMessage("Tool Response ("+toolName+")", result)
	return result, nil
}

// processToolCommand checks if the input string begins with a tool command in the format:
// "CALL TOOL: <tool-name> <tool-input>" and, if so, calls the corresponding tool.
func (a *Agent) processToolCommand(ctx context.Context, response string) (string, error) {
	// Use a regex to detect commands beginning with "CALL TOOL:"
	// Format example: "CALL TOOL: calculator 2+3"
	re := regexp.MustCompile(`(?i)^CALL TOOL:\s*(\w+)\s+(.+)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(response))
	if len(matches) != 3 {
		// If no tool command is found, return empty string.
		return "", nil
	}

	toolName := matches[1]
	toolInput := matches[2]
	return a.CallTool(ctx, toolName, toolInput)
}
