package tools

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Tool represents an external helper function that the agent can call.
type Tool interface {
	// Name returns the name of the tool.
	Name() string
	// Description returns a short description of the tool.
	Description() string
	// Execute runs the tool's command given an input string and returns its output.
	Execute(ctx context.Context, input string) (string, error)
}

// EnhancedTool is an optional extension that exposes additional metadata about a tool.
type EnhancedTool interface {
	Tool
	// Schema returns a JSON schema string that defines the expected input.
	Schema() string
	// Help returns detailed usage instructions.
	Help() string
}

// Manager manages a set of tools that an agent can use.
type Manager struct {
	tools   map[string]Tool
	metrics map[string]int // Track the number of times each tool is executed.

}

// NewManager creates a new Manager instance.
func NewManager() *Manager {
	return &Manager{
		tools:   make(map[string]Tool),
		metrics: make(map[string]int),
	}
}

// RegisterTool registers a tool with the manager.
func (m *Manager) RegisterTool(tool Tool) {
	fmt.Printf("Registering tool: %s\n", tool.Name())
	m.tools[tool.Name()] = tool
	// Initialize call count metric.
	m.metrics[tool.Name()] = 0
}

// GetTool retrieves a tool by its name.
func (m *Manager) GetTool(name string) (Tool, error) {
	tool, ok := m.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// ExecuteTool executes a registered tool by name with the provided input
// and logs execution details such as duration and errors.
// It also updates the call metrics for that tool.
func (m *Manager) ExecuteTool(ctx context.Context, name, input string) (string, error) {
	tool, err := m.GetTool(name)
	if err != nil {
		return "", err
	}
	start := time.Now()
	output, err := tool.Execute(ctx, input)
	duration := time.Since(start)
	if err != nil {
		log.Printf("[Tool Execution] Tool '%s' failed after %v: %v", name, duration, err)
		return "", err
	}
	log.Printf("[Tool Execution] Tool '%s' executed in %v", name, duration)
	// Increment call count metric.
	m.metrics[name]++
	return output, nil
}

// ListTools returns a slice of all registered tool names.
func (m *Manager) ListTools() []string {
	names := make([]string, 0, len(m.tools))
	for name := range m.tools {
		names = append(names, name)
	}
	return names
}

// ListDetailedTools returns a detailed description for all registered tools.
// For tools that implement EnhancedTool, it includes the schema and help information.
func (m *Manager) ListDetailedTools() string {
	var result string
	for name, tool := range m.tools {
		result += fmt.Sprintf("Tool: %s\n", name)
		result += fmt.Sprintf("Description: %s\n", tool.Description())
		// Check if the tool implements the EnhancedTool interface.
		if et, ok := tool.(EnhancedTool); ok {
			result += fmt.Sprintf("Schema: %s\n", et.Schema())
			result += fmt.Sprintf("Help: %s\n", et.Help())
		}
		result += "\n"
	}
	return result
}

// GetCallCount returns the number of times a tool has been executed.
// If the tool isn't found, it returns a count of 0.
func (m *Manager) GetCallCount(name string) int {
	count, ok := m.metrics[name]
	if !ok {
		return 0
	}
	return count
}
