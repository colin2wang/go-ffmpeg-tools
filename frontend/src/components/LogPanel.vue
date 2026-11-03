<template>
  <div v-if="logVisible" class="log-window">
    <div class="log-window-header">
      <span class="log-window-title">📄 {{ logTitle }}</span>
      <div class="log-window-actions">
        <label class="log-auto-popup">
          <input v-model="autoPopup" type="checkbox" /> {{ t('log.autoPopup') }}
        </label>
        <button class="btn btn-secondary btn-sm" @click="clearLog">{{ t('log.clear') }}</button>
        <button class="btn btn-secondary btn-sm" @click="closeLog">{{ t('log.close') }}</button>
      </div>
    </div>
    <div class="log-window-body">{{ logLines.join('\n') }}</div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { logVisible, logTitle, logLines, autoPopup, closeLog, clearLog } from '../composables/logPanel'

const { t } = useI18n()
</script>

<style scoped>
.log-window {
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-card);
  display: flex;
  flex-direction: column;
  height: 180px;
  flex-shrink: 0;
  transition: background 0.25s ease;
}
.log-window-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--gap-xs) var(--gap-md);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}
.log-window-title {
  font-size: 13px;
  color: var(--text-secondary);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.log-window-actions {
  display: flex;
  align-items: center;
  gap: var(--gap-xs);
  flex-shrink: 0;
}
.log-auto-popup {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  user-select: none;
}
.log-window-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-sm) var(--gap-md);
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
