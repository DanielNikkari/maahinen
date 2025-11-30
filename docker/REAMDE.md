# Docker containerization

> [!NOTE]
> It is recommended to run the tool inside a container for safety, as the tool has access to the system.

## Build & Run

Run these from the **project root**.

### Build
```bash
docker build -f docker/Dockerfile -t maahinen .
```

### Run
```bash
docker run -it maahinen
```

## Compose (recommended)

Compose runs Ollama as a separate service with a persistent volume, so models don't need to be re-downloaded between runs.

Run these from the `docker/` directory.

### Start services
```bash
docker compose up
```

### Docker build services
```bash
docker compose build maahinen
```

### Run interactively
```bash
docker compose run maahinen
```

### Start shell
```bash
docker compose exec maahinen bash
```

### Stop services
```bash
docker compose down
```

### Full reset (removes downloaded models)
```bash
docker compose down -v
```