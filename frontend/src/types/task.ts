// 单条字幕轨道（多路软字幕用）
export interface SubTrack {
  id: string
  file: string // srt 文件绝对路径
  lang: string // 语言代码：zh / eng / jpn 等，ffmpeg language tag
  title: string // 显示名称：中文 / English
}

// 字幕模式：soft = 多路 mov_text 软字幕封装（流复制，速度快）；burn = 硬字幕烧录（重编码）
export type SubMode = 'soft' | 'burn'

// 输出模式：custom = 手动选择文件；same = 和输入视频同目录；exe = 程序执行目录
export type OutputMode = 'custom' | 'same' | 'exe'

export interface SubConfig {
  mode: SubMode
  tracks: SubTrack[]
  // 硬字幕专属
  fontName?: string
  fontSize?: number
  fontFile?: string
}

// 转码配置：format 决定输出容器（扩展名）；copy 表示对应流直接复制不重编码
export interface ConvertConfig {
  format: string // mp4 / mkv / mov
  videoCodec: string // copy / libx264 / libx265
  audioCodec: string // copy / aac / ac3 / mp3
  crf?: number // 质量，仅视频重编码时生效
  scale?: string // source / 1080p / 720p / 480p
}

export interface Job {
  jobId: string
  kind: 'subtitle' | 'convert'
  inputVideo: string
  outputVideo: string
  subConfig: SubConfig
  convertConfig: ConvertConfig
  totalSeconds: number
  currentSeconds: number
  progress: number // 0~100
  status: 'pending' | 'running' | 'done' | 'error' | 'cancelled'
  log: string[]
}

export interface SubmitJobReq {
  kind: 'subtitle' | 'convert'
  inputVideo: string
  outputVideo: string
  outputMode: OutputMode
  outputSuffix: string
  subConfig: SubConfig
  convertConfig?: ConvertConfig
}

// code 为常用缩写（chs/cht 等），提交后由后端映射为 ffmpeg 的 ISO 639-2 标签
export const LANG_PRESETS: Array<{ code: string; label: string }> = [
  { code: 'chs', label: '简体中文' },
  { code: 'cht', label: '繁體中文' },
  { code: 'eng', label: 'English' },
  { code: 'jpn', label: '日本語' },
  { code: 'kor', label: '한국어' },
]

// 语言推断已移至后端 internal/ffmpeg/lang.go（ffmpeg.InferLang，整词边界匹配），
// 前端通过 App.InferSubtitleLang 调用
