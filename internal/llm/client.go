package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	model      string
	tools      []Tool
	httpClient *http.Client
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		tools:   []Tool{},
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (c *Client) RegisterTool(tool Tool) {
	c.tools = append(c.tools, tool)
}

func (c *Client) Chat(messages []Message) (*Message, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    c.tools,
		Stream:   false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal the request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp.Message, nil
}

func (m *Message) HasToolCalls() bool {
	return len(m.ToolCalls) > 0
}

func (c *Client) Model() string {
	return c.model
}

func (c *Client) SetModel(model string) {
	c.model = model
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

// StreamCallback is called for each chunk of a streaming response
type StreamCallback func(chunk string, done bool, fullMessage *Message)

// ChatStream sends a chat request with streaming enabled
// The callback is called for each chunk received
// Returns the final complete message
func (c *Client) ChatStream(messages []Message, callback StreamCallback) (*Message, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    c.tools,
		Stream:   true,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var fullMessage Message
	fullMessage.Role = RoleAssistant

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer size for potentially large responses
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var streamResp ChatResponse
		if err := json.Unmarshal(line, &streamResp); err != nil {
			// Skip malformed lines
			continue
		}

		// Accumulate content
		if streamResp.Message.Content != "" {
			fullMessage.Content += streamResp.Message.Content
			if callback != nil {
				callback(streamResp.Message.Content, false, nil)
			}
		}

		// Check for tool calls (usually comes at the end)
		if len(streamResp.Message.ToolCalls) > 0 {
			fullMessage.ToolCalls = streamResp.Message.ToolCalls
		}

		// Final message
		if streamResp.Done {
			if callback != nil {
				callback("", true, &fullMessage)
			}
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %w", err)
	}

	return &fullMessage, nil
}
