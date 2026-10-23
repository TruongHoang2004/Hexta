package types

import "strings"

// These value types describe the dormant legacy generator protocol. They do not
// grant permission to mutate business data or send external messages.
// AIPart represents a component of a multi-modal or multi-part message.
type AIPart struct {
	Text             string              `json:"text,omitempty"`
	FunctionCall     *AIFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *AIFunctionResponse `json:"functionResponse,omitempty"`
}

// AIFunctionCall represents an invocation request initiated by the model.
type AIFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// AIFunctionResponse represents the output of a client-side function execution.
type AIFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// AIContent represents a conversational turn.
type AIContent struct {
	Role  string   `json:"role,omitempty"` // "user", "model", "function"
	Parts []AIPart `json:"parts"`
}

// AIFunctionDeclaration describes a tool available for function calling.
type AIFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// AITool wraps a set of function declarations.
type AITool struct {
	FunctionDeclarations []AIFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// AIGenerationConfig contains sampling hyperparameters.
type AIGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
}

// AIGenerationRequest represents the payload sent to the Gemini API.
type AIGenerationRequest struct {
	Contents          []AIContent         `json:"contents"`
	SystemInstruction *AIContent          `json:"systemInstruction,omitempty"`
	Tools             []AITool            `json:"tools,omitempty"`
	GenerationConfig  *AIGenerationConfig `json:"generationConfig,omitempty"`
}

// AICandidate represents a generated response branch.
type AICandidate struct {
	Content      AIContent `json:"content"`
	FinishReason string    `json:"finishReason,omitempty"`
}

// AIGenerationResponse represents the payload received from Gemini API.
type AIGenerationResponse struct {
	Candidates []AICandidate `json:"candidates,omitempty"`
}

// FirstText extracts the textual output from the first candidate part if available.
func (r *AIGenerationResponse) FirstText() string {
	if r == nil || len(r.Candidates) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, part := range r.Candidates[0].Content.Parts {
		if part.Text != "" {
			sb.WriteString(part.Text)
		}
	}
	return sb.String()
}

// FirstFunctionCall extracts the first function call from candidates if present.
func (r *AIGenerationResponse) FirstFunctionCall() *AIFunctionCall {
	if r == nil || len(r.Candidates) == 0 {
		return nil
	}
	for _, part := range r.Candidates[0].Content.Parts {
		if part.FunctionCall != nil {
			return part.FunctionCall
		}
	}
	return nil
}
