import { ref } from 'vue'

// 主题状态挂在 body 上（body 背景依赖主题变量），全局共享
const isDark = ref(true)
const THEME_KEY = 'go-ffmpeg-tools-theme'

function applyTheme() {
  document.body.classList.toggle('theme-dark', isDark.value)
  document.body.classList.toggle('theme-light', !isDark.value)
}

export function useTheme() {
  // 初始化：读取上次保存的主题，首次调用时生效
  if (!document.body.classList.contains('theme-dark') && !document.body.classList.contains('theme-light')) {
    const saved = localStorage.getItem(THEME_KEY)
    isDark.value = saved ? saved === 'dark' : true
    applyTheme()
  }
  function toggleTheme() {
    isDark.value = !isDark.value
    applyTheme()
    localStorage.setItem(THEME_KEY, isDark.value ? 'dark' : 'light')
  }
  return { isDark, toggleTheme }
}
