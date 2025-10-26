# 快速开始指南

## 5分钟快速上手

### 前置条件检查

```bash
# 1. 检查Go版本 (需要 1.21+)
go version

# 2. 检查FFmpeg
ffmpeg -version

# 3. 准备OpenAI API Key
export OPENAI_API_KEY="sk-your-api-key"
```

### 快速启动步骤

#### 步骤1: 设置环境

```bash
# 克隆项目
git clone https://github.com/Jancd/1504.git
cd 1504

# 运行设置脚本
make setup
```

#### 步骤2: 配置API密钥

编辑 `configs/config.yaml`:

```yaml
openai:
  api_key: "sk-your-openai-api-key"
```

#### 步骤3: 启动Stable Diffusion (可选 - 如果你有本地GPU)

```bash
# 假设你已安装 stable-diffusion-webui
cd /path/to/stable-diffusion-webui
./webui.sh --api --listen
```

**如果没有本地GPU**: 暂时跳过此步骤,后续可以配置使用Replicate API

#### 步骤4: 启动服务

```bash
# 开发模式(推荐)
make dev

# 或编译后运行
make build
make run
```

服务将在 `http://localhost:8080` 启动

#### 步骤5: 测试API

打开新终端,测试健康检查:

```bash
curl http://localhost:8080/health
```

预期响应:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "time": "2025-10-26T14:00:00Z"
}
```

### 生成你的第一个视频

#### 方法1: 使用curl

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:宁静的校园,清晨。小樱独自走在樱花树下,花瓣随风飘落。她停下脚步,看着远处的教学楼。小樱(内心独白):今天是新学期的第一天,会发生什么有趣的事呢?突然,一个男生从她身边跑过。男生(气喘吁吁):对不起!小樱转过头,两人的目光相遇。",
    "options": {
      "duration_target": 30
    }
  }'
```

#### 方法2: 使用提供的测试脚本

```bash
# 创建测试脚本
cat > test.sh << 'EOF'
#!/bin/bash

# 创建任务
RESPONSE=$(curl -s -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:现代都市街道,傍晚。小明独自走在回家的路上。",
    "options": {"duration_target": 20}
  }')

TASK_ID=$(echo $RESPONSE | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
echo "任务已创建: $TASK_ID"

# 轮询状态
while true; do
  STATUS=$(curl -s "http://localhost:8080/api/tasks/$TASK_ID")
  STATE=$(echo $STATUS | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
  PROGRESS=$(echo $STATUS | grep -o '"progress":[0-9]*' | cut -d':' -f2)

  echo "状态: $STATE, 进度: $PROGRESS%"

  if [ "$STATE" == "completed" ]; then
    echo "视频生成完成!"
    curl -o "output_$TASK_ID.mp4" "http://localhost:8080/api/download/$TASK_ID"
    echo "视频已下载: output_$TASK_ID.mp4"
    break
  elif [ "$STATE" == "failed" ]; then
    echo "任务失败!"
    echo $STATUS
    break
  fi

  sleep 5
done
EOF

chmod +x test.sh
./test.sh
```

### 查看任务状态

```bash
# 获取任务ID (从上面的响应中)
TASK_ID="your-task-id"

# 查询状态
curl http://localhost:8080/api/tasks/$TASK_ID

# 列出所有任务
curl http://localhost:8080/api/tasks
```

### 下载完成的视频

```bash
curl http://localhost:8080/api/download/$TASK_ID -o output.mp4

# 播放视频
open output.mp4  # macOS
# 或
xdg-open output.mp4  # Linux
```

## 常见问题快速解决

### 问题1: 连接不到Stable Diffusion

**症状**: 日志显示 "Stable Diffusion API health check failed"

**解决方案**:
```bash
# 确保SD已启动
cd /path/to/stable-diffusion-webui
./webui.sh --api --listen

# 测试连接
curl http://localhost:7860/sdapi/v1/sd-models
```

### 问题2: OpenAI API调用失败

**症状**: "openai api call failed"

**解决方案**:
```bash
# 检查API Key
echo $OPENAI_API_KEY

# 或在配置文件中设置
vim configs/config.yaml
# 修改 openai.api_key
```

### 问题3: FFmpeg找不到

**症状**: "ffmpeg not found"

**解决方案**:
```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# 验证安装
ffmpeg -version
```

### 问题4: 端口已被占用

**症状**: "bind: address already in use"

**解决方案**:
```bash
# 修改配置中的端口
vim configs/config.yaml
# 修改 server.port 为其他值,如 "8081"
```

## 性能调优Tips

### 1. 加速图像生成

编辑 `internal/client/sd_client.go`:
```go
Steps: 20,  // 从30改为20,速度提升50%
```

### 2. 使用更便宜的模型

编辑 `configs/config.yaml`:
```yaml
openai:
  model: "gpt-3.5-turbo"  # 代替gpt-4,成本降低10倍
```

### 3. 减少生成的镜头数

编辑 `configs/config.yaml`:
```yaml
limits:
  max_shots_per_video: 10  # 从20改为10
```

## 下一步

1. 阅读完整的 [README_MVP.md](./README_MVP.md)
2. 查看 [MVP开发计划.md](./MVP开发计划.md) 了解架构设计
3. 尝试调整配置优化生成质量
4. 提交Issue报告问题或建议

## 获取帮助

- 查看日志: 服务运行时会输出详细日志
- 使用debug模式: `log.level: "debug"`
- 提交Issue: https://github.com/Jancd/1504/issues

---

**祝你使用愉快!** 🎉
