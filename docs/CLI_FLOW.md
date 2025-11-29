# CLI FLOW

## SETUP

```text
Start CLI
    │
    ▼
Is Ollama installed? ──No──► Install Ollama (brew or manual prompt)
    │ Yes
    ▼
Is Ollama running? ──No──► Start Ollama
    │ Yes
    ▼
Any models installed? ──No──► Show model picker → Pull selected
    │ Yes
    ▼
Enter main agent loop
```

## Ollama API ENDPOINTS NEEDED FOR SETUP

```bash
# Check if running + list models
GET http://localhost:11434/api/tags

# Pull a model (streams progress)
POST http://localhost:11434/api/pull
{"name": "qwen2.5-coder:7b"}

# Check version
GET http://localhost:11434/api/version
```