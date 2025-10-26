package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jancd/1504/internal/model"
	"github.com/Jancd/1504/pkg/ffmpeg"
	"github.com/Jancd/1504/pkg/logger"
	"github.com/Jancd/1504/pkg/utils"
	"go.uber.org/zap"
)

// RenderService 视频渲染服务
type RenderService struct {
	ffmpeg  *ffmpeg.FFmpeg
	dataDir string
	fps     int
}

// NewRenderService 创建视频渲染服务
func NewRenderService(dataDir string, fps int) *RenderService {
	return &RenderService{
		ffmpeg:  ffmpeg.New(),
		dataDir: dataDir,
		fps:     fps,
	}
}

// Render 渲染视频
func (s *RenderService) Render(ctx context.Context, taskID string, storyboard *model.Storyboard, bgmPath string) (*model.Result, error) {
	logger.Info("Starting video rendering",
		zap.String("task_id", taskID),
		zap.Int("shots", len(storyboard.Shots)),
		zap.String("bgm", bgmPath))

	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	outputPath := filepath.Join(projectDir, "output.mp4")

	// 生成FFmpeg concat文件
	concatFile := filepath.Join(projectDir, "concat.txt")
	if err := s.generateConcatFile(concatFile, storyboard); err != nil {
		return nil, fmt.Errorf("failed to generate concat file: %w", err)
	}

	// 检查BGM文件是否存在
	if bgmPath != "" {
		bgmFullPath := filepath.Join(s.dataDir, "assets", "bgm", bgmPath)
		if !utils.FileExists(bgmFullPath) {
			logger.Warn("BGM file not found, rendering without audio",
				zap.String("bgm_path", bgmFullPath))
			bgmPath = ""
		} else {
			bgmPath = bgmFullPath
		}
	}

	// 使用FFmpeg合成视频
	if err := s.ffmpeg.ConcatVideosFromImages(ctx, concatFile, bgmPath, outputPath, s.fps); err != nil {
		return nil, fmt.Errorf("failed to render video: %w", err)
	}

	// 获取文件信息
	fileSize, err := utils.GetFileSize(outputPath)
	if err != nil {
		logger.Warn("Failed to get file size", zap.Error(err))
		fileSize = 0
	}

	// 创建缩略图(可选)
	thumbnailPath := filepath.Join(projectDir, "thumbnail.jpg")
	if err := s.ffmpeg.CreateThumbnail(ctx, outputPath, thumbnailPath, 1.0); err != nil {
		logger.Warn("Failed to create thumbnail", zap.Error(err))
		thumbnailPath = ""
	}

	result := &model.Result{
		VideoPath:    outputPath,
		Duration:     storyboard.TotalDuration,
		Resolution:   "1920x1080",
		FileSize:     fileSize,
		ThumbnailURL: thumbnailPath,
		ShotCount:    len(storyboard.Shots),
	}

	logger.Info("Video rendered successfully",
		zap.String("task_id", taskID),
		zap.String("output_path", outputPath),
		zap.Int64("file_size", fileSize),
		zap.Float64("duration", storyboard.TotalDuration))

	return result, nil
}

// generateConcatFile 生成FFmpeg concat文件
func (s *RenderService) generateConcatFile(path string, storyboard *model.Storyboard) error {
	logger.Debug("Generating concat file", zap.String("path", path))

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer f.Close()

	// FFmpeg concat文件格式:
	// file 'path/to/image.png'
	// duration 3.0
	for i, shot := range storyboard.Shots {
		if shot.ImagePath == "" {
			return fmt.Errorf("shot %d has no image path", shot.ID)
		}

		// 写入文件路径
		fmt.Fprintf(f, "file '%s'\n", shot.ImagePath)

		// 写入持续时间(最后一帧除外)
		if i < len(storyboard.Shots)-1 {
			fmt.Fprintf(f, "duration %.2f\n", shot.Duration)
		}
	}

	// 最后一帧需要重复一次(FFmpeg concat协议要求)
	lastShot := storyboard.Shots[len(storyboard.Shots)-1]
	fmt.Fprintf(f, "file '%s'\n", lastShot.ImagePath)

	logger.Debug("Concat file generated successfully",
		zap.String("path", path),
		zap.Int("shots", len(storyboard.Shots)))

	return nil
}

// GenerateSubtitles 生成字幕文件(SRT格式)
func (s *RenderService) GenerateSubtitles(taskID string, storyboard *model.Storyboard) (string, error) {
	logger.Info("Generating subtitles", zap.String("task_id", taskID))

	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	subtitlePath := filepath.Join(projectDir, "subtitles.srt")

	f, err := os.Create(subtitlePath)
	if err != nil {
		return "", fmt.Errorf("failed to create subtitle file: %w", err)
	}
	defer f.Close()

	// SRT格式:
	// 1
	// 00:00:01,000 --> 00:00:04,000
	// 字幕文本

	currentTime := 0.0
	subtitleIndex := 1

	for _, shot := range storyboard.Shots {
		if shot.Dialogue != nil && shot.Dialogue.Text != "" {
			startTime := currentTime
			endTime := currentTime + shot.Duration

			// 写入字幕索引
			fmt.Fprintf(f, "%d\n", subtitleIndex)

			// 写入时间范围
			fmt.Fprintf(f, "%s --> %s\n",
				formatSRTTime(startTime),
				formatSRTTime(endTime))

			// 写入字幕文本
			fmt.Fprintf(f, "%s: %s\n\n", shot.Dialogue.Character, shot.Dialogue.Text)

			subtitleIndex++
		}

		currentTime += shot.Duration
	}

	logger.Info("Subtitles generated successfully",
		zap.String("task_id", taskID),
		zap.String("subtitle_path", subtitlePath),
		zap.Int("subtitle_count", subtitleIndex-1))

	return subtitlePath, nil
}

// formatSRTTime 格式化SRT时间格式 (HH:MM:SS,mmm)
func formatSRTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}

// RenderWithSubtitles 渲染带字幕的视频
func (s *RenderService) RenderWithSubtitles(ctx context.Context, taskID string, storyboard *model.Storyboard, bgmPath string) (*model.Result, error) {
	logger.Info("Rendering video with subtitles", zap.String("task_id", taskID))

	// 先渲染基础视频
	result, err := s.Render(ctx, taskID, storyboard, bgmPath)
	if err != nil {
		return nil, err
	}

	// 生成字幕文件
	subtitlePath, err := s.GenerateSubtitles(taskID, storyboard)
	if err != nil {
		logger.Warn("Failed to generate subtitles, returning video without subtitles", zap.Error(err))
		return result, nil
	}

	// 添加字幕到视频
	projectDir := filepath.Join(s.dataDir, "projects", taskID)
	outputWithSubtitles := filepath.Join(projectDir, "output_with_subtitles.mp4")

	if err := s.ffmpeg.AddSubtitles(ctx, result.VideoPath, subtitlePath, outputWithSubtitles); err != nil {
		logger.Warn("Failed to add subtitles to video, returning video without subtitles", zap.Error(err))
		return result, nil
	}

	// 更新结果
	result.VideoPath = outputWithSubtitles

	// 更新文件大小
	fileSize, err := utils.GetFileSize(outputWithSubtitles)
	if err == nil {
		result.FileSize = fileSize
	}

	logger.Info("Video with subtitles rendered successfully",
		zap.String("task_id", taskID),
		zap.String("output_path", outputWithSubtitles))

	return result, nil
}
