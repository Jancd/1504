import axios from 'axios'

// 创建axios实例
const api = axios.create({
    baseURL: '/api',
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json'
    }
})

// 请求拦截器
api.interceptors.request.use(
    config => {
        console.log('API Request:', config.method?.toUpperCase(), config.url, config.data)
        return config
    },
    error => {
        console.error('Request Error:', error)
        return Promise.reject(error)
    }
)

// 响应拦截器
api.interceptors.response.use(
    response => {
        console.log('API Response:', response.config.url, response.data)
        return response
    },
    error => {
        console.error('Response Error:', error.response?.data || error.message)
        return Promise.reject(error)
    }
)

// API方法
export default {
    // 健康检查
    checkHealth() {
        // 健康检查不在 /api 路径下，直接调用根路径
        return axios.get('/health')
    },

    // 生成视频
    generateVideo(data) {
        return api.post('/generate', data)
    },

    // 获取任务列表
    getTasks() {
        return api.get('/tasks')
    },

    // 获取单个任务
    getTask(taskId) {
        return api.get(`/tasks/${taskId}`)
    },

    // 下载视频
    downloadVideo(taskId) {
        return api.get(`/download/${taskId}`, {
            responseType: 'blob'
        })
    },

    // 删除任务
    deleteTask(taskId) {
        return api.delete(`/tasks/${taskId}`)
    }
}