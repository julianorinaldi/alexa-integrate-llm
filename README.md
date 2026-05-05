# 🚀 Alexa Integrate LLM (Echo Show Focused)

Este projeto integra a **Alexa** com qualquer modelo de Inteligência Artificial via **OpenRouter** (GPT-4, Claude 3, Llama 3, etc.), com suporte completo a visualização no **Echo Show** usando APL.

O projeto utiliza **Go (Golang)** como linguagem principal com foco em Cloud Functions e baixa latência (Cold Start incrivelmente rápido).

## 📂 Organização do Repositório

* **`golang/`**: Diretório contendo a stack em **Go 1.24** otimizada para Cloud Functions (GCP).
* **`skill-package/`**: Modelo de interação e assets da Alexa Skill.

---

## 🛠️ Variáveis de Ambiente (.env)

Crie um arquivo `.env` na raiz do projeto (o arquivo `.gitignore` já está configurado para não subir suas chaves):

| Variável | Descrição | Exemplo |
| :--- | :--- | :--- |
| `OPENROUTER_API_KEY` | Sua chave de API do OpenRouter. | `sk-or-v1-...` |
| `MODEL_NAME` | Nome do modelo (recomenda-se modelos Flash para baixa latência). | `google/gemini-2.0-flash-lite` |
| `ALEXA_SKILL_ID` | (Segurança) Lista de IDs permitidos (separados por vírgula). | `amzn1.ask.skill...,amzn2...` |
| `ALEXA_SECRET_TOKEN` | (Segurança) Token global compartilhado para acesso. | `UmaSenhaComplexa123` |
| `DB_PATH` | Caminho do arquivo SQLite (dentro do container). | `/data/alexa.db` |
| `DASHBOARD_USER` | Usuário para acessar o painel administrativo. | `admin` |
| `DASHBOARD_PASS` | Senha para acessar o painel administrativo. | `mudar123` |

---

## 🎙️ Configurando a Skill na Amazon Alexa

Para que a Alexa entenda o que o usuário fala e encaminhe para o nosso código, precisamos configurar a "casca" da Skill no portal da Amazon.

### Passo a Passo no Developer Console:
1. **Crie a Skill**:
   - Acesse o [Alexa Developer Console](https://developer.amazon.com/alexa/console/ask).
   - Escolha o idioma `pt-BR`, modelo *Custom* e método de hospedagem *Provision your own*.
2. **Importe o Modelo de Interação (Intents)**:
   - Vá em **Interaction Model** -> **JSON Editor**.
   - Use o conteúdo do arquivo `skill-package/interactionModel/custom/pt-BR.json`.
   - Clique em **Save Model** e em **Build Model**. **IMPORTANTE:** A intenção configurada é a `AskIntent`. Se você mudar o nome no console, deverá mudar no código Go também.
3. **Interfaces (Echo Show)**:
   - Ative a chave **Alexa Presentation Language** para suporte a telas.
4. **Endpoint**:
   - Selecione **HTTPS**. Em *Default Region*, cole a URL da sua Cloud Function com o token:
     `https://SUA_URL_GCP/alexa-llm-go?token=SEU_TOKEN`
    - SSL: Selecione *"My development endpoint is a sub-domain of a domain that has a wildcard certificate..."*.

### 🔄 Captura Total (Ask-All)
O modelo de interação foi configurado para capturar **qualquer frase ou pergunta** (via `AMAZON.SearchQuery`). Isso significa que o usuário não precisa usar comandos prefixados. Toda a fala é enviada diretamente ao LLM para processamento contextual.

---

## 🖥️ Painel Administrativo e Multi-Usuários

Agora você pode compartilhar seu backend com outras pessoas de forma segura através do Painel Administrativo embutido.

### Como funciona:
1. **Acesso**: Navegue até `https://SUA_URL_GCP/admin`.
2. **Login**: Use as credenciais configuradas nas variáveis `DASHBOARD_USER` e `DASHBOARD_PASS`.
3. **Gerenciamento**:
   - Cadastre novos **Skill IDs** e **Secret Tokens** individuais para amigos ou clientes.
   - O sistema valida a permissão consultando tanto o `.env` (acesso mestre) quanto o banco de dados SQLite em tempo real.

---

## 📦 Banco de Dados (SQLite)

O projeto utiliza **SQLite** para persistência local, dentro do container. O arquivo é criado automaticamente na primeira execução — nenhuma configuração manual de banco é necessária.

### Estrutura da Tabela:
A tabela `authorized_skills` armazena:
- `id`: UUID gerado automaticamente pelo SQLite.
- `skill_id`: O identificador exclusivo da Alexa Skill.
- `secret_token`: O token de segurança que deve ser passado na URL do webhook.
- `owner_name`: Nome da pessoa/dispositivo para identificação.

### Volume Docker (Backup):
O arquivo SQLite fica em `/data/alexa.db` dentro do container, mapeado para o volume nomeado `alexa_data`.

```bash
# Fazer backup do banco
docker run --rm \
  -v alexa_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/alexa-backup.tar.gz /data

# Restaurar backup
docker run --rm \
  -v alexa_data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/alexa-backup.tar.gz -C /
```

---

## 🚀 Ambiente de Desenvolvimento DEV Local

Nossa stack utiliza **Go 1.24**. Para rodar localmente usando o Makefile da raiz:

```bash
make build   # Cria a imagem Docker
make up      # Roda na porta 5000
```

---

## 🌍 Guia de Deploy e Free Tier (GCP)

O projeto está configurado para rodar dentro do **GCP Free Tier** (nível gratuito).

### 1. Preparação
```bash
gcloud auth login
gcloud config set project NOME_DO_PROJETO
```

### 2. Deploy com Controle de Custos
O comando `make deploy` foi otimizado com:
- `--min-instances=0`: Não cobra quando não há uso.
- `--max-instances=5`: Limita gastos em caso de picos de uso.
- `--memory=256Mi`: Equilíbrio entre performance e custo.

```bash
make deploy
```

### 3. Manutenção de Espaço (Artifact Registry)
O Google cobra pelo armazenamento de imagens antigas. O comando de deploy agora executa automaticamente o:
```bash
make clean-registry
```
Este comando apaga versões anteriores do container no **Artifact Registry**, mantendo apenas a última e evitando cobranças por armazenamento excedente (limite free de 500MB).

---

## 🧠 Personalização e Idioma
O sistema agora possui um **System Prompt** fixo no arquivo `llm/client.go` que força a IA a:
- Responder sempre em **Português do Brasil (PT-BR)**.
- Ser concisa e direta (ideal para voz).

---

## URL Importantes
AWS Developer Alexa: https://developer.amazon.com/alexa/console/ask

GCP Console: https://console.cloud.google.com/home

Endpoint publicado: https://us-central1-alexa-inteligente.cloudfunctions.net/alexa-llm-go

Volume Docker (SQLite): `docker volume inspect alexa_data`