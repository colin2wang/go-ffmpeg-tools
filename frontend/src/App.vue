<template>
  <div class="app-shell">
    <!-- 顶部导航栏：品牌 + 语言/主题切换 -->
    <header class="navbar">
      <div class="navbar-brand">FFMpeg Tools</div>
      <div class="navbar-actions">
        <select
          class="lang-switch"
          :value="locale"
          @change="switchLocale(($event.target as HTMLSelectElement).value as Locale)"
        >
          <option v-for="opt in LOCALE_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>
        <button class="btn btn-secondary btn-icon" :title="isDark ? t('app.lightMode') : t('app.darkMode')" @click="toggleTheme">
          {{ isDark ? '☀' : '🌙' }}
        </button>
      </div>
    </header>

    <div class="app-body">
      <!-- 左侧菜单 -->
      <aside class="sidebar">
        <nav class="menu">
          <button
            v-for="item in menuItems"
            :key="item.key"
            class="menu-item"
            :class="{ active: activeMenu === item.key, disabled: item.disabled }"
            :title="item.disabled ? t('common.developing') : t('menu.' + item.key)"
            :disabled="item.disabled"
            @click="!item.disabled && (activeMenu = item.key)"
          >
            <span class="menu-icon">{{ item.icon }}</span>
            <span>{{ t('menu.' + item.key) }}</span>
          </button>
        </nav>
      </aside>

      <!-- 内容区 -->
      <main class="content">
        <Index v-if="activeMenu === 'subtitle'" />
        <Version v-else-if="activeMenu === 'info'" />
        <div v-else class="placeholder">{{ t('app.placeholder') }}</div>
      </main>
    </div>

    <!-- 底部固定日志窗口（切换菜单保持不变） -->
    <LogPanel />

    <!-- 底部状态栏：左侧状态 + 日志开关按钮，右侧总数 -->
    <footer class="status-bar">
      <span>{{ runningCount > 0 ? t('status.running', { n: runningCount }) : t('status.ready') }}</span>
      <button
        class="btn-log-toggle"
        :class="{ open: logVisible }"
        :title="logVisible ? t('log.collapse') : t('log.expand')"
        @click="toggleLog"
      >
        <span class="chevron">{{ logVisible ? '▾' : '▴' }}</span> {{ t('log.title') }}
      </button>
      <span>{{ t('status.total', { n: store.jobList.length }) }}</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Index from './views/Index.vue'
import Version from './views/Version.vue'
import LogPanel from './components/LogPanel.vue'
import { useTheme } from './composables/theme'
import { logVisible, toggleLog } from './composables/logPanel'
import { LOCALE_OPTIONS, saveLocale, type Locale } from './i18n'
import { useJobsStore } from './stores/jobs'

// 挂载前先应用主题，避免闪烁
onMounted(() => {
  useTheme()
})

interface MenuItem {
  key: string
  icon: string
  disabled?: boolean
}

const menuItems: MenuItem[] = [
  { key: 'subtitle', icon: '🎬' },
  { key: 'convert', icon: '🎞', disabled: true },
  { key: 'info', icon: 'ℹ️' },
]

const { t, locale } = useI18n()
const store = useJobsStore()
const activeMenu = ref('subtitle')
const { isDark, toggleTheme } = useTheme()

const runningCount = computed(() => store.jobList.filter((j) => j.status === 'running').length)

function switchLocale(value: Locale) {
  locale.value = value
  saveLocale(value)
}
</script>

<style lang="scss">
@import './styles/app.scss';
</style>
