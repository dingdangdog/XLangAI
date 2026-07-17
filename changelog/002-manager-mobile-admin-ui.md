# 002 - Manager 全量手机端管理 UI

## 时间

2026-07-17

## 变更说明

为运营后台全部管理页面落地手机端适配：列表改为卡片、双栏改为钻取、对话框/确认框贴边全宽，并补齐移动导航能力。

## 主要改动

### 设计

- 新增 `design/manager-mobile-admin-ui.md` 交互与断点约定

### 共享基础设施

- `useIsMobile` / `useIsNarrow` 断点 composable
- `AdminMobileCard` / `AdminMobileMeta` / `AdminOverflowMenu` 移动列表组件
- `AdminTextarea`（修复练习场景页缺失组件）
- `AdminTable` 支持 `#mobile` 槽：有槽时桌面表格 / 手机卡片双模
- `AdminDialog` / `AdminConfirmHost` 手机底部贴边 + safe-area
- `AdminPageHeader` / `AdminPagination` / `AgentDrawer` / `AgentSegmentNav` / `AgentHubDrawerFooter` 窄屏布局优化
- `default` 布局：顶栏简化、抽屉补齐退出登录与版本、safe-area

### 页面与面板

- CRUD：用户、会话、消息、语言、会员、提示词、练习场景 — 移动卡片 + 操作菜单
- 跟读：分类 ↔ 词汇钻取 + 词汇卡片
- 系统变量：列表 ↔ 编辑钻取
- 语音角色、用户用量、删除记录四面板、数据备份表 — 移动卡片
- 修正 `AdminDialog` 误用 `size` 为 `width`；语音角色未定义色类改为主题色

### i18n

- 中/英/日新增 `common.back`
