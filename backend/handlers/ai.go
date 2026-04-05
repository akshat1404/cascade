package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AIHandler struct{}

func NewAIHandler() *AIHandler {
	return &AIHandler{}
}

type aiProcessRequest struct {
	Text   string `json:"text"`
	Action string `json:"action"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqChoice struct {
	Message groqMessage `json:"message"`
}

type groqResponse struct {
	Choices []groqChoice `json:"choices"`
}

func buildPrompt(action, text string) (string, error) {
	switch action {
	case "fix grammar":
		return fmt.Sprintf("Fix grammatical mistakes. Return only corrected text:\n\n%s", text), nil
	case "translate to French":
		return fmt.Sprintf("Translate to French. Return only the translation:\n\n%s", text), nil
	case "make a table":
		return fmt.Sprintf("Convert to a markdown table. Return only the table:\n\n%s", text), nil
	case "summarise":
		return fmt.Sprintf("Summarise in 2-3 sentences. Return only the summary:\n\n%s", text), nil
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (h *AIHandler) Process(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req aiProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Text == "" || req.Action == "" {
		http.Error(w, "text and action are required", http.StatusBadRequest)
		return
	}

	prompt, err := buildPrompt(req.Action, req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	groqPayload := groqRequest{
		Model: "llama-3.1-8b-instant",
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	payloadBytes, err := json.Marshal(groqPayload)
	if err != nil {
		http.Error(w, "Failed to build request", http.StatusInternalServerError)
		return
	}

	groqReq, err := http.NewRequestWithContext(
		r.Context(),
		http.MethodPost,
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		http.Error(w, "Failed to create Groq request", http.StatusInternalServerError)
		return
	}
	groqReq.Header.Set("Content-Type", "application/json")
	groqReq.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(groqReq)
	if err != nil {
		http.Error(w, "Failed to call Groq API", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read Groq response", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Groq API error: %s", string(body)), http.StatusBadGateway)
		return
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil || len(groqResp.Choices) == 0 {
		http.Error(w, "Invalid Groq response", http.StatusInternalServerError)
		return
	}

	result := groqResp.Choices[0].Message.Content

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": result})
}
