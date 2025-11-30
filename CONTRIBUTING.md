# Contributing

## Development

1. Work on `dev` branch
2. Create PR to `main` when ready
3. Wait for CI to pass
4. Merge

## Releasing

1. Merge `dev` to `main` via PR
2. Tag the release:
```bash
   git checkout main
   git pull origin main
   git tag v0.2.0
   git push origin v0.2.0
```
3. Create GitHub Release with release notes