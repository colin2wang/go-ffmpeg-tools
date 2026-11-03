//go:build windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

// hideCmdWindow 隐藏 ffmpeg 子进程的控制台窗口（GUI 程序下每次调用会闪黑框）
func hideCmdWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
