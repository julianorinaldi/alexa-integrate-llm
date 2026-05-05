package alexallm

import (
	"github.com/julianorinaldi/alexa-llm-go/alexa"
	"github.com/julianorinaldi/alexa-llm-go/llm"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func HandleAlexaRequest(w http.ResponseWriter, r *http.Request) {
	// Roteamento básico
	path := r.URL.Path
	if strings.Contains(path, "/admin") || strings.Contains(path, "/login") || strings.Contains(path, "/change-password") {
		handleAdminRouting(w, r)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "✅ O Servidor Alexa-LLM está rodando! Acesse /admin para gerenciar.")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqEnvelope alexa.RequestEnvelope
	if err := json.NewDecoder(r.Body).Decode(&reqEnvelope); err != nil {
		log.Printf("Erro ao decodificar JSON da Alexa: %v\n", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	log.Printf("Recebendo Request: Type=%s, Intent=%s", reqEnvelope.Request.Type, reqEnvelope.Request.Intent.Name)

	// ---- INÍCIO DA VERIFICAÇÃO DE SEGURANÇA ----
	
	// 1. Acesso Opcional de Token Global foi descontinuado em favor da tabela autorizada.

	// 2. Verificação oficial pelo ALEXA_SKILL_ID
	// Agora verificamos tanto na lista estática do .env quanto no banco de dados Supabase
	appID := reqEnvelope.Session.Application.ApplicationID
	if appID == "" {
		appID = reqEnvelope.Context.System.Application.ApplicationID
	}

	if !isAuthorized(appID, r.URL.Query().Get("token")) {
		log.Printf("Acesso negado para Skill ID: %s", appID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// ---- FIM DA VERIFICAÇÃO DE SEGURANÇA ----

	apiKey, modelName := GetLLMConfig()
	llmClient, err := llm.NewOpenRouterClient(apiKey, modelName)
	if err != nil {
		log.Printf("Aviso LLM INIT: %v", err)
		// We still process so that Alexa can reply there's an error rather than just 500 error
	}

	handler := alexa.NewAlexaHandler(llmClient)
	respEnvelope := handler.Handle(reqEnvelope)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respEnvelope); err != nil {
		log.Printf("Erro ao enviar resposta JSON: %v\n", err)
	}
}
