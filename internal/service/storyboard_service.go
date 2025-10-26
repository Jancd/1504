package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Jancd/1504/internal/client"
	"github.com/Jancd/1504/internal/model"
	"github.com/Jancd/1504/pkg/logger"
	"github.com/Jancd/1504/pkg/utils"
	"go.uber.org/zap"
)

// StoryboardService 分镜生成服务
type StoryboardService struct {
	openaiClient *client.OpenAIClient
	dataDir      string
}

// NewStoryboardService 创建分镜生成服务
func NewStoryboardService(openaiClient *client.OpenAIClient, dataDir string) *StoryboardService {
	return &StoryboardService{
		openaiClient: openaiClient,
		dataDir:      dataDir,
	}
}

// Generate 生成分镜脚本
func (s *StoryboardService) Generate(ctx context.Context, taskID string, parsed *model.ParsedScript, targetDuration int) (*model.Storyboard, error) {
	logger.Info("Starting storyboard generation",
		zap.String("task_id", taskID),
		zap.Int("scenes", len(parsed.Scenes)),
		zap.Int("target_duration", targetDuration))

	// 调用OpenAI生成分镜
	storyboard, err := s.openaiClient.GenerateStoryboard(ctx, parsed, targetDuration)
	if err != nil {
		logger.Error("Failed to generate storyboard", zap.String("task_id", taskID), zap.Error(err))
		return nil, fmt.Errorf("failed to generate storyboard: %w", err)
	}

	// 为每个镜头生成AI绘图Prompt
	for i := range storyboard.Shots {
		shot := &storyboard.Shots[i]
		if shot.Prompt == "" {
			shot.Prompt = s.generateImagePrompt(shot)
		}
	}

	// 保存分镜脚本
	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	storyboardPath := filepath.Join(projectDir, "storyboard.json")
	if err := utils.SaveJSON(storyboardPath, storyboard); err != nil {
		return nil, fmt.Errorf("failed to save storyboard: %w", err)
	}

	logger.Info("Storyboard generated successfully",
		zap.String("task_id", taskID),
		zap.Int("shots", len(storyboard.Shots)),
		zap.Float64("total_duration", storyboard.TotalDuration))

	return storyboard, nil
}

// generateImagePrompt 生成AI绘图Prompt
func (s *StoryboardService) generateImagePrompt(shot *model.Shot) string {
	// 基础风格描述
	basePrompt := "anime style, manga, japanese animation, high quality, detailed, cinematic, "

	// 添加镜头类型
	switch shot.Type {
	case model.ShotTypeCloseup:
		basePrompt += "close-up shot, facial expression, detailed face, "
	case model.ShotTypeMedium:
		basePrompt += "medium shot, half body, character interaction, "
	case model.ShotTypeLong:
		basePrompt += "wide shot, establishing shot, full body, environment, "
	default:
		basePrompt += "medium shot, "
	}

	// 添加场景描述
	basePrompt += shot.Description

	// 添加角色信息
	if len(shot.Characters) > 0 {
		basePrompt += fmt.Sprintf(", featuring %d character(s)", len(shot.Characters))
	}

	// 添加情绪氛围
	if shot.Dialogue != nil && shot.Dialogue.Emotion != "" {
		basePrompt += fmt.Sprintf(", %s atmosphere", shot.Dialogue.Emotion)
	}

	// 添加通用质量标签
	basePrompt += ", professional artwork, trending on pixiv"

	return basePrompt
}

// GenerateNegativePrompt 生成负面Prompt
func (s *StoryboardService) GenerateNegativePrompt() string {
	return "low quality, blurry, distorted, ugly, bad anatomy, bad proportions, " +
		"bad hands, text, error, missing fingers, extra digit, fewer digits, " +
		"cropped, worst quality, jpeg artifacts, signature, watermark, username"
}
