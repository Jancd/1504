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
	StartAt  *time.Time `json:"start_at,omitempty"`
	EndAt    *time.Time `json:"end_at,omitempty"`
}

// Input 输入参数
type Input struct {
	Text    string  `json:"text" binding:"required"`
	Options Options `json:"options"`
}

// Options 生成选项
type Options struct {
	Style          string `json:"style"`
	DurationTarget int    `json:"duration_target"`
	AspectRatio    string `json:"aspect_ratio"`
	BGM            string `json:"bgm"`
}

// Result 生成结果
type Result struct {
	VideoPath    string  `json:"video_path"`
	Duration     float64 `json:"duration"`
	Resolution   string  `json:"resolution"`
	FileSize     int64   `json:"file_size"`
	ThumbnailURL string  `json:"thumbnail_url,omitempty"`
	ShotCount    int     `json:"shot_count"`
}

// ParsedScript 解析后的剧本
type ParsedScript struct {
	Scenes     []Scene  `json:"scenes"`
	Characters []string `json:"characters"`
	Metadata   Metadata `json:"metadata"`
}

// Scene 场景
type Scene struct {
	ID         int        `json:"id"`
	Location   string     `json:"location"`
	Time       string     `json:"time"`
	Characters []string   `json:"characters"`
	Dialogues  []Dialogue `json:"dialogues"`
	Actions    []Action   `json:"actions,omitempty"`
}

// Dialogue 对话
type Dialogue struct {
	Character string `json:"character"`
	Text      string `json:"text"`
	Emotion   string `json:"emotion,omitempty"`
}

// Action 动作描述
type Action struct {
	Character   string `json:"character,omitempty"`
	Description string `json:"description"`
}

// Storyboard 分镜脚本
type Storyboard struct {
	Shots         []Shot  `json:"shots"`
	TotalDuration float64 `json:"total_duration"`
}

// Shot 镜头
type Shot struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"` // closeup, medium, long
	Description string    `json:"description"`
	Characters  []string  `json:"characters"`
	Duration    float64   `json:"duration"`
	Transition  string    `json:"transition"` // cut, fade, dissolve
	Dialogue    *Dialogue `json:"dialogue,omitempty"`
	ImagePath   string    `json:"image_path,omitempty"`
	Prompt      string    `json:"prompt,omitempty"`
}

// Metadata 元数据
type Metadata struct {
	TotalScenes       int     `json:"total_scenes"`
	TotalShots        int     `json:"total_shots"`
	EstimatedDuration float64 `json:"estimated_duration"`
	WordCount         int     `json:"word_count"`
}

// APIResponse 统一API响应
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// StepName 步骤名称常量
const (
	StepParseScript       = "parse_script"
	StepGenerateStoryboard = "generate_storyboard"
	StepGenerateImages    = "generate_images"
	StepRenderVideo       = "render_video"
)

// TaskStatus 任务状态常量
const (
	TaskStatusQueued     = "queued"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// StepStatus 步骤状态常量
const (
	StepStatusPending    = "pending"
	StepStatusProcessing = "processing"
	StepStatusCompleted  = "completed"
	StepStatusFailed     = "failed"
)

// ShotType 镜头类型常量
const (
	ShotTypeCloseup = "closeup"
	ShotTypeMedium  = "medium"
	ShotTypeLong    = "long"
)

// Transition 转场类型常量
const (
	TransitionCut      = "cut"
	TransitionFade     = "fade"
	TransitionDissolve = "dissolve"
)

// NewTask 创建新任务
func NewTask(id string, input Input) *Task {
	now := time.Now()
	return &Task{
		ID:          id,
		Status:      TaskStatusQueued,
		Progress:    0,
		CurrentStep: "queued",
		Steps: []Step{
			{Name: StepParseScript, Status: StepStatusPending},
			{Name: StepGenerateStoryboard, Status: StepStatusPending},
			{Name: StepGenerateImages, Status: StepStatusPending},
			{Name: StepRenderVideo, Status: StepStatusPending},
		},
		Input:     input,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateStep 更新步骤状态
func (t *Task) UpdateStep(stepName, status string) {
	now := time.Now()
	for i := range t.Steps {
		if t.Steps[i].Name == stepName {
			t.Steps[i].Status = status
			if status == StepStatusProcessing {
				t.Steps[i].StartAt = &now
			} else if status == StepStatusCompleted || status == StepStatusFailed {
				t.Steps[i].EndAt = &now
				if t.Steps[i].StartAt != nil {
					t.Steps[i].Duration = now.Sub(*t.Steps[i].StartAt).Seconds()
				}
			}
			break
		}
	}
	t.CurrentStep = stepName
	t.UpdatedAt = now
}

// SetStepProgress 设置步骤进度
func (t *Task) SetStepProgress(stepName string, progress int, current string) {
	for i := range t.Steps {
		if t.Steps[i].Name == stepName {
			t.Steps[i].Progress = progress
			t.Steps[i].Current = current
			break
		}
	}
	t.UpdatedAt = time.Now()
}
