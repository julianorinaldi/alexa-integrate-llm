# 🚀 Alexa Integrate LLM (Echo Show Focused)

Este projeto integra a **Alexa** com qualquer modelo de Inteligência Artificial via **OpenRouter** (GPT-4, Claude 3, Llama 3, etc.), com suporte completo a visualização no **Echo Show** usando APL.

O foco aqui é o baixo custo (AWS Free Tier) e desenvolvimento resiliente usando **containers**.

---

## 🛠️ Variáveis de Ambiente (.env)

Crie um arquivo `.env` na raiz do projeto (use o `.env.example` como guia):

| Variável | Descrição | Exemplo |
| :--- | :--- | :--- |
| `OPENROUTER_API_KEY` | Sua chave de API do OpenRouter. | `sk-or-v1-...` |
| `MODEL_NAME` | Nome do modelo que a Alexa deve usar. | `openai/gpt-3.5-turbo` |
| `ALEXA_SKILL_ID` | (Opcional) Verificação de segurança p/ sua Lambda. | `amzn1.ask.skill...` |
| `DEBUG` | Se `True`, exibe logs detalhados no CloudWatch. | `True` |

---

## 🏗️ Desenvolvimento em Container

Toda a compilação e testes são feitos dentro do Docker para evitar conflitos de versão local.

### 1. Construir o ambiente
```bash
make build
```

### 2. Depurador Local em Tempo Real (Interativo)
Se você quer ver as alterações no código refletidas **na hora** no seu dispositivo Alexa ou simulador, sem fazer deploy:

1.  **Inicie o servidor local no container:**
    ```bash
    make run-local
    ```
    *Isso iniciará um servidor Flask na porta 5000 dentro do Docker.*

2.  **Crie um túnel público (Ex: ngrok):**
    No seu terminal local (fora do docker), abra um túnel para a porta 5000:
    ```bash
    ngrok http 5000
    ```
    *Copie a URL `https` gerada (ex: `https://abcd-123.ngrok-free.app`).*

3.  **Configure no Alexa Developer Console:**
    *   Em **Skill Architecture > Endpoint**, selecione **HTTPS**.
    *   No campo **Default Region**, cole a URL do ngrok.
    *   Selecione: *"My development endpoint is a sub-domain of a domain that has a wildcard certificate from a certificate authority"*.
    *   Clique em **Save Endpoints**.

4.  **Teste agora!** 
    Agora, ao falar com a Alexa no console ou celular, o sinal chegará diretamente no código da sua máquina e você verá os logs no terminal.

### 3. Testar via Mocks (Automatizado)
Você não precisa de internet ou da Alexa para saber se a lógica fundamental está correta:
```bash
make test
```
*   **O que isso testa?** Se a chamada para a IA está correta e se os Handlers da Alexa estão gerando o JSON de resposta esperado (incluindo o visual APL).

### 4. Acessar o terminal do projeto
```bash
make shell
```

---

## 🌍 Guia de Deploy (Produção)

### 1. Gerar o Pacote para o AWS Lambda
Nosso sistema já gera o arquivo `.zip` otimizado:
```bash
make package
```
*Isso criará o arquivo `lambda_package.zip` na raiz do seu projeto.*

### 2. Configurar a Lambda na AWS
1.  Vá ao console **AWS Lambda** e clique em **Create Function**.
2.  Escolha **Author from scratch**, Nome: `alexa-llm-skill`, runtime: **Python 3.12**.
3.  Em **Configuration > Environment Variables**, adicione as chaves que você definiu no `.env`.
4.  Em **Triggers**, adicione **Alexa Skills Kit** e insira o ID da sua skill.

### 3. Configurar a Skill na Amazon (Developer Console)
1.  **Endpoint:** Aponte para o ARN da sua Lambda.
2.  **Interaction Model:** 
    *   Vá em **JSON Editor**.
    *   Arraste o arquivo `skill-package/interactionModel/custom/pt-BR.json` para o console.
    *   Clique em **Save** e **Build Model**.
3.  **Interfaces:** Ative a opção **Alexa Presentation Language** (Obrigatório p/ Echo Show).

---

## 🧠 Como funciona a Arquitetura

1.  **Voz do Usuário:** Capturada pela Alexa.
2.  **Interaction Model:** Extrai o texto e envia para nossa Lambda.
3.  **Lambda (Python):**
    *   Recupera o histórico da sessão (Contexto).
    *   Envia p/ o **OpenRouter** com o histórico.
    *   Recebe a resposta e gera o JSON da Alexa.
    *   Renderiza o **Template APL** p/ o Echo Show.
4.  **Resposta:** Alexa fala o texto e exibe o cartão visual na tela.

---

## 🧪 Testando em Desenvolvimento
Para testar novas funcionalidades:
1.  Crie um novo teste em `tests/unit/`.
2.  Use o comando `make test`.
3.  Isso garante que a lógica da LLM não quebrou antes de você atualizar a Lambda.

---
*Criado com ❤️ para integração de IA e IoT.*