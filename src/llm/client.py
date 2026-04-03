import os
import requests
import json
from dotenv import load_dotenv

# Carrega variáveis locales caso rodando fora do container
load_dotenv()

class OpenRouterClient:
    def __init__(self):
        self.api_key = os.getenv("OPENROUTER_API_KEY")
        self.model = os.getenv("MODEL_NAME", "openai/gpt-3.5-turbo")
        self.base_url = "https://openrouter.ai/api/v1/chat/completions"

        if not self.api_key:
            raise ValueError("OPENROUTER_API_KEY não encontrada no arquivo .env!")

    def ask(self, prompt: str, history: list = None) -> str:
        """
        Envia uma pergunta para o OpenRouter mantendo o histórico se fornecido.
        """
        messages = history if history else []
        messages.append({"role": "user", "content": prompt})

        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "HTTP-Referer": "https://github.com/julianorinaldi/alexa-integrate-llm", # Opcional no OpenRouter
            "Content-Type": "application/json"
        }

        payload = {
            "model": self.model,
            "messages": messages
        }

        try:
            response = requests.post(self.base_url, headers=headers, data=json.dumps(payload), timeout=15)
            response.raise_for_status()
            
            result = response.json()
            answer = result["choices"][0]["message"]["content"]
            
            # Adicionamos a resposta da IA no histórico local para retorno opcional
            messages.append({"role": "assistant", "content": answer})
            
            return answer
        except Exception as e:
            print(f"Erro ao chamar OpenRouter: {e}")
            return "Sinto muito, tive um erro ao processar sua pergunta. Pode tentar novamente?"

# Singleton simplificado
llm_client = OpenRouterClient()
