# Docker containerization

> [!NOTE]
> It is recommended to run the tool inside a container for safety of the system as the tool has access to the system.

### Build
```bash
docker build -f docker/Dockerfile -t maahinen .
```

### Run
```bash
docker run -it maahinen
```