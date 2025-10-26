# 快速启动指南

如果自动启动脚本有问题，可以按照以下步骤手动启动：

## 方式一：使用启动脚本

### Windows用户

```powershell
# 方式1：PowerShell脚本（推荐）
powershell -ExecutionPolicy Bypass -File scripts/start-dev.ps1

# 方式2：批处理脚本
scripts\start-dev.bat
```

### Linux/Mac用户

```bash
chmod +x scripts/start-dev.sh
./scripts/start-dev.sh
```

## 方式二：手动启动（推荐）

### 1. 启动后端服务

打开第一个终端窗口：

```bash
# 确保在项目根目录
cd /path/to/1504

# 启动后端
go run cmd/server/main.go
```

看到类似输出表示启动成功：
```
INFO    Starting MVP Video Generator Server
INFO    Server starting {"address": "0.0.0.0:8080", "mode": "debug"}
```

### 2. 启动前端服务

打开第二个终端窗口：

```bash
# 进入前端目录
cd frontend

# 安装依赖（首次运行）
npm install

# 启动前端开发服务器
npm run dev
```

看到类似输出表示启动成功：
```
  VITE v5.0.0  ready in 500 ms

  ➜  Local:   http://localhost:3000/
  ➜  Network: use --host to expose
```

### 3. 访问应用

打开浏览器访问：http://localhost:3000

## 方式三：使用Makefile

```bash
# 启动后端
make dev

# 启动前端（新终端）
make frontend

# 或者一键启动（如果支持）
make dev-full
```

## 验证服务状态

### 检查后端服务

```bash
curl http://localhost:8080/health
```

应该返回：
```json
{
  "status": "ok",
  "version": "1.0.0",
  "time": "2025-10-26T..."
}
```

### 检查前端服务

浏览器访问 http://localhost:3000，应该看到视频生成工具界面。

## 常见问题解决

### Q1: 端口被占用

```bash
# 查看端口占用
netstat -ano | findstr :8080
netstat -ano | findstr :3000

# 杀死占用进程（Windows）
taskkill /PID <PID> /F

# 杀死占用进程（Linux/Mac）
kill -9 <PID>
```

### Q2: Go命令找不到

确保Go已正确安装并添加到PATH：

```bash
# 检查Go版本
go version

# 如果命令不存在，请安装Go 1.21+
# Windows: https://golang.org/dl/
# Mac: brew install go
# Ubuntu: sudo apt install golang-go
```

### Q3: Node.js/npm命令找不到

确保Node.js已正确安装：

```bash
# 检查版本
node --version
npm --version

# 如果命令不存在，请安装Node.js 16+
# https://nodejs.org/
```

### Q4: 前端依赖安装失败

```bash
# 清除缓存重新安装
cd frontend
rm -rf node_modules package-lock.json
npm install

# 或使用yarn
yarn install
```

### Q5: API请求失败

检查：
1. 后端服务是否正常运行（http://localhost:8080/health）
2. 前端代理配置是否正确（vite.config.js）
3. 防火墙是否阻止了连接

### Q6: 配置文件问题

确保 `configs/config.yaml` 中的API密钥已正确配置：

```yaml
openai:
  api_key: "sk-your-actual-api-key"  # 替换为真实的API Key
  
video_generation:
  qiniu:
    api_key: "sk-your-qiniu-key"     # 替换为真实的七牛云API Key
```

## 开发环境要求

- **Go**: 1.21+
- **Node.js**: 16+
- **npm**: 8+
- **操作系统**: Windows 10+, macOS 10.15+, Linux

## 生产环境部署

参考 `frontend/deploy.md` 文档进行生产环境部署。

## 获取帮助

如果遇到问题：

1. 检查终端输出的错误信息
2. 查看浏览器开发者工具的控制台
3. 确认所有依赖都已正确安装
4. 参考项目文档：`README.md`, `frontend/README.md`

---

**提示**: 建议使用方式二（手动启动）进行开发，这样可以更好地查看日志和调试问题。