<template>
  <div class="job-card" :class="job.status">
    <div class="job-header">
      <span class="job-title" :title="job.outputVideo">{{ fileName(job.outputVideo) || '未命名输出' }}</span>
      <span class="job-tag" :class="job.status">{{ statusLabel(job.status) }}</span>
      <button
        v-if="job.status === 'pending' || job.status === 'running'"
        class="btn btn-secondary btn-sm"
        @click="cancel(job.jobId)"
      >
        取消
      </button>
    </div>
    <div class="progress-bar-wrap">
      <div class="progress-bar" :class="job.status" :style="{ width: progressPct + '%' }" />
    </div>
    <div class="job-meta">
      <span>
        {{ modeLabel(job.subConfig.mode) }} · {{ fmtTime(job.currentSeconds) }} / {{ fmtTime(job.totalSeconds) }} ·
        {{ progressPct.toFixed(1) }}%
      </span>
      <button class="btn btn-secondary btn-sm" @click="showLog = !showLog">
        {{ showLog ? '收起日志' : '▶ 日志' }}
      </button>
    </div>
    <div v-if="showLog" class="log-panel">{{ job.log.join('\n') }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { CancelJob } from '../../wailsjs/go/main/App'
import type { Job } from '../types/task'

const props = defineProps<{ job: Job }>()
const showLog = ref(false)

const progressPct = computed(() => Math.max(0, Math.min(100, props.job.progress ?? 0)))

function fileName(p: string): string {
  return (p ?? '').replace(/\\/g, '/').split('/').pop() ?? ''
}

function statusLabel(s: Job['status']): string {
  const map: Record<Job['status'], string> = {
    pending: '等待中',
    running: '运行中',
    done: '已完成',
    error: '失败',
    cancelled: '已取消',
  }
  return map[s] ?? s
}

function modeLabel(m: string): string {
  return m === 'soft' ? '软字幕封装 (mov_text)' : '硬字幕烧录'
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
