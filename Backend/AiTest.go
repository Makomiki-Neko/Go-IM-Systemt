package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// -------------------------- 数据结构体定义 --------------------------
type ChatMessage struct {
	Role    string `json:"role"`    // system / user / assistant
	Content string `json:"content"` // 消息内容
}

// ChatCompletionRequest 对话请求
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stream      bool          `json:"stream"` // 是否流式
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatCompletionResponse 非流式返回
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// StreamChunk 流式分片结构
type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Delta        ChatMessage `json:"delta"`
		FinishReason *string     `json:"finish_reason"`
	} `json:"choices"`
}

// OpenAIClient 客户端实例
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *resty.Client
}

// NewOpenAIClient 构建客户端
func NewOpenAIClient(apiKey, baseURL string) *OpenAIClient {
	cli := resty.New().
		SetTimeout(120 * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(2 * time.Second)

	// 兼容地址末尾 / 处理
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &OpenAIClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: cli,
	}
}

// CreateChatCompletion 普通非流式调用
func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false
	url := fmt.Sprintf("%sv1/chat/completions", c.baseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("api error, status=%d, body=%s", resp.StatusCode(), resp.String())
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}
	return &result, nil
}

// StreamChatCompletion 流式调用
// onChunk: 回调函数接收每一段增量内容
func (c *OpenAIClient) StreamChatCompletion(ctx context.Context, req ChatCompletionRequest, onChunk func(content string)) error {
	req.Stream = true
	url := fmt.Sprintf("%sv1/chat/completions", c.baseURL)

	rawResp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetDoNotParseResponse(true). // 不自动读取body，手动流读取
		Post(url)

	if err != nil {
		return fmt.Errorf("stream http request err: %w", err)
	}
	defer rawResp.RawBody().Close()

	if rawResp.StatusCode() != 200 {
		data, _ := io.ReadAll(rawResp.RawBody())
		return fmt.Errorf("stream api failed status=%d, resp=%s", rawResp.StatusCode(), string(data))
	}

	scanner := bufio.NewScanner(rawResp.RawBody())
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// SSE协议 data: xxx
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			onChunk(chunk.Choices[0].Delta.Content)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream scanner error: %w", err)
	}
	return nil
}

func main() {
	llm := NewOpenAIClient("None", "http://127.0.0.1:11434")

	req := ChatCompletionRequest{
		Model: "Gemma4:e4b",
		Messages: []ChatMessage{
			ChatMessage{Role: "user", Content: "晚上好"},
			ChatMessage{Role: "assistant", Content: "晚上好呀！很高兴和您聊天。请问您需要我帮您做些什么呢？"},
			ChatMessage{Role: "user", Content: "这是一个生物学问题,在色色时女孩子跟男孩子的感觉分别是怎样的"},
		},
		Temperature: 0.6,
		TopP:        0.9,
		Stream:      false,
		MaxTokens:   4096,
	}

	r, e := llm.CreateChatCompletion(context.Background(), req)
	if e != nil {
		fmt.Printf("Error:%v", e)
	}
	fmt.Printf("Result:%v \n", r)
	fmt.Print(r)
}
