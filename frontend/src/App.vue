<template>
  <div class="app-shell">
    <!-- 顶部导航栏：品牌 + 语言/主题切换 -->
    <header class="navbar">
      <div class="navbar-brand">FFMpeg Tools</div>
      <div class="navbar-actions">
        <span class="lang-icon" :title="t('app.language')"><Languages :size="16" :stroke-width="1.75" /></span>
        <select
          class="lang-switch"
          :value="locale"
          @change="switchLocale(($event.target as HTMLSelectElement).value as Locale)"
        >
          <option v-for="opt in LOCALE_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>
        <button class="btn btn-secondary btn-icon" :title="isDark ? t('app.lightMode') : t('app.darkMode')" @click="toggleTheme">
          <component :is="isDark ? Sun : Moon" :size="16" :stroke-width="1.75" />
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
            <span class="menu-icon"><component :is="item.icon" :size="16" :stroke-width="1.75" /></span>
            <span>{{ t('menu.' + item.key) }}</span>
          </button>
        </nav>
      </aside>

      <!-- 内容区 -->
      <main class="content">
        <!-- 字幕工具页用 v-show 常驻：表单状态（视频/字幕路径）跨页面保留，
             且 onMounted 里的 OnFileDrop 只注册一次，避免切页后旧回调闭包残留 -->
        <Index v-show="activeMenu === 'subtitle'" />
        <Version v-if="activeMenu === 'info'" />
        <div v-if="activeMenu === 'convert'" class="placeholder">{{ t('app.placeholder') }}</div>
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
        <component :is="logVisible ? ChevronDown : ChevronUp" :size="12" /> {{ t('log.title') }}
      </button>
      <span>{{ t('status.total', { n: store.jobList.length }) }}</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { Subtitles, Repeat, Info, Sun, Moon, ChevronDown, ChevronUp, Languages } from 'lucide-vue-next'
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
  icon: Component
  disabled?: boolean
}

// emoji 字符串 → 图标组件
const menuItems: MenuItem[] = [
  { key: 'subtitle', icon: Subtitles },
  { key: 'convert', icon: Repeat, disabled: true },
  { key: 'info', icon: Info },
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

// 语言切换图标：与下拉框垂直居中
.lang-icon {
  display: inline-flex;
  align-items: center;
  color: var(--text-secondary);
}
</style>
