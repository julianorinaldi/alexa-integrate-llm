import pytest
from unittest.mock import patch, MagicMock
from src.llm.client import OpenRouterClient

@patch("requests.post")
def test_openrouter_ask_success(mock_post):
    # Mock da resposta do OpenRouter
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.json.return_value = {
        "choices": [
            {
                "message": {
                    "content": "Olá, sou uma IA de teste!"
                }
            }
        ]
    }
    mock_post.return_value = mock_response

    client = OpenRouterClient()
    answer = client.ask("Olá!")
    
    assert answer == "Olá, sou uma IA de teste!"
    assert mock_post.called

@patch("requests.post")
def test_openrouter_ask_error(mock_post):
    # Simula erro de rede
    mock_post.side_effect = Exception("Falha de conexão")
    
    client = OpenRouterClient()
    answer = client.ask("Olá!")
    
    # Deve retornar a mensagem de erro amigável ao usuário
    assert "Sinto muito, tive um erro" in answer
