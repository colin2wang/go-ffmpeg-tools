<template>
  <div class="page">
    <!-- 卡片1：源视频（支持拖拽 + 选择） -->
    <div class="card">
      <div class="card-title">{{ t('video.title') }}</div>
      <div
        class="drop-zone video-drop"
        data-drop-zone="video"
        :class="{ 'drag-over': videoDragOver }"
        @click="pickVideo"
        @dragenter.prevent="videoDragOver = true"
        @dragover.prevent
        @dragleave.prevent="videoDragOver = false"
      >
        <template v-if="!inputVideo">
          <FileVideo :size="24" :stroke-width="1.5" />
          <span>{{ t('video.drop') }}</span>
        </template>
        <span v-else :title="inputVideo">{{ inputVideo }}</span>
      </div>
      <!-- 内嵌字幕轨道列表（探测到才显示），支持逐条导出为 srt -->
      <div v-if="subStreams.length" class="embedded-list">
        <div v-for="s in subStreams" :key="s.index" class="embedded-row">
          <span class="embedded-label" :title="`0:s:${s.relIndex}`">
            #{{ s.index }} {{ s.codecName }} · {{ s.language || 'und' }}<template v-if="s.title"> · {{ s.title }}</template>
          </span>
          <button class="btn btn-secondary btn-export" @click="exportSub(s)">{{ t('video.export') }}</button>
        </div>
      </div>
    </div>

    <!-- 卡片2：输出设置 -->
    <div class="card">
      <div class="card-title">{{ t('output.title') }}</div>
      <div class="form-row">
        <label class="form-label">{{ t('output.location') }}</label>
        <select v-model="outputMode" class="output-mode">
          <option value="custom">{{ t('output.custom') }}</option>
          <option value="same">{{ t('output.same') }}</option>
          <option value="exe">{{ t('output.exe') }}</option>
        </select>
      </div>
      <div v-if="outputMode === 'custom'" class="form-row">
        <label class="form-label">{{ t('output.file') }}</label>
        <div class="form-input-wrap">
          <input class="input-ellipsis" readonly :value="outputVideo" :title="outputVideo" :placeholder="t('output.filePlaceholder')" />
          <button class="btn btn-secondary" @click="pickOutput">{{ t('common.select') }}</button>
        </div>
      </div>
      <div v-else class="form-row">
        <label class="form-label">{{ t('output.suffix') }}</label>
        <input v-model="outputSuffix" class="output-suffix" :placeholder="defaultSuffix" />
        <span class="muted-hint">{{ t('output.autoHint') }}</span>
      </div>
    </div>

    <!-- 卡片3：字幕配置 Tab -->
    <div class="card">
      <div class="card-title">{{ t('sub.title') }}</div>
      <div class="tab-group">
        <button class="tab-item" :class="{ active: mode === 'soft' }" @click="mode = 'soft'">
          {{ t('sub.softTab') }}
        </button>
        <button class="tab-item" :class="{ active: mode === 'burn' }" @click="mode = 'burn'">
          {{ t('sub.burnTab') }}
        </button>
      </div>

      <!-- 软字幕 Tab -->
      <div v-if="mode === 'soft'" class="tab-content">
        <p class="tip-warn"><TriangleAlert :size="14" /> {{ t('sub.softWarn') }}</p>
        <div
          class="drop-zone sub-drop"
          data-drop-zone="sub"
          :class="{ 'drag-over': subDragOver }"
          @dragenter.prevent="subDragOver = true"
          @dragover.prevent
          @dragleave.prevent="subDragOver = false"
        >
          <FileText :size="14" /> {{ t('sub.softDrop') }}
        </div>
        <div class="track-list">
          <SubTrackRow
            v-for="(tr, i) in tracks"
            :key="tr.id"
            :track="tr"
            :is-first="i === 0"
            :is-last="i === tracks.length - 1"
            @remove="removeTrack"
            @move-up="moveTrackUp"
            @move-down="moveTrackDown"
          />
        </div>
        <button class="btn btn-secondary" @click="addTrackByDialog">{{ t('sub.addTrack') }}</button>
      </div>

      <!-- 硬字幕 Tab -->
      <div v-else class="tab-content">
        <div
          class="drop-zone sub-drop"
          data-drop-zone="burn"
          :class="{ 'drag-over': burnSubDragOver }"
          @dragenter.prevent="burnSubDragOver = true"
          @dragover.prevent
          @dragleave.prevent="burnSubDragOver = false"
        >
          <FileText :size="14" /> {{ t('sub.burnDrop') }}
        </div>
        <div class="form-row">
          <label class="form-label">{{ t('sub.fileLabel') }}</label>
          <div class="form-input-wrap">
            <input class="input-ellipsis" readonly :value="burnSubFile" :title="burnSubFile" :placeholder="t('sub.filePlaceholder')" />
            <button class="btn btn-secondary" @click="pickBurnSub">{{ t('common.select') }}</button>
          </div>
        </div>
        <div class="form-row">
          <label class="form-label">{{ t('sub.fontName') }}</label>
          <input v-model="fontName" class="font-name" :placeholder="t('sub.fontNamePlaceholder')" />
          <label class="form-label label-short">{{ t('sub.fontSize') }}</label>
          <input v-model.number="fontSize" type="number" class="font-size" min="8" max="96" />
        </div>
        <div class="form-row">
          <label class="form-label">{{ t('sub.fontFile') }}</label>
          <div class="form-input-wrap">
            <input class="input-ellipsis" readonly :value="fontFile" :title="fontFile" :placeholder="t('sub.fontFilePlaceholder')" />
            <button class="btn btn-secondary" @click="pickFont">{{ t('common.select') }}</button>
          </div>
        </div>
      </div>
    </div>

    <div class="action-bar">
      <button class="btn btn-primary" :disabled="!canSubmit || store.hasActive" @click="submit">
        {{ store.hasActive ? t('action.processing') : t('action.start') }}
      </button>
    </div>

    <!-- 卡片4：任务队列 -->
    <div class="card">
      <div class="card-title">{{ t('queue.title') }}</div>
      <div v-if="store.jobList.length === 0" class="empty-tip">
        <Clapperboard class="empty-icon" :size="30" :stroke-width="1.4" />
        <p>{{ t('queue.empty') }}</p>
      </div>
      <div v-else class="job-list">
        <JobCard v-for="job in store.jobList" :key="job.jobId" :job="job" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { FileVideo, FileText, TriangleAlert, Clapperboard } from 'lucide-vue-next'
import { EventsOn, OnFileDrop } from '../../wailsjs/runtime/runtime'
import { SelectFile, SelectOutputFile, SubmitJob, ProbeSubtitles, ExtractSubtitle, InferSubtitleLang } from '../../wailsjs/go/main/App'
import { main, task, ffmpeg } from '../../wailsjs/go/models'
import { useJobsStore } from '../stores/jobs'
import { openLog, logVisible } from '../composables/logPanel'
import type { Job, OutputMode, SubMode, SubTrack } from '../types/task'
import SubTrackRow from '../components/SubTrackRow.vue'
import JobCard from '../components/JobCard.vue'

const { t } = useI18n()
const store = useJobsStore()

// ========== 拖拽悬浮标记（仅用于高亮，落点分发走 Wails OnFileDrop） ==========
const videoDragOver = ref(false)
const subDragOver = ref(false)
const burnSubDragOver = ref(false)

// ========== 表单状态 ==========
const inputVideo = ref('')
const outputMode = ref<OutputMode>('same')
const outputVideo = ref('')
const outputSuffix = ref('.subbed')
const mode = ref<SubMode>('soft')
const tracks = ref<SubTrack[]>([])
// 内嵌字幕流（选定源视频后探测，0 条则不显示）
const subStreams = ref<ffmpeg.SubStream[]>([])
// 硬字幕专属
const burnSubFile = ref('')
const fontName = ref('Microsoft YaHei')
const fontSize = ref(24)
const fontFile = ref('')

let trackSeq = 0
function newTrack(file = '', lang = 'eng', title = ''): SubTrack {
  return { id: `t${++trackSeq}`, file, lang, title }
}

// ========== 拖拽事件（Wails 原生文件拖拽，OnFileDrop 返回绝对路径） ==========
// useDropTarget=false：任意位置落点都会回调，由这里自行反查所属拖放区
// （Wails 内建的 drop-target 检查只看落点元素本身样式，拖到区域内文字上会丢回调）
// 落点反查：video=源视频 / sub=软字幕轨道 / burn=硬字幕文件
function zoneAt(x: number, y: number): string | null {
  const el = document.elementFromPoint(x, y)
  return (el as HTMLElement | null)?.closest('[data-drop-zone]')?.getAttribute('data-drop-zone') ?? null
}

async function handleDroppedPaths(zone: string | null, paths: string[]) {
  videoDragOver.value = false
  subDragOver.value = false
  burnSubDragOver.value = false
  if (!zone || paths.length === 0) return
  if (zone === 'video') {
    inputVideo.value = paths[0]
    if (outputMode.value !== 'custom') outputVideo.value = ''
    probeEmbedded(paths[0])
    return
  }
  if (zone === 'sub') {
    // 批量拖拽 .srt：先填入默认的空轨道，其余逐个推断语言后新增（语言推断走后端，异步）
    for (const p of paths.filter((p) => /\.srt$/i.test(p))) {
      const lang = await InferSubtitleLang(p)
      const empty = tracks.value.find((t) => !t.file)
      if (empty) {
        empty.file = p
        empty.lang = lang
        empty.title = lang
      } else {
        tracks.value.push(newTrack(p, lang, lang))
      }
    }
    return
  }
  if (zone === 'burn') {
    burnSubFile.value = paths[0]
  }
}

// ========== 内嵌字幕探测与导出 ==========
// 选定源视频后探测内嵌字幕流；失败静默（视为无内嵌字幕）
async function probeEmbedded(video: string) {
  subStreams.value = []
  if (!video) return
  try {
    subStreams.value = (await ProbeSubtitles(video)) ?? []
  } catch {
    /* 无 ffprobe 或无字幕：忽略 */
  }
}

// 导出一条内嵌字幕：弹保存对话框（默认 <视频名>.<语言>.srt），完成后在日志面板提示
async function exportSub(s: ffmpeg.SubStream) {
  const base = decodeURIComponent(inputVideo.value.replace(/\\/g, '/').split('/').pop() ?? '').replace(/\.[^.]+$/, '')
  const out = await SelectOutputFile(t('video.exportPicker'), `${base}.${s.language || 'und'}.srt`)
  if (!out) return
  try {
    await ExtractSubtitle(inputVideo.value, s.index, out)
    openLog(t('video.embedded'), [`${t('video.exportOk')}: ${out}`])
    logVisible.value = true
  } catch (e) {
    openLog(t('video.embedded'), [`${t('video.exportFail')}: ${String(e)}`])
    logVisible.value = true
  }
}

// ========== 字幕轨道操作 ==========
function removeTrack(id: string) {
  tracks.value = tracks.value.filter((t) => t.id !== id)
}
function moveTrack(id: string, delta: -1 | 1) {
  const idx = tracks.value.findIndex((t) => t.id === id)
  const target = idx + delta
  if (idx < 0 || target < 0 || target >= tracks.value.length) return
  const [t] = tracks.value.splice(idx, 1)
  tracks.value.splice(target, 0, t)
}
function moveTrackUp(id: string) {
  moveTrack(id, -1)
}
function moveTrackDown(id: string) {
  moveTrack(id, 1)
}

// 手动添加轨道：选中文件后同样优先填入默认空轨道，再走语言推断
async function addTrackByDialog() {
  const file = await SelectFile(t('sub.pickerTitle'), [
    { DisplayName: t('sub.files'), Pattern: '*.srt;*.ass' },
    { DisplayName: t('common.allFiles'), Pattern: '*.*' },
  ])
  if (!file) return
  const lang = await InferSubtitleLang(file)
  const empty = tracks.value.find((t) => !t.file)
  if (empty) {
    empty.file = file
    empty.lang = lang
    empty.title = lang
  } else {
    tracks.value.push(newTrack(file, lang, lang))
  }
}

// 后缀只保留名字后缀部分，扩展名由后端跟随源容器决定
const defaultSuffix = computed(() => '.subbed')

// ========== 文件选择 ==========
async function pickVideo() {
  const f = await SelectFile(t('video.pickerTitle'), [
    { DisplayName: t('video.videoFiles'), Pattern: '*.mp4;*.mkv;*.mov;*.m4v' },
    { DisplayName: t('common.allFiles'), Pattern: '*.*' },
  ])
  if (f) {
    inputVideo.value = f
    probeEmbedded(f)
    if (outputMode.value !== 'custom') {
      outputVideo.value = ''
    }
  }
}

async function pickOutput() {
  const f = await SelectOutputFile(t('output.pickerTitle'), 'output.mp4')
  if (f) outputVideo.value = f
}

async function pickBurnSub() {
  const f = await SelectFile(t('sub.pickerTitle'), [
    { DisplayName: t('sub.files'), Pattern: '*.srt;*.ass' },
    { DisplayName: t('common.allFiles'), Pattern: '*.*' },
  ])
  if (f) burnSubFile.value = f
}

async function pickFont() {
  const f = await SelectFile(t('sub.fontPickerTitle'), [
    { DisplayName: t('sub.fontFiles'), Pattern: '*.ttf;*.otf;*.ttc' },
    { DisplayName: t('common.allFiles'), Pattern: '*.*' },
  ])
  if (f) fontFile.value = f
}

// ========== 任务提交 ==========
const canSubmit = computed(() => {
  if (!inputVideo.value) return false
  if (outputMode.value === 'custom' && !outputVideo.value) return false
  return mode.value === 'soft' ? tracks.value.some((t) => t.file) : !!burnSubFile.value
})

async function submit() {
  if (!canSubmit.value) return
  // 烧录模式只用第一条轨道（语言推断走后端）
  let usableTracks = tracks.value.filter((t) => t.file)
  if (mode.value !== 'soft') {
    const lang = await InferSubtitleLang(burnSubFile.value)
    usableTracks = [newTrack(burnSubFile.value, lang, lang)]
  }
  // wailsjs 生成的模型类带 convertValues 运行时方法，直接以字面量调用会被拒收，
  // 这里按生成的类型构造并做窄化断言
  const req = new main.SubmitJobReq({
    inputVideo: inputVideo.value,
    outputVideo: outputVideo.value,
    outputMode: outputMode.value,
    outputSuffix: outputSuffix.value,
    subConfig: new task.SubConfig({
      mode: mode.value,
      tracks: usableTracks.map(
        (t) =>
          new task.SubTrack({
            id: t.id,
            file: t.file,
            lang: t.lang || 'und',
            title: t.title || t.file.replace(/\\/g, '/').split('/').pop() || '',
          }),
      ),
      fontName: mode.value === 'burn' ? fontName.value : undefined,
      fontSize: mode.value === 'burn' ? fontSize.value : undefined,
      fontFile: mode.value === 'burn' ? fontFile.value : undefined,
    }),
  })
  const job = await SubmitJob(req)
  if (job) store.upsertJob(job as unknown as Job)
}

// 任务成功后清空输入：视频、输出路径与字幕轨道列表
function resetForm() {
  inputVideo.value = ''
  outputVideo.value = ''
  subStreams.value = []
  videoDragOver.value = false
  subDragOver.value = false
  burnSubDragOver.value = false
  burnSubFile.value = ''
  fontFile.value = ''
  tracks.value = [newTrack()]
}

onMounted(() => {
  EventsOn('jobUpdate', (job: any) => {
    if (job && job.jobId) {
      store.updateJob(job)
      // 日志实时流入底部日志框：每次任务更新即刷新面板内容
      const name = (job.outputVideo ?? '').replace(/\\/g, '/').split('/').pop() || t('job.unnamed')
      openLog(name, job.log ?? [])
      // 任务成功后自动清空表单，方便处理下一个视频
      if (job.status === 'done') resetForm()
    }
  })
  // Wails 原生拖拽：关闭内建 drop-target 过滤，自行按坐标反查拖放区
  OnFileDrop((x, y, paths) => {
    handleDroppedPaths(zoneAt(x, y), paths)
  }, false)
  tracks.value.push(newTrack())
})
</script>

<style lang="scss" scoped>
.page {
  max-width: 860px;
  margin: 0 auto;
}
.output-mode {
  width: 220px;
}
.output-suffix {
  width: 220px;
}
.font-name {
  width: 200px;
}
.font-size {
  width: 80px;
}
.label-short {
  width: 70px;
  margin-left: var(--gap-sm);
}
.sub-drop {
  margin-bottom: var(--gap-md);
}
.embedded-list {
  margin-top: var(--gap-sm);
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.embedded-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--gap-sm);
}
.embedded-label {
  font-size: 12px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.btn-export {
  flex-shrink: 0;
  font-size: 12px;
  padding: 2px 10px;
}
.muted-hint {
  color: var(--text-muted);
  font-size: 13px;
}
.tab-item {
  flex: 1;
}
.video-drop {
  min-height: 88px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
</style>
