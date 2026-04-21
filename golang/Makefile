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

clean-registry:
	@echo "🧹 Limpando versões antigas no Artifact Registry para economizar espaço..."
	@DIGESTS=$$(gcloud artifacts docker images list us-central1-docker.pkg.dev/alexa-inteligente/gcf-artifacts/alexa--inteligente__us--central1__alexa--llm--go --format='value(digest)' --sort-by='~UPDATE_TIME' | tail +2); \
	for digest in $$DIGESTS; do \
		echo "Deletando digest: $$digest"; \
		gcloud artifacts docker images delete "us-central1-docker.pkg.dev/alexa-inteligente/gcf-artifacts/alexa--inteligente__us--central1__alexa--llm--go@$$digest" --quiet --delete-tags; \
	done

deploy: clean-registry
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
	  --min-instances=0 \
	  --max-instances=5 \
	  --memory=256Mi \
	  --set-env-vars OPENROUTER_API_KEY="$$OPENROUTER_API_KEY",MODEL_NAME="$$MODEL_NAME",ALEXA_SKILL_ID="$$ALEXA_SKILL_ID",ALEXA_SECRET_TOKEN="$$ALEXA_SECRET_TOKEN",SUPABASE_URL="$$SUPABASE_URL",SUPABASE_SERVICE_ROLE_KEY="$$SUPABASE_SERVICE_ROLE_KEY",DASHBOARD_USER="$$DASHBOARD_USER",DASHBOARD_PASS="$$DASHBOARD_PASS"'
