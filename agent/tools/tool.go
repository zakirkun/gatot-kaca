package tools

import (
	"context"
	"fmt"
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

// Manager manages a set of tools that an agent can use.
type Manager struct {
	tools map[string]Tool
}

// NewManager creates a new Manager instance.
func NewManager() *Manager {
	return &Manager{
		tools: make(map[string]Tool),
	}
}

// RegisterTool registers a tool with the manager.
func (m *Manager) RegisterTool(tool Tool) {
	fmt.Printf("Registering tool: %s\n", tool.Name())
	m.tools[tool.Name()] = tool
}

// GetTool retrieves a tool by its name.
func (m *Manager) GetTool(name string) (Tool, error) {
	tool, ok := m.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// ListTools returns a slice of all registered tool names.
func (m *Manager) ListTools() []string {
	names := make([]string, 0, len(m.tools))
	for name := range m.tools {
		names = append(names, name)
	}
	return names
}
