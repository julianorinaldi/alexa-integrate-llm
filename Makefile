# 🚀 Alexa Integrate LLM - Root Makefile

.PHONY: help build up down logs deploy clean-registry backup restore

help:
	@echo "Uso: make [comando]"
	@echo ""
	@echo "Comandos de Desenvolvimento:"
	@echo "  build           Constrói o container da aplicação (Go)"
	@echo "  up              Inicia os containers em modo interativo"
	@echo "  down            Para e remove os containers"
	@echo "  logs            Exibe os logs dos containers"
	@echo ""
	@echo "Comandos de Deploy (GCP):"
	@echo "  deploy          Executa o deploy para o Google Cloud Functions"
	@echo "  clean-registry  Limpa versões antigas no Artifact Registry"
	@echo ""
	@echo "Comandos de Banco de Dados:"
	@echo "  backup          Cria um backup do banco SQLite (alexa-backup.tar.gz)"
	@echo "  restore         Restaura o banco a partir do backup"



build:
	$(MAKE) -C golang build

up:
	$(MAKE) -C golang up

down:
	$(MAKE) -C golang down

logs:
	$(MAKE) -C golang logs

deploy:
	$(MAKE) -C golang deploy

clean-registry:
	$(MAKE) -C golang clean-registry

backup:
	@echo "📦 Criando backup do banco de dados..."
	tar -czf alexa-backup.tar.gz -C volume/data .
	@echo "✅ Backup concluído: alexa-backup.tar.gz"

restore:
	@echo "🔄 Restaurando backup do banco de dados..."
	mkdir -p volume/data
	tar -xzf alexa-backup.tar.gz -C volume/data
	@echo "✅ Restauração concluída"
