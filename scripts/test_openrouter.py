import os
from dotenv import load_dotenv

# Carrega o .env antes de importar o client
load_dotenv()

from src.llm.client import llm_client

def test_api():
    print("=======================================")
    print("🤖 Teste de Conexão com OpenRouter")
    print("=======================================")
    
    key = llm_client.api_key
    safe_key = f"{key[:10]}...{key[-5:]}" if key and len(key) > 15 else str(key)
    print(f"🔑 Chave identificada: {safe_key}")
    print(f"🧠 Modelo configurado: {llm_client.model}")
    print("---------------------------------------")
    
    pergunta = "Responda em apenas uma frase curta: qual o tamanho da lua?"
    print(f"🗣️  Enviando pergunta: '{pergunta}'")
    print("Aguardando resposta do modelo... ⏳")
    print("---------------------------------------")
    
    try:
        resposta = llm_client.ask(pergunta)
        
        if "Sinto muito, tive um erro" in resposta:
            print("\n❌ FALHA: A API retornou erro. Verifique sua chave no OpenRouter ou se tem saldo no modelo escolhido.")
        else:
            print("\n✅ SUCCESSO! Conexão validada. O modelo respondeu:")
            print("\n" + resposta + "\n")
            print("=======================================")
            print("Pode testar na Alexa agora!")
            
    except Exception as e:
        print(f"\n❌ ERRO FATAL: Houve uma quebra no código de envio: {e}")

if __name__ == "__main__":
    test_api()
