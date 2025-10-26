package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server          ServerConfig          `mapstructure:"server"`
	Storage         StorageConfig         `mapstructure:"storage"`
	OpenAI          OpenAIConfig          `mapstructure:"openai"`
	VideoGeneration VideoGenerationConfig `mapstructure:"video_generation"`
	Video           VideoConfig           `mapstructure:"video"`
	Limits          LimitsConfig          `mapstructure:"limits"`
	Log             LogConfig             `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Mode string `mapstructure:"mode"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir       string `mapstructure:"data_dir"`
	MaxUploadSize int64  `mapstructure:"max_upload_size"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
	Timeout int    `mapstructure:"timeout"`
}

// VideoGenerationConfig 视频生成配置
type VideoGenerationConfig struct {
	Type    string        `mapstructure:"type"`
	Qiniu   QiniuConfig   `mapstructure:"qiniu"`
	LocalSD LocalSDConfig `mapstructure:"local_sd"`
}

// QiniuConfig 七牛云配置
type QiniuConfig struct {
	APIURL      string `mapstructure:"api_url"`
	APIKey      string `mapstructure:"api_key"`
	Model       string `mapstructure:"model"`
	Timeout     int    `mapstructure:"timeout"`
	MaxWaitTime int    `mapstructure:"max_wait_time"`
}

// LocalSDConfig 本地SD配置
type LocalSDConfig struct {
	APIURL  string `mapstructure:"api_url"`
	Timeout int    `mapstructure:"timeout"`
}

// VideoConfig 视频配置
type VideoConfig struct {
	DefaultBGM  string `mapstructure:"default_bgm"`
	Resolution  string `mapstructure:"resolution"`
	FPS         int    `mapstructure:"fps"`
	Quality     string `mapstructure:"quality"`
	MaxDuration int    `mapstructure:"max_duration"`
}

// LimitsConfig 限制配置
type LimitsConfig struct {
	MaxConcurrentTasks int `mapstructure:"max_concurrent_tasks"`
	MaxShotsPerVideo   int `mapstructure:"max_shots_per_video"`
	MaxTextLength      int `mapstructure:"max_text_length"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 自动读取环境变量
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 扩展环境变量
	expandEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// expandEnvVars 展开环境变量
func expandEnvVars(v *viper.Viper) {
	// OpenAI API Key
	if apiKey := v.GetString("openai.api_key"); apiKey != "" {
		v.Set("openai.api_key", os.ExpandEnv(apiKey))
	}
}

// validate 验证配置
func validate(cfg *Config) error {
	// 验证OpenAI配置
	if cfg.OpenAI.APIKey == "" {
		return fmt.Errorf("openai.api_key is required")
	}

	// 验证视频生成配置
	validTypes := []string{"qiniu", "local_sd"}
	isValid := false
	for _, t := range validTypes {
		if cfg.VideoGeneration.Type == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("video_generation.type must be one of: qiniu, local_sd")
	}

	if cfg.VideoGeneration.Type == "qiniu" && cfg.VideoGeneration.Qiniu.APIKey == "" {
		return fmt.Errorf("video_generation.qiniu.api_key is required when type is 'qiniu'")
	}

	if cfg.VideoGeneration.Type == "local_sd" && cfg.VideoGeneration.LocalSD.APIURL == "" {
		return fmt.Errorf("video_generation.local_sd.api_url is required when type is 'local_sd'")
	}

	return nil
}
