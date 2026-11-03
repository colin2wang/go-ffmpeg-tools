package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// versionRegex 从 "ffmpeg version 6.1.1-full_build ..." 第一行提取版本号
var versionRegex = regexp.MustCompile(`ffmpeg version (\S+)`)

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

// VersionInfo ffmpeg -version 输出的解析结果
type VersionInfo struct {
	Version   string `json:"version"`   // 形如 6.1.1-full_build
	FirstLine string `json:"firstLine"` // ffmpeg -version 的第一行原文
	BuiltWith string `json:"builtWith"` // "built with" 行
	Config    string `json:"config"`    // "configuration" 行
	Path      string `json:"path"`      // 实际使用的 ffmpeg 路径
	Raw       string `json:"raw"`       // 完整原始输出
}

// GetVersionInfo 运行 ffmpeg -version 并解析输出；ffmpeg 不存在时返回错误。
func GetVersionInfo() (*VersionInfo, error) {
	bin, err := GetFFmpegBinPath()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(bin, "-version")
	hideCmdWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run ffmpeg -version: %w", err)
	}

	info := &VersionInfo{Path: bin, Raw: string(out)}
	lines := strings.Split(strings.ReplaceAll(info.Raw, "\r\n", "\n"), "\n")
	for i, line := range lines {
		switch {
		case i == 0:
			info.FirstLine = line
			if m := versionRegex.FindStringSubmatch(line); m != nil {
				info.Version = m[1]
			}
		case strings.HasPrefix(line, "built with"):
			info.BuiltWith = line
		case strings.HasPrefix(line, "configuration:"):
			info.Config = line
		}
		if info.Config != "" && i > 2 {
			break
		}
	}
	return info, nil
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
