package ffmpeg

import (
	"fmt"
	"path/filepath"
	"strings"

	"go-ffmpeg-tools/internal/task"
)

// normalizeLang 将常用语言缩写映射为 ffmpeg 接受的 ISO 639-2 语言标签；
// 未识别的代码原样返回（ffmpeg 侧再校验）
func normalizeLang(code string) string {
	switch strings.ToLower(code) {
	case "chs", "zh", "zh-hans":
		return "chi"
	case "cht", "zh-hant", "zh-tw", "zh-hk":
		return "chi"
	default:
		return code
	}
}

// escapeFilterPath 转义 subtitles filter 中的文件路径（Windows 盘符冒号、filter 分隔符等）
func escapeFilterPath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.ReplaceAll(p, `\`, `/`)
	p = strings.ReplaceAll(p, `:`, `\:`)
	p = strings.ReplaceAll(p, `'`, `\'`)
	p = strings.ReplaceAll(p, `[`, `\[`)
	p = strings.ReplaceAll(p, `]`, `\]`)
	return p
}

// escapeForceStyle 转义 force_style 值中的引号
func escapeForceStyle(s string) string {
	s = strings.ReplaceAll(s, `'`, ``)
	return s
}

// BuildArgs 根据字幕配置构建完整 ffmpeg 命令行（含 ffmpeg 可执行文件路径）。
// soft 模式：多路 srt -> mov_text 软字幕封装，视频音频流复制不重编码；
// burn 模式：取第一条字幕轨道（ass/srt）通过 subtitles filter 硬烧录。
func BuildArgs(videoPath string, outPath string, cfg task.SubConfig) ([]string, error) {
	ffBin, err := GetFFmpegBinPath()
	if err != nil {
		return nil, err
	}
	args := []string{ffBin, "-y"}

	if cfg.Mode == "soft" {
		// mov_text 只支持简单文本字幕，轨道必须是 srt
		if len(cfg.Tracks) == 0 {
			return nil, fmt.Errorf("soft mode requires at least one subtitle track")
		}
		for _, tr := range cfg.Tracks {
			if strings.EqualFold(filepath.Ext(tr.File), ".ass") {
				return nil, fmt.Errorf("mov_text does not support ASS effects; use burn mode for %s", tr.File)
			}
		}

		args = append(args, "-i", videoPath)
		for _, tr := range cfg.Tracks {
			args = append(args, "-i", tr.File)
		}
		// 视频音频直接复制
		args = append(args, "-map", "0:v", "-map", "0:a?", "-c:v", "copy", "-c:a", "copy")
		// 逐个字幕流设置编码 + metadata：-c:s:0 mov_text -metadata:s:s:0 language=zh ...
		for idx, tr := range cfg.Tracks {
			// -c 用 s:N（s:s:N 会因流类型重复指定而报错），-metadata 用 s:s:N
			args = append(args,
				"-map", fmt.Sprintf("%d:0", idx+1),
				fmt.Sprintf("-c:s:%d", idx), "mov_text",
				fmt.Sprintf("-metadata:s:s:%d", idx), fmt.Sprintf("language=%s", normalizeLang(tr.Lang)),
				fmt.Sprintf("-metadata:s:s:%d", idx), fmt.Sprintf("title=%s", tr.Title),
			)
		}
		args = append(args, outPath)
		return args, nil
	}

	if cfg.Mode == "burn" {
		// 硬字幕：只取第一条轨道烧录，支持 ass/srt
		if len(cfg.Tracks) == 0 {
			return nil, fmt.Errorf("burn mode requires at least one subtitle")
		}
		subFile := cfg.Tracks[0].File

		styleParts := make([]string, 0, 3)
		if cfg.FontName != "" {
			styleParts = append(styleParts, fmt.Sprintf("FontName=%s", cfg.FontName))
		}
		if cfg.FontSize > 0 {
			styleParts = append(styleParts, fmt.Sprintf("FontSize=%d", cfg.FontSize))
		}
		style := strings.Join(styleParts, ",")
		// FontFile 指定字体文件时走 fontsdir 参数（force_style 不支持 FontFile）
		vf := fmt.Sprintf("subtitles=filename='%s'", escapeFilterPath(subFile))
		if cfg.FontFile != "" {
			vf += fmt.Sprintf(":fontsdir='%s'", escapeFilterPath(filepath.Dir(cfg.FontFile)))
		}
		if style != "" {
			vf += fmt.Sprintf(":force_style='%s'", escapeForceStyle(style))
		}

		args = append(args, "-i", videoPath)
		args = append(args, "-vf", vf, "-c:a", "copy")
		args = append(args, outPath)
		return args, nil
	}

	return nil, fmt.Errorf("unknown sub mode: %s", cfg.Mode)
}
