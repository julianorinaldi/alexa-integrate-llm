# 🚀 Alexa Integrate LLM (Echo Show Focused)

Este projeto integra a **Alexa** com qualquer modelo de Inteligência Artificial via **OpenRouter** (GPT-4, Claude 3, Llama 3, etc.), com suporte completo a visualização no **Echo Show** usando APL.

O projeto foi modernizado e agora suporta duas stacks (Go e Python), sendo **Go (Golang) a linguagem principal** com foco em Cloud Functions e baixa latência (Cold Start incrivelmente rápido).

## 📂 Organização do Repositório

* **`golang/`**: Diretório principal contendo a stack em **Go 1.24** otimizada para Cloud Functions (GCP).
* **`python/`**: Código legado contendo a versão original rodando em ambiente Python 3.12 para AWS Lambda.

---

## 🛠️ Variáveis de Ambiente (.env)

Crie um arquivo `.env` na raiz do projeto (o arquivo `.gitignore` já está configurado para não subir suas chaves):

| Variável | Descrição | Exemplo |
| :--- | :--- | :--- |
| `OPENROUTER_API_KEY` | Sua chave de API do OpenRouter. | `sk-or-v1-...` |
| `MODEL_NAME` | Nome do modelo (recomenda-se modelos Flash para baixa latência). | `google/gemini-2.0-flash-lite` |
| `ALEXA_SKILL_ID` | (Segurança) Restringe o acesso apenas à sua Alexa Skill. | `amzn1.ask.skill...` |
| `ALEXA_SECRET_TOKEN` | (Segurança) Exige `?token=` na URL do seu Webhook. | `UmaSenhaComplexa123` |

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

---

## 🚀 Ambiente de Desenvolvimento DEV Local (GO)

Nossa stack utiliza **Go 1.24**. Para rodar localmente:

```bash
cd golang/
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
cd golang/
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