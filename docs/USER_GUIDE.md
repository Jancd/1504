# 文生漫画视频工具 - 使用指南

## 📖 目录

- [快速开始](#快速开始)
- [API 使用说明](#api-使用说明)
- [配置说明](#配置说明)
- [常见问题](#常见问题)
- [最佳实践](#最佳实践)

---

## 快速开始

### 前置要求

- Go 1.21+ (如需自行编译)
- OpenAI API Key (用于文本解析和分镜生成)
- 七牛云 API Key (用于视频生成)

### 1. 配置 API 密钥

编辑 `configs/config.yaml` 文件，填入你的 API 密钥：

```yaml
openai:
  api_key: "your-openai-api-key"
  model: "deepseek-v3.1"
  base_url: "https://openai.qiniu.com/v1"

video_generation:
  type: "qiniu"
  qiniu:
    api_url: "https://openai.qiniu.com/v1/videos/generations"
    api_key: "your-qiniu-api-key"
    model: "veo-3.0-fast-generate-preview"
```

### 2. 启动服务

```bash
# 方式1: 使用 make 命令
make run

# 方式2: 直接运行编译好的二进制文件
./bin/video-generator

# 方式3: 使用 go run
go run cmd/server/main.go
```

服务启动后会监听在 `http://localhost:8080`

### 3. 检查服务状态

```bash
curl http://localhost:8080/health
```

响应示例：
```json
{
  "status": "ok",
  "time": "2025-10-26T16:00:00+08:00",
  "version": "1.0.0"
}
```

---

## API 使用说明

### 1. 创建视频生成任务

**接口**: `POST /api/generate`

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "场景:樱花飘落的校园。小樱走在路上,突然遇到了她心仪的学长。两人的目光在空中相遇,时间仿佛静止了。学长微笑着向她打招呼,小樱害羞地低下了头。",
    "options": {
      "style": "anime",
      "duration_target": 8
    }
  }'
```

**请求参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | 是 | 小说文本内容，最长 2000 字符 |
| options.style | string | 否 | 视频风格，默认 "anime" |
| options.duration_target | int | 否 | 目标时长（秒），建议设为 8 秒 |
| options.aspect_ratio | string | 否 | 画面比例，默认 "16:9" |

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "649e3603-4fb7-40ec-9b17-c68d1749250d",
    "status": "processing",
    "estimated_time": 300
  },
  "timestamp": "2025-10-26T16:00:03+08:00"
}
```

**重要提示**:
- ⚠️ 七牛云 Veo API 当前只支持 **8 秒**的视频，`duration_target` 建议设为 8
- 文本长度建议控制在 200-500 字符之间，以获得最佳效果
- 任务创建后会在后台异步处理，通过 `task_id` 查询进度

---

### 2. 查询任务状态

**接口**: `GET /api/tasks/{task_id}`

**请求示例**:

```bash
curl http://localhost:8080/api/tasks/649e3603-4fb7-40ec-9b17-c68d1749250d
```

**响应示例 (处理中)**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "649e3603-4fb7-40ec-9b17-c68d1749250d",
    "status": "processing",
    "progress": 50,
    "current_step": "generate_images",
    "steps": [
      {
        "name": "parse_script",
        "status": "completed",
        "duration": 7.57,
        "start_at": "2025-10-26T16:00:03+08:00",
        "end_at": "2025-10-26T16:00:10+08:00"
      },
      {
        "name": "generate_storyboard",
        "status": "completed",
        "duration": 14.60,
        "start_at": "2025-10-26T16:00:10+08:00",
        "end_at": "2025-10-26T16:00:25+08:00"
      },
      {
        "name": "generate_images",
        "status": "processing",
        "start_at": "2025-10-26T16:00:25+08:00"
      },
      {
        "name": "render_video",
        "status": "pending"
      }
    ],
    "created_at": "2025-10-26T16:00:03+08:00",
    "updated_at": "2025-10-26T16:00:25+08:00"
  }
}
```

**响应示例 (已完成)**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "649e3603-4fb7-40ec-9b17-c68d1749250d",
    "status": "completed",
    "progress": 100,
    "current_step": "render_video",
    "steps": [...],
    "result": {
      "video_path": "data/projects/649e3603-4fb7-40ec-9b17-c68d1749250d/output.mp4",
      "duration": 8,
      "resolution": "1920x1080",
      "file_size": 2599378,
      "shot_count": 4
    },
    "created_at": "2025-10-26T16:00:03+08:00",
    "updated_at": "2025-10-26T16:06:16+08:00"
  }
}
```

**任务状态说明**:

| 状态 | 说明 |
|------|------|
| processing | 处理中 |
| completed | 已完成 |
| failed | 失败 |

**处理步骤说明**:

| 步骤 | 说明 | 预计耗时 |
|------|------|----------|
| parse_script | 解析剧本，识别场景和角色 | 5-10 秒 |
| generate_storyboard | 生成分镜脚本 | 10-20 秒 |
| generate_images | 调用七牛云生成视频 | 4-6 分钟 |
| render_video | 保存视频文件 | 即时完成 |

---

### 3. 下载视频

**接口**: `GET /api/download/{task_id}`

**请求示例**:

```bash
# 下载视频文件
curl http://localhost:8080/api/download/649e3603-4fb7-40ec-9b17-c68d1749250d \
  -o my_video.mp4

# 或使用 wget
wget http://localhost:8080/api/download/649e3603-4fb7-40ec-9b17-c68d1749250d \
  -O my_video.mp4
```

**响应**: 返回 MP4 视频文件流

---

### 4. 查询所有任务

**接口**: `GET /api/tasks`

**请求示例**:

```bash
curl http://localhost:8080/api/tasks
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tasks": [
      {
        "task_id": "649e3603-4fb7-40ec-9b17-c68d1749250d",
        "status": "completed",
        "created_at": "2025-10-26T16:00:03+08:00"
      },
      {
        "task_id": "2fdc3a01-aeed-46ec-9bd5-968bda72e72e",
        "status": "failed",
        "created_at": "2025-10-26T15:58:06+08:00"
      }
    ],
    "total": 2
  }
}
```

---

### 5. 删除任务

**接口**: `DELETE /api/tasks/{task_id}`

**请求示例**:

```bash
curl -X DELETE http://localhost:8080/api/tasks/649e3603-4fb7-40ec-9b17-c68d1749250d
```

**响应示例**:

```json
{
  "code": 0,
  "message": "Task deleted successfully"
}
```

---

## 配置说明

### 完整配置文件示例

```yaml
# 服务器配置
server:
  port: "8080"
  host: "0.0.0.0"
  mode: "debug"  # debug, release

# 存储配置
storage:
  data_dir: "./data"
  max_upload_size: 10485760  # 10MB

# OpenAI 配置 (用于文本解析)
openai:
  api_key: "your-openai-api-key"
  model: "deepseek-v3.1"
  base_url: "https://openai.qiniu.com/v1"
  timeout: 300

# 视频生成配置
video_generation:
  type: "qiniu"  # qiniu 或 local_sd

  # 七牛云配置
  qiniu:
    api_url: "https://openai.qiniu.com/v1/videos/generations"
    api_key: "your-qiniu-api-key"
    model: "veo-3.0-fast-generate-preview"
    timeout: 600
    max_wait_time: 600  # 最大等待时间(秒)

# 视频输出配置
video:
  default_bgm: "default.mp3"
  resolution: "1920x1080"
  fps: 30
  quality: "high"
  max_duration: 120

# 系统限制
limits:
  max_concurrent_tasks: 1
  max_shots_per_video: 20
  max_text_length: 2000

# 日志配置
log:
  level: "info"  # debug, info, warn, error
  output: "stdout"
```

### 配置参数说明

#### 服务器配置 (server)

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| port | string | "8080" | 服务监听端口 |
| host | string | "0.0.0.0" | 服务监听地址 |
| mode | string | "debug" | 运行模式: debug, release |

#### OpenAI 配置 (openai)

| 参数 | 类型 | 说明 |
|------|------|------|
| api_key | string | OpenAI API 密钥 |
| model | string | 使用的模型，推荐 deepseek-v3.1 |
| base_url | string | API 基础 URL |
| timeout | int | 请求超时时间(秒) |

#### 七牛云配置 (qiniu)

| 参数 | 类型 | 说明 |
|------|------|------|
| api_url | string | 七牛云视频生成 API 地址 |
| api_key | string | 七牛云 API 密钥 |
| model | string | 使用的模型，如 veo-3.0-fast-generate-preview |
| timeout | int | 单次请求超时时间(秒) |
| max_wait_time | int | 等待视频生成完成的最大时间(秒) |

---

## 常见问题

### Q1: 视频生成失败，提示 "do not support durationSeconds != 8"

**原因**: 七牛云 Veo API 当前只支持 8 秒的视频生成。

**解决方案**:
- 将 `duration_target` 设为 8
- 服务已自动适配，实际发送给七牛云的时长固定为 8 秒

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "你的文本内容...",
    "options": {
      "duration_target": 8
    }
  }'
```

---

### Q2: 视频生成超时怎么办？

**原因**: 七牛云视频生成通常需要 4-6 分钟。

**解决方案**:
1. 检查任务状态，确认是否真的失败
2. 增加配置文件中的 `max_wait_time`

```yaml
qiniu:
  max_wait_time: 900  # 增加到 15 分钟
```

---

### Q3: 如何获取 API 密钥？

**OpenAI API Key**:
1. 访问 https://openai.qiniu.com
2. 注册并登录账号
3. 在控制台创建 API Key

**七牛云 API Key**:
1. 访问 https://portal.qiniu.com
2. 注册并登录
3. 在"密钥管理"中创建 API Key

---

### Q4: 视频生成时间大概多久？

**预计时间分布**:
- 剧本解析: 5-10 秒
- 分镜生成: 10-20 秒
- 视频生成: 4-6 分钟
- **总计**: 约 5-7 分钟

实际时间取决于:
- 文本复杂度
- 七牛云服务负载
- 网络状况

---

### Q5: 支持什么样的文本输入？

**最佳输入格式**:

```
场景: 描述场景环境和氛围
动作: 角色的行为和互动
对话: 角色的台词
情感: 角色的情绪变化
```

**示例**:

```
场景: 樱花飘落的校园小路。
小樱走在路上，突然看到心仪的学长。
两人的目光相遇，时间仿佛静止了。
学长微笑着打招呼："早上好！"
小樱害羞地低下头，脸颊泛红。
```

**建议**:
- 文本长度: 200-500 字符最佳
- 场景数量: 1-3 个场景
- 角色数量: 1-3 个主要角色
- 描述要具体形象，避免过于抽象

---

### Q6: 生成的视频可以编辑吗？

当前版本不支持视频编辑功能。生成的视频是最终成品。

**未来计划**:
- 支持视频片段合成
- 支持添加背景音乐
- 支持字幕编辑

---

### Q7: 如何查看详细日志？

**方法1**: 修改配置文件

```yaml
log:
  level: "debug"
```

**方法2**: 查看服务器日志

```bash
# 如果使用 nohup 启动
tail -f nohup.out

# 或查看 server.log
tail -f server.log
```

---

### Q8: 服务重启后之前的任务还在吗？

当前版本使用内存存储任务状态，服务重启后任务列表会清空。

但生成的视频文件仍然保存在 `data/projects/` 目录下，可以手动访问。

---

## 最佳实践

### 1. 文本编写技巧

**✅ 推荐写法**:

```
场景: 夕阳下的海滩，海浪拍打着岸边。
小美独自站在沙滩上，海风吹动她的长发。
她望着远处的落日，眼中闪烁着泪光。
突然，身后传来熟悉的脚步声。
她惊喜地转身，看到是失联多年的好友。
两人激动地拥抱在一起。
```

**❌ 不推荐写法**:

```
她很开心。他们见面了。
```

**要点**:
- 描述要具体、形象
- 包含场景、动作、情感
- 有明确的起承转合
- 避免过于抽象或简单

---

### 2. 参数设置建议

```json
{
  "text": "你的故事文本...",
  "options": {
    "style": "anime",           // 动漫风格最稳定
    "duration_target": 8,       // 固定为 8 秒
    "aspect_ratio": "16:9"      // 推荐 16:9 横屏
  }
}
```

---

### 3. 错误处理

**建议使用脚本轮询任务状态**:

```bash
#!/bin/bash
TASK_ID=$1

while true; do
  STATUS=$(curl -s http://localhost:8080/api/tasks/$TASK_ID | \
    jq -r '.data.status')

  echo "Status: $STATUS"

  if [ "$STATUS" = "completed" ]; then
    echo "Video generated successfully!"
    curl http://localhost:8080/api/download/$TASK_ID -o output.mp4
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "Generation failed!"
    break
  fi

  sleep 10
done
```

---

### 4. 性能优化建议

1. **控制并发**: 当前版本只支持单任务处理，避免同时提交多个任务
2. **合理设置超时**: 根据网络情况调整 `timeout` 和 `max_wait_time`
3. **定期清理**: 定期删除旧任务和视频文件以释放空间

```bash
# 清理 7 天前的任务
find data/projects -type d -mtime +7 -exec rm -rf {} \;
```

---

### 5. 监控和告警

**健康检查**:

```bash
# 添加到 crontab 定期检查
*/5 * * * * curl -f http://localhost:8080/health || systemctl restart video-generator
```

**磁盘空间监控**:

```bash
# 检查 data 目录大小
du -sh data/
```

---

## 示例代码

### Python 示例

```python
import requests
import time
import json

# 创建任务
def create_task(text):
    url = "http://localhost:8080/api/generate"
    payload = {
        "text": text,
        "options": {
            "style": "anime",
            "duration_target": 8
        }
    }

    response = requests.post(url, json=payload)
    data = response.json()
    return data['data']['task_id']

# 等待任务完成
def wait_for_completion(task_id, max_wait=600):
    url = f"http://localhost:8080/api/tasks/{task_id}"
    start_time = time.time()

    while time.time() - start_time < max_wait:
        response = requests.get(url)
        data = response.json()
        status = data['data']['status']

        print(f"Status: {status}")

        if status == 'completed':
            return data['data']
        elif status == 'failed':
            raise Exception(f"Task failed: {data['data'].get('error')}")

        time.sleep(10)

    raise Exception("Timeout waiting for task completion")

# 下载视频
def download_video(task_id, output_path):
    url = f"http://localhost:8080/api/download/{task_id}"
    response = requests.get(url)

    with open(output_path, 'wb') as f:
        f.write(response.content)

    print(f"Video saved to {output_path}")

# 使用示例
if __name__ == "__main__":
    text = """
    场景:樱花飘落的校园。
    小樱走在路上,突然遇到了她心仪的学长。
    两人的目光在空中相遇,时间仿佛静止了。
    学长微笑着向她打招呼,小樱害羞地低下了头。
    """

    # 创建任务
    task_id = create_task(text)
    print(f"Task created: {task_id}")

    # 等待完成
    result = wait_for_completion(task_id)
    print(f"Video info: {json.dumps(result['result'], indent=2)}")

    # 下载视频
    download_video(task_id, f"{task_id}.mp4")
```

---

### JavaScript/Node.js 示例

```javascript
const axios = require('axios');
const fs = require('fs');

const API_BASE = 'http://localhost:8080';

// 创建任务
async function createTask(text) {
  const response = await axios.post(`${API_BASE}/api/generate`, {
    text,
    options: {
      style: 'anime',
      duration_target: 8
    }
  });

  return response.data.data.task_id;
}

// 等待任务完成
async function waitForCompletion(taskId, maxWait = 600000) {
  const startTime = Date.now();

  while (Date.now() - startTime < maxWait) {
    const response = await axios.get(`${API_BASE}/api/tasks/${taskId}`);
    const { status } = response.data.data;

    console.log(`Status: ${status}`);

    if (status === 'completed') {
      return response.data.data;
    } else if (status === 'failed') {
      throw new Error(`Task failed: ${response.data.data.error}`);
    }

    await new Promise(resolve => setTimeout(resolve, 10000));
  }

  throw new Error('Timeout waiting for task completion');
}

// 下载视频
async function downloadVideo(taskId, outputPath) {
  const response = await axios.get(`${API_BASE}/api/download/${taskId}`, {
    responseType: 'stream'
  });

  const writer = fs.createWriteStream(outputPath);
  response.data.pipe(writer);

  return new Promise((resolve, reject) => {
    writer.on('finish', resolve);
    writer.on('error', reject);
  });
}

// 使用示例
(async () => {
  const text = `
    场景:樱花飘落的校园。
    小樱走在路上,突然遇到了她心仪的学长。
    两人的目光在空中相遇,时间仿佛静止了。
    学长微笑着向她打招呼,小樱害羞地低下了头。
  `;

  try {
    // 创建任务
    const taskId = await createTask(text);
    console.log(`Task created: ${taskId}`);

    // 等待完成
    const result = await waitForCompletion(taskId);
    console.log('Video info:', JSON.stringify(result.result, null, 2));

    // 下载视频
    await downloadVideo(taskId, `${taskId}.mp4`);
    console.log('Video downloaded successfully!');
  } catch (error) {
    console.error('Error:', error.message);
  }
})();
```

---

### Bash 脚本示例

```bash
#!/bin/bash

API_BASE="http://localhost:8080"

# 创建任务
create_task() {
  local text=$1

  curl -s -X POST "$API_BASE/api/generate" \
    -H "Content-Type: application/json" \
    -d "{\"text\": \"$text\", \"options\": {\"duration_target\": 8}}" \
    | jq -r '.data.task_id'
}

# 等待任务完成
wait_for_completion() {
  local task_id=$1
  local max_checks=${2:-60}

  for i in $(seq 1 $max_checks); do
    local status=$(curl -s "$API_BASE/api/tasks/$task_id" | jq -r '.data.status')
    echo "[$i/$max_checks] Status: $status"

    if [ "$status" = "completed" ]; then
      return 0
    elif [ "$status" = "failed" ]; then
      echo "Task failed!"
      return 1
    fi

    sleep 10
  done

  echo "Timeout!"
  return 1
}

# 下载视频
download_video() {
  local task_id=$1
  local output=${2:-"output.mp4"}

  curl -o "$output" "$API_BASE/api/download/$task_id"
  echo "Video saved to $output"
}

# 主函数
main() {
  local text="场景:樱花飘落的校园。小樱走在路上,突然遇到了她心仪的学长。两人的目光在空中相遇,时间仿佛静止了。学长微笑着向她打招呼,小樱害羞地低下了头。"

  # 创建任务
  local task_id=$(create_task "$text")
  echo "Task created: $task_id"

  # 等待完成
  if wait_for_completion "$task_id"; then
    # 下载视频
    download_video "$task_id" "${task_id}.mp4"
  fi
}

main
```

---

## 附录

### A. 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 404 | 任务不存在 |
| 500 | 服务器内部错误 |

### B. 支持的视频风格

当前版本主要支持 `anime` (动漫) 风格，未来计划支持:
- realistic (写实)
- cartoon (卡通)
- artistic (艺术)

### C. 文件目录结构

```
data/
├── projects/
│   └── {task_id}/
│       ├── output.mp4        # 最终视频
│       ├── script.txt        # 原始文本
│       ├── parsed.json       # 解析结果
│       └── storyboard.json   # 分镜脚本
├── uploads/                  # 上传文件(预留)
└── assets/                   # 资源文件
    ├── bgm/                  # 背景音乐
    └── fonts/                # 字体文件
```

---

## 技术支持

- **文档**: https://github.com/Jancd/1504/tree/main/docs
- **Issues**: https://github.com/Jancd/1504/issues
- **七牛云文档**: https://developer.qiniu.com/aitokenapi/13083/video-generate-api

---

**版本**: v1.0.0
**更新日期**: 2025-10-26
**作者**: Jancd
