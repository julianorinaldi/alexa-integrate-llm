.PHONY: build up down logs

DOCKER_COMPOSE=docker compose -f docker-compose.yml

build:
	@echo "Construindo container Go..."
	$(DOCKER_COMPOSE) build

up:
	@echo "Subindo aplicação Go na porta 5000..."
	$(DOCKER_COMPOSE) up

down:
	@echo "Derrubando os containers..."
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f
