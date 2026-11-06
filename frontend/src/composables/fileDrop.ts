// 全局文件拖拽分发：Wails OnFileDrop 全局只能注册一个回调（后注册会覆盖先注册），
// 这里统一由 App.vue 注册一次，按落点坐标反查 data-drop-zone 后分发给各页面注册的处理器。
// 各页面用 v-show 常驻，注册一次即长期有效。
type DropHandler = (paths: string[]) => void

const handlers = new Map<string, DropHandler>()

export function registerDropZone(zone: string, fn: DropHandler) {
  handlers.set(zone, fn)
}

export function unregisterDropZone(zone: string) {
  handlers.delete(zone)
}

// 落点反查：由坐标找到所属 data-drop-zone 元素
export function zoneAt(x: number, y: number): string | null {
  const el = document.elementFromPoint(x, y)
  return (el as HTMLElement | null)?.closest('[data-drop-zone]')?.getAttribute('data-drop-zone') ?? null
}

export function dispatchDrop(zone: string | null, paths: string[]) {
  if (!zone || paths.length === 0) return
  handlers.get(zone)?.(paths)
}
