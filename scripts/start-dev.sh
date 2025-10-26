#!/bin/bash

# 文生漫画视频工具 - 开发环境启动脚本

echo "🚀 启动文生漫画视频工具开发环境"
echo "=================================="

# 检查是否在项目根目录
if [ ! -f "go.mod" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go 1.21+"
    exit 1
fi

# 检查Node.js是否安装
if ! command -v node &> /dev/null; then
    echo "❌ Node.js未安装，请先安装Node.js"
    exit 1
fi

# 检查npm是否安装
if ! command -v npm &> /dev/null; then
    echo "❌ npm未安装，请先安装npm"
    exit 1
fi

echo "✅ 环境检查通过"
echo ""

# 安装前端依赖（如果需要）
if [ ! -d "frontend/node_modules" ]; then
    echo "📦 安装前端依赖..."
    cd frontend
    npm install
    cd ..
    echo "✅ 前端依赖安装完成"
    echo ""
fi

# 启动后端服务
echo "🔧 启动后端服务..."
echo "后端地址: http://localhost:8080"
echo "API文档: http://localhost:8080/health"
echo ""

# 在后台启动后端
go run cmd/server/main.go &
BACKEND_PID=$!

# 等待后端启动
echo "⏳ 等待后端服务启动..."
sleep 3

# 检查后端是否启动成功
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ 后端服务启动成功"
else
    echo "❌ 后端服务启动失败"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

echo ""

# 启动前端服务
echo "🎨 启动前端服务..."
echo "前端地址: http://localhost:3000"
echo ""

cd frontend
npm run dev &
FRONTEND_PID=$!

echo "🎉 开发环境启动完成！"
echo ""
echo "📋 服务信息:"
echo "   后端服务: http://localhost:8080"
echo "   前端界面: http://localhost:3000"
echo "   健康检查: http://localhost:8080/health"
echo ""
echo "💡 使用说明:"
echo "   1. 打开浏览器访问 http://localhost:3000"
echo "   2. 在文本框输入小说内容"
echo "   3. 点击'生成视频'开始创建"
echo "   4. 在右侧查看任务进度"
echo "   5. 完成后点击'下载'获取视频"
echo ""
echo "⚠️  注意事项:"
echo "   - 确保已配置OpenAI API Key"
echo "   - 七牛云模式需要配置七牛云API Key"
echo "   - 本地SD模式需要启动Stable Diffusion服务"
echo ""
echo "🛑 停止服务: 按 Ctrl+C"

# 等待用户中断
trap 'echo ""; echo "🛑 正在停止服务..."; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; echo "✅ 服务已停止"; exit 0' INT

# 保持脚本运行
wait