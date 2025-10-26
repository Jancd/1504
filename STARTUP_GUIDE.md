# 启动指南

## 🚀 快速启动

### Windows用户（推荐）

1. **使用启动器**（最简单）
   ```cmd
   scripts\start-all.bat
   ```
   选择选项1自动启动，会打开两个服务窗口和浏览器。

2. **直接启动**
   ```cmd
   scripts\start-dev.bat
   ```

3. **分别启动**（如果自动启动有问题）
   ```cmd
   # 终端1
   scripts\start-backend.bat
   
   # 终端2  
   scripts\start-frontend.bat
   ```

### Linux/Mac用户

```bash
chmod +x scripts/start-dev.sh
./scripts/start-dev.sh
```

## 🔧 故障排除

### 常见问题

#### 1. 脚本无法运行
**症状**: 双击bat文件闪退或报错
**解决**: 
- 右键bat文件 → "以管理员身份运行"
- 或在cmd中运行: `scripts\start-all.bat`

#### 2. Go命令找不到
**症状**: `'go' is not recognized as an internal or external command`
**解决**: 
- 安装Go: https://golang.org/dl/
- 确保Go在系统PATH中

#### 3. Node.js命令找不到  
**症状**: `'node' is not recognized as an internal or external command`
**解决**:
- 安装Node.js: https://nodejs.org/
- 重启命令行窗口

#### 4. 端口被占用
**症状**: `bind: address already in use`
**解决**:
```cmd
# 查看端口占用
netstat -ano | findstr :8080
netstat -ano | findstr :3000

# 结束占用进程
taskkill /PID <进程ID> /F
```

#### 5. 前端依赖安装失败
**症状**: npm install报错
**解决**:
```cmd
cd frontend
rmdir /s node_modules
del package-lock.json
npm install
```

#### 6. 权限问题
**症状**: 文件创建失败或访问被拒绝
**解决**:
- 以管理员身份运行
- 检查防病毒软件是否阻止

### 验证服务状态

#### 检查后端
```cmd
curl http://localhost:8080/health
```
或浏览器访问: http://localhost:8080/health

#### 检查前端
浏览器访问: http://localhost:3000

## 📋 启动脚本说明

### 脚本文件

- `start-all.bat` - 启动器，提供多种启动选项
- `start-dev.bat` - 自动启动前后端
- `start-backend.bat` - 只启动后端服务
- `start-frontend.bat` - 只启动前端服务
- `simple-start.bat` - 简化版本，用于调试

### 启动流程

1. **环境检查** - 验证Go、Node.js、npm是否安装
2. **依赖安装** - 自动安装前端依赖（如果需要）
3. **启动后端** - 在新窗口启动Go服务
4. **启动前端** - 在新窗口启动Vue开发服务器
5. **打开浏览器** - 自动打开前端界面

### 服务窗口

启动后会看到3个窗口：
- **启动器窗口** - 可以关闭
- **后端服务窗口** - 显示后端日志，关闭会停止后端
- **前端服务窗口** - 显示前端日志，关闭会停止前端

## 🎯 使用建议

### 开发环境
- 使用 `start-all.bat` 启动器
- 保持服务窗口开启以查看日志
- 修改代码后前端会自动重载

### 生产环境
- 参考 `frontend/deploy.md`
- 使用 `npm run build` 构建前端
- 使用 `go build` 编译后端

### 调试问题
- 查看服务窗口的日志输出
- 使用浏览器开发者工具
- 检查网络连接和防火墙

## 📞 获取帮助

如果遇到问题：

1. **查看日志** - 检查后端和前端窗口的输出
2. **检查文档** - 阅读 README.md 和相关文档
3. **重新启动** - 关闭所有服务窗口，重新运行启动脚本
4. **清理环境** - 删除 `frontend/node_modules` 重新安装

---

**提示**: 首次启动可能需要几分钟来安装依赖，请耐心等待。