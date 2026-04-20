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

deploy:
	@echo "📌 Iniciando deploy para o Google Cloud Functions..."
	@bash -c 'set -a; source ../.env; set +a; \
	gcloud functions deploy alexa-llm-go \
	  --gen2 \
	  --runtime=go124 \
	  --region=us-central1 \
	  --source=. \
	  --entry-point=HandleAlexaRequest \
	  --trigger-http \
	  --allow-unauthenticated \
	  --set-env-vars OPENROUTER_API_KEY="$$OPENROUTER_API_KEY",MODEL_NAME="$$MODEL_NAME",ALEXA_SKILL_ID="$$ALEXA_SKILL_ID",ALEXA_SECRET_TOKEN="$$ALEXA_SECRET_TOKEN"'
