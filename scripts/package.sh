#!/bin/bash
set -e

echo "📦 Iniciando processo de empacotamento para AWS Lambda..."

# Local da saída
DIST_DIR="dist"
PACKAGE_ZIP="lambda_package.zip"

# Limpar versões anteriores
rm -rf $DIST_DIR $PACKAGE_ZIP
mkdir -p $DIST_DIR

# 1. Instalar dependências na pasta dist (importante: sem as de teste p/ economizar espaço)
echo "📥 Instalando dependências em $DIST_DIR..."
pip install --no-cache-dir -r requirements.txt -t $DIST_DIR

# 2. Copiar o código fonte (src) para a raiz do pacote (exigência da Alexa SDK no Lambda)
echo "📂 Copiando código fonte..."
cp -r src/* $DIST_DIR/

# 3. Gerar o arquivo ZIP
echo "🤐 Gerando $PACKAGE_ZIP..."
cd $DIST_DIR
zip -r ../$PACKAGE_ZIP . -x "**/__pycache__/*" "*.dist-info/*" "*.pyc"
cd ..

echo "✅ Pacote pronto! Tamanho: $(du -sh $PACKAGE_ZIP | cut -f1)"
echo "🚀 Agora basta fazer o upload de '$PACKAGE_ZIP' no console da AWS Lambda."
