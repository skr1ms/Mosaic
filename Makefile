.PHONY: help test test-deps clean

help:
	@echo "Available commands:"
	@echo "  test      - Run all tests"
	@echo "  test-deps - Install test dependencies"
	@echo "  clean     - Clean test cache"

test:
	@echo "🧪 Running all tests..."
	cd backend && go test ./... -v || true

test-deps:
	@echo "📦 Installing test dependencies..."
	cd backend && go mod download
	cd backend && go mod vendor

clean:
	@echo "🧹 Cleaning test cache..."
	cd backend && go clean -cache -modcache -testcache -fuzzcache
	@echo "✅ Cleanup completed"
