<template>
  <div v-if="dockOpen" class="bottom-dock">
    <div class="dock-header">
      <!-- 标签：任务队列 / 日志 -->
      <div class="dock-tabs">
        <button class="dock-tab" :class="{ active: dockTab === 'queue' }" @click="dockTab = 'queue'">
          <Clapperboard :size="13" /> {{ t('queue.title') }}
          <span v-if="runningCount > 0" class="dock-badge">{{ runningCount }}</span>
        </button>
        <button class="dock-tab" :class="{ active: dockTab === 'log' }" @click="dockTab = 'log'">
          <FileText :size="13" /> {{ t('log.title') }}
        </button>
      </div>
      <div class="dock-actions">
        <!-- 统一只有一个清空按钮：队列标签清空任务，日志标签清空日志 -->
        <button class="btn btn-secondary btn-sm" @click="dockTab === 'queue' ? store.clearJobs() : clearLog()">
          {{ dockTab === 'queue' ? t('queue.clear') : t('log.clear') }}
        </button>
      </div>
    </div>

    <!-- 统一内容区：队列标签 = 任务卡片列表，日志标签 = 日志文本 -->
    <div class="dock-body">
      <div v-if="dockTab === 'queue'" class="dock-queue">
        <div v-if="store.jobList.length === 0" class="empty-tip">
          <Clapperboard class="empty-icon" :size="30" :stroke-width="1.4" />
          <p>{{ t('queue.empty') }}</p>
        </div>
        <div v-else class="job-list">
          <JobCard v-for="job in store.jobList" :key="job.jobId" :job="job" />
        </div>
      </div>
      <div v-else class="dock-log">{{ logLines.join('\n') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Clapperboard, FileText } from 'lucide-vue-next'
import { dockOpen, dockTab, logLines, clearLog } from '../composables/bottomDock'
import { useJobsStore } from '../stores/jobs'
import JobCard from './JobCard.vue'

const { t } = useI18n()
const store = useJobsStore()

const runningCount = computed(() => store.jobList.filter((j) => j.status === 'running').length)
</script>

<style lang="scss" scoped>
.bottom-dock {
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-card);
  display: flex;
  flex-direction: column;
  height: 240px;
  flex-shrink: 0;
  transition: background 0.25s ease;
}
.dock-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--gap-xs) var(--gap-md);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}
.dock-tabs {
  display: flex;
  align-items: center;
  gap: var(--gap-xs);
}
.dock-tab {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 26px;
  padding: 0 12px;
  font-size: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
  &:hover {
    color: var(--text-primary);
    border-color: var(--border-hover);
  }
  &.active {
    background: var(--bg-hover);
    border-color: var(--primary-border);
    color: var(--primary);
  }
}
.dock-badge {
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--primary);
  color: #fff;
  font-size: 11px;
}
.dock-actions {
  display: flex;
  align-items: center;
  gap: var(--gap-xs);
  flex-shrink: 0;
}
.dock-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-sm) var(--gap-md);
}
.dock-log {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-secondary);
}
.btn-sm {
  height: 22px;
  padding: 0 8px;
  font-size: 12px;
}
</style>
