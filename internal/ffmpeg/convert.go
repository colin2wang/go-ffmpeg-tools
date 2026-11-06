package ffmpeg

import (
	"fmt"
	"strconv"

	"go-ffmpeg-tools/internal/task"
)

// 支持的输出容器（决定输出扩展名与封装格式）
var convertFormats = map[string]bool{
	".mp4": true,
	".mkv": true,
	".mov": true,
}

// 支持的视频编码；copy 表示直接复制视频流不重编码
var convertVideoCodecs = map[string]bool{
	"copy":    true,
	"libx264": true,
	"libx265": true,
}

// 支持的音频编码；copy 表示直接复制音频流不重编码
var convertAudioCodecs = map[string]bool{
	"copy": true,
	"aac":  true,
	"ac3":  true,
	"mp3":  true,
}

// scaleArgs 将分辨率预设转换为 scale 滤镜参数：
// 宽度按 16 对齐（-2）自动计算，保持源宽高比
func scaleArgs(scale string) (string, error) {
	switch scale {
	case "", "source":
		return "", nil
	case "1080p":
		return "scale=-2:1080", nil
	case "720p":
		return "scale=-2:720", nil
	case "480p":
		return "scale=-2:480", nil
	default:
		return "", fmt.Errorf("unsupported scale preset: %s", scale)
	}
}

// ResolveConvertFormat 校验并返回规范化输出扩展名（含点），如 ".mp4"
func ResolveConvertFormat(format string) (string, error) {
	ext := "." + format
	if !convertFormats[ext] {
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
	return ext, nil
}

// BuildConvertArgs 根据转码配置构建完整 ffmpeg 命令行（含 ffmpeg 可执行文件路径）。
// 视频流 copy 时忽略 CRF 与分辨率（复制不重编码，两者均无意义）。
func BuildConvertArgs(videoPath string, outPath string, cfg task.ConvertConfig) ([]string, error) {
	if !convertVideoCodecs[cfg.VideoCodec] {
		return nil, fmt.Errorf("unsupported video codec: %s", cfg.VideoCodec)
	}
	if !convertAudioCodecs[cfg.AudioCodec] {
		return nil, fmt.Errorf("unsupported audio codec: %s", cfg.AudioCodec)
	}

	ffBin, err := GetFFmpegBinPath()
	if err != nil {
		return nil, err
	}
	args := []string{ffBin, "-y", "-i", videoPath}

	// 视频流：copy 或 libx264/libx265（CRF 质量 + 可选缩放）
	if cfg.VideoCodec == "copy" {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args, "-c:v", cfg.VideoCodec)
		if vf, err := scaleArgs(cfg.Scale); err != nil {
			return nil, err
		} else if vf != "" {
			args = append(args, "-vf", vf)
		}
		if cfg.Crf > 0 {
			// x264 CRF 合理区间 0~51，越低质量越高；这里限制到常用区间避免误填
			crf := cfg.Crf
			if crf < 14 {
				crf = 14
			} else if crf > 32 {
				crf = 32
			}
			args = append(args, "-crf", strconv.Itoa(crf))
		}
	}

	// 音频流：copy 或指定编码
	args = append(args, "-c:a", cfg.AudioCodec)

	// mp4 移动端友好：moov 前置，边下边播
	if outPath[len(outPath)-4:] == ".mp4" {
		args = append(args, "-movflags", "+faststart")
	}

	args = append(args, outPath)
	return args, nil
}
