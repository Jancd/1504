package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Jancd/1504/internal/model"
	"github.com/Jancd/1504/pkg/logger"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// OpenAIClient OpenAI客户端
type OpenAIClient struct {
	client  *openai.Client
	model   string
	timeout time.Duration
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(apiKey, modelName, baseURL string, timeout int) *OpenAIClient {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	return &OpenAIClient{
		client:  openai.NewClientWithConfig(config),
		model:   modelName,
		timeout: time.Duration(timeout) * time.Second,
	}
}

// ParseScript 解析剧本
func (c *OpenAIClient) ParseScript(ctx context.Context, text string) (*model.ParsedScript, error) {
	logger.Info("Calling OpenAI to parse script", zap.Int("text_length", len(text)))

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	prompt := fmt.Sprintf(`你是一个专业的剧本分析师。请分析以下小说文本,提取关键信息并生成结构化数据。

文本:
%s

要求:
1. 识别所有场景,包括地点和时间
2. 提取所有出现的角色
3. 识别每个场景中的对话和动作
4. 分析角色的情绪状态
5. 统计元数据信息

请严格按照以下JSON格式返回(不要添加任何markdown标记):
{
    "scenes": [
        {
            "id": 1,
            "location": "场景地点描述",
            "time": "时间描述(如:傍晚、清晨等)",
            "characters": ["角色1", "角色2"],
            "dialogues": [
                {
                    "character": "角色名",
                    "text": "对话内容",
                    "emotion": "情绪(如:高兴、悲伤、紧张等)"
                }
            ],
            "actions": [
                {
                    "character": "角色名",
                    "description": "动作描述"
                }
            ]
        }
    ],
    "characters": ["角色1", "角色2", "角色3"],
    "metadata": {
        "total_scenes": 1,
        "word_count": 100
    }
}`, text)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一个专业的剧本分析师,擅长从文本中提取结构化信息。请始终以JSON格式返回结果,不要添加任何markdown代码块标记。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: 0.3, // 降低温度以获得更稳定的输出
	})

	if err != nil {
		logger.Error("Failed to call OpenAI API", zap.Error(err))
		return nil, fmt.Errorf("openai api call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	content := resp.Choices[0].Message.Content
	logger.Debug("OpenAI response received", zap.String("content", content))

	// 解析JSON响应
	var parsed model.ParsedScript
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		logger.Error("Failed to parse OpenAI response", zap.Error(err), zap.String("content", content))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.Info("Script parsed successfully",
		zap.Int("scenes", len(parsed.Scenes)),
		zap.Int("characters", len(parsed.Characters)))

	return &parsed, nil
}

// GenerateStoryboard 生成分镜脚本
func (c *OpenAIClient) GenerateStoryboard(ctx context.Context, parsed *model.ParsedScript, targetDuration int) (*model.Storyboard, error) {
	logger.Info("Calling OpenAI to generate storyboard",
		zap.Int("scenes", len(parsed.Scenes)),
		zap.Int("target_duration", targetDuration))

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 将解析结果转为JSON
	parsedJSON, err := json.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parsed script: %w", err)
	}

	prompt := fmt.Sprintf(`你是一个专业的分镜师。基于以下剧本解析结果,设计详细的分镜脚本。

剧本解析结果:
%s

目标视频时长: %d秒

要求:
1. 为每个重要对话和动作创建独立镜头
2. 合理分配镜头类型(特写closeup、中景medium、远景long)
3. 为每个镜头分配合适的时长(2-8秒)
4. 设计转场效果(cut直切、fade淡入淡出、dissolve溶解)
5. 为每个镜头生成详细的画面描述,用于AI绘图
6. 总时长应接近目标时长

镜头类型选择原则:
- 对话场景: 多用特写(closeup)展现表情
- 动作场景: 使用中景(medium)
- 场景介绍: 使用远景(long)

转场选择原则:
- 同一场景内对话: cut直切
- 场景转换: fade或dissolve
- 时间跳跃: fade

请严格按照以下JSON格式返回(不要添加任何markdown标记):
{
    "shots": [
        {
            "id": 1,
            "type": "closeup",
            "description": "详细的画面描述,包含环境、角色外貌、动作、表情等",
            "characters": ["角色1"],
            "duration": 3.0,
            "transition": "cut",
            "dialogue": {
                "character": "角色1",
                "text": "对话内容",
                "emotion": "情绪"
            }
        }
    ],
    "total_duration": 60.0
}`, string(parsedJSON), targetDuration)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一个专业的分镜师,擅长将剧本转换为视觉化的分镜脚本。请始终以JSON格式返回结果,不要添加任何markdown代码块标记。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: 0.5,
	})

	if err != nil {
		logger.Error("Failed to call OpenAI API for storyboard", zap.Error(err))
		return nil, fmt.Errorf("openai api call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	content := resp.Choices[0].Message.Content
	logger.Debug("OpenAI storyboard response received", zap.String("content", content))

	// 解析JSON响应
	var storyboard model.Storyboard
	if err := json.Unmarshal([]byte(content), &storyboard); err != nil {
		logger.Error("Failed to parse storyboard response", zap.Error(err), zap.String("content", content))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.Info("Storyboard generated successfully",
		zap.Int("shots", len(storyboard.Shots)),
		zap.Float64("total_duration", storyboard.TotalDuration))

	return &storyboard, nil
}
