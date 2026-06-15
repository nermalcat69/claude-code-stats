# Development

## Setup

```bash
make install   # install deps (npm + go mod)
make dev       # backend :6967 + frontend :5173
```

## Build

```bash
make build     # produces ./dist/claude-stats (embeds built frontend)
```

## Release

Push a semver tag — GitHub Actions builds and publishes binaries automatically:

```bash
git tag v1.0.0 && git push origin v1.0.0
```

Binaries are cross-compiled for macOS arm64/amd64, Linux amd64, and Windows amd64.
