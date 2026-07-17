# 001 - 补全 practice-scenarios 管理端 API 路由

**日期：** 2026-07-17

## 变更说明

`/api/admin/practice-scenarios` 此前仅在 `adminResource.ts` 中注册了资源元数据，前端管理页也已调用该接口，但缺少对应的 Nitro 路由文件，导致登录后访问返回 Nuxt 404 错误页。

## 新增文件

- `manager/server/api/admin/practice-scenarios/index.get.ts` — 列表查询
- `manager/server/api/admin/practice-scenarios/index.post.ts` — 创建
- `manager/server/api/admin/practice-scenarios/[id].put.ts` — 更新
- `manager/server/api/admin/practice-scenarios/[id].delete.ts` — 删除

实现方式与同目录下 `prompt-templates` 等资源的 CRUD 路由一致，复用 `adminCrudHandlers` 通用处理器。

## 部署

需重新构建并部署 manager 服务后生效。
