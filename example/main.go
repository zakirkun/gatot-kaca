package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zakirkun/gatot-kaca/agent"
	"github.com/zakirkun/gatot-kaca/config"
	"github.com/zakirkun/gatot-kaca/integration"
	"github.com/zakirkun/gatot-kaca/llm"
	"github.com/zakirkun/gatot-kaca/workflow"
)

func main() {
	ctx := context.Background()

	// -------------------------------------------------------------------
	// Set Up LLM Client and Agent
	// -------------------------------------------------------------------
	// Load the LLM configuration.
	llmConfig, err := config.LoadLLMConfig("./config_llm.json")
	if err != nil {
		log.Fatalf("Error loading LLM config: %v", err)
	}

	// Configure the LLM client.
	client, err := config.ConfigureLLMClient(llmConfig)
	if err != nil {
		log.Fatalf("Error configuring LLM client: %v", err)
	}

	// Create a new agent instance (using a placeholder model name, e.g., "gpt-4").
	agentInstance := agent.NewAgent(client, "gpt-4")

	// Register available tools.
	agentInstance.RegisterTool(WeatherTool{})
	agentInstance.RegisterTool(CalculatorTool{})

	// -------------------------------------------------------------------
	// Weather Tool Example
	// -------------------------------------------------------------------
	// Direct tool call: Fetch the weather in London.
	weatherResult, err := agentInstance.CallTool(ctx, "weather", "London")
	if err != nil {
		log.Fatalf("Error calling weather tool: %v", err)
	}
	fmt.Println("Direct Tool Call - Weather in London:", weatherResult)

	// Workflow integration: Build a ToolNode to fetch weather for Paris.
	weatherToolNode := &workflow.ToolNode{
		Agent:       agentInstance,
		ToolName:    "weather",
		Instruction: "Fetch the current weather for the following city:",
	}
	weatherFlow := workflow.NewFlow([]workflow.Node{weatherToolNode})
	flowOutput, err := weatherFlow.Run(ctx, "Paris")
	if err != nil {
		log.Fatalf("Error running weather workflow: %v", err)
	}
	fmt.Println("Workflow ToolNode - Weather in Paris:", flowOutput)

	// -------------------------------------------------------------------
	// Balancing Node Example
	// -------------------------------------------------------------------
	// Create two function nodes with simple processing.
	leftNode := &workflow.FuncNode{
		Process: func(ctx context.Context, input string) (string, error) {
			return "Left Node Processed: " + input, nil
		},
	}
	rightNode := &workflow.FuncNode{
		Process: func(ctx context.Context, input string) (string, error) {
			return "Right Node Processed: " + input, nil
		},
	}
	// Create a BalancingNode to distribute execution between the two function nodes.
	// Setting weights to favor the right node (weight: 1 for left, 2 for right).
	balancingNode := &workflow.BalancingNode{
		Nodes:   []workflow.Node{leftNode, rightNode},
		Weights: []int{1, 2},
	}

	// Build a workflow that uses the balancing node.
	balanceFlow := workflow.NewFlow([]workflow.Node{balancingNode})

	// Run the workflow multiple times to see the balancing in action.
	fmt.Println("Balancing Node Workflow Outputs:")
	for i := 0; i < 5; i++ {
		output, err := balanceFlow.Run(ctx, "Sample Input")
		if err != nil {
			log.Fatalf("Error running balancing workflow: %v", err)
		}
		fmt.Printf("Run %d: %s\n", i+1, output)
	}

	// -------------------------------------------------------------------
	// Integrated Model Example with Embedded Tool Commands
	// -------------------------------------------------------------------
	// Retrieve an underlying LLM model (we use "gpt-4" here as an example).
	innerModel, err := client.GetModel("gpt-4o")
	if err != nil {
		log.Fatalf("Error retrieving model: %v", err)
	}

	// Wrap the inner model with agent integration.
	integratedModel := integration.NewAgentModel(agentInstance, innerModel)

	// Now use the integrated model to generate a response.
	// The prompt includes two embedded tool commands:
	// 1. Calculator tool to compute 2+2
	// 2. Weather tool to fetch current weather in New York.
	req := llm.ModelRequest{
		Prompt:      "Hello!\nCALL TOOL: calculator 2+2\nCALL TOOL: weather New York",
		Temperature: 0.7,
		MaxTokens:   100,
		TopP:        0.9,
	}

	resp, err := integratedModel.Generate(ctx, req)
	if err != nil {
		log.Fatalf("Error generating response: %v", err)
	}

	fmt.Println("Final Integrated Response:")
	fmt.Println(resp.Text)
}
