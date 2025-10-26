package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Jancd/1504/internal/model"
	"github.com/Jancd/1504/internal/service"
	"github.com/Jancd/1504/internal/task"
	"github.com/Jancd/1504/pkg/config"
	"github.com/Jancd/1504/pkg/logger"
	"github.com/Jancd/1504/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VideoHandler 视频处理器
type VideoHandler struct {
	taskManager       *task.Manager
	parserService     *service.ParserService
	storyboardService *service.StoryboardService
	imageService      *service.ImageService
	renderService     *service.RenderService
	qiniuVideoService *service.QiniuVideoService
	config            *config.Config
	useQiniuMode      bool // 是否使用七牛云直接生成视频模式
}

// NewVideoHandler 创建视频处理器
func NewVideoHandler(
	taskManager *task.Manager,
	parserService *service.ParserService,
	storyboardService *service.StoryboardService,
	imageService *service.ImageService,
	renderService *service.RenderService,
	qiniuVideoService *service.QiniuVideoService,
	cfg *config.Config,
) *VideoHandler {
	// 判断使用哪种模式
	useQiniu := cfg.VideoGeneration.Type == "qiniu"

	return &VideoHandler{
		taskManager:       taskManager,
		parserService:     parserService,
		storyboardService: storyboardService,
		imageService:      imageService,
		renderService:     renderService,
		qiniuVideoService: qiniuVideoService,
		config:            cfg,
		useQiniuMode:      useQiniu,
	}
}

// Generate 创建生成任务
func (h *VideoHandler) Generate(c *gin.Context) {
	var req model.Input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:      400,
			Message:   "Invalid request",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 验证输入
	if len(req.Text) == 0 {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:      400,
			Message:   "Text is required",
			Error:     "text field cannot be empty",
			Timestamp: time.Now(),
		})
		return
	}

	if len(req.Text) > h.config.Limits.MaxTextLength {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:      400,
			Message:   "Text too long",
			Error:     fmt.Sprintf("text length exceeds maximum of %d characters", h.config.Limits.MaxTextLength),
			Timestamp: time.Now(),
		})
		return
	}

	// 设置默认选项
	if req.Options.Style == "" {
		req.Options.Style = "anime"
	}
	if req.Options.DurationTarget == 0 {
		req.Options.DurationTarget = 60
	}
	if req.Options.AspectRatio == "" {
		req.Options.AspectRatio = "16:9"
	}
	if req.Options.BGM == "" {
		req.Options.BGM = h.config.Video.DefaultBGM
	}

	// 创建任务
	taskID := uuid.New().String()
	t := model.NewTask(taskID, req)

	h.taskManager.Create(t)

	logger.Info("Task created",
		zap.String("task_id", taskID),
		zap.Int("text_length", len(req.Text)))

	// 异步处理任务
	go h.processTask(taskID)

	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"task_id":        taskID,
			"status":         "processing",
			"estimated_time": 300,
		},
		Timestamp: time.Now(),
	})
}

// processTask 处理任务
func (h *VideoHandler) processTask(taskID string) {
	ctx := context.Background()

	t, ok := h.taskManager.Get(taskID)
	if !ok {
		logger.Error("Task not found in processTask", zap.String("task_id", taskID))
		return
	}

	// 更新任务状态为处理中
	t.Status = model.TaskStatusProcessing
	h.taskManager.Update(t)

	logger.Info("Starting task processing", zap.String("task_id", taskID))

	// 步骤1: 解析剧本
	t.UpdateStep(model.StepParseScript, model.StepStatusProcessing)
	h.taskManager.Update(t)

	parsed, err := h.parserService.Parse(ctx, taskID, t.Input.Text)
	if err != nil {
		h.failTask(taskID, model.StepParseScript, fmt.Sprintf("Failed to parse script: %v", err))
		return
	}

	t.UpdateStep(model.StepParseScript, model.StepStatusCompleted)
	h.taskManager.Update(t)

	// 步骤2: 生成分镜
	t.UpdateStep(model.StepGenerateStoryboard, model.StepStatusProcessing)
	h.taskManager.Update(t)

	storyboard, err := h.storyboardService.Generate(ctx, taskID, parsed, t.Input.Options.DurationTarget)
	if err != nil {
		h.failTask(taskID, model.StepGenerateStoryboard, fmt.Sprintf("Failed to generate storyboard: %v", err))
		return
	}

	// 检查镜头数量限制
	if len(storyboard.Shots) > h.config.Limits.MaxShotsPerVideo {
		h.failTask(taskID, model.StepGenerateStoryboard,
			fmt.Sprintf("Too many shots generated (%d), maximum is %d",
				len(storyboard.Shots), h.config.Limits.MaxShotsPerVideo))
		return
	}

	t.UpdateStep(model.StepGenerateStoryboard, model.StepStatusCompleted)
	h.taskManager.Update(t)

	var result *model.Result

	// 根据模式选择不同的处理流程
	if h.useQiniuMode {
		// 七牛云模式：直接生成视频
		logger.Info("Using Qiniu Video Generation mode", zap.String("task_id", taskID))

		// 步骤3: 生成视频(跳过图像生成步骤)
		t.UpdateStep(model.StepGenerateImages, model.StepStatusProcessing)
		h.taskManager.Update(t)

		videoPath, err := h.qiniuVideoService.GenerateFromStoryboard(ctx, taskID, storyboard)
		if err != nil {
			h.failTask(taskID, model.StepGenerateImages, fmt.Sprintf("Failed to generate video with Qiniu: %v", err))
			return
		}

		t.UpdateStep(model.StepGenerateImages, model.StepStatusCompleted)
		t.UpdateStep(model.StepRenderVideo, model.StepStatusCompleted) // 视频已经生成,跳过渲染步骤
		h.taskManager.Update(t)

		// 获取文件大小
		fileSize, _ := utils.GetFileSize(videoPath)

		result = &model.Result{
			VideoPath:  videoPath,
			Duration:   storyboard.TotalDuration,
			Resolution: "1920x1080",
			FileSize:   fileSize,
			ShotCount:  len(storyboard.Shots),
		}
	} else {
		// SD模式：图像生成 + 渲染
		logger.Info("Using SD + Render mode", zap.String("task_id", taskID))

		// 步骤3: 生成图像
		t.UpdateStep(model.StepGenerateImages, model.StepStatusProcessing)
		h.taskManager.Update(t)

		err = h.imageService.GenerateAll(ctx, taskID, storyboard, func(current, total int) {
			// 更新进度
			progress := (current * 100) / total
			t.Progress = progress
			t.SetStepProgress(model.StepGenerateImages, progress, fmt.Sprintf("%d/%d shots", current, total))
			h.taskManager.Update(t)

			logger.Info("Image generation progress",
				zap.String("task_id", taskID),
				zap.Int("current", current),
				zap.Int("total", total),
				zap.Int("progress", progress))
		})

		if err != nil {
			h.failTask(taskID, model.StepGenerateImages, fmt.Sprintf("Failed to generate images: %v", err))
			return
		}

		t.UpdateStep(model.StepGenerateImages, model.StepStatusCompleted)
		h.taskManager.Update(t)

		// 步骤4: 渲染视频
		t.UpdateStep(model.StepRenderVideo, model.StepStatusProcessing)
		h.taskManager.Update(t)

		var renderErr error
		result, renderErr = h.renderService.RenderWithSubtitles(ctx, taskID, storyboard, t.Input.Options.BGM)
		if renderErr != nil {
			h.failTask(taskID, model.StepRenderVideo, fmt.Sprintf("Failed to render video: %v", renderErr))
			return
		}

		t.UpdateStep(model.StepRenderVideo, model.StepStatusCompleted)
		h.taskManager.Update(t)
	}

	// 任务完成
	t.Status = model.TaskStatusCompleted
	t.Progress = 100
	t.Result = result
	h.taskManager.Update(t)

	logger.Info("Task completed successfully",
		zap.String("task_id", taskID),
		zap.String("video_path", result.VideoPath),
		zap.Int64("file_size", result.FileSize))
}

// failTask 标记任务失败
func (h *VideoHandler) failTask(taskID, step, errMsg string) {
	logger.Error("Task failed",
		zap.String("task_id", taskID),
		zap.String("step", step),
		zap.String("error", errMsg))

	t, ok := h.taskManager.Get(taskID)
	if !ok {
		return
	}

	t.UpdateStep(step, model.StepStatusFailed)
	t.Status = model.TaskStatusFailed
	t.Error = errMsg
	h.taskManager.Update(t)
}

// GetTask 获取任务状态
func (h *VideoHandler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")

	t, ok := h.taskManager.Get(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Code:      404,
			Message:   "Task not found",
			Error:     fmt.Sprintf("task %s does not exist", taskID),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Code:      0,
		Message:   "success",
		Data:      t,
		Timestamp: time.Now(),
	})
}

// Download 下载视频
func (h *VideoHandler) Download(c *gin.Context) {
	taskID := c.Param("task_id")

	t, ok := h.taskManager.Get(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Code:      404,
			Message:   "Task not found",
			Timestamp: time.Now(),
		})
		return
	}

	if t.Status != model.TaskStatusCompleted || t.Result == nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:      400,
			Message:   "Video not ready",
			Error:     fmt.Sprintf("task status is %s", t.Status),
			Timestamp: time.Now(),
		})
		return
	}

	// 检查文件是否存在
	if !utils.FileExists(t.Result.VideoPath) {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Code:      404,
			Message:   "Video file not found",
			Error:     "video file has been deleted or moved",
			Timestamp: time.Now(),
		})
		return
	}

	// 设置下载响应头
	filename := filepath.Base(t.Result.VideoPath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "video/mp4")

	logger.Info("Video downloaded",
		zap.String("task_id", taskID),
		zap.String("file_path", t.Result.VideoPath))

	c.File(t.Result.VideoPath)
}

// DeleteTask 删除任务
func (h *VideoHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("task_id")

	_, ok := h.taskManager.Get(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Code:      404,
			Message:   "Task not found",
			Timestamp: time.Now(),
		})
		return
	}

	// 删除项目文件
	projectDir := filepath.Join(h.config.Storage.DataDir, "projects", taskID)
	if utils.FileExists(projectDir) {
		if err := utils.RemoveDir(projectDir); err != nil {
			logger.Warn("Failed to remove project directory",
				zap.String("task_id", taskID),
				zap.Error(err))
		}
	}

	// 从任务管理器中删除
	if err := h.taskManager.Delete(taskID); err != nil {
		logger.Error("Failed to delete task from manager",
			zap.String("task_id", taskID),
			zap.Error(err))
	}

	logger.Info("Task deleted", zap.String("task_id", taskID))

	c.JSON(http.StatusOK, model.APIResponse{
		Code:      0,
		Message:   "Task deleted successfully",
		Timestamp: time.Now(),
	})
}

// ListTasks 列出所有任务
func (h *VideoHandler) ListTasks(c *gin.Context) {
	tasks := h.taskManager.List()

	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"tasks": tasks,
			"total": len(tasks),
		},
		Timestamp: time.Now(),
	})
}
