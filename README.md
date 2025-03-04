# gatot-kaca
Gatot Kaca is AI Agent powered by Golang stack

```json
{
    "models": [
        {
            "provider": "openai",
            "model_name": "gpt-40",
            "api_key": "${OPENAI_API_KEY}",
            "options": {
            "timeout": 30
            }
        },
        {
            "provider": "anthropic",
            "model_name": "claude-3-opus-20240229",
            "api_key": "${ANTHROPIC_API_KEY}"
        },
        {
            "provider": "gemini",
            "model_name": "gemini-pro",
            "api_key": "${GEMINI_API_KEY}"
        }
    ],
    "default": "gpt-4o"
}```