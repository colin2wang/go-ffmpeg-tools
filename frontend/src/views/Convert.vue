<template>
  <div class="page">
    <!-- 卡片1：源视频（支持拖拽 + 选择） -->
    <div class="card">
      <div class="card-title">{{ t('video.title') }}</div>
      <div
        class="drop-zone video-drop"
        data-drop-zone="convert-video"
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
        <span class="muted-hint">{{ t('convert.autoHint') }}</span>
      </div>
    </div>

    <!-- 卡片3：转码设置 -->
    <div class="card">
      <div class="card-title">{{ t('convert.title') }}</div>
      <div class="form-row">
        <label class="form-label">{{ t('convert.format') }}</label>
        <select v-model="format" class="convert-select">
          <option value="mp4">MP4</option>
          <option value="mkv">MKV</option>
          <option value="mov">MOV</option>
        </select>
      </div>
      <div class="form-row">
        <label class="form-label">{{ t('convert.videoCodec') }}</label>
        <select v-model="videoCodec" class="convert-select">
          <option value="copy">{{ t('convert.codecCopy') }}</option>
          <option value="libx264">H.264 (libx264)</option>
          <option value="libx265">H.265 (libx265)</option>
        </select>
      </div>
      <div class="form-row">
        <label class="form-label">{{ t('convert.audioCodec') }}</label>
        <select v-model="audioCodec" class="convert-select">
          <option value="copy">{{ t('convert.codecCopy') }}</option>
          <option value="aac">AAC</option>
          <option value="ac3">AC3</option>
          <option value="mp3">MP3</option>
        </select>
      </div>
      <div class="form-row">
        <label class="form-label">{{ t('convert.quality') }}</label>
        <select v-model="quality" class="convert-select" :disabled="videoCodec === 'copy'">
          <option value="18">{{ t('convert.qualityHigh') }}</option>
          <option value="23">{{ t('convert.qualityStandard') }}</option>
          <option value="28">{{ t('convert.qualityCompact') }}</option>
        </select>
        <span class="muted-hint">{{ videoCodec === 'copy' ? t('convert.copyHint') : t('convert.qualityHint') }}</span>
      </div>
      <div class="form-row">
        <label class="form-label">{{ t('convert.resolution') }}</label>
        <select v-model="scale" class="convert-select" :disabled="videoCodec === 'copy'">
          <option value="source">{{ t('convert.resSource') }}</option>
          <option value="1080p">1080p</option>
          <option value="720p">720p</option>
          <option value="480p">480p</option>
        </select>
      </div>
    </div>

    <div class="action-bar">
      <button class="btn btn-primary" :disabled="!canSubmit || store.hasActive" @click="submit">
        {{ store.hasActive ? t('action.processing') : t('action.start') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { FileVideo } from 'lucide-vue-next'
import { SelectFile, SelectOutputFile, SubmitJob } from '../../wailsjs/go/main/App'
import { main, task } from '../../wailsjs/go/models'
import { useJobsStore } from '../stores/jobs'
import { registerDropZone } from '../composables/fileDrop'
import type { Job, OutputMode } from '../types/task'

const { t } = useI18n()
const store = useJobsStore()

// ========== 拖拽悬浮标记（仅用于高亮，落点分发走全局 fileDrop） ==========
const videoDragOver = ref(false)

// ========== 表单状态 ==========
const inputVideo = ref('')
const outputMode = ref<OutputMode>('same')
const outputVideo = ref('')
const outputSuffix = ref('.conv')
// 转码参数
const format = ref('mp4')
const videoCodec = ref('copy')
const audioCodec = ref('copy')
const quality = ref('23')
const scale = ref('source')

// 拖拽 / 选择源视频后统一走这里
function setVideo(path: string) {
  inputVideo.value = path
  if (outputMode.value !== 'custom') outputVideo.value = ''
}

// ========== 文件选择 ==========
async function pickVideo() {
  const f = await SelectFile(t('video.pickerTitle'), [
    { DisplayName: t('video.videoFiles'), Pattern: '*.mp4;*.mkv;*.mov;*.m4v' },
    { DisplayName: t('common.allFiles'), Pattern: '*.*' },
  ])
  if (f) setVideo(f)
}

async function pickOutput() {
  // 保存对话框的默认文件名跟随当前所选容器格式
  const f = await SelectOutputFile(t('output.pickerTitle'), `output.${format.value}`)
  if (f) outputVideo.value = f
}

// 后缀只保留名字后缀部分，扩展名由后端按所选容器决定
const defaultSuffix = computed(() => '.conv')

// ========== 任务提交 ==========
const canSubmit = computed(() => {
  if (!inputVideo.value) return false
  if (outputMode.value === 'custom' && !outputVideo.value) return false
  return true
})

async function submit() {
  if (!canSubmit.value) return
  // wailsjs 生成的模型类带 convertValues 运行时方法，直接以字面量调用会被拒收，
  // 这里按生成的类型构造
  const req = new main.SubmitJobReq({
    kind: 'convert',
    inputVideo: inputVideo.value,
    outputVideo: outputVideo.value,
    outputMode: outputMode.value,
    outputSuffix: outputSuffix.value,
    convertConfig: new task.ConvertConfig({
      format: format.value,
      videoCodec: videoCodec.value,
      audioCodec: audioCodec.value,
      // 视频流复制时质量/分辨率无效，不提交
      crf: videoCodec.value === 'copy' ? undefined : Number(quality.value),
      scale: videoCodec.value === 'copy' ? undefined : scale.value,
    }),
  })
  const job = await SubmitJob(req)
  if (job) store.upsertJob(job as unknown as Job)
}

// 任务成功后清空输入
function resetForm() {
  inputVideo.value = ''
  outputVideo.value = ''
  videoDragOver.value = false
}

onMounted(() => {
  registerDropZone('convert-video', (paths) => {
    videoDragOver.value = false
    if (paths.length > 0) setVideo(paths[0])
  })
  store.registerDoneHandler('convert', resetForm)
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
.convert-select {
  width: 220px;
}
.muted-hint {
  color: var(--text-muted);
  font-size: 13px;
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
