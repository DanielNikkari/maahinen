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