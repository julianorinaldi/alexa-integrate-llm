package alexa

import "alexa-llm-go/llm"

type RequestEnvelope struct {
	Session Session `json:"session"`
	Request Request `json:"request"`
	Context Context `json:"context,omitempty"`
}

type Session struct {
	New         bool                   `json:"new"`
	SessionID   string                 `json:"sessionId"`
	Application Application            `json:"application"`
	Attributes  map[string]interface{} `json:"attributes"`
}

type Context struct {
	System System `json:"System"`
}

type System struct {
	Application Application `json:"application"`
	Device      Device      `json:"device"`
}

type Application struct {
	ApplicationID string `json:"applicationId"`
}

type Device struct {
	DeviceID            string                 `json:"deviceId"`
	SupportedInterfaces map[string]interface{} `json:"supportedInterfaces"`
}

func (d Device) SupportsAPL() bool {
	_, ok := d.SupportedInterfaces["Alexa.Presentation.APL"]
	return ok
}

type Request struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
	Intent    Intent `json:"intent,omitempty"`
}

type Intent struct {
	Name               string          `json:"name"`
	ConfirmationStatus string          `json:"confirmationStatus,omitempty"`
	Slots              map[string]Slot `json:"slots,omitempty"`
}

type Slot struct {
	Name               string `json:"name"`
	Value              string `json:"value"`
	ConfirmationStatus string `json:"confirmationStatus,omitempty"`
}

type ResponseEnvelope struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          Response               `json:"response"`
}

type Response struct {
	OutputSpeech     *OutputSpeech `json:"outputSpeech,omitempty"`
	Card             *Card         `json:"card,omitempty"`
	Reprompt         *Reprompt     `json:"reprompt,omitempty"`
	Directives       []interface{} `json:"directives,omitempty"`
	ShouldEndSession bool          `json:"shouldEndSession"`
}

type OutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

type Reprompt struct {
	OutputSpeech OutputSpeech `json:"outputSpeech"`
}

type Card struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
}

type DirectiveAPL struct {
	Type        string                 `json:"type"`
	Document    map[string]interface{} `json:"document"`
	Datasources map[string]interface{} `json:"datasources,omitempty"`
}

// Helper methods to get chat history safely
func GetChatHistory(attrs map[string]interface{}) []llm.Message {
	if attrs == nil {
		return []llm.Message{}
	}
	historyRaw, ok := attrs["chat_history"].([]interface{})
	if !ok {
		return []llm.Message{}
	}

	var history []llm.Message
	for _, item := range historyRaw {
		m, ok2 := item.(map[string]interface{})
		if ok2 {
			r, _ := m["role"].(string)
			c, _ := m["content"].(string)
			history = append(history, llm.Message{Role: r, Content: c})
		}
	}
	return history
}
