package alexa

import (
	"alexa-llm-go/llm"
	"embed"
	"encoding/json"
	"fmt"
	"log"
)

//go:embed templates/*.json
var templatesFS embed.FS

type AlexaHandler struct {
	LLMClient *llm.OpenRouterClient
}

func NewAlexaHandler(llmClient *llm.OpenRouterClient) *AlexaHandler {
	return &AlexaHandler{LLMClient: llmClient}
}

func (h *AlexaHandler) Handle(req RequestEnvelope) ResponseEnvelope {
	resp := ResponseEnvelope{
		Version:           "1.0",
		SessionAttributes: req.Session.Attributes,
	}
	if resp.SessionAttributes == nil {
		resp.SessionAttributes = make(map[string]interface{})
	}

	switch req.Request.Type {
	case "LaunchRequest":
		h.handleLaunchRequest(&req, &resp)
	case "IntentRequest":
		switch req.Request.Intent.Name {
		case "AskIntent":
			h.handleAskIntent(&req, &resp)
		case "AMAZON.CancelIntent", "AMAZON.StopIntent":
			h.handleCancelOrStopIntent(&req, &resp)
		default:
			h.handleFallback(&req, &resp)
		}
	case "SessionEndedRequest":
		h.handleSessionEnded(&req, &resp)
	default:
		h.handleFallback(&req, &resp)
	}

	return resp
}

func (h *AlexaHandler) setSpeak(resp *ResponseEnvelope, text string, endSession bool) {
	resp.Response.OutputSpeech = &OutputSpeech{
		Type: "PlainText",
		Text: text,
	}
	resp.Response.ShouldEndSession = endSession
}

func (h *AlexaHandler) setAsk(resp *ResponseEnvelope, txt string, repromptTxt string) {
	h.setSpeak(resp, txt, false)
	resp.Response.Reprompt = &Reprompt{
		OutputSpeech: OutputSpeech{
			Type: "PlainText",
			Text: repromptTxt,
		},
	}
}

func (h *AlexaHandler) addAPLDirective(resp *ResponseEnvelope, templateName string, datasources map[string]interface{}) {
	templateData, err := templatesFS.ReadFile("templates/" + templateName)
	if err != nil {
		log.Printf("Error reading template file %s: %v", templateName, err)
		return
	}

	var document map[string]interface{}
	if err := json.Unmarshal(templateData, &document); err != nil {
		log.Printf("Error decoding template %s: %v", templateName, err)
		return
	}

	directive := DirectiveAPL{
		Type:        "Alexa.Presentation.APL.RenderDocument",
		Document:    document,
		Datasources: datasources,
	}
	resp.Response.Directives = append(resp.Response.Directives, directive)
}

func (h *AlexaHandler) handleLaunchRequest(req *RequestEnvelope, resp *ResponseEnvelope) {
	speakOutput := "Olá! Sou seu assistente inteligente. O que deseja me perguntar hoje?"
	resp.SessionAttributes["chat_history"] = []llm.Message{}

	if req.Context.System.Device.SupportsAPL() {
		h.addAPLDirective(resp, "welcome_template.json", map[string]interface{}{
			"welcomeData": map[string]interface{}{
				"header": "Bem-vindo!",
				"text":   "Estou pronto para ouvir. Faça sua pergunta.",
			},
		})
	}

	h.setAsk(resp, speakOutput, speakOutput)
}

func (h *AlexaHandler) handleAskIntent(req *RequestEnvelope, resp *ResponseEnvelope) {
	userQuery := "O que você pode fazer?"
	if slot, ok := req.Request.Intent.Slots["query"]; ok && slot.Value != "" {
		userQuery = slot.Value
	}

	history := GetChatHistory(req.Session.Attributes)
	var answer string
	if h.LLMClient != nil {
		answer, history = h.LLMClient.Ask(userQuery, history)
	} else {
		answer = "O cliente LLM não está configurado."
		history = append(history, llm.Message{Role: "user", Content: userQuery}, llm.Message{Role: "assistant", Content: answer})
	}

	resp.SessionAttributes["chat_history"] = history

	if req.Context.System.Device.SupportsAPL() {
		h.addAPLDirective(resp, "response_template.json", map[string]interface{}{
			"chatData": map[string]interface{}{
				"question": userQuery,
				"answer":   answer,
			},
		})
	}

	h.setAsk(resp, answer, "Deseja me perguntar mais alguma coisa?")
}

func (h *AlexaHandler) handleCancelOrStopIntent(req *RequestEnvelope, resp *ResponseEnvelope) {
	h.setSpeak(resp, "Até logo! Estarei aqui se precisar.", true)
}

func (h *AlexaHandler) handleSessionEnded(req *RequestEnvelope, resp *ResponseEnvelope) {
	resp.Response.ShouldEndSession = true
}

func (h *AlexaHandler) handleFallback(req *RequestEnvelope, resp *ResponseEnvelope) {
	intentName := req.Request.Intent.Name
	var speakOutput string
	if intentName != "" {
		speakOutput = fmt.Sprintf("A intenção %s foi chamada, mas não tenho um handler para ela.", intentName)
	} else {
		speakOutput = fmt.Sprintf("Recebi um evento do tipo %s, que não sei como processar.", req.Request.Type)
	}
	h.setAsk(resp, speakOutput, speakOutput)
}
