<template>
  <div class="app-shell">
    <!-- 左侧菜单 -->
    <aside class="sidebar">
      <div class="app-brand">FFMpeg Tools</div>
      <nav class="menu">
        <button
          v-for="item in menuItems"
          :key="item.key"
          class="menu-item"
          :class="{ active: activeMenu === item.key, disabled: item.disabled }"
          :title="item.disabled ? '开发中' : item.label"
          :disabled="item.disabled"
          @click="!item.disabled && (activeMenu = item.key)"
        >
          <span class="menu-icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </button>
      </nav>
      <div class="sidebar-footer">
        <button class="btn btn-secondary btn-icon" :title="isDark ? '浅色模式' : '深色模式'" @click="toggleTheme">
          {{ isDark ? '☀' : '🌙' }}
        </button>
      </div>
    </aside>

    <!-- 内容区 -->
    <main class="content">
      <Index v-if="activeMenu === 'subtitle'" />
      <div v-else class="placeholder">功能开发中…</div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Index from './views/Index.vue'
import { useTheme } from './composables/theme'

// 挂载前先应用主题，避免闪烁
onMounted(() => {
  useTheme()
})

interface MenuItem {
  key: string
  label: string
  icon: string
  disabled?: boolean
}

const menuItems: MenuItem[] = [
  { key: 'subtitle', label: '字幕工具', icon: '🎬' },
  { key: 'convert', label: '转码工具', icon: '🎞', disabled: true },
  { key: 'info', label: '媒体信息', icon: 'ℹ️', disabled: true },
]

const activeMenu = ref('subtitle')
const { isDark, toggleTheme } = useTheme()
</script>

<style lang="scss">
@import './styles/app.scss';
</style>
