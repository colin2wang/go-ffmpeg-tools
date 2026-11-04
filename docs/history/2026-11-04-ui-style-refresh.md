# go-ffmpeg-tools UI 去模板感改造计划
> 目标：去掉"AI 生成/模板站"的视觉信号，对齐 Linear、GitHub Desktop 一类专业桌面工具的质感。
> 范围：仅动样式、图标、文案，不改任何业务逻辑与 Wails 绑定。

## 一、现状问题清单（"UI 味"来源）
| # | 问题 | 位置 |
|---|------|------|
| 1 | Emoji 当图标（🎬🎞ℹ️📂📄▶×↑↓），彩色 emoji 与单色 UI 冲突、跨平台渲染不一致 | App.vue / Index.vue / LogPanel.vue / SubTrackRow.vue / i18n 文案 |
| 2 | 默认 Tailwind 蓝 `#3b82f6`，AI 生成界面标配色 | app.scss |
| 3 | 纯平面：卡片无阴影，深色下与背景几乎同灰度 | app.scss |
| 4 | 静态微交互：进度条是色块、running 状态无动态语义 | app.scss |
| 5 | 药丸 tag（999px 圆角）、纯文字空状态、⚠️ 黄字警告条 | app.scss / Index.vue |

## 二、设计原则
- **单一强调色**：全站只有主色一个强调色，状态色（成功/警告/危险）只做语义不做装饰
- **图标统一**：全部替换为 lucide SVG，`stroke-width: 1.75`，跟随 `currentColor`
- **动效克制**：全部 ≤1.4s 循环/0.2s 过渡，尊重 `prefers-reduced-motion`
- **零框架**：不引入 UI 组件库，仅新增一个图标依赖
- **纯 CSS 优先**：呼吸点、流光等尽量用 CSS 伪元素实现，减少模板改动

## 三、改动总览
| 文件 | 改动 | 类型 | 风险 |
|------|------|------|------|
| `frontend/package.json` | + `lucide-vue-next` | 依赖 | 无 |
| `frontend/src/styles/app.scss` | 设计令牌重定义 + 全量样式升级（完整替换） | 核心 | 低，选择器全部保留 |
| `frontend/src/App.vue` | 菜单/主题/日志按钮 emoji → SVG | 模板 | 低 |
| `frontend/src/views/Index.vue` | 拖拽区图标化 + 加高、空状态 CTA、警告条图标 | 模板 | 低 |
| `frontend/src/components/SubTrackRow.vue` | ↑↓× → ChevronUp/ChevronDown/X | 模板 | 低 |
| `frontend/src/components/LogPanel.vue` | 📄 → FileText | 模板 | 低 |
| `frontend/src/components/JobCard.vue` | **无需改动**（呼吸点/流光纯 CSS，基于已有 status class） | — | — |
| `frontend/src/i18n/locales/*.ts` | 文案剥离 emoji + 新增 `queue.pickVideo` | 文案 | 低 |

## 四、设计令牌变更对照
| 令牌 | 旧值 | 新值（深色） | 说明 |
|------|------|------|------|
| `--primary` | `#3b82f6` | `#34d399` | 去默认蓝，翡翠绿 |
| `--primary`（浅色） | `#3b82f6` | `#059669` | 浅色下压深保证对比度 |
| `--primary-dim` | 无 | `rgba(52,211,153,.13)` | 新增：选中态淡底色，替代实心色块 |
| `--primary-contrast` | 无 | `#0a2418` | 新增：主按钮文字色（亮底配深字） |
| `--primary-border` | 无 | `rgba(52,211,153,.38)` | 新增：运行中任务卡描边 |
| `--radius-sm/md` | 4/8 | 6/10 | 圆角层级化：控件 < 卡片 < 大容器 |
| `--shadow-card` | 无 | 双层阴影 | 新增：卡片与背景分离 |
| `--glass` | 无 | `rgba(24,27,34,.78)` | 新增：顶栏毛玻璃 |
| `--warn-dim` | 无 | `rgba(251,191,36,.10)` | 新增：callout 警告条底色 |
| `.job-tag` 圆角 | 999px | 4px | 药丸 → 工程风格方块 tag |

不使用 `color-mix()`，所有混色用预定义 rgba 令牌，兼容 Wails 三端 WebView。

## 五、实施顺序
1. `pnpm add lucide-vue-next`（实际安装 0.577.0，1.x 为弃用版本）
2. 完整替换 `app.scss`（此时 UI 已更新，图标暂缺不影响功能）
3. 改 `App.vue` → `Index.vue` → `SubTrackRow.vue` → `LogPanel.vue` 图标
4. 更新三个 locale 文案（剥离 emoji + 新增 key）
5. 验收

## 六、验证清单
- [ ] 深浅两主题逐页检查：主按钮文字对比度、focus 光圈、卡片阴影
- [ ] 任务运行中：进度条流光、状态 tag 呼吸点、任务卡主色描边
- [ ] 拖拽：落点蚂蚁线动画；`OnFileDrop` 落点反查逻辑未动，video/sub/burn 三区仍正确分发
- [ ] 取消/失败/完成三种终态颜色正确
- [ ] 系统开启"减少动态效果"时动画全部停用
- [ ] 日志面板浅色主题下保持深色终端风格

## 七、回滚方案
- 样式：`git revert` app.scss 单文件即可整体回滚
- 图标：模板级改动，随各自 commit revert
- 文案：i18n 独立 commit，可单独回滚

## 八、本次明确不做（后续可选）
- 源视频与输出设置卡片合并（布局重构，改动面大）
- 字幕轨道拖拽排序（SortableJS）
- 主题/语言跟随系统设置的深度联动
