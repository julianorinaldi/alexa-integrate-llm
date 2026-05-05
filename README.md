# 🚀 Alexa Integrate LLM (Echo Show Focused)

Este projeto integra a **Alexa** com qualquer modelo de Inteligência Artificial via **OpenRouter** (GPT-4, Claude 3, Llama 3, etc.), com suporte completo a visualização no **Echo Show** usando APL.

O projeto utiliza **Go (Golang)** como linguagem principal com foco em Cloud Functions e baixa latência (Cold Start incrivelmente rápido).

## 📂 Organização do Repositório

* **`golang/`**: Diretório contendo a stack em **Go 1.24** otimizada para Cloud Functions (GCP).
* **`skill-package/`**: Modelo de interação e assets da Alexa Skill.

---
## 🛠️ Configuração Inicial (Sem Variáveis de Ambiente!)

O sistema evoluiu para **eliminar a necessidade de arquivos `.env`** com credenciais difíceis de manter.
Toda a configuração de chaves (API) e cadastro de Skills agora é feita via **Painel Administrativo Web**, que persiste as informações em um banco de dados SQLite local mapeado por volume.

**Única variável possível no `.env` (Opcional)**:
`PORT=5000` (Define a porta local para debug).

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
   - Selecione **HTTPS**. Em *Default Region*, cole a URL da sua Cloud Function limpa:
     `https://SUA_URL_GCP/alexa-llm-go`
    - SSL: Selecione *"My development endpoint is a sub-domain of a domain that has a wildcard certificate..."*.

### 🔄 Captura Total (Ask-All)
O modelo de interação foi configurado para capturar **qualquer frase ou pergunta** (via `AMAZON.SearchQuery`). Isso significa que o usuário não precisa usar comandos prefixados. Toda a fala é enviada diretamente ao LLM para processamento contextual.

---

## 🖥️ Painel Administrativo e Multi-Usuários

Agora você pode compartilhar seu backend com outras pessoas de forma segura através do Painel Administrativo embutido.

### Como funciona:
1. **Acesso**: Navegue até `https://SUA_URL_GCP/admin`.
2. **Login Inicial**: No primeiro acesso, utilize o usuário padrão **`admin`** e a senha **`admin`**.
3. **Segurança**: O sistema detectará o acesso inicial e forçará a troca de senha imediatamente. A nova senha será salva de forma segura no SQLite.
3. **Gerenciamento Completo**:
   - **Modelos LLM:** Cadastre o nome do serviço, sua chave (`OPENROUTER_API_KEY`) e o modelo (`MODEL_NAME`) que a IA vai usar.
   - **Dispositivos/Skills:** Cadastre novos **Skill IDs** e **Secret Tokens** individuais para amigos ou clientes.
   - Apenas dispositivos com `Skill ID` e `Token` correspondentes ao cadastrado no banco SQLite poderão acessar a IA.

---

## 📦 Banco de Dados (SQLite)

O projeto utiliza **SQLite** para persistência local, dentro do container. O arquivo é criado automaticamente na primeira execução — nenhuma configuração manual de banco é necessária.

### Estrutura da Tabela:
A tabela `authorized_skills` armazena as skills autorizadas, enquanto a tabela `dashboard_users` armazena as credenciais de acesso ao painel. Ambos os dados persistem no volume Docker.

### Mapeamento de Volume Local (Persistência):
O arquivo SQLite e demais dados persistentes ficam armazenados no diretório **`volume/data`** na raiz do projeto. Este diretório é mapeado para `/data` dentro do container no `docker-compose`. 

Isso significa que você tem acesso direto ao arquivo do banco pelo seu sistema host.

```bash
# Fazer backup do banco (cria um arquivo alexa-backup.tar.gz)
make backup

# Restaurar backup (descompacta de volta para volume/data)
make restore
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