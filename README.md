# 文生漫画视频工具 - MVP版本

> 基于文本生成漫画风格视频的AI工具，支持七牛云和本地Stable Diffusion两种模式

## 项目概述

这是一个将小说文本转换为漫画风格视频的工具，使用AI技术自动完成剧本解析、分镜生成和视频制作的全流程。

### 核心特性

- ✅ **智能剧本解析** - 使用OpenAI/DeepSeek自动分析文本结构
- ✅ **自动分镜生成** - AI生成专业的视频分镜脚本
- ✅ **双模式支持** - 七牛云API或本地Stable Diffusion
- ✅ **异步任务处理** - 后台处理，实时状态查询
- ✅ **RESTful API** - 简单易用的HTTP接口

### 技术栈

- **后端**: Go 1.21+, Gin
- **AI服务**: OpenAI/DeepSeek (剧本解析)
- **视频生成**: 七牛云API / Stable Diffusion + FFmpeg
- **配置管理**: Viper
- **日志**: Zap

## 快速开始

### 前置要求

- Go 1.21+
- OpenAI API Key (用于剧本解析)
- 七牛云API Key (推荐) 或本地GPU + Stable Diffusion

### 5分钟快速启动

```bash
# 1. 克隆项目
git clone https://github.com/Jancd/1504.git
cd 1504

# 2. 配置API密钥
# 编辑 configs/config.yaml，填入你的API密钥

# 3. 安装依赖并运行
make dev
```

完整安装指南请查看 [快速开始文档](docs/QUICKSTART.md)

## 使用示例

### 创建视频生成任务

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:樱花飘落的校园。小樱走在路上,遇到了心仪的学长。两人的目光相遇,时间仿佛静止了。",
    "options": {
      "style": "anime",
      "duration_target": 30
    }
  }'
```

### 查询任务状态

```bash
curl http://localhost:8080/api/tasks/{task_id}
```

### 下载视频

```bash
curl http://localhost:8080/api/download/{task_id} -o output.mp4
```

## 两种模式对比

### 七牛云模式 (推荐)

```yaml
video_generation:
  type: "qiniu"
```

**优点:**
- ❌ 无需本地GPU
- ⚡ 处理速度快 (5-15分钟)
- ☁️ 云端运行，稳定可靠
- 📦 部署简单

**适合:** 没有GPU或希望快速开始的用户

### 本地SD模式

```yaml
video_generation:
  type: "local_sd"
```

**优点:**
- 🎨 完全可控，自定义程度高
- 💰 本地运行免费
- 🔒 数据完全私密

**适合:** 有GPU且需要高度自定义的用户

详细对比请查看 [七牛云集成文档](docs/README_QINIU.md)

## 文档导航

### 新手入门
- [快速开始](docs/QUICKSTART.md) - 5分钟快速上手
- [七牛云版本说明](docs/README_QINIU.md) - 无需GPU的快速方案
- [MVP版本文档](docs/README_MVP.md) - 完整功能说明

### 开发者文档
- [架构设计文档](docs/架构设计文档.md) - 系统架构和技术选型
- [七牛云集成指南](docs/QINIU_INTEGRATION.md) - 七牛云API集成详解
- [MVP开发计划](docs/MVP开发计划.md) - 项目开发规划
- [项目统计](docs/PROJECT_SUMMARY.md) - 代码统计和文件结构

### 原始设计
- [产品设计文档](docs/文生漫画视频工具-产品设计文档.md) - 最初的产品设计

## 系统架构

```
┌─────────────┐
│  用户输入    │ 小说文本
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ OpenAI GPT  │ 剧本解析 + 分镜生成
└──────┬──────┘
       │
       ├─────────────────┬─────────────────┐
       │                 │                 │
       ▼                 ▼                 ▼
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│  七牛云API  │   │   SD图像    │   │  FFmpeg     │
│  生成视频    │   │   生成      │   │  渲染视频   │
└──────┬──────┘   └──────┬──────┘   └──────┬──────┘
       │                 │                 │
       └────────┬────────┴────────┬────────┘
                │                 │
                ▼                 ▼
         ┌─────────────────────────────┐
         │      下载并返回视频          │
         └─────────────────────────────┘
```

## API接口

### 创建任务
- **POST** `/api/generate` - 创建视频生成任务

### 查询任务
- **GET** `/api/tasks/:task_id` - 查询任务状态
- **GET** `/api/tasks` - 列出所有任务

### 下载和管理
- **GET** `/api/download/:task_id` - 下载生成的视频
- **DELETE** `/api/tasks/:task_id` - 删除任务

### 健康检查
- **GET** `/health` - 服务健康状态

完整API文档请查看 [MVP版本文档](docs/README_MVP.md#api接口)

## 配置说明

主要配置项在 `configs/config.yaml`:

```yaml
# OpenAI配置
openai:
  api_key: "your-api-key"
  model: "deepseek-v3.1"
  base_url: "https://openai.qiniu.com"

# 视频生成模式
video_generation:
  type: "qiniu"  # qiniu 或 local_sd

  # 七牛云配置
  qiniu:
    api_url: "https://api.qiniu.com/v1/video/generate"
    api_key: "your-qiniu-api-key"
    timeout: 600
    max_wait_time: 600

# 视频参数
video:
  resolution: "1920x1080"
  fps: 30
  quality: "high"
  max_duration: 120

# 限制
limits:
  max_concurrent_tasks: 1
  max_shots_per_video: 20
  max_text_length: 2000
```

## 常见问题

### Q: 如何获取七牛云API Key?
A: 访问 [七牛云文生视频API文档](https://developer.qiniu.com/aitokenapi/13083/video-generate-api) 申请API访问权限。

### Q: 七牛云模式还需要FFmpeg吗?
A: 不需要。七牛云直接生成完整视频。但如果切换到本地SD模式，仍需要FFmpeg。

### Q: 可以同时使用两种模式吗?
A: 可以通过修改配置文件切换，但同一时间只能使用一种模式。

### Q: 生成一个视频需要多长时间?
A: 七牛云模式: 5-15分钟，本地SD模式: 10-30分钟，具体取决于视频长度和镜头数量。

更多问题请查看 [七牛云版本说明](docs/README_QINIU.md#常见问题)

## 项目结构

```
1504/
├── cmd/                    # 应用入口
│   └── server/            # 服务器主程序
├── internal/              # 内部包
│   ├── client/           # 外部API客户端
│   ├── handler/          # HTTP处理器
│   ├── model/            # 数据模型
│   ├── service/          # 业务逻辑
│   └── task/             # 任务管理
├── pkg/                   # 公共包
│   ├── config/           # 配置管理
│   ├── logger/           # 日志
│   ├── ffmpeg/           # FFmpeg封装
│   └── utils/            # 工具函数
├── configs/               # 配置文件
├── data/                  # 数据目录
├── scripts/               # 脚本
└── docs/                  # 文档
```

## 性能指标

| 指标 | 七牛云模式 | 本地SD模式 |
|------|-----------|-----------|
| 处理时间 | 5-15分钟 | 10-30分钟 |
| GPU要求 | ❌ 不需要 | ✅ RTX 3060+ |
| 稳定性 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| 可扩展性 | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| 成本 | 按量付费 | 固定成本 |

## 开发计划

当前版本: v1.1.0 (七牛云集成版)

### 已完成
- ✅ MVP核心功能
- ✅ 七牛云API集成
- ✅ 双模式支持
- ✅ 完整文档

### 计划中
- 🔲 更多视频风格支持
- 🔲 批量处理
- 🔲 Web管理界面
- 🔲 Docker部署

详细开发计划请查看 [MVP开发计划](docs/MVP开发计划.md)

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License

## 相关链接

- [七牛云文生视频API](https://developer.qiniu.com/aitokenapi/13083/video-generate-api)
- [OpenAI API](https://platform.openai.com/)
- [Stable Diffusion](https://github.com/AUTOMATIC1111/stable-diffusion-webui)
- [FFmpeg](https://ffmpeg.org/)

## 联系方式

- GitHub: https://github.com/Jancd/1504
- Issues: https://github.com/Jancd/1504/issues

---

**版本**: v1.1.0
**更新日期**: 2025-10-26
