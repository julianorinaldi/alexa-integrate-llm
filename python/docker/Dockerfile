FROM python:3.12-slim

# Evitar que o Python gere arquivos .pyc e garantir que logs saiam na hora certa
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

# Instalar dependências do sistema necessárias para compilação (se houver)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    make \
    zip \
    && rm -rf /var/lib/apt/lists/*

# Copiar e instalar dependências Python
COPY ../requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiar o código do projeto
COPY .. .

# Comando padrão: Rodar testes
CMD ["pytest", "tests/"]
