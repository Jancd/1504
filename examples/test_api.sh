#!/bin/bash

# 文生漫画视频工具 - API测试脚本

set -e

API_URL="http://localhost:8080"
SAMPLE_TEXT="场景:宁静的高中校园,春天的清晨。小樱独自走在樱花树下,心里想着新学期的事情。突然,一个男生从她身边跑过。男生(气喘吁吁):对不起!要迟到了!小樱(惊讶):啊!两人的目光相遇了。"

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================"
echo "   API测试脚本"
echo -e "======================================${NC}"
echo ""

# 1. 健康检查
echo -e "${YELLOW}1. 健康检查...${NC}"
HEALTH=$(curl -s "$API_URL/health")
if echo "$HEALTH" | grep -q "ok"; then
    echo -e "${GREEN}✓ 服务正常运行${NC}"
    echo "$HEALTH" | jq '.'
else
    echo -e "${RED}✗ 服务未响应${NC}"
    exit 1
fi
echo ""

# 2. 创建生成任务
echo -e "${YELLOW}2. 创建视频生成任务...${NC}"
CREATE_RESPONSE=$(curl -s -X POST "$API_URL/api/generate" \
  -H "Content-Type: application/json" \
  -d "{
    \"text\": \"$SAMPLE_TEXT\",
    \"options\": {
      \"style\": \"anime\",
      \"duration_target\": 20,
      \"aspect_ratio\": \"16:9\",
      \"bgm\": \"default.mp3\"
    }
  }")

if echo "$CREATE_RESPONSE" | grep -q "task_id"; then
    TASK_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.task_id')
    echo -e "${GREEN}✓ 任务创建成功${NC}"
    echo "任务ID: $TASK_ID"
    echo "$CREATE_RESPONSE" | jq '.'
else
    echo -e "${RED}✗ 任务创建失败${NC}"
    echo "$CREATE_RESPONSE"
    exit 1
fi
echo ""

# 3. 轮询任务状态
echo -e "${YELLOW}3. 监控任务进度...${NC}"
echo "等待视频生成完成 (预计3-10分钟)..."
echo ""

START_TIME=$(date +%s)
while true; do
    STATUS_RESPONSE=$(curl -s "$API_URL/api/tasks/$TASK_ID")

    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.data.status')
    PROGRESS=$(echo "$STATUS_RESPONSE" | jq -r '.data.progress')
    CURRENT_STEP=$(echo "$STATUS_RESPONSE" | jq -r '.data.current_step')

    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))

    echo -ne "\r状态: $STATUS | 进度: $PROGRESS% | 当前步骤: $CURRENT_STEP | 已用时: ${ELAPSED}秒  "

    if [ "$STATUS" == "completed" ]; then
        echo ""
        echo -e "${GREEN}✓ 视频生成完成!${NC}"
        echo ""
        echo "任务详情:"
        echo "$STATUS_RESPONSE" | jq '.data'
        break
    elif [ "$STATUS" == "failed" ]; then
        echo ""
        echo -e "${RED}✗ 任务失败${NC}"
        echo "错误信息:"
        echo "$STATUS_RESPONSE" | jq '.data.error'
        exit 1
    fi

    sleep 5
done
echo ""

# 4. 下载视频
echo -e "${YELLOW}4. 下载视频...${NC}"
OUTPUT_FILE="output_${TASK_ID}.mp4"

curl -s "$API_URL/api/download/$TASK_ID" -o "$OUTPUT_FILE"

if [ -f "$OUTPUT_FILE" ]; then
    FILE_SIZE=$(ls -lh "$OUTPUT_FILE" | awk '{print $5}')
    echo -e "${GREEN}✓ 视频下载成功${NC}"
    echo "文件: $OUTPUT_FILE"
    echo "大小: $FILE_SIZE"
else
    echo -e "${RED}✗ 视频下载失败${NC}"
    exit 1
fi
echo ""

# 5. 列出所有任务
echo -e "${YELLOW}5. 列出所有任务...${NC}"
TASKS=$(curl -s "$API_URL/api/tasks")
TOTAL=$(echo "$TASKS" | jq -r '.data.total')
echo -e "${GREEN}总任务数: $TOTAL${NC}"
echo "$TASKS" | jq '.data.tasks[] | {task_id, status, progress}'
echo ""

# 完成
echo -e "${BLUE}======================================"
echo "   测试完成!"
echo -e "======================================${NC}"
echo ""
echo "生成的视频: $OUTPUT_FILE"
echo ""
echo "播放视频:"
echo "  macOS:  open $OUTPUT_FILE"
echo "  Linux:  xdg-open $OUTPUT_FILE"
echo ""
echo "清理任务:"
echo "  curl -X DELETE $API_URL/api/tasks/$TASK_ID"
echo ""
