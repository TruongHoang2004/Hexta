package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTransport struct {
	fn func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func TestGeminiClient_GenerateContent_Success(t *testing.T) {
	expectedText := "Hello, I can help you draft your order."
	transport := &mockTransport{
		fn: func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, "/v1beta/models/gemini-2.0-flash:generateContent", r.URL.Path)
			assert.Equal(t, "test-key", r.URL.Query().Get("key"))

			resp := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Role: "model",
							Parts: []Part{
								{Text: expectedText},
							},
						},
						FinishReason: "STOP",
					},
				},
			}
			b, _ := json.Marshal(resp)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(b)),
				Header:     make(http.Header),
			}, nil
		},
	}

	httpClient := &http.Client{Transport: transport}
	client := NewCustomGeminiClient("test-key", "https://mock.gemini.api", "gemini-2.0-flash", httpClient)
	req := &GenerateContentRequest{
		Contents: []Content{
			{Role: "user", Parts: []Part{{Text: "Hi"}}},
		},
	}

	res, err := client.GenerateContent(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expectedText, res.FirstText())
	assert.Nil(t, res.FirstFunctionCall())
}

func TestGeminiClient_GenerateContent_FunctionCall(t *testing.T) {
	transport := &mockTransport{
		fn: func(r *http.Request) (*http.Response, error) {
			resp := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Role: "model",
							Parts: []Part{
								{
									FunctionCall: &FunctionCall{
										Name: "search_products",
										Args: map[string]interface{}{"query": "sữa bắp"},
									},
								},
							},
						},
					},
				},
			}
			b, _ := json.Marshal(resp)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(b)),
				Header:     make(http.Header),
			}, nil
		},
	}

	httpClient := &http.Client{Transport: transport}
	client := NewCustomGeminiClient("test-key", "https://mock.gemini.api", "gemini-2.0-flash", httpClient)
	req := &GenerateContentRequest{
		Contents: []Content{
			{Role: "user", Parts: []Part{{Text: "Tìm sữa bắp"}}},
		},
	}

	res, err := client.GenerateContent(context.Background(), req)
	require.NoError(t, err)
	fc := res.FirstFunctionCall()
	require.NotNil(t, fc)
	assert.Equal(t, "search_products", fc.Name)
	assert.Equal(t, "sữa bắp", fc.Args["query"])
}

func TestGeminiClient_StreamGenerateContent_Success(t *testing.T) {
	chunk1 := GenerateContentResponse{
		Candidates: []Candidate{
			{Content: Content{Parts: []Part{{Text: "Đang "}}}},
		},
	}
	chunk2 := GenerateContentResponse{
		Candidates: []Candidate{
			{Content: Content{Parts: []Part{{Text: "xử lý"}}}},
		},
	}
	b1, _ := json.Marshal(chunk1)
	b2, _ := json.Marshal(chunk2)

	ssePayload := "data: " + string(b1) + "\n\ndata: " + string(b2) + "\n\ndata: [DONE]\n\n"

	transport := &mockTransport{
		fn: func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, "/v1beta/models/gemini-2.0-flash:streamGenerateContent", r.URL.Path)
			assert.Equal(t, "sse", r.URL.Query().Get("alt"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(ssePayload))),
				Header:     make(http.Header),
			}, nil
		},
	}

	httpClient := &http.Client{Transport: transport}
	client := NewCustomGeminiClient("test-key", "https://mock.gemini.api", "gemini-2.0-flash", httpClient)
	req := &GenerateContentRequest{
		Contents: []Content{
			{Role: "user", Parts: []Part{{Text: "Chào"}}},
		},
	}

	var collected string
	err := client.StreamGenerateContent(context.Background(), req, func(chunk *GenerateContentResponse) error {
		collected += chunk.FirstText()
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, "Đang xử lý", collected)
}
