# 文生漫画视频工具 - MVP开发计划

## 1. MVP目标

**核心功能**: 输入小说文本 → 输出漫画视频

**使用场景**: 个人本地使用,无需用户管理

**预期效果**:
- 输入: 500-1000字的小说片段
- 输出: 30秒-1分钟的漫画视频(MP4)
- 处理时间: 5-10分钟

---

## 2. 功能范围

### 2.1 包含功能(IN)

- ✅ 小说文本解析
- ✅ 自动分镜生成
- ✅ AI图像生成(漫画风格)
- ✅ 角色一致性保持(基础)
- ✅ 视频合成(图片序列+转场)
- ✅ BGM添加(预设音乐)
- ✅ 字幕自动生成
- ✅ 导出MP4视频

### 2.2 不包含功能(OUT)

- ❌ 用户注册/登录
- ❌ 订阅管理
- ❌ 项目保存/管理
- ❌ 在线编辑器
- ❌ TTS语音合成
- ❌ 音效系统
- ❌ 实时预览
- ❌ 多种画风选择(只支持1种)
- ❌ 手动调整功能
- ❌ Web前端界面

### 2.3 简化策略

| 完整版 | MVP版 |
|-------|-------|
| 微服务架构 | 单体应用 |
| PostgreSQL + MongoDB | SQLite(可选) |
| RabbitMQ消息队列 | Go Channel |
| 多种画风 | 固定日系漫画风格 |
| 用户管理 | 无需认证 |
| 在线编辑 | 命令行交互 |
| 分布式任务队列 | 本地同步处理 |
| 对象存储S3 | 本地文件系统 |

---

## 3. 技术架构(MVP)

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    命令行客户端                          │
│              (CLI / 简单HTTP API)                       │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP REST
┌────────────────────────┴────────────────────────────────┐
│                 Go API Server (单体应用)                 │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │           Handler Layer                          │  │
│  │  - 文本上传接口                                   │  │
│  │  - 任务查询接口                                   │  │
│  │  - 视频下载接口                                   │  │
│  └──────────────────┬───────────────────────────────┘  │
│                     │                                   │
│  ┌──────────────────┴───────────────────────────────┐  │
│  │          Service Layer                           │  │
│  │                                                   │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐       │  │
│  │  │ 剧本解析 │  │ 分镜生成 │  │ 图像生成 │       │  │
│  │  └──────────┘  └──────────┘  └──────────┘       │  │
│  │                                                   │  │
│  │  ┌──────────┐  ┌──────────┐                     │  │
│  │  │ 视频合成 │  │ 字幕生成 │                     │  │
│  │  └──────────┘  └──────────┘                     │  │
│  └──────────────────┬───────────────────────────────┘  │
│                     │                                   │
│  ┌──────────────────┴───────────────────────────────┐  │
│  │          AI Client Layer                         │  │
│  │  - Stable Diffusion (本地/API)                   │  │
│  │  - OpenAI API (剧本解析)                         │  │
│  └──────────────────┬───────────────────────────────┘  │
└─────────────────────┼───────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────┐
│                   本地存储                               │
│  - 上传的文本文件                                        │
│  - 生成的图片                                           │
│  - 合成的视频                                           │
│  - 配置文件(config.yaml)                                │
└─────────────────────────────────────────────────────────┘
```

### 3.2 技术栈

#### 后端
- **语言**: Go 1.21+
- **Web框架**: Gin (轻量级HTTP框架)
- **配置管理**: Viper
- **日志**: Zap
- **视频处理**: FFmpeg (通过exec调用)

#### AI服务
- **图像生成**:
  - 方案1: Stable Diffusion Web UI API (本地部署)
  - 方案2: Replicate API (云端,需付费)
- **文本处理**:
  - OpenAI GPT-4 API (剧本解析和分镜生成)
  - 或本地LLM: Ollama (免费但效果略差)

#### 存储
- **文件系统**: 本地文件存储
- **目录结构**:
  ```
  ./data/
    ├── uploads/        # 上传的文本
    ├── projects/       # 项目数据
    │   └── {task_id}/
    │       ├── script.txt
    │       ├── parsed.json
    │       ├── storyboard.json
    │       ├── images/
    │       │   ├── shot_001.png
    │       │   └── ...
    │       └── output.mp4
    └── assets/         # 公共资源
        ├── bgm/        # 背景音乐
        └── fonts/      # 字体文件
  ```

#### 工具
- **FFmpeg**: 视频合成
- **ImageMagick**: 图片处理(可选)

---

## 4. 数据流程

```
1. 用户上传小说文本
   ↓
2. Go服务解析文本
   └─> 调用OpenAI API识别场景、角色、对话
   └─> 生成结构化数据(JSON)
   ↓
3. 生成分镜脚本
   └─> 调用OpenAI API根据解析结果设计分镜
   └─> 确定每个镜头的描述、时长、类型
   ↓
4. 生成图像(串行/并行)
   └─> 为每个镜头生成Prompt
   └─> 调用Stable Diffusion API生成图片
   └─> 保存图片到本地
   └─> (可选)角色一致性检查
   ↓
5. 生成字幕
   └─> 根据对话和时间轴生成SRT文件
   ↓
6. 视频合成
   └─> FFmpeg读取图片序列
   └─> 添加转场效果
   └─> 混入BGM
   └─> 烧录字幕
   └─> 导出MP4
   ↓
7. 返回视频文件路径
```

---

## 5. API设计(极简版)

### 5.1 接口列表

```
POST   /api/generate         # 创建生成任务
GET    /api/tasks/{task_id}  # 查询任务状态
GET    /api/download/{task_id} # 下载视频
DELETE /api/tasks/{task_id}  # 删除任务(清理文件)
GET    /health               # 健康检查
```

### 5.2 详细接口

#### 创建生成任务

```bash
POST /api/generate
Content-Type: application/json

Request:
{
    "text": "场景:现代都市街道,傍晚。小明独自走在回家的路上...",
    "options": {
        "style": "anime",           # 画风(MVP只支持anime)
        "duration_target": 60,      # 目标时长(秒)
        "aspect_ratio": "16:9",     # 比例
        "bgm": "default"            # BGM选择
    }
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "task_id": "uuid-1234",
        "status": "processing",
        "estimated_time": 300  # 预计处理时间(秒)
    }
}
```

#### 查询任务状态

```bash
GET /api/tasks/{task_id}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "task_id": "uuid-1234",
        "status": "processing",  # queued, processing, completed, failed
        "progress": 45,          # 0-100
        "current_step": "generating_images",  # 当前步骤
        "steps": [
            {
                "name": "parse_script",
                "status": "completed",
                "duration": 2.5
            },
            {
                "name": "generate_storyboard",
                "status": "completed",
                "duration": 3.2
            },
            {
                "name": "generate_images",
                "status": "processing",
                "progress": 45,
                "current": "6/12 shots"
            },
            {
                "name": "render_video",
                "status": "pending"
            }
        ],
        "result": null,  # 完成后包含视频信息
        "error": null,
        "created_at": "2025-10-26T12:00:00Z",
        "updated_at": "2025-10-26T12:05:00Z"
    }
}
```

#### 下载视频

```bash
GET /api/download/{task_id}

Response:
# 直接返回视频文件流
Content-Type: video/mp4
Content-Disposition: attachment; filename="output.mp4"

或

Response: (如果未完成)
{
    "code": 40001,
    "message": "Video not ready yet"
}
```

---

## 6. 核心模块实现

### 6.1 项目结构

```
mvp-video-generator/
├── cmd/
│   └── server/
│       └── main.go                 # 入口文件
│
├── internal/
│   ├── handler/
│   │   └── video_handler.go       # HTTP处理器
│   │
│   ├── service/
│   │   ├── parser_service.go      # 剧本解析
│   │   ├── storyboard_service.go  # 分镜生成
│   │   ├── image_service.go       # 图像生成
│   │   └── render_service.go      # 视频渲染
│   │
│   ├── client/
│   │   ├── openai_client.go       # OpenAI客户端
│   │   └── sd_client.go           # Stable Diffusion客户端
│   │
│   ├── model/
│   │   └── types.go               # 数据结构定义
│   │
│   └── task/
│       └── manager.go             # 任务管理器
│
├── pkg/
│   ├── config/
│   │   └── config.go              # 配置管理
│   ├── logger/
│   │   └── logger.go              # 日志
│   ├── ffmpeg/
│   │   └── ffmpeg.go              # FFmpeg封装
│   └── utils/
│       └── file.go                # 文件工具
│
├── configs/
│   └── config.yaml                # 配置文件
│
├── data/                          # 数据目录(gitignore)
│   ├── uploads/
│   ├── projects/
│   └── assets/
│
├── scripts/
│   ├── setup.sh                   # 环境设置脚本
│   └── download_assets.sh         # 下载BGM等资源
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 6.2 数据结构定义

```go
// internal/model/types.go

package model

import "time"

// Task 任务
type Task struct {
    ID          string    `json:"task_id"`
    Status      string    `json:"status"` // queued, processing, completed, failed
    Progress    int       `json:"progress"` // 0-100
    CurrentStep string    `json:"current_step"`
    Steps       []Step    `json:"steps"`
    Input       Input     `json:"input"`
    Result      *Result   `json:"result,omitempty"`
    Error       string    `json:"error,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// Step 处理步骤
type Step struct {
    Name     string  `json:"name"`
    Status   string  `json:"status"` // pending, processing, completed, failed
    Progress int     `json:"progress,omitempty"`
    Current  string  `json:"current,omitempty"` // 当前进度描述
    Duration float64 `json:"duration,omitempty"` // 耗时(秒)
}

// Input 输入参数
type Input struct {
    Text    string  `json:"text"`
    Options Options `json:"options"`
}

// Options 生成选项
type Options struct {
    Style          string `json:"style" default:"anime"`
    DurationTarget int    `json:"duration_target" default:"60"`
    AspectRatio    string `json:"aspect_ratio" default:"16:9"`
    BGM            string `json:"bgm" default:"default"`
}

// Result 生成结果
type Result struct {
    VideoPath    string  `json:"video_path"`
    Duration     float64 `json:"duration"`
    Resolution   string  `json:"resolution"`
    FileSize     int64   `json:"file_size"`
    ThumbnailURL string  `json:"thumbnail_url,omitempty"`
}

// ParsedScript 解析后的剧本
type ParsedScript struct {
    Scenes     []Scene  `json:"scenes"`
    Characters []string `json:"characters"`
    Metadata   Metadata `json:"metadata"`
}

// Scene 场景
type Scene struct {
    ID         int      `json:"id"`
    Location   string   `json:"location"`
    Time       string   `json:"time"`
    Characters []string `json:"characters"`
    Dialogues  []Dialogue `json:"dialogues"`
}

// Dialogue 对话
type Dialogue struct {
    Character string `json:"character"`
    Text      string `json:"text"`
    Emotion   string `json:"emotion,omitempty"`
}

// Storyboard 分镜脚本
type Storyboard struct {
    Shots []Shot `json:"shots"`
}

// Shot 镜头
type Shot struct {
    ID          int     `json:"id"`
    Type        string  `json:"type"` // closeup, medium, long
    Description string  `json:"description"`
    Characters  []string `json:"characters"`
    Duration    float64 `json:"duration"`
    Transition  string  `json:"transition"` // cut, fade, dissolve
    Dialogue    *Dialogue `json:"dialogue,omitempty"`
    ImagePath   string  `json:"image_path,omitempty"`
    Prompt      string  `json:"prompt,omitempty"`
}

// Metadata 元数据
type Metadata struct {
    TotalScenes         int     `json:"total_scenes"`
    TotalShots          int     `json:"total_shots"`
    EstimatedDuration   float64 `json:"estimated_duration"`
    WordCount           int     `json:"word_count"`
}
```

---

## 7. 开发步骤

### Phase 1: 环境搭建 (1天)

#### 步骤1.1: 初始化项目

```bash
# 创建项目目录
mkdir mvp-video-generator
cd mvp-video-generator

# 初始化Go模块
go mod init github.com/yourusername/mvp-video-generator

# 创建目录结构
mkdir -p cmd/server
mkdir -p internal/{handler,service,client,model,task}
mkdir -p pkg/{config,logger,ffmpeg,utils}
mkdir -p configs
mkdir -p data/{uploads,projects,assets/{bgm,fonts}}
mkdir -p scripts
```

#### 步骤1.2: 安装依赖

```bash
# Go依赖
go get github.com/gin-gonic/gin
go get github.com/spf13/viper
go get go.uber.org/zap
go get github.com/google/uuid
go get github.com/sashabaranov/go-openai

# 系统依赖
# macOS
brew install ffmpeg imagemagick

# Linux
sudo apt install ffmpeg imagemagick

# 验证安装
ffmpeg -version
```

#### 步骤1.3: 配置文件

```yaml
# configs/config.yaml

server:
  port: 8080
  host: "0.0.0.0"

storage:
  data_dir: "./data"
  max_upload_size: 10485760  # 10MB

openai:
  api_key: "your-api-key"
  model: "gpt-4"
  base_url: "https://api.openai.com/v1"

stable_diffusion:
  type: "local"  # local or api
  # 本地SD Web UI
  local:
    api_url: "http://127.0.0.1:7860"
  # 或使用Replicate API
  replicate:
    api_token: "your-replicate-token"

video:
  default_bgm: "default.mp3"
  resolution: "1920x1080"
  fps: 30
  quality: "high"

limits:
  max_concurrent_tasks: 1  # MVP单任务处理
  max_shots_per_video: 20
  max_video_duration: 120  # 秒
```

---

### Phase 2: 核心功能开发 (5-7天)

#### 步骤2.1: 基础框架搭建 (0.5天)

**任务清单**:
- [ ] 创建main.go入口
- [ ] 配置加载(Viper)
- [ ] 日志初始化(Zap)
- [ ] HTTP服务器(Gin)
- [ ] 健康检查接口

**代码示例**:

```go
// cmd/server/main.go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/mvp-video-generator/internal/handler"
    "github.com/yourusername/mvp-video-generator/pkg/config"
    "github.com/yourusername/mvp-video-generator/pkg/logger"
)

func main() {
    // 加载配置
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 初始化日志
    logger.Init(cfg.Log)
    defer logger.Sync()

    // 创建HTTP服务器
    r := gin.Default()

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 注册路由
    api := r.Group("/api")
    {
        videoHandler := handler.NewVideoHandler(cfg)
        api.POST("/generate", videoHandler.Generate)
        api.GET("/tasks/:task_id", videoHandler.GetTask)
        api.GET("/download/:task_id", videoHandler.Download)
        api.DELETE("/tasks/:task_id", videoHandler.DeleteTask)
    }

    // 启动服务器
    addr := cfg.Server.Host + ":" + cfg.Server.Port
    logger.Info("Starting server on " + addr)
    if err := r.Run(addr); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

#### 步骤2.2: 任务管理器 (0.5天)

**任务清单**:
- [ ] 实现内存任务队列
- [ ] 任务状态管理
- [ ] 任务CRUD操作

```go
// internal/task/manager.go
package task

import (
    "sync"

    "github.com/yourusername/mvp-video-generator/internal/model"
)

type Manager struct {
    tasks map[string]*model.Task
    mu    sync.RWMutex
}

func NewManager() *Manager {
    return &Manager{
        tasks: make(map[string]*model.Task),
    }
}

func (m *Manager) Create(task *model.Task) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.tasks[task.ID] = task
}

func (m *Manager) Get(taskID string) (*model.Task, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    task, ok := m.tasks[taskID]
    return task, ok
}

func (m *Manager) Update(task *model.Task) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.tasks[task.ID] = task
}

func (m *Manager) Delete(taskID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.tasks, taskID)
}
```

#### 步骤2.3: OpenAI客户端封装 (0.5天)

**任务清单**:
- [ ] 封装OpenAI API调用
- [ ] 实现剧本解析Prompt
- [ ] 实现分镜生成Prompt

```go
// internal/client/openai_client.go
package client

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/sashabaranov/go-openai"
    "github.com/yourusername/mvp-video-generator/internal/model"
)

type OpenAIClient struct {
    client *openai.Client
    model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
    return &OpenAIClient{
        client: openai.NewClient(apiKey),
        model:  model,
    }
}

// ParseScript 解析剧本
func (c *OpenAIClient) ParseScript(ctx context.Context, text string) (*model.ParsedScript, error) {
    prompt := fmt.Sprintf(`
你是一个专业的剧本分析师。请分析以下小说文本,提取关键信息:

文本:
%s

请以JSON格式返回以下信息:
{
    "scenes": [
        {
            "id": 1,
            "location": "场景地点",
            "time": "时间",
            "characters": ["角色1", "角色2"],
            "dialogues": [
                {
                    "character": "角色名",
                    "text": "对话内容",
                    "emotion": "情绪"
                }
            ]
        }
    ],
    "characters": ["角色1", "角色2"],
    "metadata": {
        "total_scenes": 1,
        "word_count": 100
    }
}
`, text)

    resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: c.model,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: prompt,
            },
        },
        ResponseFormat: &openai.ChatCompletionResponseFormat{
            Type: openai.ChatCompletionResponseFormatTypeJSONObject,
        },
    })

    if err != nil {
        return nil, err
    }

    var parsed model.ParsedScript
    if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &parsed); err != nil {
        return nil, err
    }

    return &parsed, nil
}

// GenerateStoryboard 生成分镜
func (c *OpenAIClient) GenerateStoryboard(ctx context.Context, parsed *model.ParsedScript, targetDuration int) (*model.Storyboard, error) {
    // 类似实现...
    return nil, nil
}
```

#### 步骤2.4: Stable Diffusion客户端 (1天)

**任务清单**:
- [ ] 封装SD API调用
- [ ] Prompt生成优化
- [ ] 图像生成与保存

```go
// internal/client/sd_client.go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

type SDClient struct {
    apiURL string
    client *http.Client
}

func NewSDClient(apiURL string) *SDClient {
    return &SDClient{
        apiURL: apiURL,
        client: &http.Client{},
    }
}

// GenerateImage 生成图像
func (c *SDClient) GenerateImage(prompt, negativePrompt string, width, height int) ([]byte, error) {
    reqBody := map[string]interface{}{
        "prompt":          prompt,
        "negative_prompt": negativePrompt,
        "steps":           30,
        "cfg_scale":       7.5,
        "width":           width,
        "height":          height,
        "sampler_name":    "DPM++ 2M Karras",
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    resp, err := c.client.Post(
        c.apiURL+"/sdapi/v1/txt2img",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Images []string `json:"images"` // base64 encoded
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    if len(result.Images) == 0 {
        return nil, fmt.Errorf("no image generated")
    }

    // Decode base64
    // ... (省略base64解码逻辑)

    return imageBytes, nil
}
```

#### 步骤2.5: 剧本解析服务 (1天)

**任务清单**:
- [ ] 实现解析逻辑
- [ ] 场景识别
- [ ] 角色提取
- [ ] 对话分析

```go
// internal/service/parser_service.go
package service

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/yourusername/mvp-video-generator/internal/client"
    "github.com/yourusername/mvp-video-generator/internal/model"
    "github.com/yourusername/mvp-video-generator/pkg/logger"
    "github.com/yourusername/mvp-video-generator/pkg/utils"
)

type ParserService struct {
    openaiClient *client.OpenAIClient
}

func NewParserService(openaiClient *client.OpenAIClient) *ParserService {
    return &ParserService{
        openaiClient: openaiClient,
    }
}

func (s *ParserService) Parse(ctx context.Context, taskID, text string) (*model.ParsedScript, error) {
    logger.Info(fmt.Sprintf("Parsing script for task %s", taskID))

    // 调用OpenAI解析
    parsed, err := s.openaiClient.ParseScript(ctx, text)
    if err != nil {
        return nil, fmt.Errorf("failed to parse script: %w", err)
    }

    // 保存解析结果
    projectDir := filepath.Join("data", "projects", taskID)
    if err := os.MkdirAll(projectDir, 0755); err != nil {
        return nil, err
    }

    if err := utils.SaveJSON(filepath.Join(projectDir, "parsed.json"), parsed); err != nil {
        return nil, err
    }

    logger.Info(fmt.Sprintf("Script parsed successfully: %d scenes, %d characters",
        len(parsed.Scenes), len(parsed.Characters)))

    return parsed, nil
}
```

#### 步骤2.6: 分镜生成服务 (1天)

**任务清单**:
- [ ] 分镜算法实现
- [ ] 镜头类型决策
- [ ] 时长分配
- [ ] Prompt生成

```go
// internal/service/storyboard_service.go
package service

import (
    "context"
    "fmt"

    "github.com/yourusername/mvp-video-generator/internal/client"
    "github.com/yourusername/mvp-video-generator/internal/model"
)

type StoryboardService struct {
    openaiClient *client.OpenAIClient
}

func NewStoryboardService(openaiClient *client.OpenAIClient) *StoryboardService {
    return &StoryboardService{
        openaiClient: openaiClient,
    }
}

func (s *StoryboardService) Generate(ctx context.Context, parsed *model.ParsedScript, targetDuration int) (*model.Storyboard, error) {
    // 使用OpenAI生成分镜
    storyboard, err := s.openaiClient.GenerateStoryboard(ctx, parsed, targetDuration)
    if err != nil {
        return nil, err
    }

    // 为每个镜头生成详细的Prompt
    for i := range storyboard.Shots {
        shot := &storyboard.Shots[i]
        shot.Prompt = s.generatePrompt(shot)
    }

    return storyboard, nil
}

func (s *StoryboardService) generatePrompt(shot *model.Shot) string {
    basePrompt := "anime style, manga, high quality, detailed, "

    // 添加镜头类型
    switch shot.Type {
    case "closeup":
        basePrompt += "close-up shot, "
    case "medium":
        basePrompt += "medium shot, "
    case "long":
        basePrompt += "wide shot, establishing shot, "
    }

    // 添加场景描述
    basePrompt += shot.Description

    // 添加角色
    if len(shot.Characters) > 0 {
        basePrompt += fmt.Sprintf(", characters: %v", shot.Characters)
    }

    // 添加情绪
    if shot.Dialogue != nil && shot.Dialogue.Emotion != "" {
        basePrompt += fmt.Sprintf(", %s emotion", shot.Dialogue.Emotion)
    }

    return basePrompt
}
```

#### 步骤2.7: 图像生成服务 (1天)

**任务清单**:
- [ ] 图像生成逻辑
- [ ] 批量/串行生成
- [ ] 错误重试机制
- [ ] 进度更新

```go
// internal/service/image_service.go
package service

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/yourusername/mvp-video-generator/internal/client"
    "github.com/yourusername/mvp-video-generator/internal/model"
    "github.com/yourusername/mvp-video-generator/pkg/logger"
)

type ImageService struct {
    sdClient *client.SDClient
}

func NewImageService(sdClient *client.SDClient) *ImageService {
    return &ImageService{
        sdClient: sdClient,
    }
}

func (s *ImageService) GenerateAll(ctx context.Context, taskID string, storyboard *model.Storyboard, progressCallback func(int, int)) error {
    imagesDir := filepath.Join("data", "projects", taskID, "images")
    if err := os.MkdirAll(imagesDir, 0755); err != nil {
        return err
    }

    totalShots := len(storyboard.Shots)

    // 串行生成(MVP简化,避免GPU内存不足)
    for i := range storyboard.Shots {
        shot := &storyboard.Shots[i]

        logger.Info(fmt.Sprintf("Generating image %d/%d: %s", i+1, totalShots, shot.Description))

        // 生成图像
        imageData, err := s.sdClient.GenerateImage(
            shot.Prompt,
            "low quality, blurry, distorted", // negative prompt
            1920, 1080,
        )
        if err != nil {
            return fmt.Errorf("failed to generate image for shot %d: %w", i, err)
        }

        // 保存图像
        imagePath := filepath.Join(imagesDir, fmt.Sprintf("shot_%03d.png", i+1))
        if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
            return err
        }

        shot.ImagePath = imagePath

        // 更新进度
        if progressCallback != nil {
            progressCallback(i+1, totalShots)
        }
    }

    return nil
}
```

#### 步骤2.8: 视频渲染服务 (1.5天)

**任务清单**:
- [ ] FFmpeg命令封装
- [ ] 图片序列转视频
- [ ] 添加转场效果
- [ ] 混入BGM
- [ ] 烧录字幕

```go
// internal/service/render_service.go
package service

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/yourusername/mvp-video-generator/internal/model"
    "github.com/yourusername/mvp-video-generator/pkg/logger"
)

type RenderService struct {
    dataDir string
}

func NewRenderService(dataDir string) *RenderService {
    return &RenderService{
        dataDir: dataDir,
    }
}

func (s *RenderService) Render(taskID string, storyboard *model.Storyboard, bgmPath string) (string, error) {
    projectDir := filepath.Join(s.dataDir, "projects", taskID)
    imagesDir := filepath.Join(projectDir, "images")
    outputPath := filepath.Join(projectDir, "output.mp4")

    // 生成FFmpeg输入文件列表
    concatFile := filepath.Join(projectDir, "concat.txt")
    if err := s.generateConcatFile(concatFile, storyboard); err != nil {
        return "", err
    }

    // 构建FFmpeg命令
    // ffmpeg -f concat -safe 0 -i concat.txt -i bgm.mp3 -c:v libx264 -c:a aac -shortest output.mp4
    cmd := exec.Command("ffmpeg",
        "-f", "concat",
        "-safe", "0",
        "-i", concatFile,
        "-i", bgmPath,
        "-c:v", "libx264",
        "-preset", "medium",
        "-crf", "23",
        "-c:a", "aac",
        "-b:a", "192k",
        "-shortest",
        "-y", // 覆盖已存在文件
        outputPath,
    )

    logger.Info("Running FFmpeg command: " + cmd.String())

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
    }

    return outputPath, nil
}

func (s *RenderService) generateConcatFile(path string, storyboard *model.Storyboard) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    for _, shot := range storyboard.Shots {
        // file 'images/shot_001.png'
        // duration 3.0
        fmt.Fprintf(f, "file '%s'\n", shot.ImagePath)
        fmt.Fprintf(f, "duration %.2f\n", shot.Duration)
    }

    // 最后一帧需要重复
    lastShot := storyboard.Shots[len(storyboard.Shots)-1]
    fmt.Fprintf(f, "file '%s'\n", lastShot.ImagePath)

    return nil
}
```

#### 步骤2.9: HTTP Handler实现 (1天)

**任务清单**:
- [ ] 实现Generate接口
- [ ] 实现GetTask接口
- [ ] 实现Download接口
- [ ] 异步任务处理

```go
// internal/handler/video_handler.go
package handler

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/yourusername/mvp-video-generator/internal/model"
    "github.com/yourusername/mvp-video-generator/internal/service"
    "github.com/yourusername/mvp-video-generator/internal/task"
    "github.com/yourusername/mvp-video-generator/pkg/logger"
)

type VideoHandler struct {
    taskManager       *task.Manager
    parserService     *service.ParserService
    storyboardService *service.StoryboardService
    imageService      *service.ImageService
    renderService     *service.RenderService
}

func NewVideoHandler(cfg *config.Config) *VideoHandler {
    // 初始化各种服务...
    return &VideoHandler{
        taskManager: task.NewManager(),
        // ... 初始化其他服务
    }
}

func (h *VideoHandler) Generate(c *gin.Context) {
    var req model.Input
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 创建任务
    taskID := uuid.New().String()
    t := &model.Task{
        ID:          taskID,
        Status:      "queued",
        Progress:    0,
        CurrentStep: "queued",
        Input:       req,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    h.taskManager.Create(t)

    // 异步处理
    go h.processTask(taskID)

    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "message": "success",
        "data": gin.H{
            "task_id":        taskID,
            "status":         "processing",
            "estimated_time": 300,
        },
    })
}

func (h *VideoHandler) processTask(taskID string) {
    ctx := context.Background()

    t, _ := h.taskManager.Get(taskID)
    t.Status = "processing"
    h.taskManager.Update(t)

    // 1. 解析剧本
    h.updateStep(taskID, "parse_script", "processing")
    parsed, err := h.parserService.Parse(ctx, taskID, t.Input.Text)
    if err != nil {
        h.failTask(taskID, "parse_script", err)
        return
    }
    h.updateStep(taskID, "parse_script", "completed")

    // 2. 生成分镜
    h.updateStep(taskID, "generate_storyboard", "processing")
    storyboard, err := h.storyboardService.Generate(ctx, parsed, t.Input.Options.DurationTarget)
    if err != nil {
        h.failTask(taskID, "generate_storyboard", err)
        return
    }
    h.updateStep(taskID, "generate_storyboard", "completed")

    // 3. 生成图像
    h.updateStep(taskID, "generate_images", "processing")
    err = h.imageService.GenerateAll(ctx, taskID, storyboard, func(current, total int) {
        t, _ := h.taskManager.Get(taskID)
        t.Progress = (current * 100) / total
        h.taskManager.Update(t)
    })
    if err != nil {
        h.failTask(taskID, "generate_images", err)
        return
    }
    h.updateStep(taskID, "generate_images", "completed")

    // 4. 渲染视频
    h.updateStep(taskID, "render_video", "processing")
    outputPath, err := h.renderService.Render(taskID, storyboard, "data/assets/bgm/default.mp3")
    if err != nil {
        h.failTask(taskID, "render_video", err)
        return
    }
    h.updateStep(taskID, "render_video", "completed")

    // 完成
    t, _ = h.taskManager.Get(taskID)
    t.Status = "completed"
    t.Progress = 100
    t.Result = &model.Result{
        VideoPath: outputPath,
    }
    h.taskManager.Update(t)

    logger.Info(fmt.Sprintf("Task %s completed successfully", taskID))
}

func (h *VideoHandler) GetTask(c *gin.Context) {
    taskID := c.Param("task_id")

    t, ok := h.taskManager.Get(taskID)
    if !ok {
        c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "message": "success",
        "data": t,
    })
}

func (h *VideoHandler) Download(c *gin.Context) {
    taskID := c.Param("task_id")

    t, ok := h.taskManager.Get(taskID)
    if !ok {
        c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
        return
    }

    if t.Status != "completed" || t.Result == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "video not ready"})
        return
    }

    c.File(t.Result.VideoPath)
}

// 辅助方法
func (h *VideoHandler) updateStep(taskID, step, status string) {
    // 实现略...
}

func (h *VideoHandler) failTask(taskID, step string, err error) {
    // 实现略...
}
```

---

### Phase 3: 测试与优化 (2天)

#### 步骤3.1: 单元测试

**任务清单**:
- [ ] Parser服务测试
- [ ] Storyboard服务测试
- [ ] Image服务测试(mock)
- [ ] Render服务测试

#### 步骤3.2: 集成测试

**任务清单**:
- [ ] 端到端流程测试
- [ ] 错误处理测试
- [ ] 性能测试

#### 步骤3.3: 优化

**任务清单**:
- [ ] Prompt优化(提高生成质量)
- [ ] 错误重试机制
- [ ] 日志完善
- [ ] 清理临时文件

---

### Phase 4: 文档与部署 (1天)

#### 步骤4.1: 编写文档

**任务清单**:
- [ ] README.md
- [ ] API文档
- [ ] 部署指南
- [ ] 使用示例

#### 步骤4.2: Docker化(可选)

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY --from=builder /app/server .
COPY configs/ ./configs/
COPY data/assets/ ./data/assets/

EXPOSE 8080
CMD ["./server"]
```

#### 步骤4.3: 部署脚本

```bash
# scripts/setup.sh

#!/bin/bash

echo "Setting up MVP Video Generator..."

# 创建数据目录
mkdir -p data/{uploads,projects,assets/{bgm,fonts}}

# 下载示例BGM
echo "Downloading sample BGM..."
# wget -O data/assets/bgm/default.mp3 "https://example.com/bgm.mp3"

# 安装Go依赖
echo "Installing Go dependencies..."
go mod download

# 构建
echo "Building..."
go build -o bin/server cmd/server/main.go

echo "Setup complete!"
echo "Please update configs/config.yaml with your API keys"
```

---

## 8. 使用示例

### 8.1 启动服务

```bash
# 配置API密钥
vim configs/config.yaml

# 启动服务
go run cmd/server/main.go

# 或使用构建后的二进制
./bin/server
```

### 8.2 API调用

```bash
# 1. 创建生成任务
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:现代都市街道,傍晚。小明独自走在回家的路上,心里想着今天发生的事情。突然,他看到了小红站在路边等他。小明(惊讶):你怎么在这里?小红(微笑):我在等你啊。小明的心跳加速了。",
    "options": {
      "style": "anime",
      "duration_target": 30,
      "aspect_ratio": "16:9",
      "bgm": "default"
    }
  }'

# Response:
# {
#   "code": 0,
#   "message": "success",
#   "data": {
#     "task_id": "abc-123-def",
#     "status": "processing",
#     "estimated_time": 300
#   }
# }

# 2. 查询任务状态
curl http://localhost:8080/api/tasks/abc-123-def

# 3. 下载视频(完成后)
curl http://localhost:8080/api/download/abc-123-def -o output.mp4
```

### 8.3 命令行客户端(可选)

```bash
# 创建简单的CLI工具
cat > cli.sh << 'EOF'
#!/bin/bash

API_URL="http://localhost:8080/api"

# 生成视频
generate() {
    local text_file=$1
    local text=$(cat "$text_file")

    response=$(curl -s -X POST "$API_URL/generate" \
        -H "Content-Type: application/json" \
        -d "{\"text\": \"$text\", \"options\": {\"duration_target\": 60}}")

    task_id=$(echo "$response" | jq -r '.data.task_id')
    echo "Task created: $task_id"

    # 轮询状态
    while true; do
        status=$(curl -s "$API_URL/tasks/$task_id" | jq -r '.data.status')
        progress=$(curl -s "$API_URL/tasks/$task_id" | jq -r '.data.progress')

        echo "Status: $status, Progress: $progress%"

        if [ "$status" == "completed" ]; then
            echo "Downloading video..."
            curl -o "output_${task_id}.mp4" "$API_URL/download/$task_id"
            echo "Done!"
            break
        elif [ "$status" == "failed" ]; then
            echo "Task failed!"
            break
        fi

        sleep 5
    done
}

# 使用
generate "my_novel.txt"
EOF

chmod +x cli.sh
```

---

## 9. 开发时间估算

| 阶段 | 任务 | 预计时间 |
|-----|------|---------|
| Phase 1 | 环境搭建 | 1天 |
| Phase 2 | 基础框架 | 0.5天 |
| | 任务管理器 | 0.5天 |
| | OpenAI客户端 | 0.5天 |
| | SD客户端 | 1天 |
| | 剧本解析服务 | 1天 |
| | 分镜生成服务 | 1天 |
| | 图像生成服务 | 1天 |
| | 视频渲染服务 | 1.5天 |
| | HTTP Handler | 1天 |
| Phase 3 | 测试与优化 | 2天 |
| Phase 4 | 文档与部署 | 1天 |
| **总计** | | **10-12天** |

---

## 10. 成本估算(MVP)

### 10.1 开发成本

- 个人开发: 10-12天
- 硬件要求: 支持CUDA的NVIDIA GPU(RTX 3060以上)

### 10.2 运行成本

#### 本地部署方案(推荐)

| 项目 | 成本 |
|-----|------|
| 硬件(一次性) | 已有GPU:¥0 / 购买GPU:¥2000-5000 |
| OpenAI API | ~¥0.5/视频 (GPT-4调用) |
| 电费 | ~¥0.2/视频 (GPU功耗) |
| **单视频成本** | **¥0.7** |

#### 云端API方案

| 项目 | 成本 |
|-----|------|
| OpenAI API | ~¥0.5/视频 |
| Replicate API | ~¥3/视频 (Stable Diffusion) |
| 服务器 | ¥100/月 (无GPU) |
| **单视频成本** | **¥3.5** |

---

## 11. 潜在问题与解决方案

### 11.1 技术问题

| 问题 | 解决方案 |
|-----|---------|
| SD生成速度慢 | 使用低步数(20步)预览,优化后再高质量生成 |
| GPU内存不足 | 串行生成图片,或使用更小的模型 |
| 角色一致性差 | 在Prompt中强化角色特征描述 |
| FFmpeg转场效果单调 | 使用xfade滤镜添加多种转场 |
| 视频时长不准 | 调整每张图片的duration参数 |

### 11.2 优化方向

1. **性能优化**
   - 并行生成图片(如果GPU允许)
   - 缓存已生成的角色图片
   - 使用TensorRT加速SD推理

2. **质量优化**
   - 使用ControlNet保持角色一致性
   - 优化Prompt模板
   - 添加图像后处理(超分辨率)

3. **功能扩展**
   - 添加简单的Web UI
   - 支持多种画风
   - 添加TTS语音
   - 实现项目保存功能

---

## 12. MVP验收标准

### 12.1 功能验收

- [ ] 能够接收文本输入(500-1000字)
- [ ] 自动生成5-15个分镜
- [ ] 为每个分镜生成漫画风格图片
- [ ] 合成视频(30秒-1分钟)
- [ ] 添加BGM
- [ ] 添加字幕
- [ ] 导出MP4格式视频
- [ ] 处理时间<10分钟

### 12.2 质量验收

- [ ] 图片清晰度满足1080p要求
- [ ] 角色外观基本一致(相似度>70%)
- [ ] 分镜转场自然
- [ ] BGM音量适中
- [ ] 字幕与画面同步
- [ ] 无明显错误和崩溃

### 12.3 用户体验

- [ ] API接口响应正常
- [ ] 任务状态实时更新
- [ ] 错误信息清晰易懂
- [ ] 文档完整,易于上手

---

## 13. 下一步计划

完成MVP后可以考虑:

1. **添加Web界面** - 使用React/Vue创建简单前端
2. **数据持久化** - 使用SQLite保存项目历史
3. **多用户支持** - 添加简单的用户管理
4. **角色库功能** - 保存和复用角色设定
5. **批量生成** - 支持一次处理多个章节
6. **云端部署** - 部署到云服务器供远程访问
7. **高级编辑** - 手动调整分镜、替换图片等

---

**文档版本**: v1.0
**创建日期**: 2025-10-26
**预计完成**: 2025-11-07 (10-12个工作日)
