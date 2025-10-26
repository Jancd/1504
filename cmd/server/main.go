package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jancd/1504/internal/client"
	"github.com/Jancd/1504/internal/handler"
	"github.com/Jancd/1504/internal/service"
	"github.com/Jancd/1504/internal/task"
	"github.com/Jancd/1504/pkg/config"
	"github.com/Jancd/1504/pkg/ffmpeg"
	"github.com/Jancd/1504/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Output, cfg.Log.FilePath); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting MVP Video Generator Server")

	// 检查FFmpeg是否已安装（仅在local_sd模式下必需）
	ff := ffmpeg.New()
	if err := ff.CheckInstalled(); err != nil {
		if cfg.VideoGeneration.Type == "local_sd" {
			logger.Fatal("FFmpeg check failed (required for local_sd mode)", zap.Error(err))
		} else {
			logger.Warn("FFmpeg not found (not required for qiniu mode)", zap.Error(err))
		}
	}

	// 创建OpenAI客户端
	openaiClient := client.NewOpenAIClient(
		cfg.OpenAI.APIKey,
		cfg.OpenAI.Model,
		cfg.OpenAI.BaseURL,
		cfg.OpenAI.Timeout,
	)
	logger.Info("OpenAI client initialized", zap.String("model", cfg.OpenAI.Model))

	// 创建视频生成客户端
	var sdClient *client.SDClient
	var qiniuVideoClient *client.QiniuVideoClient

	switch cfg.VideoGeneration.Type {
	case "qiniu":
		// 七牛云文生视频
		qiniuVideoClient = client.NewQiniuVideoClient(
			cfg.VideoGeneration.Qiniu.APIURL,
			cfg.VideoGeneration.Qiniu.APIKey,
			cfg.VideoGeneration.Qiniu.Model,
			cfg.VideoGeneration.Qiniu.Timeout,
		)
		logger.Info("Qiniu Video client initialized",
			zap.String("api_url", cfg.VideoGeneration.Qiniu.APIURL),
			zap.String("model", cfg.VideoGeneration.Qiniu.Model))

		// 检查健康状态
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := qiniuVideoClient.CheckHealth(ctx); err != nil {
			logger.Warn("Qiniu Video API health check failed", zap.Error(err))
		}

	case "local_sd":
		// 本地Stable Diffusion
		sdClient = client.NewSDClient(
			cfg.VideoGeneration.LocalSD.APIURL,
			cfg.VideoGeneration.LocalSD.Timeout,
		)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sdClient.CheckHealth(ctx); err != nil {
			logger.Warn("Stable Diffusion API health check failed (service may not be running)", zap.Error(err))
		} else {
			logger.Info("Stable Diffusion client initialized",
				zap.String("api_url", cfg.VideoGeneration.LocalSD.APIURL))
		}

	default:
		logger.Fatal("Unsupported video generation type", zap.String("type", cfg.VideoGeneration.Type))
	}

	// 确保七牛云服务可用
	if cfg.VideoGeneration.Type == "qiniu" && qiniuVideoClient == nil {
		logger.Fatal("Qiniu video client is required but not initialized")
	}

	// 创建任务管理器
	taskManager := task.NewManager()

	// 创建服务
	parserService := service.NewParserService(openaiClient, cfg.Storage.DataDir)
	storyboardService := service.NewStoryboardService(openaiClient, cfg.Storage.DataDir)

	// 解析视频分辨率
	var width, height int
	switch cfg.Video.Resolution {
	case "1920x1080":
		width, height = 1920, 1080
	case "1280x720":
		width, height = 1280, 720
	case "3840x2160":
		width, height = 3840, 2160
	default:
		width, height = 1920, 1080
	}

	imageService := service.NewImageService(sdClient, storyboardService, cfg.Storage.DataDir, width, height)
	renderService := service.NewRenderService(cfg.Storage.DataDir, cfg.Video.FPS)

	// 创建七牛云视频服务
	var qiniuVideoService *service.QiniuVideoService
	
	if qiniuVideoClient != nil {
		qiniuVideoService = service.NewQiniuVideoService(
			qiniuVideoClient,
			cfg.Storage.DataDir,
			cfg.VideoGeneration.Qiniu.MaxWaitTime,
		)
		logger.Info("Qiniu Video Service initialized")
	}

	// 创建HTTP处理器
	videoHandler := handler.NewVideoHandler(
		taskManager,
		parserService,
		storyboardService,
		imageService,
		renderService,
		qiniuVideoService,
		cfg,
	)

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由器
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
			"mode":    cfg.VideoGeneration.Type,
		})
	})

	// API路由
	api := r.Group("/api")
	{
		api.POST("/generate", videoHandler.Generate)
		api.GET("/tasks/:task_id", videoHandler.GetTask)
		api.GET("/tasks", videoHandler.ListTasks)
		api.GET("/download/:task_id", videoHandler.Download)
		api.DELETE("/tasks/:task_id", videoHandler.DeleteTask)
	}

	// 启动服务器
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	logger.Info("Server starting",
		zap.String("address", addr),
		zap.String("mode", cfg.Server.Mode))

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
