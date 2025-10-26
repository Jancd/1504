package task

import (
	"fmt"
	"sync"

	"github.com/Jancd/1504/internal/model"
)

// Manager 任务管理器
type Manager struct {
	tasks map[string]*model.Task
	mu    sync.RWMutex
}

// NewManager 创建任务管理器
func NewManager() *Manager {
	return &Manager{
		tasks: make(map[string]*model.Task),
	}
}

// Create 创建任务
func (m *Manager) Create(task *model.Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
}

// Get 获取任务
func (m *Manager) Get(taskID string) (*model.Task, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.tasks[taskID]
	return task, ok
}

// Update 更新任务
func (m *Manager) Update(task *model.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[task.ID]; !ok {
		return fmt.Errorf("task not found: %s", task.ID)
	}
	m.tasks[task.ID] = task
	return nil
}

// Delete 删除任务
func (m *Manager) Delete(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[taskID]; !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	delete(m.tasks, taskID)
	return nil
}

// List 列出所有任务
func (m *Manager) List() []*model.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*model.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// UpdateTaskStatus 更新任务状态
func (m *Manager) UpdateTaskStatus(taskID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Status = status
	return nil
}

// UpdateTaskProgress 更新任务进度
func (m *Manager) UpdateTaskProgress(taskID string, progress int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Progress = progress
	return nil
}

// SetTaskError 设置任务错误
func (m *Manager) SetTaskError(taskID, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Status = model.TaskStatusFailed
	task.Error = errMsg
	return nil
}

// SetTaskResult 设置任务结果
func (m *Manager) SetTaskResult(taskID string, result *model.Result) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Result = result
	task.Status = model.TaskStatusCompleted
	task.Progress = 100
	return nil
}
