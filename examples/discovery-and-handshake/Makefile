.PHONY: build up test down clean logs all

# Build the lango Docker image from repo root
build:
	docker compose build

# Start both agents
up:
	docker compose up -d

# Run integration tests (agents must be healthy first)
test:
	@echo "Waiting for all agents to be healthy..."
	@sh scripts/wait-for-health.sh http://localhost:18789/health 90
	@sh scripts/wait-for-health.sh http://localhost:18790/health 90
	@echo "Running integration tests..."
	@sh scripts/test-discovery.sh

# Stop and remove all containers
down:
	docker compose down

# Full cleanup: remove containers, volumes, and orphans
clean:
	docker compose down -v --remove-orphans

# Tail logs from all services
logs:
	docker compose logs -f

# One-shot: build, start, test, then stop
all: build up test down
