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

export interface Job {
  jobId: string
  inputVideo: string
  outputVideo: string
  subConfig: SubConfig
  totalSeconds: number
  currentSeconds: number
  progress: number // 0~100
  status: 'pending' | 'running' | 'done' | 'error' | 'cancelled'
  log: string[]
}

export interface SubmitJobReq {
  inputVideo: string
  outputVideo: string
  outputMode: OutputMode
  outputSuffix: string
  subConfig: SubConfig
}

// code 为常用缩写（chs/cht 等），提交后由后端映射为 ffmpeg 的 ISO 639-2 标签
export const LANG_PRESETS: Array<{ code: string; label: string }> = [
  { code: 'chs', label: '简体中文' },
  { code: 'cht', label: '繁體中文' },
  { code: 'eng', label: 'English' },
  { code: 'jpn', label: '日本語' },
  { code: 'kor', label: '한국어' },
]

// 从字幕文件名推断语言代码，失败返回默认值 eng
export function inferLang(file: string): string {
  const name = decodeURIComponent(file.replace(/\\/g, '/').split('/').pop() ?? '').toLowerCase()
  if (/(chs|zh-hans|zh-cn|简体|简中|sc)|chinese.?\(simplified\)|(?<![a-z])gb(?![a-z0-9])/.test(name)) return 'chs'
  if (/(cht|zh-hant|zh-tw|zh-hk|繁体|繁體|繁中|tc|big5)|chinese.?\(traditional\)/.test(name)) return 'cht'
  if (/(eng|english|英文|英语)|english.?\(united states\)/.test(name)) return 'eng'
  if (/(jpn|japanese|日文|日语|日本語)/.test(name)) return 'jpn'
  if (/(kor|korean|韩文|韩语|韓文|한국)/.test(name)) return 'kor'
  return 'eng'
}
