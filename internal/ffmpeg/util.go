package ffmpeg

import (
	"encoding/json"
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

// SubStream 视频内嵌字幕流信息
type SubStream struct {
	Index     int    `json:"index"`     // ffprobe 中的绝对流索引
	RelIndex  int    `json:"relIndex"`  // 在字幕流中的相对序号（-map 0:s:N 用）
	CodecName string `json:"codecName"` // subrip / mov_text / ass / hdmv_pgs_subtitle 等
	Language  string `json:"language"`  // language tag，可能为空
	Title     string `json:"title"`     // 轨道标题，可能为空
}

// ProbeSubtitles 用 ffprobe 列出视频内嵌字幕流；无字幕时返回空切片。
func ProbeSubtitles(videoPath string) ([]SubStream, error) {
	bin, err := GetFFprobeBinPath()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(bin,
		"-v", "error",
		"-select_streams", "s",
		"-show_entries", "stream=index,codec_name:stream_tags=language,title",
		"-of", "json",
		videoPath,
	)
	hideCmdWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("probe subtitle streams: %w", err)
	}
	var parsed struct {
		Streams []struct {
			Index     int    `json:"index"`
			CodecName string `json:"codec_name"`
			Tags      struct {
				Language string `json:"language"`
				Title    string `json:"title"`
			} `json:"tags"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}
	streams := make([]SubStream, 0, len(parsed.Streams))
	for i, s := range parsed.Streams {
		streams = append(streams, SubStream{
			Index:     s.Index,
			RelIndex:  i,
			CodecName: s.CodecName,
			Language:  s.Tags.Language,
			Title:     s.Tags.Title,
		})
	}
	return streams, nil
}

// ExtractSubtitle 导出第 streamIndex（绝对流索引）路内嵌字幕为 srt；
// 位图字幕（PGS / DVD / DVB 等）无法转为文本，返回错误。
func ExtractSubtitle(videoPath string, streamIndex int, outPath string) error {
	streams, err := ProbeSubtitles(videoPath)
	if err != nil {
		return err
	}
	var target *SubStream
	for i := range streams {
		if streams[i].Index == streamIndex {
			target = &streams[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("subtitle stream 0:%d not found", streamIndex)
	}
	switch strings.ToLower(target.CodecName) {
	case "dvd_subtitle", "hdmv_pgs_subtitle", "dvb_subtitle", "dvb_teletext":
		return fmt.Errorf("bitmap subtitle codec %s cannot be exported as srt", target.CodecName)
	}

	ffBin, err := GetFFmpegBinPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(ffBin, "-y",
		"-i", videoPath,
		"-map", fmt.Sprintf("0:s:%d", target.RelIndex),
		"-c:s", "srt",
		outPath,
	)
	hideCmdWindow(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extract subtitle: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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
