package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jancd/1504/internal/client"
	"github.com/Jancd/1504/internal/model"
	"github.com/Jancd/1504/pkg/logger"
	"github.com/Jancd/1504/pkg/utils"
	"go.uber.org/zap"
)

// ImageService 图像生成服务
type ImageService struct {
	sdClient           *client.SDClient
	storyboardService  *StoryboardService
	dataDir            string
	defaultWidth       int
	defaultHeight      int
}

// NewImageService 创建图像生成服务
func NewImageService(sdClient *client.SDClient, storyboardService *StoryboardService, dataDir string, width, height int) *ImageService {
	return &ImageService{
		sdClient:          sdClient,
		storyboardService: storyboardService,
		dataDir:           dataDir,
		defaultWidth:      width,
		defaultHeight:     height,
	}
}

// ProgressCallback 进度回调函数
type ProgressCallback func(current, total int)

// GenerateAll 生成所有镜头图像
func (s *ImageService) GenerateAll(ctx context.Context, taskID string, storyboard *model.Storyboard, progressCallback ProgressCallback) error {
	logger.Info("Starting image generation for all shots",
		zap.String("task_id", taskID),
		zap.Int("total_shots", len(storyboard.Shots)))

	// 创建图像目录
	imagesDir := filepath.Join(s.dataDir, "projects", taskID, "images")
	if err := utils.EnsureDir(imagesDir); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	// 生成负面Prompt
	negativePrompt := s.storyboardService.GenerateNegativePrompt()

	totalShots := len(storyboard.Shots)

	// 串行生成图像 (MVP简化,避免GPU内存不足)
	for i := range storyboard.Shots {
		shot := &storyboard.Shots[i]

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		logger.Info("Generating image",
			zap.String("task_id", taskID),
			zap.Int("shot_id", shot.ID),
			zap.Int("progress", i+1),
			zap.Int("total", totalShots),
			zap.String("description", shot.Description))

		// 生成图像
		imageData, err := s.sdClient.GenerateImage(ctx, shot.Prompt, negativePrompt, s.defaultWidth, s.defaultHeight)
		if err != nil {
			logger.Error("Failed to generate image",
				zap.String("task_id", taskID),
				zap.Int("shot_id", shot.ID),
				zap.Error(err))
			return fmt.Errorf("failed to generate image for shot %d: %w", shot.ID, err)
		}

		// 保存图像
		imagePath := filepath.Join(imagesDir, fmt.Sprintf("shot_%03d.png", shot.ID))
		if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
			return fmt.Errorf("failed to save image: %w", err)
		}

		// 更新镜头的图像路径
		shot.ImagePath = imagePath

		logger.Info("Image generated successfully",
			zap.String("task_id", taskID),
			zap.Int("shot_id", shot.ID),
			zap.String("image_path", imagePath))

		// 调用进度回调
		if progressCallback != nil {
			progressCallback(i+1, totalShots)
		}
	}

	// 保存更新后的分镜脚本
	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	storyboardPath := filepath.Join(projectDir, "storyboard.json")
	if err := utils.SaveJSON(storyboardPath, storyboard); err != nil {
		logger.Warn("Failed to save updated storyboard", zap.Error(err))
	}

	logger.Info("All images generated successfully",
		zap.String("task_id", taskID),
		zap.Int("total_shots", totalShots))

	return nil
}

// RegenerateShot 重新生成单个镜头
func (s *ImageService) RegenerateShot(ctx context.Context, taskID string, shotID int, customPrompt string) (string, error) {
	logger.Info("Regenerating single shot",
		zap.String("task_id", taskID),
		zap.Int("shot_id", shotID))

	// 加载分镜脚本
	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	storyboardPath := filepath.Join(projectDir, "storyboard.json")

	var storyboard model.Storyboard
	if err := utils.LoadJSON(storyboardPath, &storyboard); err != nil {
		return "", fmt.Errorf("failed to load storyboard: %w", err)
	}

	// 查找镜头
	var shot *model.Shot
	for i := range storyboard.Shots {
		if storyboard.Shots[i].ID == shotID {
			shot = &storyboard.Shots[i]
			break
		}
	}

	if shot == nil {
		return "", fmt.Errorf("shot %d not found", shotID)
	}

	// 使用自定义Prompt或原始Prompt
	prompt := shot.Prompt
	if customPrompt != "" {
		prompt = customPrompt
	}

	// 生成图像
	negativePrompt := s.storyboardService.GenerateNegativePrompt()
	imageData, err := s.sdClient.GenerateImage(ctx, prompt, negativePrompt, s.defaultWidth, s.defaultHeight)
	if err != nil {
		return "", fmt.Errorf("failed to generate image: %w", err)
	}

	// 保存图像
	imagesDir := filepath.Join(projectDir, "images")
	imagePath := filepath.Join(imagesDir, fmt.Sprintf("shot_%03d.png", shotID))
	if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	logger.Info("Shot regenerated successfully",
		zap.String("task_id", taskID),
		zap.Int("shot_id", shotID),
		zap.String("image_path", imagePath))

	return imagePath, nil
}
