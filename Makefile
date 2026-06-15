.PHONY: install dev backend frontend build release

PORT := 6967

install:
	cd frontend && npm install
	cd backend && go mod download

# Run both dev servers (frontend proxies /api to backend)
dev:
	@echo "Backend → http://localhost:$(PORT)  |  Frontend → http://localhost:5173"
	@trap 'kill 0' INT; \
	(cd backend && go run .) & \
	(cd frontend && npm run dev) & \
	wait

# Production: build frontend then run Go server (serves everything on :PORT)
backend:
	cd backend && go run .

frontend:
	cd frontend && npm run dev

# Build optimised release binary (embeds built frontend)
build: build-frontend
	cd backend && go build -ldflags="-s -w" -o ../dist/claude-stats .
	@echo "Binary: ./dist/claude-stats"
	@echo "Run:    ./dist/claude-stats"

build-frontend:
	cd frontend && npm run build

# Cross-compile for macOS arm64 (Apple Silicon)
release-macos-arm64: build-frontend
	cd backend && GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ../dist/claude-stats-macos-arm64 .

# Cross-compile for macOS amd64 (Intel)
release-macos-amd64: build-frontend
	cd backend && GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ../dist/claude-stats-macos-amd64 .

# Cross-compile for Linux amd64
release-linux: build-frontend
	cd backend && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../dist/claude-stats-linux-amd64 .

release: release-macos-arm64 release-macos-amd64 release-linux
	@echo "Release binaries in ./dist/"
	@ls -lh dist/
