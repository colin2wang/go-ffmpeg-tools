<template>
  <div class="job-card" :class="job.status">
    <div class="job-header">
      <span class="job-title" :title="job.outputVideo">{{ fileName(job.outputVideo) || t('job.unnamed') }}</span>
      <span class="job-tag" :class="job.status">{{ statusLabel(job.status) }}</span>
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
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { CancelJob } from '../../wailsjs/go/main/App'
import type { Job } from '../types/task'

const { t } = useI18n()
const props = defineProps<{ job: Job }>()

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
</style>
