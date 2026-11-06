import { ref } from 'vue'

// 底部抽屉状态：任务队列 / 日志 两个标签共用一个窗口
// （模块级，跨页面/菜单共享，切换菜单不丢失）
export type DockTab = 'queue' | 'log'
export const dockOpen = ref(false)
export const dockTab = ref<DockTab>('log')
export const logLines = ref<string[]>([])

export function openDock(tab: DockTab) {
  dockTab.value = tab
  dockOpen.value = true
}

// 写入日志内容（不自动展开，展开由状态栏按钮控制）
export function openLog(lines: string[]) {
  logLines.value = lines
}

export function clearLog() {
  logLines.value = []
}
