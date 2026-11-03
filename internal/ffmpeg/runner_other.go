//go:build !windows

package ffmpeg

import "os/exec"

// hideCmdWindow 非 Windows 平台无需隐藏窗口
func hideCmdWindow(cmd *exec.Cmd) {}
