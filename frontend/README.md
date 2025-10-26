# 文生漫画视频工具 - 前端

基于 Vue3 + Element Plus 的前端界面，提供直观的视频生成和任务管理功能。

## 功能特性

- 📝 **文本输入** - 支持长文本输入，实时字数统计
- ⚙️ **生成选项** - 视频风格、时长、画面比例等参数配置
- 📊 **任务管理** - 实时查看任务进度和状态
- 📥 **视频下载** - 一键下载生成的视频文件
- 🔄 **自动刷新** - 自动更新任务状态，无需手动刷新

## 快速开始

### 1. 安装依赖

```bash
cd frontend
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

前端将在 http://localhost:3000 启动

### 3. 确保后端服务运行

确保后端服务在 http://localhost:8080 运行：

```bash
# 在项目根目录
make dev
```

## 使用说明

### 创建视频

1. 在左侧文本框输入小说文本
2. 可以点击"加载示例文本"查看格式
3. 调整生成选项（风格、时长、比例）
4. 点击"生成视频"按钮

### 管理任务

1. 右侧显示所有任务列表
2. 实时显示任务进度和当前步骤
3. 任务完成后可以下载视频
4. 可以删除不需要的任务

### 任务状态说明

- **排队中** - 任务已创建，等待处理
- **处理中** - 正在生成视频
- **已完成** - 视频生成完成，可以下载
- **失败** - 生成失败，查看错误信息

### 处理步骤

1. **解析剧本** - AI分析文本结构
2. **生成分镜** - 创建视频分镜脚本
3. **生成图像** - 生成漫画风格图片（本地SD模式）
4. **渲染视频** - 合成最终视频

## 技术栈

- **Vue 3** - 渐进式JavaScript框架
- **Element Plus** - Vue 3 UI组件库
- **Axios** - HTTP客户端
- **Vite** - 现代前端构建工具

## 项目结构

```
frontend/
├── src/
│   ├── api/           # API接口
│   ├── App.vue        # 主应用组件
│   └── main.js        # 应用入口
├── index.html         # HTML模板
├── package.json       # 项目配置
├── vite.config.js     # Vite配置
└── README.md          # 说明文档
```

## 开发说明

### 代理配置

开发环境下，前端请求会自动代理到后端：

```javascript
// vite.config.js
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true
  }
}
```

### API接口

所有API调用都在 `src/api/index.js` 中定义：

- `checkHealth()` - 健康检查
- `generateVideo(data)` - 生成视频
- `getTasks()` - 获取任务列表
- `getTask(taskId)` - 获取单个任务
- `downloadVideo(taskId)` - 下载视频
- `deleteTask(taskId)` - 删除任务

### 自动刷新

任务列表每3秒自动刷新一次，确保状态实时更新。

## 构建部署

### 构建生产版本

```bash
npm run build
```

构建产物在 `dist/` 目录下。

### 预览构建结果

```bash
npm run preview
```

### 部署到服务器

将 `dist/` 目录内容部署到Web服务器，并配置API代理：

```nginx
# Nginx配置示例
location /api {
    proxy_pass http://localhost:8080/api;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

## 常见问题

### Q: 前端无法连接后端？
A: 检查后端服务是否在8080端口运行，确认CORS配置正确。

### Q: 任务状态不更新？
A: 检查浏览器控制台是否有错误，确认网络连接正常。

### Q: 视频下载失败？
A: 确认任务已完成且视频文件存在，检查浏览器下载权限。

### Q: 界面显示异常？
A: 清除浏览器缓存，刷新页面重试。

## 更新日志

### v1.0.0 (2025-10-26)
- ✅ 基础界面和功能
- ✅ 任务管理和进度显示
- ✅ 视频下载功能
- ✅ 响应式设计

## 许可证

MIT License