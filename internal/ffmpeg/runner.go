package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

var timeRegex = regexp.MustCompile(`time=(\d+):(\d+):(\d+)\.(\d+)`)

// scanLinesCR 按 \r 或 \n 分行：ffmpeg 用 \r 原地刷新进度行（不换行），
// 默认的 ScanLines 只按 \n 切分，会导致进度输出直到进程结束才被读到。
func scanLinesCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		advance = i + 1
		if data[i] == '\r' && advance < len(data) && data[advance] == '\n' {
			advance++
		}
		return advance, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// ProgressCallback 进度回调：currentSec > 0 表示时间推进；logLine 非空表示一行 stderr 日志
type ProgressCallback func(currentSec float64, logLine string)

// RunJob 启动 ffmpeg 进程，解析 stderr 中的 time= 进度并通过 cb 回报。
// ctx 取消时 kill 进程并返回 cancelled 错误。
func RunJob(ctx context.Context, args []string, totalSec float64, cb ProgressCallback) error {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	// GUI 模式下不弹出 ffmpeg 控制台黑框（Windows）
	hideCmdWindow(cmd)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stderr)
	// ffmpeg 单行输出可能很长（进度行、filter 描述），放大缓冲上限
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	scanner.Split(scanLinesCR)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		cb(0, line)
		if m := timeRegex.FindStringSubmatch(line); m != nil {
			h, _ := strconv.Atoi(m[1])
			mi, _ := strconv.Atoi(m[2])
			s, _ := strconv.Atoi(m[3])
			centis, _ := strconv.Atoi(m[4])
			current := float64(h*3600+mi*60+s) + float64(centis)/100
			cb(current, "")
		}
	}

	waitErr := cmd.Wait()
	if ctx.Err() == context.Canceled {
		return fmt.Errorf("cancelled")
	}
	if scanErr := scanner.Err(); scanErr != nil && waitErr == nil {
		return scanErr
	}
	return waitErr
}
