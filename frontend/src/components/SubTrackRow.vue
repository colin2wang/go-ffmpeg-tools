<template>
  <div class="track-row">
    <span class="track-path" :title="track.file || t('track.noFile')">{{ track.file || t('track.noFile') }}</span>
    <select v-model="track.lang" class="lang" @change="fillTitle">
      <option v-for="p in LANG_PRESETS" :key="p.code" :value="p.code">{{ p.label }} ({{ p.code }})</option>
    </select>
    <input v-model="track.title" class="title" :placeholder="t('track.titlePlaceholder')" />
    <button class="btn btn-secondary btn-icon" :disabled="isFirst" :title="t('track.moveUp')" @click="$emit('move-up', track.id)">
      <ChevronUp :size="14" />
    </button>
    <button class="btn btn-secondary btn-icon" :disabled="isLast" :title="t('track.moveDown')" @click="$emit('move-down', track.id)">
      <ChevronDown :size="14" />
    </button>
    <button class="btn btn-secondary btn-icon" :title="t('track.remove')" @click="$emit('remove', track.id)">
      <X :size="14" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ChevronUp, ChevronDown, X } from 'lucide-vue-next'
import { LANG_PRESETS, type SubTrack } from '../types/task'

const { t } = useI18n()
const props = defineProps<{ track: SubTrack; isFirst: boolean; isLast: boolean }>()
defineEmits<{
  (e: 'remove', id: string): void
  (e: 'move-up', id: string): void
  (e: 'move-down', id: string): void
}>()

// 切换语言时自动填入轨道显示名称（简写代码，如 chs/eng）
function fillTitle(e: Event) {
  const track = props.track
  track.title = LANG_PRESETS.find((p) => p.code === (e.target as HTMLSelectElement).value)?.code ?? ''
}
</script>

<style scoped>
.lang {
  width: 130px;
}
.title {
  width: 140px;
}
</style>
