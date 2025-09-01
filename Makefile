.PHONY: help test test-deps test-format test-vet test-full clean

help:
	@echo "Available commands:"
	@echo "  test        - Run all tests"
	@echo "  test-deps   - Install test dependencies"
	@echo "  test-format - Format code with go fmt"
	@echo "  test-vet    - Run go vet static analysis"
	@echo "  test-full   - Run full test suite (format + vet + tests)"
	@echo "  clean       - Clean test cache"

test:
	@echo "🧪 Running all tests..."
	cd backend && go test ./... -v

test-format:
	@echo "📝 Formatting code..."
	cd backend && go fmt ./...

test-vet:
	@echo "🔍 Running static analysis..."
	cd backend && go vet ./...

test-full: test-deps test-format test-vet test
	@echo "✅ Full test suite completed"

test-deps:
	@echo "📦 Installing test dependencies..."
	cd backend && go mod download
	cd backend && go mod vendor

clean:
	@echo "🧹 Cleaning test cache..."
	cd backend && go clean -cache -modcache -testcache -fuzzcache
	@echo "✅ Cleanup completed"
