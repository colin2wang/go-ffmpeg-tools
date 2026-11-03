package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"go-ffmpeg-tools/internal/ffmpeg"
	"go-ffmpeg-tools/internal/task"
)

// App Wails 绑定入口，暴露给前端调用的方法
type App struct {
	ctx    context.Context
	jobMgr *task.Manager
}

func NewApp() *App {
	return &App{
		jobMgr: task.NewManager(),
	}
}

// Startup 由 Wails 在应用启动时调用
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// SubmitJobReq 提交任务请求
type SubmitJobReq struct {
	InputVideo   string         `json:"inputVideo"`
	OutputVideo  string         `json:"outputVideo"`
	OutputMode   string         `json:"outputMode"`   // custom / same / exe
	OutputSuffix string         `json:"outputSuffix"` // same/exe 模式下拼接在原文件名后
	SubConfig    task.SubConfig `json:"subConfig"`
}

// resolveOutputPath 按输出模式计算最终输出路径：
// same = 输入视频同目录；exe = 程序可执行文件目录；custom（默认）= 直接使用 OutputVideo
func (a *App) resolveOutputPath(req SubmitJobReq) (string, error) {
	base := strings.TrimSuffix(filepath.Base(req.InputVideo), filepath.Ext(req.InputVideo))
	suffix := req.OutputSuffix
	if suffix == "" {
		suffix = ".subbed.mp4"
	}
	switch req.OutputMode {
	case "same":
		return filepath.Join(filepath.Dir(req.InputVideo), base+suffix), nil
	case "exe":
		exePath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("resolve executable dir: %w", err)
		}
		return filepath.Join(filepath.Dir(exePath), base+suffix), nil
	default:
		return req.OutputVideo, nil
	}
}

// SubmitJob 创建并异步执行一个字幕任务，立即返回任务快照
func (a *App) SubmitJob(req SubmitJobReq) (*task.Job, error) {
	if req.InputVideo == "" {
		return nil, fmt.Errorf("inputVideo is required")
	}
	outputVideo, err := a.resolveOutputPath(req)
	if err != nil {
		return nil, err
	}
	if outputVideo == "" {
		return nil, fmt.Errorf("outputVideo is required")
	}

	jobId := uuid.NewString()
	ctxJob, cancel := context.WithCancel(context.Background())
	job := &task.Job{
		JobId:       jobId,
		InputVideo:  req.InputVideo,
		OutputVideo: outputVideo,
		SubConfig:   req.SubConfig,
		Status:      "pending",
		Log:         []string{},
	}
	job.BindCtx(ctxJob, cancel)
	a.jobMgr.AddJob(job)
	a.emitJob(job)

	// 异步跑任务
	go func() {
		job.Status = "running"

		// 尝试探测视频总时长（失败不阻塞任务）
		if job.TotalSeconds == 0 {
			if d := ffmpeg.ProbeDuration(job.InputVideo); d > 0 {
				job.TotalSeconds = d
			}
		}

		args, err := ffmpeg.BuildArgs(job.InputVideo, job.OutputVideo, job.SubConfig)
		if err != nil {
			job.Status = "error"
			job.AppendLog(err.Error())
			a.emitJob(job)
			return
		}
		job.AppendLog("$ " + joinArgs(args))
		a.emitJob(job)

		// 进度回调本身被 ffmpeg runner 逐行调用，节流后再 emit
		var lastEmit time.Time
		err = ffmpeg.RunJob(ctxJob, args, job.TotalSeconds, func(currentSec float64, logLine string) {
			if logLine != "" {
				job.AppendLog(logLine)
			}
			if currentSec > 0 {
				job.CurrentSeconds = currentSec
				if job.TotalSeconds > 0 {
					p := currentSec / job.TotalSeconds * 100
					if p > 100 {
						p = 100
					}
					job.Progress = p
				}
			}
			// 至少间隔 200ms 才向前端推送一次，避免刷爆事件通道
			if time.Since(lastEmit) > 200*time.Millisecond {
				lastEmit = time.Now()
				a.emitJob(job)
			}
		})

		if err != nil {
			if err.Error() == "cancelled" {
				job.Status = "cancelled"
			} else {
				job.Status = "error"
				job.AppendLog(err.Error())
			}
		} else {
			job.Status = "done"
			job.Progress = 100
		}
		a.emitJob(job)
	}()

	return job, nil
}

// CancelJob 取消任务，返回是否成功触发
func (a *App) CancelJob(jobId string) bool {
	return a.jobMgr.CancelJob(jobId)
}

// GetJobs 返回全部任务（按提交顺序）
func (a *App) GetJobs() []*task.Job {
	jobs := a.jobMgr.ListJobs()
	out := make([]*task.Job, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, j.Snapshot())
	}
	return out
}

// SelectFile 打开文件选择对话框（通用）
func (a *App) SelectFile(title string, filters []runtime.FileFilter) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: filters,
	})
}

// SelectOutputFile 打开保存对话框
func (a *App) SelectOutputFile(title string, defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
	})
}

// emitJob 通过 Wails 事件推送任务状态到前端
func (a *App) emitJob(job *task.Job) {
	runtime.EventsEmit(a.ctx, "jobUpdate", job.Snapshot())
}

func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		out += a
	}
	return out
}
