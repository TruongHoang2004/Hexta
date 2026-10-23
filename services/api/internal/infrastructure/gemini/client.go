package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

// Provider wire values use the core generation protocol.
type Part = types.AIPart
type FunctionCall = types.AIFunctionCall
type FunctionResponse = types.AIFunctionResponse
type Content = types.AIContent
type FunctionDeclaration = types.AIFunctionDeclaration
type Tool = types.AITool
type GenerationConfig = types.AIGenerationConfig
type GenerateContentRequest = types.AIGenerationRequest
type Candidate = types.AICandidate
type GenerateContentResponse = types.AIGenerationResponse

// IGeminiClient defines the interface for interacting with Google Gemini API.
type IGeminiClient interface {
	GenerateContent(ctx context.Context, req *GenerateContentRequest) (*GenerateContentResponse, error)
	StreamGenerateContent(ctx context.Context, req *GenerateContentRequest, callback func(*GenerateContentResponse) error) error
}

// GeminiClient implements IGeminiClient over standard HTTP.
type GeminiClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewGeminiClient instantiates a new GeminiClient from application configuration.
func NewGeminiClient() (IGeminiClient, error) {
	apiKey := config.AppConfig.AI.ApiKey
	baseURL := strings.TrimRight(config.AppConfig.AI.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	model := config.AppConfig.AI.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	timeoutSec := config.AppConfig.AI.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 30
	}

	return &GeminiClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
	}, nil
}

// NewCustomGeminiClient creates a GeminiClient with an explicit baseURL and httpClient (useful for testing).
func NewCustomGeminiClient(apiKey, baseURL, model string, client *http.Client) *GeminiClient {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &GeminiClient{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		httpClient: client,
	}
}

// GenerateContent sends a synchronous generation request to Gemini.
func (c *GeminiClient) GenerateContent(ctx context.Context, req *GenerateContentRequest) (*GenerateContentResponse, error) {
	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.baseURL, c.model)
	if c.apiKey != "" {
		endpoint = fmt.Sprintf("%s?key=%s", endpoint, c.apiKey)
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Gemini request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var genResp GenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	return &genResp, nil
}

// StreamGenerateContent streams generation response chunks using Server-Sent Events (SSE).
func (c *GeminiClient) StreamGenerateContent(ctx context.Context, req *GenerateContentRequest, callback func(*GenerateContentResponse) error) error {
	endpoint := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse", c.baseURL, c.model)
	if c.apiKey != "" {
		endpoint = fmt.Sprintf("%s&key=%s", endpoint, c.apiKey)
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal streaming request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create streaming request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("streaming api call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini streaming returned status %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		jsonData := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if jsonData == "" || jsonData == "[DONE]" {
			continue
		}

		var chunk GenerateContentResponse
		if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
			log.Warn(ctx, "failed to unmarshal streaming chunk: %v", err)
			continue
		}

		if err := callback(&chunk); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error during gemini streaming: %w", err)
	}

	return nil
}
