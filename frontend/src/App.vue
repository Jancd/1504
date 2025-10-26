<template>
  <div id="app">
    <el-container>
      <!-- 头部 -->
      <el-header class="header">
        <div class="header-content">
          <h1>
            <el-icon><VideoCamera /></el-icon>
            文生漫画视频工具
          </h1>
          <div class="header-info">
            <el-tag type="success">{{ serverStatus }}</el-tag>
            <el-button @click="checkHealth" :loading="checking" size="small">
              <el-icon><Refresh /></el-icon>
              检查状态
            </el-button>
          </div>
        </div>
      </el-header>

      <!-- 主体内容 -->
      <el-main class="main-content">
        <el-row :gutter="20">
          <!-- 左侧：文本输入和生成 -->
          <el-col :span="12">
            <el-card class="input-card">
              <template #header>
                <div class="card-header">
                  <span>创建视频</span>
                  <el-button 
                    type="primary" 
                    @click="generateVideo" 
                    :loading="generating"
                    :disabled="!inputText.trim()"
                  >
                    <el-icon><VideoPlay /></el-icon>
                    生成视频
                  </el-button>
                </div>
              </template>

              <!-- 文本输入 -->
              <div class="input-section">
                <el-input
                  v-model="inputText"
                  type="textarea"
                  :rows="12"
                  placeholder="请输入小说文本，例如：

场景:宁静的高中校园,春天的清晨。樱花树下,花瓣随风飘落。

小樱独自走在樱花树下,背着书包,神情有些紧张。她停下脚步,看着远处的教学楼。

小樱(内心独白,紧张):今天是新学期的第一天,不知道会遇到什么样的同学呢?

..."
                  maxlength="2000"
                  show-word-limit
                />
              </div>



              <!-- 生成选项 -->
              <div class="options-section">
                <el-row :gutter="16">
                  <el-col :span="8">
                    <el-form-item label="视频风格">
                      <el-select v-model="options.style" style="width: 100%">
                        <el-option label="日系动漫" value="anime" />
                        <el-option label="写实风格" value="realistic" />
                        <el-option label="水彩画风" value="watercolor" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="目标时长">
                      <el-input-number 
                        v-model="options.duration_target" 
                        :min="10" 
                        :max="120" 
                        style="width: 100%"
                      />
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="画面比例">
                      <el-select v-model="options.aspect_ratio" style="width: 100%">
                        <el-option label="16:9" value="16:9" />
                        <el-option label="4:3" value="4:3" />
                        <el-option label="1:1" value="1:1" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
              </div>

              <!-- 示例文本按钮 -->
              <div class="example-section">
                <el-button @click="loadExample" size="small" type="info" plain>
                  <el-icon><Document /></el-icon>
                  加载示例文本
                </el-button>
                <el-button @click="clearText" size="small" plain>
                  <el-icon><Delete /></el-icon>
                  清空文本
                </el-button>
              </div>
            </el-card>
          </el-col>

          <!-- 右侧：任务列表和进度 -->
          <el-col :span="12">
            <el-card class="task-card">
              <template #header>
                <div class="card-header">
                  <span>任务管理</span>
                  <el-button @click="refreshTasks" size="small">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </template>

              <!-- 任务列表 -->
              <div class="task-list">
                <div v-if="tasks.length === 0" class="empty-state">
                  <el-empty description="暂无任务" />
                </div>
                
                <div v-for="task in tasks" :key="task.task_id" class="task-item">
                  <el-card shadow="hover">
                    <div class="task-header">
                      <div class="task-info">
                        <span class="task-id">{{ task.task_id.substring(0, 8) }}...</span>
                        <el-tag 
                          :type="getStatusType(task.status)" 
                          size="small"
                        >
                          {{ getStatusText(task.status) }}
                        </el-tag>
                      </div>
                      <div class="task-actions">
                        <el-button 
                          v-if="task.status === 'completed'" 
                          @click="downloadVideo(task.task_id)"
                          type="success" 
                          size="small"
                        >
                          <el-icon><Download /></el-icon>
                          下载
                        </el-button>
                        <el-button 
                          @click="deleteTask(task.task_id)"
                          type="danger" 
                          size="small"
                          plain
                        >
                          <el-icon><Delete /></el-icon>
                        </el-button>
                      </div>
                    </div>

                    <!-- 进度条 -->
                    <div class="task-progress">
                      <el-progress 
                        :percentage="task.progress" 
                        :status="task.status === 'failed' ? 'exception' : 
                                task.status === 'completed' ? 'success' : ''"
                      />
                      <div class="progress-text">
                        {{ task.current_step }}
                      </div>
                    </div>

                    <!-- 步骤详情 -->
                    <div class="task-steps">
                      <el-steps :active="getActiveStep(task)" size="small">
                        <el-step 
                          v-for="step in task.steps" 
                          :key="step.name"
                          :title="getStepTitle(step.name)"
                          :status="getStepStatus(step.status)"
                        />
                      </el-steps>
                    </div>

                    <!-- 错误信息 -->
                    <div v-if="task.error" class="task-error">
                      <el-alert 
                        :title="task.error" 
                        type="error" 
                        :closable="false"
                        show-icon
                      />
                    </div>

                    <!-- 结果信息 -->
                    <div v-if="task.result" class="task-result">
                      <el-descriptions :column="2" size="small">
                        <el-descriptions-item label="时长">
                          {{ Math.round(task.result.duration) }}秒
                        </el-descriptions-item>
                        <el-descriptions-item label="分辨率">
                          {{ task.result.resolution }}
                        </el-descriptions-item>
                        <el-descriptions-item label="文件大小">
                          {{ formatFileSize(task.result.file_size) }}
                        </el-descriptions-item>
                        <el-descriptions-item label="镜头数">
                          {{ task.result.shot_count }}个
                        </el-descriptions-item>
                      </el-descriptions>
                    </div>
                  </el-card>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-main>
    </el-container>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from './api/index.js'

export default {
  name: 'App',
  setup() {
    // 响应式数据
    const serverStatus = ref('检查中...')
    const checking = ref(false)
    const generating = ref(false)
    const inputText = ref('')
    const tasks = ref([])
    const options = ref({
      style: 'anime',
      duration_target: 60,
      aspect_ratio: '16:9'
    })

    let refreshTimer = null

    // 检查服务器健康状态
    const checkHealth = async () => {
      checking.value = true
      try {
        const response = await api.checkHealth()
        serverStatus.value = '服务正常'
      } catch (error) {
        serverStatus.value = '服务异常'
        console.error('Health check failed:', error)
      } finally {
        checking.value = false
      }
    }

    // 生成视频
    const generateVideo = async () => {
      if (!inputText.value.trim()) {
        ElMessage.warning('请输入文本内容')
        return
      }

      generating.value = true
      try {
        const response = await api.generateVideo({
          text: inputText.value,
          options: options.value
        })
        
        ElMessage.success('任务创建成功！')
        await refreshTasks()
      } catch (error) {
        ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
      } finally {
        generating.value = false
      }
    }

    // 刷新任务列表
    const refreshTasks = async () => {
      try {
        const response = await api.getTasks()
        // 后端返回的数据结构是 { code, message, data: { tasks: [...] } }
        tasks.value = response.data.data?.tasks || []
      } catch (error) {
        console.error('Failed to refresh tasks:', error)
      }
    }

    // 下载视频
    const downloadVideo = async (taskId) => {
      try {
        const response = await api.downloadVideo(taskId)
        
        // 创建下载链接
        const url = window.URL.createObjectURL(new Blob([response.data]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `video_${taskId.substring(0, 8)}.mp4`)
        document.body.appendChild(link)
        link.click()
        link.remove()
        window.URL.revokeObjectURL(url)
        
        ElMessage.success('视频下载成功！')
      } catch (error) {
        ElMessage.error('下载失败: ' + (error.response?.data?.error || error.message))
      }
    }

    // 删除任务
    const deleteTask = async (taskId) => {
      try {
        await ElMessageBox.confirm('确定要删除这个任务吗？', '确认删除', {
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          type: 'warning'
        })
        
        await api.deleteTask(taskId)
        ElMessage.success('任务删除成功！')
        await refreshTasks()
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
        }
      }
    }

    // 加载示例文本
    const loadExample = () => {
      inputText.value = `场景:宁静的高中校园,春天的清晨。樱花树下,花瓣随风飘落。

小樱独自走在樱花树下,背着书包,神情有些紧张。她停下脚步,看着远处的教学楼。

小樱(内心独白,紧张):今天是新学期的第一天,不知道会遇到什么样的同学呢?

突然,一个男生从她身边快速跑过,差点撞到她。

男生(气喘吁吁):对不起!要迟到了!

小樱(惊讶):啊!

男生回头看了一眼,两人的目光相遇了一瞬间。樱花花瓣在他们之间飘落。

小樱脸红了,心跳加速。

小樱(内心独白,害羞):好帅...

男生已经跑远了,小樱呆呆地站在原地。

小樱(自言自语,期待):新的学期,好像会很有趣呢。

她笑了笑,继续向教室走去。远处的教学楼在晨光中闪闪发光。`
    }

    // 清空文本
    const clearText = () => {
      inputText.value = ''
    }

    // 工具函数
    const getStatusType = (status) => {
      const types = {
        'queued': 'info',
        'processing': 'warning', 
        'completed': 'success',
        'failed': 'danger'
      }
      return types[status] || 'info'
    }

    const getStatusText = (status) => {
      const texts = {
        'queued': '排队中',
        'processing': '处理中',
        'completed': '已完成', 
        'failed': '失败'
      }
      return texts[status] || status
    }

    const getStepTitle = (stepName) => {
      const titles = {
        'parse_script': '解析剧本',
        'generate_storyboard': '生成分镜',
        'generate_images': '生成图像',
        'render_video': '渲染视频'
      }
      return titles[stepName] || stepName
    }

    const getStepStatus = (status) => {
      const statuses = {
        'pending': 'wait',
        'processing': 'process',
        'completed': 'finish',
        'failed': 'error'
      }
      return statuses[status] || 'wait'
    }

    const getActiveStep = (task) => {
      const stepNames = ['parse_script', 'generate_storyboard', 'generate_images', 'render_video']
      const currentIndex = stepNames.indexOf(task.current_step)
      return currentIndex >= 0 ? currentIndex : 0
    }

    const formatFileSize = (bytes) => {
      if (!bytes) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    // 生命周期
    onMounted(() => {
      checkHealth()
      refreshTasks()
      
      // 定时刷新任务状态
      refreshTimer = setInterval(() => {
        refreshTasks()
      }, 3000)
    })

    onUnmounted(() => {
      if (refreshTimer) {
        clearInterval(refreshTimer)
      }
    })

    return {
      serverStatus,
      checking,
      generating,
      inputText,
      tasks,
      options,
      checkHealth,
      generateVideo,
      refreshTasks,
      downloadVideo,
      deleteTask,
      loadExample,
      clearText,
      getStatusType,
      getStatusText,
      getStepTitle,
      getStepStatus,
      getActiveStep,
      formatFileSize
    }
  }
}
</script>

<style scoped>
.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
  padding: 0 20px;
}

.header-content h1 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.main-content {
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.input-card, .task-card {
  height: calc(100vh - 120px);
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.input-section {
  margin-bottom: 20px;
}



.options-section {
  margin-bottom: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 6px;
}

.example-section {
  display: flex;
  gap: 10px;
}

.task-list {
  height: calc(100vh - 200px);
  overflow-y: auto;
}

.task-item {
  margin-bottom: 15px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.task-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.task-id {
  font-family: monospace;
  font-size: 12px;
  color: #666;
}

.task-actions {
  display: flex;
  gap: 5px;
}

.task-progress {
  margin-bottom: 15px;
}

.progress-text {
  font-size: 12px;
  color: #666;
  margin-top: 5px;
}

.task-steps {
  margin-bottom: 15px;
}

.task-error {
  margin-bottom: 15px;
}

.task-result {
  background: #f0f9ff;
  padding: 10px;
  border-radius: 6px;
}

.empty-state {
  text-align: center;
  padding: 50px 0;
}
</style>