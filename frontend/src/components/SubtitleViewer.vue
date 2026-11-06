<template>
  <Teleport to="body">
    <Transition name="viewer">
      <div v-if="visible" class="viewer-mask" @click.self="emit('close')">
        <div class="viewer-dialog" role="dialog" aria-modal="true" :aria-label="t('viewer.title')">
          <header class="viewer-header">
            <span class="viewer-title"><Subtitles :size="16" /> {{ t('viewer.title') }}</span>
            <select v-if="streams.length > 1" v-model.number="currentIndex" class="viewer-stream-select">
              <option v-for="s in streams" :key="s.index" :value="s.index">{{ streamLabel(s) }}</option>
            </select>
            <span v-else class="viewer-stream">{{ currentStream ? streamLabel(currentStream) : '' }}</span>
            <button class="btn btn-secondary btn-icon viewer-close" :title="t('log.close')" @click="emit('close')">
              <X :size="15" />
            </button>
          </header>

          <div class="viewer-toolbar">
            <div class="viewer-search-wrap">
              <Search :size="13" />
              <input v-model="keyword" class="viewer-search" type="text" :placeholder="t('viewer.searchPlaceholder')" />
            </div>
            <span class="viewer-count">{{ t('viewer.count', { n: filtered.length, total: cues.length }) }}</span>
            <button class="btn btn-secondary viewer-btn" :disabled="!text" @click="copyAll">
              {{ copied ? t('viewer.copyOk') : t('viewer.copy') }}
            </button>
            <button class="btn btn-secondary viewer-btn" :disabled="!text" @click="exportCurrent">
              {{ t('viewer.export') }}
            </button>
          </div>

          <div class="viewer-body">
            <div v-if="loading" class="viewer-state">{{ t('viewer.loading') }}</div>
            <div v-else-if="error" class="viewer-state viewer-error">{{ error }}</div>
            <div v-else-if="isBitmap" class="viewer-state">{{ t('viewer.bitmapHint') }}</div>
            <div v-else-if="!cues.length" class="viewer-state">{{ t('viewer.empty') }}</div>
            <template v-else>
              <div v-for="(cue, i) in filtered" :key="i" class="cue-row">
                <span class="cue-time">{{ timeRange(cue.time) }}</span>
                <span class="cue-text">{{ cue.text }}</span>
              </div>
              <div v-if="!filtered.length" class="viewer-state">{{ t('viewer.noMatch') }}</div>
            </template>
          </div>

          <footer class="viewer-footer">
            <span>{{ t('viewer.closeHint') }}</span>
            <button class="btn btn-secondary viewer-btn" @click="emit('close')">{{ t('log.close') }}</button>
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Subtitles, X, Search } from 'lucide-vue-next'
import { ReadSubtitle } from '../../wailsjs/go/main/App'
import { ffmpeg } from '../../wailsjs/go/models'

const props = defineProps<{
  visible: boolean
  videoPath: string
  streams: ffmpeg.SubStream[]
  initialIndex: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'export', stream: ffmpeg.SubStream): void
}>()

const { t } = useI18n()

const BITMAP_CODECS = ['dvd_subtitle', 'hdmv_pgs_subtitle', 'dvb_subtitle', 'dvb_teletext']

interface Cue {
  time: string
  text: string
}

// ffmpeg 输出标准 SRT：块之间以空行分隔，块内依次为序号（可省略）、时间轴、文本
function parseSrt(raw: string): Cue[] {
  if (!raw.trim()) return []
  return raw
    .replace(/^\uFEFF/, '')
    .replace(/\r\n?/g, '\n')
    .trim()
    .split(/\n{2,}/)
    .map((block) => {
      const lines = block.split('\n')
      if (/^\d+$/.test(lines[0].trim())) lines.shift()
      const time = (lines[0] ?? '').includes('-->') ? lines.shift()!.trim() : ''
      // ass/ssa 转 srt 后可能残留 {\an8} 之类的样式标签
      const text = lines.join('\n').trim().replace(/\{\\[^}]*\}/g, '')
      return { time, text }
    })
    .filter((cue) => cue.text || cue.time)
}

function timeRange(time: string) {
  const parts = time.split(/\s*-->\s*/)
  if (parts.length !== 2) return time
  const short = (v: string) => v.trim().replace(/,\d+$/, '')
  return `${short(parts[0])} → ${short(parts[1])}`
}

function streamLabel(s: ffmpeg.SubStream) {
  return `#${s.index} ${s.codecName} · ${s.language || 'und'}${s.title ? ` · ${s.title}` : ''}`
}

const currentIndex = ref(props.initialIndex)
const keyword = ref('')
const copied = ref(false)
const loading = ref(false)
const error = ref('')
const text = ref('')

const currentStream = computed(
  () => props.streams.find((s) => s.index === currentIndex.value) ?? props.streams[0] ?? null,
)
const isBitmap = computed(() => BITMAP_CODECS.includes(currentStream.value?.codecName?.toLowerCase() ?? ''))
const cues = computed<Cue[]>(() => parseSrt(text.value))
const filtered = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  if (!k) return cues.value
  return cues.value.filter((c) => c.text.toLowerCase().includes(k) || c.time.toLowerCase().includes(k))
})

// 同一个视频 + 同一条流只解码一次，切换轨道来回查看时直接复用
const cache = new Map<string, string>()
let requestSeq = 0

async function load() {
  const stream = currentStream.value
  const seq = ++requestSeq
  error.value = ''
  text.value = ''
  loading.value = false
  if (!props.videoPath || !stream || isBitmap.value) return
  const key = `${props.videoPath}::${stream.index}`
  const cached = cache.get(key)
  if (cached !== undefined) {
    text.value = cached
    return
  }
  loading.value = true
  try {
    const raw = await ReadSubtitle(props.videoPath, stream.index)
    if (seq !== requestSeq) return
    cache.set(key, raw)
    text.value = raw
  } catch (e) {
    if (seq !== requestSeq) return
    error.value = `${t('viewer.loadFail')}: ${String(e)}`
  } finally {
    if (seq === requestSeq) loading.value = false
  }
}

function exportCurrent() {
  if (currentStream.value) emit('export', currentStream.value)
}

async function copyAll() {
  if (!text.value) return
  try {
    await navigator.clipboard.writeText(text.value)
    copied.value = true
    window.setTimeout(() => (copied.value = false), 1600)
  } catch {
    copied.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (props.visible && e.key === 'Escape') emit('close')
}

watch(
  () => props.visible,
  (v) => {
    if (!v) return
    currentIndex.value = props.initialIndex
    keyword.value = ''
    copied.value = false
  },
)
watch(() => props.initialIndex, (i) => {
  currentIndex.value = i
})
watch(
  () => props.videoPath,
  () => cache.clear(),
)
// 打开对话框或切换轨道/视频时重新读取；合并到一个 watch 中避免重复请求
watch(
  () => [props.visible, props.videoPath, currentIndex.value] as const,
  () => {
    if (props.visible) {
      load()
    } else {
      // 关闭时丢弃仍在途的解码结果，避免后续写回已关闭的对话框
      requestSeq++
      loading.value = false
    }
  },
)

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<style lang="scss" scoped>
.viewer-mask {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--gap-lg);
  background: rgba(0, 0, 0, 0.55);
}
.viewer-dialog {
  width: min(760px, 100%);
  height: min(640px, 82vh);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-pop);
}
.viewer-header {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-md) var(--gap-lg);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}
.viewer-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
}
.viewer-stream {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: var(--text-muted);
}
.viewer-stream-select {
  flex: 1;
  min-width: 0;
  max-width: 400px;
  height: 28px;
  padding: 0 8px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-secondary);
  font-size: 12px;
}
.viewer-close {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}
.viewer-toolbar {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-sm) var(--gap-lg);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}
.viewer-search-wrap {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-muted);
  &:focus-within { border-color: var(--primary-border); }
}
.viewer-search {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 13px;
}
.viewer-count {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
}
.viewer-btn {
  height: 28px;
  padding: 0 10px;
  font-size: 12px;
  flex-shrink: 0;
  &:disabled { opacity: 0.5; }
}
.viewer-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-sm) var(--gap-lg);
}
.viewer-state {
  padding: 40px 0;
  text-align: center;
  font-size: 13px;
  color: var(--text-muted);
}
.viewer-error {
  color: var(--danger);
  word-break: break-word;
}
.cue-row {
  display: flex;
  gap: var(--gap-md);
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  &:hover { background: var(--bg-card-hover); }
  & + & { border-top: 1px solid var(--border-subtle); }
}
.cue-time {
  flex: 0 0 148px;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 11px;
  line-height: 1.7;
  color: var(--primary);
  white-space: nowrap;
}
.cue-text {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}
.viewer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--gap-sm) var(--gap-lg);
  border-top: 1px solid var(--border-subtle);
  font-size: 12px;
  color: var(--text-muted);
  flex-shrink: 0;
}
.viewer-enter-active,
.viewer-leave-active { transition: opacity 0.18s ease; }
.viewer-enter-from,
.viewer-leave-to { opacity: 0; }
.viewer-enter-active .viewer-dialog,
.viewer-leave-active .viewer-dialog { transition: transform 0.18s ease; }
.viewer-enter-from .viewer-dialog,
.viewer-leave-to .viewer-dialog { transform: translateY(10px) scale(0.985); }
</style>
