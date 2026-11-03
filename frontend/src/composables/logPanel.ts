import { ref } from 'vue'

// 全局日志面板状态（模块级，跨页面/菜单共享，切换菜单不丢失）
export const logVisible = ref(false)
// "自动弹出"复选框：勾选后任务日志更新才自动展开日志框，默认不勾选
export const autoPopup = ref(false)
export const logTitle = ref('')
export const logLines = ref<string[]>([])

export function openLog(title: string, lines: string[]) {
  logTitle.value = title
  logLines.value = lines
  // 勾选"自动弹出"才自动展开，否则保持折叠
  if (autoPopup.value) logVisible.value = true
}

export function closeLog() {
  logVisible.value = false
}

// 状态栏按钮：切换日志面板展开/折叠
export function toggleLog() {
  logVisible.value = !logVisible.value
}

export function clearLog() {
  logLines.value = []
}
