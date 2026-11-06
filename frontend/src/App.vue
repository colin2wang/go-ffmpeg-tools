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

      <!-- 内容区：滚动容器，供回到顶部按钮定位与滚动 -->
      <main ref="contentEl" class="content" @scroll.passive="onContentScroll">
        <!-- 字幕工具页用 v-show 常驻：表单状态（视频/字幕路径）跨页面保留，
             且 onMounted 里的 OnFileDrop 只注册一次，避免切页后旧回调闭包残留 -->
        <Index v-show="activeMenu === 'subtitle'" />
        <Convert v-show="activeMenu === 'convert'" />
        <Version v-if="activeMenu === 'info'" />
      </main>
    </div>

    <!-- 回到顶部：悬浮于内容区右下角，滚动超过一屏的 1/3 时出现 -->
    <Transition name="btt">
      <button
        v-if="showBackTop"
        class="back-to-top"
        :class="{ 'above-dock': dockOpen }"
        :title="t('app.backToTop')"
        @click="scrollToTop"
      >
        <ArrowUp :size="16" :stroke-width="1.75" />
      </button>
    </Transition>

    <!-- 底部固定抽屉：任务队列 / 日志 标签切换（切换菜单保持不变） -->
    <BottomDock />

    <!-- 底部状态栏：左侧状态 + 抽屉展开/隐藏按钮，右侧总数 -->
    <footer class="status-bar">
      <div class="status-left">
        <span>{{ runningCount > 0 ? t('status.running', { n: runningCount }) : t('status.ready') }}</span>
        <button
          class="btn-log-toggle"
          :title="dockOpen ? t('dock.hide') : t('dock.expand')"
          @click="dockOpen = !dockOpen"
        >
          <component :is="dockOpen ? ChevronDown : ChevronUp" :size="12" />
          {{ dockOpen ? t('dock.hide') : t('dock.expand') }}
        </button>
      </div>
      <span>{{ t('status.total', { n: store.jobList.length }) }}</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { Subtitles, Repeat, Info, Sun, Moon, ChevronDown, ChevronUp, Languages, ArrowUp } from 'lucide-vue-next'
import { EventsOn, OnFileDrop } from '../wailsjs/runtime/runtime'
import Index from './views/Index.vue'
import Convert from './views/Convert.vue'
import Version from './views/Version.vue'
import BottomDock from './components/BottomDock.vue'
import { useTheme } from './composables/theme'
import { dockOpen, openLog } from './composables/bottomDock'
import { zoneAt, dispatchDrop } from './composables/fileDrop'
import { LOCALE_OPTIONS, saveLocale, type Locale } from './i18n'
import { useJobsStore } from './stores/jobs'

// 挂载前先应用主题；同时注册全局任务事件与文件拖拽分发
onMounted(() => {
  useTheme()
  // 全局任务事件：更新 store + 日志流；任务完成按类型分发给对应页面的回调（清空表单）
  EventsOn('jobUpdate', (job: any) => {
    if (job && job.jobId) {
      store.updateJob(job)
      openLog(job.log ?? [])
      if (job.status === 'done') store.doneHandlers[job.kind]?.()
    }
  })
  // Wails 原生拖拽：全局只注册一次，按坐标反查 data-drop-zone 后分发给各页面处理器
  OnFileDrop((x: number, y: number, paths: string[]) => {
    dispatchDrop(zoneAt(x, y), paths)
  }, false)
})

interface MenuItem {
  key: string
  icon: Component
  disabled?: boolean
}

// emoji 字符串 → 图标组件
const menuItems: MenuItem[] = [
  { key: 'subtitle', icon: Subtitles },
  { key: 'convert', icon: Repeat },
  { key: 'info', icon: Info },
]

const { t, locale } = useI18n()
const store = useJobsStore()
const activeMenu = ref('subtitle')
const { isDark, toggleTheme } = useTheme()

const runningCount = computed(() => store.jobList.filter((j) => j.status === 'running').length)

// ========== 回到顶部（框架层统一实现：滚动 .content 容器） ==========
const contentEl = ref<HTMLElement | null>(null)
const showBackTop = ref(false)
// 滚动超过 50px 即显示（内容不长时滚到底也要能看到）
function onContentScroll(e: Event) {
  const el = e.target as HTMLElement
  showBackTop.value = el.scrollTop > 50
}
function scrollToTop() {
  contentEl.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

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

// 回到顶部按钮：悬浮于内容区右下角
.back-to-top {
  position: fixed;
  right: 20px;
  bottom: 40px; // 状态栏上方
  z-index: 30;
  width: 34px;
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-secondary);
  cursor: pointer;
  box-shadow: var(--shadow-card);
  transition: all 0.18s ease;
  &:hover {
    color: var(--primary);
    border-color: var(--primary-border);
  }
  // 底部抽屉（任务队列/日志，240px）打开时，抬到抽屉上方，避免盖在日志窗口上
  &.above-dock {
    bottom: 280px;
  }
}
// 出现/消失过渡
.btt-enter-active,
.btt-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.btt-enter-from,
.btt-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>
