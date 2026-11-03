package task

import (
	"sync"
)

// Manager 任务管理器：保存全部任务并提供取消能力
type Manager struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	// 插入顺序，用于 GetJobs 按提交顺序返回
	order []string
}

func NewManager() *Manager {
	return &Manager{
		jobs: make(map[string]*Job),
	}
}

func (m *Manager) AddJob(job *Job) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.JobId] = job
	m.order = append(m.order, job.JobId)
}

func (m *Manager) GetJob(id string) *Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobs[id]
}

// CancelJob 取消任务；只有运行中/等待中的任务可取消
func (m *Manager) CancelJob(jobId string) bool {
	m.mu.RLock()
	job, ok := m.jobs[jobId]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	switch job.Status {
	case "running", "pending":
		return job.Cancel()
	}
	return false
}

// ListJobs 按提交顺序返回任务快照（浅拷贝，日志另取）
func (m *Manager) ListJobs() []*Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Job, 0, len(m.order))
	for _, id := range m.order {
		if j, ok := m.jobs[id]; ok {
			out = append(out, j)
		}
	}
	return out
}
