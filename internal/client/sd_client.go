package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Jancd/1504/pkg/logger"
	"go.uber.org/zap"
)

// SDClient Stable Diffusion客户端
type SDClient struct {
	apiURL  string
	client  *http.Client
	timeout time.Duration
}

// NewSDClient 创建SD客户端
func NewSDClient(apiURL string, timeout int) *SDClient {
	return &SDClient{
		apiURL: apiURL,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		timeout: time.Duration(timeout) * time.Second,
	}
}

// Txt2ImgRequest SD文生图请求
type Txt2ImgRequest struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	SamplerName    string  `json:"sampler_name"`
	Seed           int64   `json:"seed,omitempty"`
}

// Txt2ImgResponse SD文生图响应
type Txt2ImgResponse struct {
	Images     []string               `json:"images"`
	Parameters map[string]interface{} `json:"parameters"`
	Info       string                 `json:"info"`
}

// GenerateImage 生成图像
func (c *SDClient) GenerateImage(ctx context.Context, prompt, negativePrompt string, width, height int) ([]byte, error) {
	logger.Info("Generating image with Stable Diffusion",
		zap.String("prompt", prompt),
		zap.Int("width", width),
		zap.Int("height", height))

	// 构建请求
	req := Txt2ImgRequest{
		Prompt:         prompt,
		NegativePrompt: negativePrompt,
		Steps:          30,
		CFGScale:       7.5,
		Width:          width,
		Height:         height,
		SamplerName:    "DPM++ 2M Karras",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/sdapi/v1/txt2img", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	resp, err := c.client.Do(httpReq)
	if err != nil {
		logger.Error("Failed to call SD API", zap.Error(err))
		return nil, fmt.Errorf("sd api call failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("SD API returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("sd api returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result Txt2ImgResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Images) == 0 {
		return nil, fmt.Errorf("no image generated")
	}

	// 解码base64图像
	imageData, err := base64.StdEncoding.DecodeString(result.Images[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	logger.Info("Image generated successfully",
		zap.Duration("duration", duration),
		zap.Int("image_size", len(imageData)))

	return imageData, nil
}

// CheckHealth 检查SD服务健康状态
func (c *SDClient) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL+"/sdapi/v1/sd-models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SD API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SD API health check failed with status %d", resp.StatusCode)
	}

	logger.Info("SD API is healthy")
	return nil
}

// GetProgress 获取生成进度(如果SD支持)
func (c *SDClient) GetProgress(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL+"/sdapi/v1/progress", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get progress: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Progress float64 `json:"progress"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode progress: %w", err)
	}

	return result.Progress, nil
}
