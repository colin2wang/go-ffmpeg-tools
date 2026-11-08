<template>
  <div class="job-card" :class="job.status">
    <div class="job-header">
      <span class="job-title" :title="job.outputVideo">{{ fileName(job.outputVideo) || t('job.unnamed') }}</span>
      <span class="job-tag" :class="job.status">{{ statusLabel(job.status) }}</span>
      <button
        v-if="job.log && job.log.length > 0"
        class="btn btn-secondary btn-sm"
        @click="consoleOpen = true"
      >
        {{ t('job.console') }}
      </button>
      <button
        v-if="job.status === 'pending' || job.status === 'running'"
        class="btn btn-secondary btn-sm"
        @click="cancel(job.jobId)"
      >
        {{ t('job.cancel') }}
      </button>
    </div>
    <div class="progress-bar-wrap">
      <div class="progress-bar" :class="job.status" :style="{ width: progressPct + '%' }" />
    </div>
    <div class="job-meta">
      <span>
        {{ modeLabel(job) }} · {{ fmtTime(job.currentSeconds) }} / {{ fmtTime(job.totalSeconds) }} ·
        {{ progressPct.toFixed(1) }}%
      </span>
    </div>

    <!-- FFmpeg 控制台对话框：实时显示该任务 stderr 日志 -->
    <Teleport to="body">
      <div v-if="consoleOpen" class="console-mask" @click.self="consoleOpen = false">
        <div class="console-dialog" role="dialog" aria-modal="true" :aria-label="t('job.consoleTitle')">
          <header class="console-header">
            <span class="console-title"><Terminal :size="15" /> {{ t('job.consoleTitle') }}</span>
            <button class="btn btn-secondary btn-sm" @click="consoleOpen = false">{{ t('log.close') }}</button>
          </header>
          <pre ref="logEl" class="console-body">{{ consoleText }}</pre>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Terminal } from 'lucide-vue-next'
import { CancelJob } from '../../wailsjs/go/main/App'
import type { Job } from '../types/task'

const { t } = useI18n()
const props = defineProps<{ job: Job }>()

const consoleOpen = ref(false)
const logEl = ref<HTMLElement | null>(null)

const consoleText = computed(() => (props.job.log ?? []).join('\n'))

// 对话框打开时跟随新日志滚动到底部
watch(
  () => [consoleOpen.value, props.job.log?.length] as const,
  ([open]) => {
    if (!open) return
    nextTick(() => {
      if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight
    })
  },
)

const progressPct = computed(() => Math.max(0, Math.min(100, props.job.progress ?? 0)))

function fileName(p: string): string {
  return (p ?? '').replace(/\\/g, '/').split('/').pop() ?? ''
}

function statusLabel(s: Job['status']): string {
  return t('job.' + s)
}

function modeLabel(job: Job): string {
  if (job.kind === 'convert') {
    const c = job.convertConfig
    const parts = [c.format?.toUpperCase(), c.videoCodec === 'copy' ? t('convert.codecCopy') : c.videoCodec]
    return `${t('job.convertMode')} · ${parts.filter(Boolean).join(' / ')}`
  }
  return job.subConfig.mode === 'soft' ? t('job.softMode') : t('job.burnMode')
}

function fmtTime(sec: number): string {
  const s = Math.floor(sec ?? 0)
  const h = String(Math.floor(s / 3600)).padStart(2, '0')
  const m = String(Math.floor((s % 3600) / 60)).padStart(2, '0')
  const ss = String(s % 60).padStart(2, '0')
  return `${h}:${m}:${ss}`
}

function cancel(jobId: string) {
  CancelJob(jobId)
}
</script>

<style scoped>
.btn-sm {
  height: 24px;
  padding: 0 10px;
  font-size: 12px;
}

.console-mask {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--gap-lg);
  background: rgba(0, 0, 0, 0.55);
}
.console-dialog {
  width: min(860px, 100%);
  height: min(760px, 90vh);
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}
.console-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--gap-sm) var(--gap-md);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}
.console-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}
.console-body {
  flex: 1;
  margin: 0;
  overflow: auto;
  padding: var(--gap-sm) var(--gap-md);
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-secondary);
}
</style>
