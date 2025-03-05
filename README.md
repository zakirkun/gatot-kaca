# Gatot Kaca

Gatot Kaca is an AI Agent framework powered by Golang. It brings together multiple LLM providers, an intelligent agent with tool command processing, robust workflow orchestration, and an evaluation framework to build advanced AI applications.

## Overview

Gatot Kaca integrates several powerful features:

- **Multi-LLM Integration:**  
  Support for multiple language model providers (e.g., OpenAI, Anthropic, Gemini). The client is configured via a JSON file that automatically substitutes environment variables for secure API keys.

- **Agent with Tool Command Processing & Middleware:**  
  The agent architecture not only handles conversation history and LLM calls but also:
  - Detects and processes embedded tool commands (e.g., `CALL TOOL: weather London`).
  - Supports middleware hooks for pre- and post-processing, including system prompt support.
  - Allows direct tool invocations and integration within workflows.

- **Tool Management:**  
  Tools are implemented through a defined interface and can optionally expose additional metadata with the extended tool interface. Built-in sample tools include:
  - **WeatherTool:** Fetches current weather information from wttr.in.
  - **CalculatorTool:** Evaluates simple arithmetic expressions (e.g., addition).

- **Workflow Engine (Wordflow):**  
  Build powerful workflows using a series of modular nodes:
  - **LLMNode:** Uses the agent to generate responses with LLMs.
  - **ToolNode:** Calls registered tools based on a given instruction.
  - **FuncNode & ConditionalNode:** Execute custom functions or branch the flow based on conditions.
  - **BalancingNode:** Supports weighted random or round-robin selection among multiple nodes.
  - **RetryNode:** Retries node execution upon failure.
  - **ParallelNode:** Executes child nodes concurrently and merges their outputs.

- **Integrated Model:**  
  Wrap an LLM model with the agent to process prompts that include embedded tool commands. The model automatically scans for tool commands, invokes the corresponding tools, and integrates their outputs back into the response.

- **(Optional) Evaluation Features:**  
  Enhance your application with an evaluation framework that supports multiple evaluators, including composite and weighted evaluators, plus LLM-based grading to continuously monitor output quality.

## Examples

The framework comes with end-to-end examples demonstrating:

1. **Direct Tool Calls:**  
   Invoke the WeatherTool and CalculatorTool directly via the agent.

2. **Workflow-Based Execution:**  
   Build workflows using ToolNodes and BalancingNodes. For instance:
   - A workflow that fetches the weather for a city.
   - A balancing workflow distributing input between two function nodes (demonstrates weighted selection).

3. **Integrated Model Usage:**  
   Wrap an inner LLM model with the agent so that prompts with embedded tool commands (e.g., `CALL TOOL: calculator 2+2` and `CALL TOOL: weather New York`) are automatically processed and replaced with tool outputs.

Check out the example source files in the `example/` directory:
- [main.go](example/main.go): Demonstrates setting up the LLM client, agent, tool registration, workflow execution, and integrated model with embedded tool commands.
- [weather.go](example/weather.go): Implements the WeatherTool.
- [calculator.go](example/calculator.go): Implements the CalculatorTool.

## Example Configuration

Below is a sample configuration file (`config_llm.json`) that configures three LLM providers:
