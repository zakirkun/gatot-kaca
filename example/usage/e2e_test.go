// This file contains end-to-end feature tests for Gatot Kaca.
// It tests direct tool invocations, workflows (including balancing nodes),
// and an integrated model that processes embedded tool commands.

package usage_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/zakirkun/gatot-kaca/agent"
	"github.com/zakirkun/gatot-kaca/integration"
	"github.com/zakirkun/gatot-kaca/llm"
	"github.com/zakirkun/gatot-kaca/workflow"
)

//////////////////////
// Fake LLM and Client for Testing
//////////////////////

// FakeLLM implements the llm.Model interface and returns the request prompt as its output.
type FakeLLM struct{}

func (f *FakeLLM) Generate(ctx context.Context, req llm.ModelRequest) (llm.ModelResponse, error) {
	// Simply echo back the prompt so that embedded tool commands are present.
	return llm.ModelResponse{
		Text:       req.Prompt,
		ModelName:  "fake",
		Provider:   llm.ModelProvider("fake"),
		FinishType: "completed",
		Usage:      llm.Usage{},
	}, nil
}

func (f *FakeLLM) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	return []float64{}, nil
}

func (f *FakeLLM) GetProvider() llm.ModelProvider { return llm.ModelProvider("fake") }
func (f *FakeLLM) GetModelName() string           { return "fake" }

// FakeLLMClient implements a minimal llm.Client needed for testing.
type FakeLLMClient struct{}

func (f *FakeLLMClient) GetModel(name string) (llm.Model, error) {
	return &FakeLLM{}, nil
}

func (f *FakeLLMClient) ConfigureFromOptions(options []llm.ModelConfig) error {
	return nil
}

func (f *FakeLLMClient) SetFallbackModel(model llm.Model) {
	// no-op for test
}

//////////////////////
// Tool Implementations for Testing
//////////////////////

// CalculatorTool is a minimal implementation that supports simple addition.
type CalculatorTool struct{}

func (c CalculatorTool) Name() string { return "calculator" }

func (c CalculatorTool) Description() string {
	return "Evaluates basic arithmetic expressions (supports addition in the format 'number+number')."
}

func (c CalculatorTool) Execute(ctx context.Context, input string) (string, error) {
	// Remove spaces and newlines.
	expr := strings.ReplaceAll(input, " ", "")
	expr = strings.TrimSpace(expr)
	expr = strings.ReplaceAll(expr, "\n", "")
	if strings.Contains(expr, ":") {
		if strings.Contains(expr, "+") {
			split := strings.Split(expr, ":")
			parts := strings.Split(split[1], "+")
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid expression: %s", input)
			}
			a, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return "", err
			}
			b, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%v", a+b), nil
		}
	}
	return "", fmt.Errorf("unsupported expression: %s", input)
}

// WeatherTool returns a dummy weather response.
type WeatherTool struct{}

func (w WeatherTool) Name() string { return "weather" }

func (w WeatherTool) Description() string {
	return "Fetches current weather information for a given city."
}

func (w WeatherTool) Execute(ctx context.Context, input string) (string, error) {
	// For testing, simply return a dummy string.
	city := strings.TrimSpace(input)
	if city == "" {
		return "", fmt.Errorf("city name must be provided")
	}
	return fmt.Sprintf("Weather in %s: Sunny 25°C", city), nil
}

//////////////////////
// End-to-End Feature Tests
//////////////////////

// TestDirectToolCall verifies that a tool call (CalculatorTool) returns the expected result.
func TestDirectToolCall(t *testing.T) {
	ctx := context.Background()
	fakeClient := &llm.Client{} // Assuming llm.Client is the correct type
	agentInstance := agent.NewAgent(fakeClient, "fake")
	agentInstance.RegisterTool(CalculatorTool{})

	// Directly invoke CalculatorTool using a simple addition.
	result, err := agentInstance.CallTool(ctx, "calculator", "2+2")
	if err != nil {
		t.Fatalf("Direct tool call failed: %v", err)
	}

	expected := "4"
	if strings.TrimSpace(result) != expected {
		t.Errorf("Expected result '%s', but got '%s'", expected, result)
	}

	t.Log("Passed")
}

// TestWorkflowToolNode uses a ToolNode to invoke the CalculatorTool within a workflow.
func TestWorkflowToolNode(t *testing.T) {
	ctx := context.Background()
	fakeClient := &llm.Client{} // Assuming llm.Client is the correct type
	agentInstance := agent.NewAgent(fakeClient, "fake")
	agentInstance.RegisterTool(CalculatorTool{})

	toolNode := &workflow.ToolNode{
		Agent:       agentInstance,
		ToolName:    "calculator",
		Instruction: "Calculate: ",
	}
	flowInstance := workflow.NewFlow([]workflow.Node{toolNode})
	result, err := flowInstance.Run(ctx, "2+2")
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	expected := "4"
	if strings.TrimSpace(result) != expected {
		t.Errorf("Expected workflow result '%s', got '%s'", expected, result)
	}

	t.Log("Passed")
}

// TestBalancingNode verifies that a BalancingNode properly selects one of the provided function nodes.
func TestBalancingNode(t *testing.T) {
	ctx := context.Background()

	// Define two function nodes that yield distinct outputs.
	leftNode := &workflow.FuncNode{
		Process: func(ctx context.Context, input string) (string, error) {
			return "Left: " + input, nil
		},
	}
	rightNode := &workflow.FuncNode{
		Process: func(ctx context.Context, input string) (string, error) {
			return "Right: " + input, nil
		},
	}

	balancingNode := &workflow.BalancingNode{
		Nodes:   []workflow.Node{leftNode, rightNode},
		Weights: []int{1, 2}, // Favor the right node.
	}
	flowInstance := workflow.NewFlow([]workflow.Node{balancingNode})
	result, err := flowInstance.Run(ctx, "test input")
	if err != nil {
		t.Fatalf("Balancing node execution failed: %v", err)
	}

	if !strings.HasPrefix(result, "Left:") && !strings.HasPrefix(result, "Right:") {
		t.Errorf("Unexpected balancing node result: %s", result)
	} else {
		t.Logf("Balancing node returned: %s", result)
	}
}

// TestIntegratedModel verifies that the integrated model processes embedded tool commands.
func TestIntegratedModel(t *testing.T) {
	ctx := context.Background()
	fakeLLMClient := &llm.Client{} // Assuming llm.Client is the correct type
	agentInstance := agent.NewAgent(fakeLLMClient, "fake")
	agentInstance.RegisterTool(CalculatorTool{})
	agentInstance.RegisterTool(WeatherTool{})

	// Use our FakeLLM as the inner model.
	fakeLLM := &FakeLLM{}
	integratedModel := integration.NewAgentModel(agentInstance, fakeLLM)

	// The prompt contains embedded tool commands.
	prompt := "Hello!\nCALL TOOL: calculator 2+2\nCALL TOOL: weather TestCity"
	req := llm.ModelRequest{
		Prompt:      prompt,
		Temperature: 0.5,
		MaxTokens:   50,
		TopP:        1.0,
	}

	resp, err := integratedModel.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Integrated model generation failed: %v", err)
	}

	// Ensure that raw tool command texts are replaced.
	if strings.Contains(resp.Text, "CALL TOOL: calculator") || strings.Contains(resp.Text, "CALL TOOL: weather") {
		t.Errorf("Embedded tool commands were not replaced correctly in integrated response: %s", resp.Text)
	}

	// Check for calculator output ("4") and weather output.
	if !strings.Contains(resp.Text, "4") {
		t.Errorf("Expected CalculatorTool output '4' missing in integrated response: %s", resp.Text)
	}
	if !strings.Contains(resp.Text, "Sunny") {
		t.Errorf("Expected WeatherTool output missing in integrated response: %s", resp.Text)
	}
}
