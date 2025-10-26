# 文生漫画视频工具 - MVP版本

一个基于Go的AI驱动工具,可以将文字小说自动转换为漫画风格的短视频。

## 功能特性

- ✅ 自动剧本解析 (使用OpenAI GPT-4)
- ✅ 智能分镜生成
- ✅ AI图像生成 (使用Stable Diffusion)
- ✅ 视频自动合成 (FFmpeg)
- ✅ 字幕自动生成
- ✅ BGM背景音乐

## 系统要求

### 必需
- Go 1.21+
- FFmpeg
- OpenAI API Key
- Stable Diffusion WebUI (本地部署) 或 Replicate API

### 推荐硬件
- CPU: 4核心+
- 内存: 8GB+
- GPU: NVIDIA GPU (RTX 3060或更高,用于本地SD)
- 存储: 10GB+ 可用空间

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/Jancd/1504.git
cd 1504
```

### 2. 环境设置

```bash
# 运行设置脚本
make setup

# 或手动执行
chmod +x scripts/setup.sh
./scripts/setup.sh
```

### 3. 配置API密钥

编辑 `configs/config.yaml`:

```yaml
openai:
  api_key: "your-openai-api-key"

stable_diffusion:
  type: "local"  # 或 "replicate"
  local:
    api_url: "http://127.0.0.1:7860"
```

或通过环境变量:

```bash
export OPENAI_API_KEY="your-api-key"
```

### 4. 启动Stable Diffusion (本地部署)

```bash
# 如果你有stable-diffusion-webui
cd /path/to/stable-diffusion-webui
./webui.sh --api --listen
```

### 5. 运行服务

```bash
# 开发模式
make dev

# 或编译后运行
make build
make run
```

服务将在 `http://localhost:8080` 启动

## API使用

### 创建视频生成任务

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:现代都市街道,傍晚。小明独自走在回家的路上,心里想着今天发生的事情。突然,他看到了小红站在路边等他。小明(惊讶):你怎么在这里?小红(微笑):我在等你啊。",
    "options": {
      "style": "anime",
      "duration_target": 30,
      "aspect_ratio": "16:9",
      "bgm": "default.mp3"
    }
  }'
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "uuid-1234",
    "status": "processing",
    "estimated_time": 300
  },
  "timestamp": "2025-10-26T14:00:00Z"
}
```

### 查询任务状态

```bash
curl http://localhost:8080/api/tasks/{task_id}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "uuid-1234",
    "status": "processing",
    "progress": 45,
    "current_step": "generate_images",
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
    ]
  }
}
```

### 下载视频

```bash
curl http://localhost:8080/api/download/{task_id} -o output.mp4
```

### 列出所有任务

```bash
curl http://localhost:8080/api/tasks
```

### 删除任务

```bash
curl -X DELETE http://localhost:8080/api/tasks/{task_id}
```

## 项目结构

```
.
├── cmd/
│   └── server/          # 主程序入口
├── internal/
│   ├── client/          # 外部API客户端
│   ├── handler/         # HTTP处理器
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑服务
│   └── task/            # 任务管理
├── pkg/
│   ├── config/          # 配置管理
│   ├── logger/          # 日志
│   ├── ffmpeg/          # FFmpeg工具
│   └── utils/           # 工具函数
├── configs/
│   └── config.yaml      # 配置文件
├── data/
│   ├── uploads/         # 上传文件
│   ├── projects/        # 项目数据
│   └── assets/          # 资源文件
│       ├── bgm/         # 背景音乐
│       └── fonts/       # 字体
└── scripts/             # 脚本文件
```

## 配置说明

### 服务器配置

```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  mode: "debug"  # debug, release
```

### OpenAI配置

```yaml
openai:
  api_key: "your-api-key"
  model: "gpt-4"
  timeout: 300
```

### Stable Diffusion配置

```yaml
stable_diffusion:
  type: "local"  # local 或 replicate
  local:
    api_url: "http://127.0.0.1:7860"
    timeout: 300
```

### 视频配置

```yaml
video:
  default_bgm: "default.mp3"
  resolution: "1920x1080"  # 1920x1080, 1280x720, 3840x2160
  fps: 30
  max_duration: 120
```

### 限制配置

```yaml
limits:
  max_concurrent_tasks: 1
  max_shots_per_video: 20
  max_text_length: 2000
```

## 常见问题

### 1. Stable Diffusion连接失败

确保SD WebUI已启动且开启了API模式:
```bash
./webui.sh --api --listen
```

检查SD服务健康状态:
```bash
curl http://localhost:7860/sdapi/v1/sd-models
```

### 2. OpenAI API调用失败

- 检查API Key是否正确
- 确认账户有足够的余额
- 检查网络连接

### 3. FFmpeg找不到

安装FFmpeg:
```bash
# macOS
brew install ffmpeg

# Ubuntu
sudo apt install ffmpeg
```

### 4. 生成的图片不一致

- 在Prompt中强化角色特征描述
- 使用更详细的场景描述
- 调整Stable Diffusion的参数

### 5. 视频渲染失败

- 检查所有图片是否成功生成
- 确认FFmpeg正确安装
- 查看日志文件获取详细错误信息

## 开发指南

### Make命令

```bash
make help          # 显示帮助
make setup         # 初始化环境
make build         # 编译项目
make run           # 运行服务
make dev           # 开发模式
make test          # 运行测试
make clean         # 清理文件
make check         # 检查依赖
make lint          # 代码检查
```

### 日志查看

日志输出到stdout,可以重定向到文件:

```bash
make run > logs/app.log 2>&1
```

或在配置中设置文件输出:

```yaml
log:
  output: "file"
  file_path: "./logs/app.log"
```

### 调试模式

设置日志级别为debug:

```yaml
log:
  level: "debug"
```

## 性能优化建议

### 1. 图像生成优化

- 使用更快的采样器 (如DPM++ 2M Karras)
- 减少生成步数 (20-25步用于预览)
- 使用较小的图像分辨率进行测试

### 2. 降低API成本

- 使用GPT-3.5-turbo代替GPT-4 (成本降低10倍)
- 优化Prompt减少token使用
- 缓存常用的解析结果

### 3. 并发处理

当前MVP版本限制为单任务处理,生产环境可以:
- 使用任务队列支持多任务
- 实现GPU任务池
- 添加任务优先级

## 限制

当前MVP版本的限制:

- 仅支持单任务处理
- 不支持用户认证
- 不支持项目持久化
- 固定使用日系漫画风格
- 最大20个镜头/视频
- 最大120秒视频时长

## 路线图

- [ ] 添加Web前端界面
- [ ] 支持多种画风选择
- [ ] 实现TTS语音合成
- [ ] 角色一致性优化
- [ ] 支持手动编辑分镜
- [ ] 批量生成功能
- [ ] 用户系统
- [ ] 项目管理功能

## 贡献

欢迎提交Issue和Pull Request!

## 许可证

MIT License

## 联系方式

- 项目主页: https://github.com/Jancd/1504
- 问题反馈: https://github.com/Jancd/1504/issues

---

**版本**: v1.0.0 (MVP)
**更新日期**: 2025-10-26
