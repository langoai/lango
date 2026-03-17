.PHONY: build up test down clean logs all

build:
	docker compose build

up:
	docker compose up -d

test:
	@echo "Waiting for all agents to be healthy..."
	@sh scripts/wait-for-health.sh http://localhost:18789/health 120
	@sh scripts/wait-for-health.sh http://localhost:18790/health 120
	@sh scripts/wait-for-health.sh http://localhost:18791/health 120
	@sh scripts/wait-for-health.sh http://localhost:18792/health 120
	@echo "Running integration tests..."
	@sh scripts/test-team.sh

down:
	docker compose down

clean:
	docker compose down -v --remove-orphans

logs:
	docker compose logs -f

all: build up test down
