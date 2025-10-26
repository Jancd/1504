# 前端部署指南

## 开发环境部署

### 1. 快速启动

```bash
# 方式一：使用启动脚本（推荐）
# Windows
scripts\start-dev.bat

# Linux/Mac  
./scripts/start-dev.sh

# 方式二：使用Makefile
make dev-full

# 方式三：分别启动
make dev          # 启动后端
make frontend     # 启动前端
```

### 2. 访问地址

- 前端界面: http://localhost:3000
- 后端API: http://localhost:8080
- 健康检查: http://localhost:8080/health

## 生产环境部署

### 1. 构建前端

```bash
cd frontend
npm install
npm run build
```

构建产物在 `frontend/dist/` 目录。

### 2. 部署到Nginx

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    # 前端静态文件
    location / {
        root /path/to/frontend/dist;
        try_files $uri $uri/ /index.html;
    }
    
    # API代理到后端
    location /api {
        proxy_pass http://localhost:8080/api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # 健康检查
    location /health {
        proxy_pass http://localhost:8080/health;
    }
}
```

### 3. 使用Docker部署

创建 `frontend/Dockerfile`:

```dockerfile
FROM node:18-alpine as builder

WORKDIR /app
COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

创建 `frontend/nginx.conf`:

```nginx
server {
    listen 80;
    
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://backend:8080/api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4. Docker Compose部署

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  backend:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs
      - ./data:/app/data
    environment:
      - GIN_MODE=release
    
  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend
```

启动：

```bash
docker-compose up -d
```

## 环境变量配置

### 前端环境变量

创建 `frontend/.env.production`:

```env
VITE_API_BASE_URL=https://your-api-domain.com
VITE_APP_TITLE=文生漫画视频工具
```

### 后端环境变量

```bash
export OPENAI_API_KEY="your-openai-key"
export QINIU_API_KEY="your-qiniu-key"
export GIN_MODE="release"
```

## 性能优化

### 1. 前端优化

- 启用Gzip压缩
- 配置CDN加速
- 使用HTTP/2
- 启用浏览器缓存

### 2. Nginx配置优化

```nginx
# 启用Gzip
gzip on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

# 缓存静态资源
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

## 监控和日志

### 1. 前端错误监控

可以集成Sentry等错误监控服务：

```javascript
// main.js
import * as Sentry from "@sentry/vue";

Sentry.init({
  app,
  dsn: "YOUR_SENTRY_DSN",
});
```

### 2. 访问日志

Nginx访问日志配置：

```nginx
log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                '$status $body_bytes_sent "$http_referer" '
                '"$http_user_agent" "$http_x_forwarded_for"';

access_log /var/log/nginx/access.log main;
```

## 常见问题

### Q: 前端构建失败？
A: 检查Node.js版本，确保使用16+版本，清除node_modules重新安装。

### Q: API请求跨域？
A: 确认后端CORS配置正确，或使用Nginx代理。

### Q: 静态资源404？
A: 检查Nginx配置，确保try_files配置正确。

### Q: 页面刷新404？
A: SPA应用需要配置fallback到index.html。

## 更新部署

### 1. 前端更新

```bash
cd frontend
git pull
npm install
npm run build
# 复制dist到服务器
```

### 2. 零停机更新

使用蓝绿部署或滚动更新策略，确保服务不中断。