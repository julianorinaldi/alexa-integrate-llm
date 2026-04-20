# 🚀 Alexa Integrate LLM (Echo Show Focused)

Este projeto integra a **Alexa** com qualquer modelo de Inteligência Artificial via **OpenRouter** (GPT-4, Claude 3, Llama 3, etc.), com suporte completo a visualização no **Echo Show** usando APL.

O projeto foi modernizado e agora suporta duas stacks (Go e Python), sendo **Go (Golang) a linguagem principal** com foco em Cloud Functions e baixa latência (Cold Start incrivelmente rápido).

---

## 📂 Organização do Repositório

* **`golang/`**: Diretório principal contendo a stack em **Go** otimizada para Cloud Functions (GCP/AWS).
* **`python/`**: Código legado contendo a versão original rodando em ambiente Python 3.12 para AWS Lambda.

---

## 🛠️ Variáveis de Ambiente (.env)

Crie um arquivo `.env` na raiz do projeto:

| Variável | Descrição | Exemplo |
| :--- | :--- | :--- |
| `OPENROUTER_API_KEY` | Sua chave de API do OpenRouter. | `sk-or-v1-...` |
| `MODEL_NAME` | Nome do modelo que a Alexa deve usar. | `openai/gpt-3.5-turbo` |
| `PORT` | Porta HTTP local (padrão 5000). | `5000` |
| `ALEXA_SKILL_ID` | (Segurança) Restringe o acesso apenas à sua Alexa Skill. | `amzn1.ask.skill...` |
| `ALEXA_SECRET_TOKEN` | (Segurança) Exige `?token=` na URL do seu Webhook. | `UmaSenhaComplexa123` |

---

## 🎙️ Configurando a Skill na Amazon Alexa

Para que a Alexa entenda o que o usuário fala e encaminhe para o nosso código, precisamos configurar a "casca" da Skill no portal da Amazon. Todos os arquivos necessários para isso estão na pasta **`skill-package/`** (na raiz do projeto).

### Passo a Passo no Developer Console:
1. **Crie a Skill**:
   - Acesse o [Alexa Developer Console](https://developer.amazon.com/alexa/console/ask).
   - Clique em **Create Skill**, dê um nome (ex: "Meu Assistente LLM"), escolha o idioma principal (ex: `pt-BR`), modelo *Custom* e método de hospedagem *Provision your own*.
2. **Pegue o Skill ID para Segurança (Opcional, mas recomendado)**:
   - Na listagem das suas Skills, clique em **Copy Skill ID** (ex: `amzn1.ask.skill.xxxx...`) e cole-o no seu arquivo `.env` na variável `ALEXA_SKILL_ID`.
3. **Importe o Modelo de Interação (Intents)**:
   - No menu lateral da sua skill, vá em **Interaction Model** -> **JSON Editor**.
   - Arraste ou copie todo o conteúdo do arquivo localizado em `skill-package/interactionModel/custom/pt-BR.json` para dentro do editor.
   - Clique em **Save Model** e em seguida **Build Model**. Isso ensina a Alexa a ouvir frases como *"pergunte {query}"*.
4. **Habilite a Tela (Visuais APL)**:
   - No menu lateral, acesse **Tools** -> **Interfaces**.
   - Ative a chave **Alexa Presentation Language** (isso é obrigatório para as telas do Echo Show funcionarem sem erro).
5. **Aponte para o seu Servidor Go**:
   - No menu lateral, acesse **Endpoint**.
   - Selecione **HTTPS**. Em *Default Region*, cole a URL onde o seu serviço Go está rodando (via Cloudflare Tunnel para testes locais, ou a URL oficial da sua Google Cloud Function em produção).
   - *Lembrete de Segurança*: Não se esqueça de adicionar o `?token=UmaSenhaComplexa123` no final da URL se configurou o `ALEXA_SECRET_TOKEN`. Atualmente estou usando um subdomínio e tratando nele o endereço correto: https://alexa.numeric.com.br?token=UmaSenhaMuitoComplexa123

---

## 🚀 Ambiente de Desenvolvimento DEV Local (GO)

Nossa stack de produção atual roda em Go, o que nos garante agilidade e altíssima performance. Usamos o **Docker** para padronizar e testar localmente.

### 1. Construir e Rodar o Container (GO)
Dentro da pasta `golang/`, execute os comandos:
```bash
cd golang/
make build   # Compila a imagem Docker Multistage do Go
make up      # Sobe a aplicação escutando na porta 5000
```

*Dica: Você pode usar `make logs` para monitorar os acessos em tempo real da sua aplicação.*

### 2. Tunelamento para Testes na Alexa (Cloudflare Tunnel)
Com seu servidor rodando localmente (porta 5000), você pode mapear sua porta de maneira segura para a internet usando o Cloudflare Tunnel (antigo Argo Tunnel).

A partir de outro terminal, chame o utilitário `cloudflared`:
```bash
cloudflared tunnel --url http://localhost:5000
```

1. Observe a saída do terminal; o Cloudflare irá gerar uma URL pública aleatória (ex: `https://palavra-aleatoria-outra.trycloudflare.com`).
2. Copie essa URL segura `HTTPS`.
3. Vá no console de desenvolvimento da Alexa, na seção **Endpoint** da aba de Build.
4. Marque como Default Region e cole a URL fornecida pelo Cloudflare.
5. No menu de validação de SSL, selecione a opção: *"My development endpoint is a sub-domain of a domain that has a wildcard certificate from a certificate authority"*.
6. Salve o endpoint (*Save Endpoints*).

A partir deste momento, a sua Alexa enviará todos os diálogos de voz tunelados direto para a sua máquina local rodando o Go.

---

## 🌍 Guia de Deploy em Produção (GO)

Como a API é escrita nativamente em Go adotando a estrutura do `net/http`, você possui total liberdade para deployar a função em qualquer lugar (Docker, Cloud Run, AWS Lambda Webhook), mas o código já vem **preparado nativamente para Google Cloud Functions**.

### Google Cloud Functions (Recomendado)
A nossa função se chama `HandleAlexaRequest` e está configurada nativamente para atender à assinatura exigida pelo GCP ("HTTP Functions"). Nós criamos um script automatizado no `Makefile` que extrai suas variáveis do arquivo `.env` para te poupar tempo!

**Passo a passo do primeiro deploy:**
1. Autentique-se e defina seu projeto no CLI:
   ```bash
   gcloud auth login
   gcloud config set project O_NOME_DO_SEU_PROJETO
   ```
2. Ative as APIs obrigatórias do Google (só na primeira vez):
   ```bash
   gcloud services enable cloudfunctions.googleapis.com \
     cloudbuild.googleapis.com \
     run.googleapis.com \
     artifactregistry.googleapis.com
   ```
3. Preencha seu arquivo `.env` com sua chave do OpenRouter, IDs da Alexa e Tokens.
4. Faça o Deploy Automático!
   ```bash
   cd golang/
   make deploy
   ```
*O comando injetará sozinho suas variáveis secretas e após 2 minutos te retornará a URL oficial da sua função (ex: `https://us-central1-...cloudfunctions.net/alexa-llm-go`). Essa é a URL para botar no Portal da Alexa!*

### Deploy Como Container (Google Cloud Run / AWS ECS / VPS)
É possível hospedar seu container nativamente onde preferir graças ao `Dockerfile` de Multi-stage, que gera uma imagem final minúscula:
```bash
cd golang/
docker build -t meu-usuario/alexa-llm-go:latest .
docker run -p 8080:5000 -e PORT=5000 -e OPENROUTER_API_KEY="..." meu-usuario/alexa-llm-go:latest
```

---

## 🧠 Arquitetura Legada em Python (`python/`)

Todo o arcabouço antigo feito primariamente para AWS Lambda em Python foi empacotado para a pasta `/python/`.

Se você preferir continuar usando o provedor nativo em Python, os comandos de desenvolvimento legado `make shell`, `make build`, e empacotadores da AWS Lambda continuam existindo de forma idêntica e sem conflito dentro da pasta `python/`.

---
*Criado com ❤️ para integração de IA usando alta performance em Go.*

## URL Importantes
AWS Developer Alexa: https://developer.amazon.com/alexa/console/ask
GCP Console: https://console.cloud.google.com/home
