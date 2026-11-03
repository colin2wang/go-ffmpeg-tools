<template>
  <div class="page">
    <!-- 卡片1：源视频（支持拖拽 + 选择） -->
    <div class="card">
      <div class="card-title">源视频</div>
      <div
        class="drop-zone"
        data-drop-zone="video"
        :class="{ 'drag-over': videoDragOver }"
        @click="pickVideo"
        @dragenter.prevent="videoDragOver = true"
        @dragover.prevent
        @dragleave.prevent="videoDragOver = false"
      >
        <span v-if="!inputVideo">📂 点击选择视频，或将视频拖拽到此处</span>
        <span v-else :title="inputVideo">{{ inputVideo }}</span>
      </div>
    </div>

    <!-- 卡片2：输出设置 -->
    <div class="card">
      <div class="card-title">输出设置</div>
      <div class="form-row">
        <label class="form-label">输出位置</label>
        <select v-model="outputMode" class="output-mode">
          <option value="custom">自定义文件路径</option>
          <option value="same">和输入视频同目录</option>
          <option value="exe">程序执行目录</option>
        </select>
      </div>
      <div v-if="outputMode === 'custom'" class="form-row">
        <label class="form-label">输出文件</label>
        <div class="form-input-wrap">
          <input class="input-ellipsis" readonly :value="outputVideo" :title="outputVideo" placeholder="设置输出 mp4 路径..." />
          <button class="btn btn-secondary" @click="pickOutput">选择</button>
        </div>
      </div>
      <div v-else class="form-row">
        <label class="form-label">文件名后缀</label>
        <input v-model="outputSuffix" class="output-suffix" placeholder="例如 .subbed.mp4" />
        <span class="muted-hint">自动拼接生成输出路径</span>
      </div>
    </div>

    <!-- 卡片3：字幕配置 Tab -->
    <div class="card">
      <div class="card-title">字幕配置</div>
      <div class="tab-group">
        <button class="tab-item" :class="{ active: mode === 'soft' }" @click="mode = 'soft'">
          软字幕封装（mov_text 多轨道，流复制，速度快）
        </button>
        <button class="tab-item" :class="{ active: mode === 'burn' }" @click="mode = 'burn'">
          硬字幕烧录（ass/srt 特效，需重编码）
        </button>
      </div>

      <!-- 软字幕 Tab -->
      <div v-if="mode === 'soft'" class="tab-content">
        <p class="tip-warn">⚠️ mov_text 不支持 ASS 特效，此模式只能使用 srt 文本字幕；播放器需支持切换字幕轨道。</p>
        <div
          class="drop-zone sub-drop"
          data-drop-zone="sub"
          :class="{ 'drag-over': subDragOver }"
          @dragenter.prevent="subDragOver = true"
          @dragover.prevent
          @dragleave.prevent="subDragOver = false"
        >
          📄 拖拽 SRT 字幕文件到这里，快速添加字幕轨道
        </div>
        <div class="track-list">
          <SubTrackRow
            v-for="(t, i) in tracks"
            :key="t.id"
            :track="t"
            :is-first="i === 0"
            :is-last="i === tracks.length - 1"
            @remove="removeTrack"
            @move-up="moveTrackUp"
            @move-down="moveTrackDown"
          />
        </div>
        <button class="btn btn-secondary" @click="addTrackByDialog">+ 添加字幕轨道（手动选择）</button>
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
          📄 拖拽 SRT / ASS 字幕文件到此处
        </div>
        <div class="form-row">
          <label class="form-label">字幕文件</label>
          <div class="form-input-wrap">
            <input class="input-ellipsis" readonly :value="burnSubFile" :title="burnSubFile" placeholder="选择 srt/ass 字幕" />
            <button class="btn btn-secondary" @click="pickBurnSub">选择</button>
          </div>
        </div>
        <div class="form-row">
          <label class="form-label">字体名称</label>
          <input v-model="fontName" class="font-name" placeholder="如 Microsoft YaHei" />
          <label class="form-label label-short">字体大小</label>
          <input v-model.number="fontSize" type="number" class="font-size" min="8" max="96" />
        </div>
        <div class="form-row">
          <label class="form-label">字体文件</label>
          <div class="form-input-wrap">
            <input class="input-ellipsis" readonly :value="fontFile" :title="fontFile" placeholder="可选，指定 .ttf/.otf" />
            <button class="btn btn-secondary" @click="pickFont">选择</button>
          </div>
        </div>
      </div>
    </div>

    <div class="action-bar">
      <button class="btn btn-primary" :disabled="!canSubmit || store.hasActive" @click="submit">
        {{ store.hasActive ? '处理中...' : '开始处理' }}
      </button>
    </div>

    <!-- 卡片4：任务队列 -->
    <div class="card">
      <div class="card-title">任务队列</div>
      <div v-if="store.jobList.length === 0" class="empty-tip">暂无任务，添加视频开始字幕封装</div>
      <div v-else class="job-list">
        <JobCard v-for="job in store.jobList" :key="job.jobId" :job="job" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { EventsOn, OnFileDrop } from '../../wailsjs/runtime/runtime'
import { SelectFile, SelectOutputFile, SubmitJob } from '../../wailsjs/go/main/App'
import { main, task } from '../../wailsjs/go/models'
import { useJobsStore } from '../stores/jobs'
import type { Job, OutputMode, SubMode, SubTrack } from '../types/task'
import { inferLang } from '../types/task'
import SubTrackRow from '../components/SubTrackRow.vue'
import JobCard from '../components/JobCard.vue'

const store = useJobsStore()

// ========== 拖拽悬浮标记（仅用于高亮，落点分发走 Wails OnFileDrop） ==========
const videoDragOver = ref(false)
const subDragOver = ref(false)
const burnSubDragOver = ref(false)

// ========== 表单状态 ==========
const inputVideo = ref('')
const outputMode = ref<OutputMode>('same')
const outputVideo = ref('')
const outputSuffix = ref('.subbed.mp4')
const mode = ref<SubMode>('soft')
const tracks = ref<SubTrack[]>([])
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

function handleDroppedPaths(zone: string | null, paths: string[]) {
  videoDragOver.value = false
  subDragOver.value = false
  burnSubDragOver.value = false
  if (!zone || paths.length === 0) return
  if (zone === 'video') {
    inputVideo.value = paths[0]
    if (outputMode.value !== 'custom') outputVideo.value = ''
    return
  }
  if (zone === 'sub') {
    // 批量拖拽 .srt：逐个推断语言后添加轨道
    for (const p of paths.filter((p) => /\.srt$/i.test(p))) {
      const lang = inferLang(p)
      tracks.value.push(newTrack(p, lang, lang))
    }
    return
  }
  if (zone === 'burn') {
    burnSubFile.value = paths[0]
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

// 手动添加轨道：选中文件后同样走语言推断
async function addTrackByDialog() {
  const file = await SelectFile('选择字幕文件 (srt/ass)', [
    { DisplayName: '字幕文件 (*.srt; *.ass)', Pattern: '*.srt;*.ass' },
    { DisplayName: '所有文件 (*.*)', Pattern: '*.*' },
  ])
  if (!file) return
  const lang = inferLang(file)
  tracks.value.push(newTrack(file, lang, lang))
}

// ========== 文件选择 ==========
async function pickVideo() {
  const f = await SelectFile('选择输入视频', [
    { DisplayName: '视频文件 (*.mp4; *.mkv; *.mov; *.m4v)', Pattern: '*.mp4;*.mkv;*.mov;*.m4v' },
    { DisplayName: '所有文件 (*.*)', Pattern: '*.*' },
  ])
  if (f) {
    inputVideo.value = f
    if (outputMode.value !== 'custom') {
      outputVideo.value = ''
    }
  }
}

async function pickOutput() {
  const f = await SelectOutputFile('选择输出文件', 'output.mp4')
  if (f) outputVideo.value = f
}

async function pickBurnSub() {
  const f = await SelectFile('选择字幕文件 (srt/ass)', [
    { DisplayName: '字幕文件 (*.srt; *.ass)', Pattern: '*.srt;*.ass' },
    { DisplayName: '所有文件 (*.*)', Pattern: '*.*' },
  ])
  if (f) burnSubFile.value = f
}

async function pickFont() {
  const f = await SelectFile('选择字体文件', [
    { DisplayName: '字体文件 (*.ttf; *.otf; *.ttc)', Pattern: '*.ttf;*.otf;*.ttc' },
    { DisplayName: '所有文件 (*.*)', Pattern: '*.*' },
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
  // 烧录模式只用第一条轨道
  const usableTracks =
    mode.value === 'soft'
      ? tracks.value.filter((t) => t.file)
      : [newTrack(burnSubFile.value, inferLang(burnSubFile.value), inferLang(burnSubFile.value))]
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

onMounted(() => {
  EventsOn('jobUpdate', (job: any) => {
    if (job && job.jobId) store.updateJob(job)
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
.muted-hint {
  color: var(--text-muted);
  font-size: 13px;
}
.tab-item {
  flex: 1;
}
</style>


