# 文生视频工具 - 快速使用指南

> 5分钟上手，从文字生成漫画风格视频

---

## 🚀 一分钟快速开始

### 1. 配置密钥

编辑 `configs/config.yaml`，填入你的 API Key：

```yaml
openai:
  api_key: "sk-your-openai-key"    # 七牛云 OpenAI Key

video_generation:
  qiniu:
    api_key: "sk-your-qiniu-key"   # 七牛云视频 API Key
```

### 2. 启动服务

```bash
# 启动服务
./bin/video-generator

# 或使用 make
make run
```

### 3. 生成第一个视频

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "樱花飘落的校园，少女与学长的邂逅。两人目光相遇，时间仿佛静止。",
    "options": {
      "duration_target": 8
    }
  }'
```

**返回**：
```json
{
  "code": 0,
  "data": {
    "task_id": "649e3603-4fb7-40ec-9b17-c68d1749250d",
    "status": "processing"
  }
}
```

### 4. 查询进度

```bash
# 查询任务状态
curl http://localhost:8080/api/tasks/649e3603-4fb7-40ec-9b17-c68d1749250d
```

### 5. 下载视频

```bash
# 任务完成后下载
curl http://localhost:8080/api/download/649e3603-4fb7-40ec-9b17-c68d1749250d -o video.mp4
```

---

## 📖 核心功能

### 功能1：生成视频

**POST** `/api/generate`

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "你的故事文本...",
    "options": {
      "style": "anime",
      "duration_target": 8
    }
  }'
```

| 参数 | 说明 | 必填 | 默认值 |
|------|------|------|--------|
| text | 故事文本（最长2000字） | 是 | - |
| options.style | 视频风格 | 否 | anime |
| options.duration_target | 视频时长（秒） | 否 | 8 |

**⚠️ 重要提示**：
- 七牛云 Veo API 当前只支持 **8秒** 视频
- 建议文本长度 **200-500字**
- 生成时间约 **5-7分钟**

---

### 功能2：查询状态

**GET** `/api/tasks/{task_id}`

```bash
curl http://localhost:8080/api/tasks/{task_id}
```

**状态说明**：

| 状态 | 说明 | 下一步 |
|------|------|--------|
| processing | 处理中 | 继续等待 |
| completed | 已完成 | 下载视频 |
| failed | 失败 | 查看错误信息 |

**处理步骤**：

```
parse_script (7秒)
    ↓
generate_storyboard (15秒)
    ↓
generate_images (5-6分钟)
    ↓
render_video (即时)
    ↓
completed ✅
```

---

### 功能3：下载视频

**GET** `/api/download/{task_id}`

```bash
# 方式1：curl
curl http://localhost:8080/api/download/{task_id} -o my_video.mp4

# 方式2：wget
wget http://localhost:8080/api/download/{task_id} -O my_video.mp4

# 方式3：浏览器直接访问
http://localhost:8080/api/download/{task_id}
```

---

### 功能4：管理任务

```bash
# 列出所有任务
curl http://localhost:8080/api/tasks

# 删除任务
curl -X DELETE http://localhost:8080/api/tasks/{task_id}
```

---

## 💡 实用脚本

### 自动化脚本（一键生成并下载）

创建文件 `generate_video.sh`：

```bash
#!/bin/bash

# 配置
API_BASE="http://localhost:8080"
TEXT="$1"

if [ -z "$TEXT" ]; then
  echo "用法: ./generate_video.sh '你的故事文本'"
  exit 1
fi

echo "📝 创建任务..."
TASK_ID=$(curl -s -X POST "$API_BASE/api/generate" \
  -H "Content-Type: application/json" \
  -d "{\"text\": \"$TEXT\", \"options\": {\"duration_target\": 8}}" \
  | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['task_id'])")

echo "✅ 任务已创建: $TASK_ID"
echo "⏳ 等待生成中（预计5-7分钟）..."

# 等待完成
while true; do
  STATUS=$(curl -s "$API_BASE/api/tasks/$TASK_ID" \
    | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['status'])")

  if [ "$STATUS" = "completed" ]; then
    echo "✅ 生成完成！"
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "❌ 生成失败"
    curl -s "$API_BASE/api/tasks/$TASK_ID" | python3 -m json.tool
    exit 1
  fi

  echo "   状态: $STATUS"
  sleep 10
done

# 下载视频
OUTPUT="${TASK_ID}.mp4"
echo "📥 下载视频到 $OUTPUT..."
curl -s "$API_BASE/api/download/$TASK_ID" -o "$OUTPUT"

echo "🎉 完成！视频已保存为 $OUTPUT"

# 显示视频信息
echo ""
echo "视频信息:"
curl -s "$API_BASE/api/tasks/$TASK_ID" \
  | python3 -c "import sys, json; result=json.load(sys.stdin)['data']['result']; print(f\"时长: {result['duration']}秒\\n分辨率: {result['resolution']}\\n文件大小: {result['file_size']/1024/1024:.2f}MB\")"
```

**使用方法**：

```bash
chmod +x generate_video.sh

./generate_video.sh "樱花飘落的校园，少女与学长的邂逅。"
```

---

### Python 一键脚本

创建文件 `generate_video.py`：

```python
#!/usr/bin/env python3
import requests
import time
import sys

API_BASE = "http://localhost:8080"

def generate_video(text):
    """生成视频并自动下载"""

    # 1. 创建任务
    print(f"📝 创建任务...")
    response = requests.post(f"{API_BASE}/api/generate", json={
        "text": text,
        "options": {"duration_target": 8}
    })
    task_id = response.json()['data']['task_id']
    print(f"✅ 任务已创建: {task_id}")

    # 2. 等待完成
    print(f"⏳ 等待生成中（预计5-7分钟）...")
    while True:
        response = requests.get(f"{API_BASE}/api/tasks/{task_id}")
        data = response.json()['data']
        status = data['status']

        if status == 'completed':
            print("✅ 生成完成！")
            break
        elif status == 'failed':
            print(f"❌ 生成失败: {data.get('error')}")
            return None

        print(f"   状态: {status}")
        time.sleep(10)

    # 3. 下载视频
    output = f"{task_id}.mp4"
    print(f"📥 下载视频到 {output}...")
    response = requests.get(f"{API_BASE}/api/download/{task_id}")
    with open(output, 'wb') as f:
        f.write(response.content)

    # 4. 显示信息
    result = data['result']
    print(f"""
🎉 完成！视频已保存为 {output}

视频信息:
  时长: {result['duration']}秒
  分辨率: {result['resolution']}
  文件大小: {result['file_size']/1024/1024:.2f}MB
  镜头数: {result['shot_count']}
    """)

    return output

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("用法: python3 generate_video.py '你的故事文本'")
        sys.exit(1)

    text = sys.argv[1]
    generate_video(text)
```

**使用方法**：

```bash
chmod +x generate_video.py

python3 generate_video.py "樱花飘落的校园，少女与学长的邂逅。"
```

---

## 📝 文本编写技巧

### ✅ 好的文本示例

```
场景：夕阳西下的海边，海浪轻拍沙滩。

小美独自站在岸边，海风吹动她的长发。
她望着远方的落日，眼中闪烁着思念的泪光。

突然，身后传来熟悉的脚步声。
她惊喜地转过身，看到是失联多年的挚友小雪。

两人激动地拥抱在一起，泪水模糊了视线。
落日的余晖洒在两人身上，温暖而美好。
```

**特点**：
- ✅ 有明确的场景描述
- ✅ 包含角色动作和情感
- ✅ 有故事的起承转合
- ✅ 描述具体形象

---

### ❌ 不好的文本示例

```
她很高兴。他们见面了。很感动。
```

**问题**：
- ❌ 过于简短抽象
- ❌ 缺少场景描述
- ❌ 没有具体动作
- ❌ 情感表达单薄

---

### 📋 文本模板

**模板1：校园青春**
```
场景：[季节] + [地点] + [环境描述]
[主角]独自[动作]，[心理状态]。
突然，[事件发生]。
[主角反应]，[情感变化]。
[结果/结局]。
```

**示例**：
```
场景：春天的校园，樱花树下，花瓣随风飘落。
小樱独自走在小路上，心事重重。
突然，她看到心仪的学长迎面走来。
小樱脸颊泛红，害羞地低下了头。
学长温柔地微笑着打招呼，春风吹过，樱花纷飞。
```

---

**模板2：浪漫邂逅**
```
场景：[时间] + [地点] + [氛围]
[主角A] [状态/动作]。
[主角B] [出场方式]。
[两人互动]，[情感碰撞]。
[浪漫时刻]。
```

**示例**：
```
场景：黄昏时分，咖啡馆的露台，暖色灯光。
小林独自坐在角落，翻看着旧照片。
小雪推门而入，目光不经意间与小林相遇。
两人都愣住了，空气中弥漫着说不出的默契。
夕阳洒在桌上，照亮了彼此的笑容。
```

---

## ⚙️ 配置说明

### 最小配置（必需）

```yaml
openai:
  api_key: "sk-xxx"
  base_url: "https://openai.qiniu.com/v1"

video_generation:
  type: "qiniu"
  qiniu:
    api_url: "https://openai.qiniu.com/v1/videos/generations"
    api_key: "sk-xxx"
```

### 完整配置（可选）

```yaml
server:
  port: "8080"           # 服务端口
  mode: "debug"          # debug 或 release

openai:
  api_key: "sk-xxx"
  model: "deepseek-v3.1"
  base_url: "https://openai.qiniu.com/v1"
  timeout: 300

video_generation:
  type: "qiniu"
  qiniu:
    api_url: "https://openai.qiniu.com/v1/videos/generations"
    api_key: "sk-xxx"
    model: "veo-3.0-fast-generate-preview"
    timeout: 600
    max_wait_time: 600   # 最大等待时间(秒)

log:
  level: "info"          # debug, info, warn, error
```

---

## 🐛 常见问题

### Q1: 视频只能生成8秒吗？

**A**: 是的，七牛云 Veo API 当前限制为 8 秒。即使设置更长时间，实际也会调整为 8 秒。

---

### Q2: 生成时间为什么这么长？

**A**: 视频生成是复杂的 AI 计算过程：
- 文本解析：7秒
- 分镜生成：15秒
- AI视频生成：5-6分钟（七牛云处理）
- **总计**：约 5-7 分钟

---

### Q3: 如何加快生成速度？

**A**:
- ✅ 简化文本描述（200-300字最佳）
- ✅ 减少场景数量（1-2个场景）
- ✅ 确保网络稳定
- ❌ 不要并发提交多个任务

---

### Q4: 视频生成失败了怎么办？

**A**: 检查失败原因：

```bash
# 查看错误信息
curl http://localhost:8080/api/tasks/{task_id} | python3 -m json.tool
```

**常见错误**：
- `do not support durationSeconds != 8`：时长必须为8秒
- `API key invalid`：检查配置文件中的 API Key
- `timeout`：网络问题或七牛云服务繁忙

---

### Q5: 如何获取 API Key？

**七牛云 API Key**：
1. 访问 https://portal.qiniu.com
2. 注册并登录
3. 在"个人中心 → 密钥管理"创建 API Key
4. 将 Key 填入配置文件

---

### Q6: 可以批量生成吗？

**A**: 当前版本仅支持单任务处理。建议使用脚本依次提交：

```bash
# 批量生成脚本
for text in "故事1" "故事2" "故事3"; do
  ./generate_video.sh "$text"
  sleep 10  # 等待任务间隔
done
```

---

### Q7: 生成的视频在哪里？

**A**: 视频保存在：
```
data/projects/{task_id}/output.mp4
```

也可以通过 API 下载：
```bash
curl http://localhost:8080/api/download/{task_id} -o video.mp4
```

---

### Q8: 可以修改视频分辨率吗？

**A**: 当前固定为 1920x1080。未来版本会支持自定义分辨率。

---

## 📊 性能参考

### 测试环境
- 网络：100Mbps
- API：七牛云 Veo 3.0
- 文本：200字

### 测试结果

| 步骤 | 耗时 | 占比 |
|------|------|------|
| 剧本解析 | 7.57秒 | 2% |
| 分镜生成 | 14.60秒 | 4% |
| 视频生成 | 351秒 (5分51秒) | 93% |
| 视频保存 | <1秒 | <1% |
| **总计** | **6分14秒** | **100%** |

### 资源占用

| 资源 | 用量 |
|------|------|
| 视频文件 | 约 2.5MB (8秒) |
| 内存占用 | < 100MB |
| CPU占用 | < 5% (等待期间) |

---

## 🔧 高级用法

### 1. 自定义分镜

虽然系统会自动生成分镜，但你可以通过详细描述来引导：

```json
{
  "text": "第一幕：远景，樱花飘落的校园，小樱走在路上。\n第二幕：中景，小樱抬头看到学长。\n第三幕：特写，两人目光相遇。\n第四幕：近景，学长微笑打招呼。",
  "options": {
    "duration_target": 8
  }
}
```

---

### 2. 添加情感描述

增加情感关键词可以提升画面表现力：

```
小樱【害羞地】低下头，脸颊【泛起红晕】。
学长【温柔地】微笑，眼神【充满关怀】。
春风【轻柔地】吹过，樱花【缓缓】飘落。
```

---

### 3. 环境渲染

详细的环境描述有助于生成更好的画面：

```
时间：黄昏
天气：微风、晴朗
光线：夕阳余晖，暖色调
音效：风声、鸟鸣（AI会自动生成音频）
```

---

## 📞 技术支持

### 文档
- 完整文档：[USER_GUIDE.md](./USER_GUIDE.md)
- 快速开始：[QUICKSTART.md](./QUICKSTART.md)
- 七牛云集成：[README_QINIU.md](./README_QINIU.md)

### 问题反馈
- GitHub Issues: https://github.com/Jancd/1504/issues

### 相关链接
- 七牛云文档: https://developer.qiniu.com/aitokenapi/13083/video-generate-api
- 项目主页: https://github.com/Jancd/1504

---

## 📄 License

MIT License

---

**版本**: v1.0.0
**更新日期**: 2025-10-26
**维护者**: Jancd
