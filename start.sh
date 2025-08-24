#!/bin/bash

echo "🦄 启动Wallet项目"
echo "===================="

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go 1.18+"
    exit 1
fi

# 检查Node.js是否安装
if ! command -v node &> /dev/null; then
    echo "❌ Node.js未安装，请先安装Node.js 16+"
    exit 1
fi

echo "✅ Go和Node.js已安装"

# 启动后端服务
echo "🚀 启动后端服务..."
echo "后端将在 http://localhost:8087 运行"
echo "按 Ctrl+C 停止后端服务"
echo ""

# 在后台启动Go服务
go run main.go &
BACKEND_PID=$!

# 等待后端启动
sleep 3

# 启动前端
echo "🎨 启动前端服务..."
echo "请在新终端窗口中运行以下命令："
echo ""
echo "  cd web"
echo "  npm install  # 首次运行时需要"
echo "  npm run dev"
echo ""
echo "前端将在 http://localhost:3000 运行"
echo ""
echo "====================="
echo "🎉 项目启动完成！"
echo "后端: http://localhost:8087"
echo "前端: http://localhost:3000"
echo "====================="
echo ""
echo "按任意键停止后端服务..."
read -n 1

# 停止后端服务
kill $BACKEND_PID
echo "后端服务已停止"