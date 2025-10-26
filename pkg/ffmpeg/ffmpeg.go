package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Jancd/1504/pkg/logger"
	"go.uber.org/zap"
)

// FFmpeg FFmpeg工具
type FFmpeg struct {
	binaryPath string
}

// New 创建FFmpeg实例
func New() *FFmpeg {
	return &FFmpeg{
		binaryPath: "ffmpeg",
	}
}

// CheckInstalled 检查FFmpeg是否已安装
func (f *FFmpeg) CheckInstalled() error {
	cmd := exec.Command(f.binaryPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}

	logger.Info("FFmpeg version", zap.String("version", strings.Split(string(output), "\n")[0]))
	return nil
}

// ConcatVideosFromImages 从图片序列创建视频
func (f *FFmpeg) ConcatVideosFromImages(ctx context.Context, concatFile, bgmPath, outputPath string, fps int) error {
	logger.Info("Creating video from images",
		zap.String("concat_file", concatFile),
		zap.String("bgm", bgmPath),
		zap.String("output", outputPath),
		zap.Int("fps", fps))

	// 构建FFmpeg命令
	// ffmpeg -f concat -safe 0 -i concat.txt -i bgm.mp3 -c:v libx264 -pix_fmt yuv420p -r 30 -c:a aac -b:a 192k -shortest -y output.mp4
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
	}

	// 如果有BGM则添加音频输入
	if bgmPath != "" {
		args = append(args, "-i", bgmPath)
	}

	args = append(args,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p", // 兼容性格式
		"-r", fmt.Sprintf("%d", fps),
		"-preset", "medium",
		"-crf", "23", // 质量控制,范围0-51,越小质量越高
	)

	// 如果有BGM则添加音频编码
	if bgmPath != "" {
		args = append(args,
			"-c:a", "aac",
			"-b:a", "192k",
			"-shortest", // 视频和音频以最短的为准
		)
	}

	args = append(args, "-y", outputPath) // -y 覆盖已存在的文件

	cmd := exec.CommandContext(ctx, f.binaryPath, args...)

	logger.Debug("Executing FFmpeg command", zap.String("command", cmd.String()))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("FFmpeg command failed",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	logger.Info("Video created successfully",
		zap.String("output", outputPath),
		zap.Int("output_length", len(output)))

	return nil
}

// AddSubtitles 添加字幕
func (f *FFmpeg) AddSubtitles(ctx context.Context, inputVideo, subtitleFile, outputVideo string) error {
	logger.Info("Adding subtitles to video",
		zap.String("input", inputVideo),
		zap.String("subtitle", subtitleFile),
		zap.String("output", outputVideo))

	// ffmpeg -i input.mp4 -vf subtitles=subtitle.srt -c:a copy -y output.mp4
	cmd := exec.CommandContext(ctx, f.binaryPath,
		"-i", inputVideo,
		"-vf", fmt.Sprintf("subtitles=%s", subtitleFile),
		"-c:a", "copy",
		"-y", outputVideo,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to add subtitles",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("failed to add subtitles: %w\nOutput: %s", err, string(output))
	}

	logger.Info("Subtitles added successfully", zap.String("output", outputVideo))
	return nil
}

// GetVideoInfo 获取视频信息
func (f *FFmpeg) GetVideoInfo(videoPath string) (map[string]interface{}, error) {
	// 使用ffprobe获取视频信息
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// 简单返回(完整版应该解析JSON)
	info := map[string]interface{}{
		"raw_output": string(output),
	}

	return info, nil
}

// CreateThumbnail 创建视频缩略图
func (f *FFmpeg) CreateThumbnail(ctx context.Context, videoPath, thumbnailPath string, timeOffset float64) error {
	logger.Info("Creating thumbnail",
		zap.String("video", videoPath),
		zap.String("thumbnail", thumbnailPath),
		zap.Float64("time", timeOffset))

	// ffmpeg -i input.mp4 -ss 00:00:01 -vframes 1 -y thumbnail.jpg
	cmd := exec.CommandContext(ctx, f.binaryPath,
		"-i", videoPath,
		"-ss", fmt.Sprintf("%.2f", timeOffset),
		"-vframes", "1",
		"-y", thumbnailPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to create thumbnail",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("failed to create thumbnail: %w", err)
	}

	logger.Info("Thumbnail created successfully", zap.String("output", thumbnailPath))
	return nil
}
