<template>
  <div class="track-row">
    <span class="track-path" :title="track.file || '未选择文件'">{{ track.file || '未选择文件' }}</span>
    <select v-model="track.lang" class="lang">
      <option v-for="p in LANG_PRESETS" :key="p.code" :value="p.code">{{ p.label }} ({{ p.code }})</option>
    </select>
    <input v-model="track.title" class="title" placeholder="轨道显示名称" />
    <button class="btn btn-secondary btn-icon" :disabled="isFirst" title="上移" @click="$emit('move-up', track.id)">↑</button>
    <button class="btn btn-secondary btn-icon" :disabled="isLast" title="下移" @click="$emit('move-down', track.id)">↓</button>
    <button class="btn btn-secondary btn-icon" title="删除" @click="$emit('remove', track.id)">×</button>
  </div>
</template>

<script setup lang="ts">
import { LANG_PRESETS, type SubTrack } from '../types/task'

defineProps<{ track: SubTrack; isFirst: boolean; isLast: boolean }>()
defineEmits<{
  (e: 'remove', id: string): void
  (e: 'move-up', id: string): void
  (e: 'move-down', id: string): void
}>()
</script>

<style scoped>
.lang {
  width: 130px;
}
.title {
  width: 140px;
}
</style>
