package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Jancd/1504/pkg/logger"
	"go.uber.org/zap"
)

// QiniuVideoClient 七牛云文生视频客户端
type QiniuVideoClient struct {
	apiURL  string
	apiKey  string
	model   string
	client  *http.Client
	timeout time.Duration
}

// NewQiniuVideoClient 创建七牛云视频客户端
func NewQiniuVideoClient(apiURL, apiKey, model string, timeout int) *QiniuVideoClient {
	return &QiniuVideoClient{
		apiURL: apiURL,
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		timeout: time.Duration(timeout) * time.Second,
	}
}

// VideoGenerateRequest 视频生成请求 (Veo API格式)
type VideoGenerateRequest struct {
	Instances  []VideoInstance `json:"instances"`
	Parameters VideoParameters `json:"parameters"`
	Model      string          `json:"model"`
}

// VideoInstance 视频实例
type VideoInstance struct {
	Prompt string `json:"prompt"`
}

// VideoParameters 视频参数
type VideoParameters struct {
	GenerateAudio   bool `json:"generateAudio"`
	DurationSeconds int  `json:"durationSeconds"`
	SampleCount     int  `json:"sampleCount"`
}

// VideoGenerateResponse 视频生成响应
type VideoGenerateResponse struct {
	ID        string                 `json:"id"`
	Model     string                 `json:"model,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Data      *VideoGenerationData   `json:"data,omitempty"`
	CreatedAt string                 `json:"created_at,omitempty"`
	UpdatedAt string                 `json:"updated_at,omitempty"`
}

// VideoGenerationData 视频生成数据
type VideoGenerationData struct {
	RaiMediaFilteredCount int           `json:"raiMediaFilteredCount"`
	Videos                []VideoResult `json:"videos"`
}

// VideoResult 视频结果
type VideoResult struct {
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
}

// GetTaskID 获取任务ID (兼容性方法)
func (r *VideoGenerateResponse) GetTaskID() string {
	return r.ID
}

// GetVideoURL 获取视频URL (兼容性方法)
func (r *VideoGenerateResponse) GetVideoURL() string {
	if r.Data != nil && len(r.Data.Videos) > 0 {
		return r.Data.Videos[0].URL
	}
	return ""
}

// GenerateVideo 生成视频
func (c *QiniuVideoClient) GenerateVideo(ctx context.Context, prompt string, duration int) (*VideoGenerateResponse, error) {
	logger.Info("Calling Qiniu Video Generation API",
		zap.String("prompt", prompt),
		zap.Int("duration", duration))

	// 构建请求 (Veo API格式)
	req := VideoGenerateRequest{
		Instances: []VideoInstance{
			{
				Prompt: prompt,
			},
		},
		Parameters: VideoParameters{
			GenerateAudio:   true,
			DurationSeconds: duration,
			SampleCount:     1,
		},
		Model: c.model,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Debug("Qiniu API request",
		zap.String("url", c.apiURL),
		zap.String("model", c.model),
		zap.String("request_body", string(jsonData)))

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	startTime := time.Now()
	resp, err := c.client.Do(httpReq)
	if err != nil {
		logger.Error("Failed to call Qiniu Video API", zap.Error(err))
		return nil, fmt.Errorf("qiniu video api call failed: %w", err)
	}
	defer resp.Body.Close()

	duration = int(time.Since(startTime).Seconds())

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Qiniu Video API returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("qiniu video api returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result VideoGenerateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Info("Video generation task created",
		zap.String("task_id", result.ID),
		zap.String("status", result.Status),
		zap.Int("duration", duration))

	return &result, nil
}

// QueryTaskStatus 查询任务状态
func (c *QiniuVideoClient) QueryTaskStatus(ctx context.Context, taskID string) (*VideoGenerateResponse, error) {
	logger.Debug("Querying video generation task status", zap.String("task_id", taskID))

	// 构建查询URL
	queryURL := fmt.Sprintf("%s/%s", c.apiURL, taskID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query task status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query returned status %d: %s", resp.StatusCode, string(body))
	}

	var result VideoGenerateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// WaitForCompletion 等待视频生成完成
func (c *QiniuVideoClient) WaitForCompletion(ctx context.Context, taskID string, maxWaitTime time.Duration) (*VideoGenerateResponse, error) {
	logger.Info("Waiting for video generation to complete",
		zap.String("task_id", taskID),
		zap.Duration("max_wait_time", maxWaitTime))

	deadline := time.Now().Add(maxWaitTime)
	ticker := time.NewTicker(10 * time.Second) // 增加查询间隔到10秒
	defer ticker.Stop()

	checkCount := 0
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			checkCount++
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for video generation after %d checks", checkCount)
			}

			result, err := c.QueryTaskStatus(ctx, taskID)
			if err != nil {
				logger.Warn("Failed to query task status", 
					zap.Error(err),
					zap.Int("check_count", checkCount))
				continue
			}

			logger.Info("Task status check",
				zap.String("task_id", taskID),
				zap.String("status", result.Status),
				zap.String("message", result.Message),
				zap.Int("check_count", checkCount))

			// 检查是否完成 (七牛云API状态)
			if result.Status == "Completed" || result.Status == "completed" || result.Status == "success" {
				videoURL := result.GetVideoURL()
				if videoURL == "" {
					return nil, fmt.Errorf("video generation completed but no video URL found")
				}
				logger.Info("Video generation completed",
					zap.String("task_id", taskID),
					zap.String("video_url", videoURL))
				return result, nil
			}

			// 检查是否失败
			if result.Status == "Failed" || result.Status == "failed" || result.Status == "error" {
				errorMsg := result.Message
				if errorMsg == "" {
					errorMsg = "unknown error"
				}
				return nil, fmt.Errorf("video generation failed: %s", errorMsg)
			}

			// 如果长时间处于Queued状态，给出提示
			if result.Status == "Queued" && checkCount > 6 { // 超过1分钟还在排队
				logger.Warn("Task has been queued for a long time",
					zap.String("task_id", taskID),
					zap.Int("check_count", checkCount),
					zap.String("suggestion", "Qiniu service may be busy"))
			}
		}
	}
}

// DownloadVideo 下载视频
func (c *QiniuVideoClient) DownloadVideo(ctx context.Context, videoURL string) ([]byte, error) {
	logger.Info("Downloading video", zap.String("url", videoURL))

	req, err := http.NewRequestWithContext(ctx, "GET", videoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read video data: %w", err)
	}

	logger.Info("Video downloaded successfully", zap.Int("size", len(data)))
	return data, nil
}

// CheckHealth 检查服务健康状态
func (c *QiniuVideoClient) CheckHealth(ctx context.Context) error {
	// 简单的健康检查 - 尝试调用API
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Qiniu Video API: %w", err)
	}
	defer resp.Body.Close()

	logger.Info("Qiniu Video API is reachable")
	return nil
}
