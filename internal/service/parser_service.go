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

// ParserService 剧本解析服务
type ParserService struct {
	openaiClient *client.OpenAIClient
	dataDir      string
}

// NewParserService 创建剧本解析服务
func NewParserService(openaiClient *client.OpenAIClient, dataDir string) *ParserService {
	return &ParserService{
		openaiClient: openaiClient,
		dataDir:      dataDir,
	}
}

// Parse 解析剧本
func (s *ParserService) Parse(ctx context.Context, taskID, text string) (*model.ParsedScript, error) {
	logger.Info("Starting script parsing", zap.String("task_id", taskID), zap.Int("text_length", len(text)))

	// 调用OpenAI解析
	parsed, err := s.openaiClient.ParseScript(ctx, text)
	if err != nil {
		logger.Error("Failed to parse script", zap.String("task_id", taskID), zap.Error(err))
		return nil, fmt.Errorf("failed to parse script: %w", err)
	}

	// 补充元数据
	if parsed.Metadata.WordCount == 0 {
		parsed.Metadata.WordCount = len([]rune(text))
	}
	if parsed.Metadata.TotalScenes == 0 {
		parsed.Metadata.TotalScenes = len(parsed.Scenes)
	}

	// 保存解析结果
	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	if err := utils.EnsureDir(projectDir); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// 保存原始文本
	scriptPath := filepath.Join(projectDir, "script.txt")
	if err := utils.SaveJSON(scriptPath, map[string]string{"text": text}); err != nil {
		logger.Warn("Failed to save original script", zap.Error(err))
	}

	// 保存解析结果
	parsedPath := filepath.Join(projectDir, "parsed.json")
	if err := utils.SaveJSON(parsedPath, parsed); err != nil {
		return nil, fmt.Errorf("failed to save parsed script: %w", err)
	}

	logger.Info("Script parsed successfully",
		zap.String("task_id", taskID),
		zap.Int("scenes", len(parsed.Scenes)),
		zap.Int("characters", len(parsed.Characters)),
		zap.Int("word_count", parsed.Metadata.WordCount))

	return parsed, nil
}
