package task

import (
	"context"
	"sync"
)

// SubTrack 单条字幕轨道（多路软字幕用）
type SubTrack struct {
	File  string `json:"file"`  // srt 文件绝对路径
	Lang  string `json:"lang"`  // 语言代码：zh / eng / jpn 等，ffmpeg language tag
	Title string `json:"title"` // 显示名称：中文 / English
}

// SubConfig 字幕配置
type SubConfig struct {
	Mode     string     `json:"mode"` // soft / burn
	Tracks   []SubTrack `json:"tracks"`
	FontName string     `json:"fontName,omitempty"` // 硬字幕专属
	FontSize int        `json:"fontSize,omitempty"`
	FontFile string     `json:"fontFile,omitempty"`
}

// Job 一次转码/封装任务
type Job struct {
	JobId          string    `json:"jobId"`
	InputVideo     string    `json:"inputVideo"`
	OutputVideo    string    `json:"outputVideo"`
	SubConfig      SubConfig `json:"subConfig"`
	TotalSeconds   float64   `json:"totalSeconds"`
	CurrentSeconds float64   `json:"currentSeconds"`
	Progress       float64   `json:"progress"` // 0~100
	Status         string    `json:"status"`   // pending / running / done / error / cancelled
	Log            []string  `json:"log"`

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// Snapshot 返回一个用于序列化/事件推送的浅拷贝（手动逐字段复制，避免拷贝内部的锁与 context）
func (j *Job) Snapshot() *Job {
	j.mu.Lock()
	defer j.mu.Unlock()
	return &Job{
		JobId:          j.JobId,
		InputVideo:     j.InputVideo,
		OutputVideo:    j.OutputVideo,
		SubConfig:      j.SubConfig,
		TotalSeconds:   j.TotalSeconds,
		CurrentSeconds: j.CurrentSeconds,
		Progress:       j.Progress,
		Status:         j.Status,
		Log:            append([]string(nil), j.Log...),
	}
}

// AppendLog 线程安全地追加一行日志
func (j *Job) AppendLog(line string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Log = append(j.Log, line)
}

// Logs 线程安全地读取日志副本
func (j *Job) Logs() []string {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make([]string, len(j.Log))
	copy(out, j.Log)
	return out
}

// BindCtx 绑定取消用的 context（创建任务时调用一次）
func (j *Job) BindCtx(ctx context.Context, cancel context.CancelFunc) {
	j.ctx = ctx
	j.cancel = cancel
}

// Cancel 触发任务取消
func (j *Job) Cancel() bool {
	if j.cancel == nil {
		return false
	}
	j.cancel()
	return true
}
