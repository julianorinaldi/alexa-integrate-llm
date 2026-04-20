package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type OpenRouterClient struct {
	APIKey  string
	Model   string
	BaseURL string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatPayload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func NewOpenRouterClient() (*OpenRouterClient, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	model := os.Getenv("MODEL_NAME")
	if model == "" {
		model = "openai/gpt-3.5-turbo"
	}

	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY não encontrada nas variáveis de ambiente")
	}

	return &OpenRouterClient{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: "https://openrouter.ai/api/v1/chat/completions",
	}, nil
}

func (c *OpenRouterClient) Ask(prompt string, history []Message) (string, []Message) {
	start := time.Now()
	log.Printf("Iniciando pergunta ao LLM: %s", prompt)

	if history == nil || len(history) == 0 {
		history = []Message{
			{Role: "system", Content: "Você é um assistente simpático para a Alexa. Responda sempre em português do Brasil (PT-BR), de forma concisa e direta, adequada para voz."},
		}
	}
	history = append(history, Message{Role: "user", Content: prompt})

	payload := ChatPayload{
		Model:    c.Model,
		Messages: history,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Erro ao serializar payload: %v\n", err)
		return "Erro na formatação da requisição.", history
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Printf("Erro ao criar request: %v\n", err)
		return "Erro ao contatar o servidor LLM.", history
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("HTTP-Referer", "https://github.com/julianorinaldi/alexa-integrate-llm")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro na chamada OpenRouter: %v\n", err)
		return "Sinto muito, tive um erro ao processar sua pergunta.", history
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Erro OpenRouter HTTP %d\n", resp.StatusCode)
		return "A inteligência artificial retornou um erro.", history
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		fmt.Printf("Erro no decode da resposa LLM: %v\n", err)
		return "Erro ao entender a resposta da API.", history
	}

	if len(chatResp.Choices) == 0 {
		return "Resposta vazia da inteligência artificial.", history
	}

	answer := chatResp.Choices[0].Message.Content
	history = append(history, Message{Role: "assistant", Content: answer})

	log.Printf("LLM respondeu em %v", time.Since(start))
	return answer, history
}
