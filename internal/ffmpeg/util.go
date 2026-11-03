package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetFFmpegBinPath 返回 ffmpeg 可执行文件路径。
// 优先取可执行文件同目录下的 build/ffmpeg/<goos>/，找不到则回退到 PATH 中的 ffmpeg。
func GetFFmpegBinPath() (string, error) {
	binName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binName = "ffmpeg.exe"
	}

	var roots []string
	if exe, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}

	for _, root := range roots {
		p := filepath.Join(root, "build", "ffmpeg", runtime.GOOS, binName)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}

	// fallback: 从 PATH 找
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("ffmpeg not found: expected %s or ffmpeg in PATH", filepath.Join("build", "ffmpeg", runtime.GOOS, binName))
}

// GetFFprobeBinPath 返回 ffprobe 可执行文件路径，查找规则同 ffmpeg。
func GetFFprobeBinPath() (string, error) {
	binName := "ffprobe"
	if runtime.GOOS == "windows" {
		binName = "ffprobe.exe"
	}

	var roots []string
	if exe, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}

	for _, root := range roots {
		p := filepath.Join(root, "build", "ffmpeg", runtime.GOOS, binName)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	if p, err := exec.LookPath("ffprobe"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("ffprobe not found")
}

// ProbeDuration 用 ffprobe 探测视频总时长（秒）；探测失败返回 0，由调用方容忍。
func ProbeDuration(videoPath string) float64 {
	bin, err := GetFFprobeBinPath()
	if err != nil {
		return 0
	}
	cmd := exec.Command(bin,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	// 同样隐藏 ffprobe 的控制台窗口
	hideCmdWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	var sec float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%f", &sec); err != nil {
		return 0
	}
	return sec
}
