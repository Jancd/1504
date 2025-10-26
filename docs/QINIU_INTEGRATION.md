# 七牛云文生视频集成指南

## 概述

本项目已集成七牛云文生视频API,可以直接从文本生成视频,无需本地GPU和Stable Diffusion。

## 主要变更

### 1. 新增七牛云客户端

创建了 `internal/client/qiniu_video_client.go`,实现了:
- 视频生成API调用
- 任务状态查询
- 视频下载
- 健康检查

### 2. 双模式支持

系统现在支持两种视频生成模式:

#### 模式A: 七牛云直接生成(推荐)
```
文本 → GPT解析 → GPT分镜 → 七牛云生成视频 → 完成
```
- **优点**: 简单快速,无需本地GPU
- **处理时间**: 约5-15分钟
- **成本**: 根据七牛云定价

#### 模式B: SD图像+FFmpeg渲染(原有模式)
```
文本 → GPT解析 → GPT分镜 → SD生成图片 → FFmpeg渲染 → 完成
```
- **优点**: 完全可控,自定义程度高
- **处理时间**: 约10-30分钟
- **成本**: 本地运行免费(需GPU)

## 配置说明

### 配置文件 (configs/config.yaml)

```yaml
video_generation:
  type: "qiniu"  # 选择模式: qiniu, local_sd

  # 七牛云配置
  qiniu:
    api_url: "https://api.qiniu.com/v1/video/generate"
    api_key: "your-qiniu-api-key"
    timeout: 600
    max_wait_time: 600  # 最大等待时间(秒)

  # 本地SD配置(备选)
  local_sd:
    api_url: "http://127.0.0.1:7860"
    timeout: 300
```

### 环境变量

```bash
export QINIU_API_KEY="your-qiniu-api-key"
```

## 使用步骤

### 1. 配置API密钥

有两种方式配置:

**方式1**: 直接在配置文件中
```yaml
qiniu:
  api_key: "your-actual-api-key"
```

**方式2**: 使用环境变量(推荐)
```bash
export QINIU_API_KEY="your-api-key"
```

然后配置文件中使用:
```yaml
qiniu:
  api_key: "${QINIU_API_KEY}"
```

### 2. 启动服务

```bash
# 开发模式
make dev

# 或编译运行
make build
./bin/video-generator
```

服务启动时会显示:
```
Qiniu Video client initialized api_url=https://api.qiniu.com/v1/video/generate
Qiniu Video Service initialized
```

### 3. 测试API

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:春天的校园,樱花树下。小樱独自走着,突然遇到了她心仪的学长。",
    "options": {
      "duration_target": 30
    }
  }'
```

### 4. 查看进度

```bash
# 获取返回的task_id
TASK_ID="your-task-id"

# 查询状态
curl http://localhost:8080/api/tasks/$TASK_ID

# 下载视频(完成后)
curl http://localhost:8080/api/download/$TASK_ID -o output.mp4
```

## API接口规范

### 七牛云视频生成API (需要根据实际文档调整)

本项目假设七牛云API遵循以下规范:

#### 创建视频生成任务

```http
POST /v1/video/generate
Authorization: Bearer <api_key>
Content-Type: application/json

{
  "prompt": "视频描述...",
  "duration": 30,
  "style": "anime"
}
```

响应:
```json
{
  "task_id": "xxx-xxx-xxx",
  "status": "processing"
}
```

#### 查询任务状态

```http
GET /v1/video/generate/{task_id}
Authorization: Bearer <api_key>
```

响应:
```json
{
  "task_id": "xxx-xxx-xxx",
  "status": "completed",
  "progress": 100,
  "video_url": "https://..."
}
```

## 重要提示

⚠️ **API端点配置**

由于我无法访问七牛云的实际API文档,配置中的API端点 `https://api.qiniu.com/v1/video/generate` 是示例地址。

**请根据实际的七牛云文生视频API文档更新以下内容:**

1. **API端点URL** (`api_url`)
2. **认证方式** (目前使用Bearer Token)
3. **请求/响应格式** (可能需要调整 `qiniu_video_client.go`)

### 如何获取正确的API信息

1. 访问: https://developer.qiniu.com/aitokenapi/13083/video-generate-api
2. 查看完整的API文档
3. 获取:
   - 正确的API端点
   - 认证方式
   - 请求参数格式
   - 响应数据结构
4. 根据实际文档修改 `internal/client/qiniu_video_client.go`

## 切换模式

### 切换到七牛云模式

```yaml
video_generation:
  type: "qiniu"
```

### 切换回本地SD模式

```yaml
video_generation:
  type: "local_sd"
```

然后重启服务。

## 工作流程对比

### 七牛云模式

```
1. 用户提交文本
   ↓
2. OpenAI解析剧本 (5秒)
   ↓
3. OpenAI生成分镜 (8秒)
   ↓
4. 七牛云生成视频 (5-15分钟)
   ├─ 创建任务
   ├─ 等待完成 (轮询状态)
   └─ 下载视频
   ↓
5. 保存到本地
   ↓
6. 返回结果
```

### 本地SD模式

```
1. 用户提交文本
   ↓
2. OpenAI解析剧本 (5秒)
   ↓
3. OpenAI生成分镜 (8秒)
   ↓
4. SD生成图片 (10-20分钟)
   ├─ 串行生成每个镜头
   └─ 保存图片
   ↓
5. FFmpeg渲染视频 (30秒)
   ├─ 合成图片序列
   ├─ 添加BGM
   └─ 烧录字幕
   ↓
6. 返回结果
```

## 优势对比

| 特性 | 七牛云模式 | 本地SD模式 |
|------|-----------|-----------|
| GPU要求 | ❌ 不需要 | ✅ 需要 (RTX 3060+) |
| 处理速度 | ⭐⭐⭐⭐ 较快 | ⭐⭐⭐ 中等 |
| 成本 | 按使用付费 | 本地免费 |
| 自定义程度 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 稳定性 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| 部署难度 | ⭐⭐ 简单 | ⭐⭐⭐⭐ 复杂 |

## 故障排查

### 1. 七牛云API连接失败

**症状**: "Qiniu Video API health check failed"

**解决方案**:
- 检查API端点是否正确
- 确认API key有效
- 检查网络连接
- 查看七牛云服务状态

### 2. 视频生成超时

**症状**: "timeout waiting for video generation"

**解决方案**:
```yaml
qiniu:
  max_wait_time: 1200  # 增加到20分钟
```

### 3. 认证失败

**症状**: "status 401" 或 "unauthorized"

**解决方案**:
- 确认API key正确
- 检查认证方式(Bearer Token vs API Key)
- 查看七牛云账户状态

## 代码示例

### 直接调用七牛云客户端

```go
package main

import (
    "context"
    "github.com/Jancd/1504/internal/client"
)

func main() {
    // 创建客户端
    qiniuClient := client.NewQiniuVideoClient(
        "https://api.qiniu.com/v1/video/generate",
        "your-api-key",
        600,
    )

    // 生成视频
    result, err := qiniuClient.GenerateVideo(
        context.Background(),
        "一个关于春天的故事...",
        30, // 30秒
    )

    if err != nil {
        panic(err)
    }

    // 等待完成
    result, err = qiniuClient.WaitForCompletion(
        context.Background(),
        result.TaskID,
        10 * time.Minute,
    )

    if err != nil {
        panic(err)
    }

    // 下载视频
    videoData, err := qiniuClient.DownloadVideo(
        context.Background(),
        result.VideoURL,
    )

    // 保存视频
    os.WriteFile("output.mp4", videoData, 0644)
}
```

## 未来改进

- [ ] 支持更多七牛云参数配置
- [ ] 添加视频质量选项
- [ ] 支持自定义视频风格
- [ ] 实现视频编辑功能
- [ ] 添加成本统计

## 相关链接

- 七牛云文档: https://developer.qiniu.com/aitokenapi/13083/video-generate-api
- 项目主页: https://github.com/Jancd/1504
- Issues: https://github.com/Jancd/1504/issues

---

**版本**: v1.1.0
**更新日期**: 2025-10-26
