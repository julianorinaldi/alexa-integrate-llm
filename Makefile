# Atalhos para o projeto Alexa LLM

DOCKER_COMPOSE=docker compose -f docker/docker-compose.yml

.PHONY: build test run shell clean package run-local

build:
	@echo "Construindo o container de desenvolvimento..."
	$(DOCKER_COMPOSE) build

run-local:
	@echo "Iniciando servidor de debug local na porta 5000..."
	$(DOCKER_COMPOSE) run --rm --service-ports app python src/local_server.py

package:
	@echo "Empacotando projeto no container para o AWS Lambda..."
	$(DOCKER_COMPOSE) run --rm app bash scripts/package.sh

test:
	@echo "Rodando testes unitários e de integração no container..."
	$(DOCKER_COMPOSE) run --rm app pytest tests/

# Caso decida rodar um servidor de desenvolvimento local para Alexa (ex: com ngrok)
run:
	@echo "Rodando aplicação localmente..."
	$(DOCKER_COMPOSE) up

shell:
	@echo "Abrindo terminal dentro do container..."
	$(DOCKER_COMPOSE) run --rm app /bin/bash

clean:
	@echo "Limpando arquivos temporários..."
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
