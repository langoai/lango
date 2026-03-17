.PHONY: build up test down clean logs all

# Build the lango Docker image from repo root
build:
	docker compose build

# Start all services (anvil → setup → agents)
up:
	docker compose up -d

# Run integration tests (agents must be healthy first)
test:
	@echo "Waiting for agents to be healthy..."
	@sh scripts/wait-for-health.sh http://localhost:18789/health 120
	@sh scripts/wait-for-health.sh http://localhost:18790/health 120
	@echo "Running integration tests..."
	@sh scripts/test-escrow.sh

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
