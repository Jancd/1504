# 文生漫画视频工具 - 七牛云版本

## 新特性 🎉

- ✅ **无需本地GPU** - 使用七牛云API直接生成视频
- ✅ **更快速** - 5-15分钟生成视频(vs. 原来的10-30分钟)
- ✅ **更简单** - 无需安装Stable Diffusion
- ✅ **更稳定** - 云端服务,7x24小时可用
- ✅ **双模式** - 支持七牛云和本地SD两种模式

## 快速开始(七牛云模式)

### 前置要求

- Go 1.21+
- OpenAI API Key (用于剧本解析)
- 七牛云API Key (用于视频生成)

### 1. 克隆项目

```bash
git clone https://github.com/Jancd/1504.git
cd 1504
```

### 2. 配置API密钥

编辑 `configs/config.yaml`:

```yaml
openai:
  api_key: "your-openai-api-key"
  base_url: "https://openai.qiniu.com"  # 如果使用七牛云OpenAI代理

video_generation:
  type: "qiniu"  # 使用七牛云模式
  qiniu:
    api_url: "https://api.qiniu.com/v1/video/generate"
    api_key: "your-qiniu-api-key"
```

**注意**: 七牛云的实际API端点请参考官方文档更新。

### 3. 运行服务

```bash
# 方式1: 开发模式
make dev

# 方式2: 编译运行
make build
./bin/video-generator
```

### 4. 生成视频

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:樱花飘落的校园。小樱走在路上,遇到了心仪的学长。两人的目光相遇,时间仿佛静止了。",
    "options": {
      "duration_target": 20
    }
  }'
```

响应:
```json
{
  "code": 0,
  "data": {
    "task_id": "xxx-xxx-xxx",
    "status": "processing"
  }
}
```

### 5. 查看进度

```bash
curl http://localhost:8080/api/tasks/{task_id}
```

### 6. 下载视频

```bash
curl http://localhost:8080/api/download/{task_id} -o output.mp4
```

## 配置说明

### 视频生成模式

系统支持两种模式,在 `configs/config.yaml` 中配置:

#### 模式1: 七牛云(推荐,无需GPU)

```yaml
video_generation:
  type: "qiniu"
  qiniu:
    api_url: "https://api.qiniu.com/v1/video/generate"
    api_key: "your-api-key"
    timeout: 600
    max_wait_time: 600
```

#### 模式2: 本地Stable Diffusion(需要GPU)

```yaml
video_generation:
  type: "local_sd"
  local_sd:
    api_url: "http://127.0.0.1:7860"
    timeout: 300
```

## 系统架构

### 七牛云模式

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
       ▼
┌─────────────┐
│  七牛云API  │ 直接生成视频
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  下载视频    │ 保存到本地
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  返回用户    │
└─────────────┘
```

## 成本对比

### 七牛云模式
- OpenAI API: ~¥0.5/视频
- 七牛云视频生成: 根据定价(请咨询七牛云)
- 服务器: 可使用低配置服务器

### 本地SD模式
- OpenAI API: ~¥0.5/视频
- 本地GPU: 免费(但需要硬件投资)
- 电费: ~¥0.2/视频

## 性能对比

| 指标 | 七牛云模式 | 本地SD模式 |
|------|-----------|-----------|
| 处理时间 | 5-15分钟 | 10-30分钟 |
| GPU要求 | ❌ 不需要 | ✅ RTX 3060+ |
| 稳定性 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| 可扩展性 | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| 成本 | 按量付费 | 固定成本 |

## API文档

### 创建生成任务

```http
POST /api/generate
Content-Type: application/json

{
  "text": "小说文本内容...",
  "options": {
    "style": "anime",
    "duration_target": 60,
    "aspect_ratio": "16:9"
  }
}
```

### 查询任务状态

```http
GET /api/tasks/{task_id}
```

响应:
```json
{
  "code": 0,
  "data": {
    "task_id": "xxx",
    "status": "processing",
    "progress": 45,
    "current_step": "generate_images",
    "steps": [...]
  }
}
```

### 下载视频

```http
GET /api/download/{task_id}
```

## 常见问题

### Q: 如何获取七牛云API Key?

A:
1. 注册七牛云账号
2. 访问文生视频API文档
3. 申请API访问权限
4. 获取API Key

### Q: 七牛云模式下是否还需要FFmpeg?

A: 不需要。七牛云直接生成完整视频,无需本地渲染。但如果要切换到本地SD模式,仍需要FFmpeg。

### Q: 可以同时使用两种模式吗?

A: 可以通过修改配置文件切换,但同一时间只能使用一种模式。

### Q: 七牛云生成的视频质量如何?

A: 七牛云使用专业的视频生成模型,质量稳定可靠。具体效果取决于输入的文本质量。

### Q: 如何调试七牛云API调用?

A:
1. 设置日志级别为debug
2. 查看日志输出
3. 使用curl直接测试API

```yaml
log:
  level: "debug"
```

## 故障排查

### 1. 七牛云API连接失败

```
错误: Qiniu Video API health check failed
```

**解决方案**:
- 检查API端点URL是否正确
- 确认API Key有效
- 测试网络连接
- 查看七牛云服务状态

### 2. 视频生成超时

```
错误: timeout waiting for video generation
```

**解决方案**:
```yaml
qiniu:
  max_wait_time: 1200  # 增加到20分钟
```

### 3. OpenAI调用失败

```
错误: openai api call failed
```

**解决方案**:
- 检查OpenAI API Key
- 确认base_url配置正确
- 检查账户余额

## 部署建议

### 开发环境
```bash
# 本地开发
make dev
```

### 生产环境
```bash
# Docker部署
docker build -t video-generator .
docker run -p 8080:8080 \
  -e OPENAI_API_KEY="your-key" \
  -e QINIU_API_KEY="your-key" \
  video-generator

# 或直接运行
make build
./bin/video-generator
```

## 环境变量

```bash
# OpenAI配置
export OPENAI_API_KEY="your-openai-key"

# 七牛云配置
export QINIU_API_KEY="your-qiniu-key"

# 服务器配置
export SERVER_PORT=8080
export SERVER_MODE=release
```

## 限制

当前版本限制:
- 单任务处理(串行)
- 最大视频时长: 120秒
- 最大文本长度: 2000字
- 最大镜头数: 20个

## 升级路径

从本地SD模式升级到七牛云模式:

1. 更新配置:
```yaml
video_generation:
  type: "qiniu"  # 改为qiniu
```

2. 添加七牛云配置
3. 重启服务
4. 测试接口

无需修改代码,系统会自动选择对应模式。

## 相关文档

- [七牛云集成指南](./QINIU_INTEGRATION.md)
- [快速开始](./QUICKSTART.md)
- [完整文档](./README_MVP.md)
- [开发计划](./MVP开发计划.md)

## 支持

- GitHub Issues: https://github.com/Jancd/1504/issues
- 七牛云文档: https://developer.qiniu.com/aitokenapi/13083/video-generate-api

---

**版本**: v1.1.0 (七牛云集成版)
**更新日期**: 2025-10-26
