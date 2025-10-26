package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Jancd/1504/internal/client"
	"github.com/Jancd/1504/internal/model"
	"github.com/Jancd/1504/pkg/logger"
	"github.com/Jancd/1504/pkg/utils"
	"go.uber.org/zap"
)

// QiniuVideoService 七牛云视频生成服务
type QiniuVideoService struct {
	qiniuClient *client.QiniuVideoClient
	dataDir     string
	maxWaitTime time.Duration
}

// NewQiniuVideoService 创建七牛云视频生成服务
func NewQiniuVideoService(qiniuClient *client.QiniuVideoClient, dataDir string, maxWaitTimeSec int) *QiniuVideoService {
	return &QiniuVideoService{
		qiniuClient: qiniuClient,
		dataDir:     dataDir,
		maxWaitTime: time.Duration(maxWaitTimeSec) * time.Second,
	}
}

// GenerateFromStoryboard 从分镜脚本生成视频
func (s *QiniuVideoService) GenerateFromStoryboard(ctx context.Context, taskID string, storyboard *model.Storyboard) (string, error) {
	logger.Info("Starting video generation with Qiniu",
		zap.String("task_id", taskID),
		zap.Int("shots", len(storyboard.Shots)))

	// 将分镜转换为视频生成prompt
	prompt := s.buildVideoPrompt(storyboard)

	logger.Debug("Generated prompt for video",
		zap.String("task_id", taskID),
		zap.String("prompt", prompt))

	// 调用七牛云API生成视频
	// 注意：七牛云Veo API当前只支持8秒视频
	videoDuration := 8
	result, err := s.qiniuClient.GenerateVideo(ctx, prompt, videoDuration)
	if err != nil {
		return "", fmt.Errorf("failed to start video generation: %w", err)
	}

	logger.Info("Video generation task created",
		zap.String("task_id", taskID),
		zap.String("qiniu_task_id", result.ID))

	// 等待视频生成完成
	result, err = s.qiniuClient.WaitForCompletion(ctx, result.ID, s.maxWaitTime)
	if err != nil {
		return "", fmt.Errorf("video generation failed: %w", err)
	}

	// 下载视频
	videoURL := result.GetVideoURL()
	if videoURL == "" {
		return "", fmt.Errorf("no video URL in response")
	}
	videoData, err := s.qiniuClient.DownloadVideo(ctx, videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	// 保存视频到本地
	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	if err := utils.EnsureDir(projectDir); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	videoPath := filepath.Join(projectDir, "output.mp4")
	if err := os.WriteFile(videoPath, videoData, 0644); err != nil {
		return "", fmt.Errorf("failed to save video: %w", err)
	}

	logger.Info("Video saved successfully",
		zap.String("task_id", taskID),
		zap.String("video_path", videoPath),
		zap.Int("file_size", len(videoData)))

	return videoPath, nil
}

// buildVideoPrompt 从分镜脚本构建视频生成prompt
func (s *QiniuVideoService) buildVideoPrompt(storyboard *model.Storyboard) string {
	// 构建详细的视频描述
	prompt := "Create an anime-style video with the following scenes:\n\n"

	for i, shot := range storyboard.Shots {
		prompt += fmt.Sprintf("Scene %d (%s shot, %.1f seconds):\n", i+1, shot.Type, shot.Duration)
		prompt += fmt.Sprintf("%s\n", shot.Description)

		if shot.Dialogue != nil {
			prompt += fmt.Sprintf("Dialogue: %s says \"%s\" (%s emotion)\n",
				shot.Dialogue.Character, shot.Dialogue.Text, shot.Dialogue.Emotion)
		}

		if len(shot.Characters) > 0 {
			prompt += fmt.Sprintf("Characters: %v\n", shot.Characters)
		}

		prompt += fmt.Sprintf("Transition: %s\n\n", shot.Transition)
	}

	prompt += "\nStyle: Japanese anime/manga art style, high quality, cinematic"

	return prompt
}

// GenerateSimple 简化版本：直接从文本生成视频
func (s *QiniuVideoService) GenerateSimple(ctx context.Context, taskID, text string, duration int) (string, error) {
	logger.Info("Starting simple video generation",
		zap.String("task_id", taskID),
		zap.Int("text_length", len(text)),
		zap.Int("duration", duration))

	// 构建简单prompt
	prompt := fmt.Sprintf("Create an anime-style video based on this story:\n\n%s\n\nStyle: Japanese anime/manga art style, cinematic, high quality", text)

	// 调用七牛云API
	// 注意：七牛云Veo API当前只支持8秒视频
	videoDuration := 8
	result, err := s.qiniuClient.GenerateVideo(ctx, prompt, videoDuration)
	if err != nil {
		return "", fmt.Errorf("failed to start video generation: %w", err)
	}

	logger.Info("Video generation task created",
		zap.String("task_id", taskID),
		zap.String("qiniu_task_id", result.ID))

	// 等待完成
	result, err = s.qiniuClient.WaitForCompletion(ctx, result.ID, s.maxWaitTime)
	if err != nil {
		return "", fmt.Errorf("video generation failed: %w", err)
	}

	// 下载并保存
	videoURL := result.GetVideoURL()
	if videoURL == "" {
		return "", fmt.Errorf("no video URL in response")
	}
	videoData, err := s.qiniuClient.DownloadVideo(ctx, videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	if err := utils.EnsureDir(projectDir); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	videoPath := filepath.Join(projectDir, "output.mp4")
	if err := os.WriteFile(videoPath, videoData, 0644); err != nil {
		return "", fmt.Errorf("failed to save video: %w", err)
	}

	logger.Info("Video saved successfully",
		zap.String("task_id", taskID),
		zap.String("video_path", videoPath))

	return videoPath, nil
}
