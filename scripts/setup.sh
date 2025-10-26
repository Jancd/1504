#!/bin/bash

set -e

echo "======================================"
echo "   MVP视频生成器 - 环境设置脚本"
echo "======================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 打印成功消息
success() {
    echo -e "${GREEN}✓${NC} $1"
}

# 打印警告消息
warning() {
    echo -e "${YELLOW}!${NC} $1"
}

# 打印错误消息
error() {
    echo -e "${RED}✗${NC} $1"
}

echo "1. 检查系统依赖..."
echo "-----------------------------------"

# 检查Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    success "Go已安装: $GO_VERSION"
else
    error "Go未安装,请先安装Go 1.21或更高版本"
    echo "  下载地址: https://golang.org/dl/"
    exit 1
fi

# 检查FFmpeg
if command_exists ffmpeg; then
    FFMPEG_VERSION=$(ffmpeg -version 2>&1 | head -n1 | awk '{print $3}')
    success "FFmpeg已安装: $FFMPEG_VERSION"
else
    error "FFmpeg未安装"
    echo "  macOS: brew install ffmpeg"
    echo "  Ubuntu: sudo apt install ffmpeg"
    exit 1
fi

# 检查Git
if command_exists git; then
    success "Git已安装"
else
    warning "Git未安装(可选)"
fi

echo ""
echo "2. 创建数据目录..."
echo "-----------------------------------"

# 创建必要的目录
mkdir -p data/{uploads,projects,assets/{bgm,fonts}}
mkdir -p logs
success "数据目录已创建"

echo ""
echo "3. 下载示例资源..."
echo "-----------------------------------"

# 检查BGM目录
if [ ! -f "data/assets/bgm/default.mp3" ]; then
    warning "未找到默认BGM文件"
    echo "  请手动将BGM文件放置到: data/assets/bgm/default.mp3"
    echo "  或修改配置文件中的BGM设置"
else
    success "默认BGM文件已存在"
fi

echo ""
echo "4. 安装Go依赖..."
echo "-----------------------------------"

go mod download
success "Go依赖已安装"

echo ""
echo "5. 环境变量配置..."
echo "-----------------------------------"

# 检查环境变量
if [ -z "$OPENAI_API_KEY" ]; then
    warning "未设置OPENAI_API_KEY环境变量"
    echo "  请执行: export OPENAI_API_KEY='your-api-key'"
    echo "  或在configs/config.yaml中直接配置"
else
    success "OPENAI_API_KEY已设置"
fi

echo ""
echo "6. 配置文件检查..."
echo "-----------------------------------"

if [ -f "configs/config.yaml" ]; then
    success "配置文件已存在: configs/config.yaml"
    echo "  请确保已正确配置:"
    echo "  - OpenAI API Key"
    echo "  - Stable Diffusion API URL (如果使用本地部署)"
else
    error "配置文件不存在"
    exit 1
fi

echo ""
echo "======================================"
echo "   环境设置完成!"
echo "======================================"
echo ""
echo "下一步:"
echo "  1. 配置API密钥: 编辑 configs/config.yaml"
echo "  2. 启动Stable Diffusion (如果使用本地):"
echo "     cd /path/to/stable-diffusion-webui"
echo "     ./webui.sh --api"
echo "  3. 编译项目: make build"
echo "  4. 运行服务: make run"
echo ""
echo "快速开发模式:"
echo "  make dev"
echo ""
