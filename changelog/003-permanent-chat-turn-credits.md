# 003 — 永久对话次数额度（turn_balance）

## 变更摘要

引入与「会员档日/月日历限额」并存的**永久对话次数**余额：不过期，用一次扣一次；运营可对指定用户加次。`token_balance` 仍仅表示 LLM token 钱包，语义不变。

## 表结构（需自行执行 Prisma migrate / generate）

- `usr_users.turn_balance` `Int NOT NULL DEFAULT 0`（Prisma：`User.turnBalance`）
- **未**提交 migration SQL；请在 `manager` 目录自行生成并应用迁移，再 `prisma generate`

建议命令（由开发者执行）：

```bash
cd manager
npx prisma migrate dev --name add_user_turn_balance
# 或你们现有的 migrate / db push 流程
```

## 额度规则

| 有效档位 | 行为 |
|----------|------|
| 日、月限额均为空/0（如 free） | 校验并消耗 `turn_balance`；用尽 → `QUOTA_TURNS` |
| 有日/月限额（如 plus/pro） | 原逻辑：日硬限、月包含轮次、月满后 `token_balance` 兜底 |

成功 AI 回复后：永久次数模式扣 `turn_balance` 1；月超额模式仍按 LLM token 扣 `token_balance`。`usr_user_usage` 继续记统计。

## Server

- `authz.EnsureChatQuota` / `DeductChatTurn` / `QUOTA_TURNS`
- 注册（密码 / 短信）按 `quota.signup_turn_grant`（默认 20）写入初始余额
- `GET /users/me` 与登录用户对象返回 `turn_balance`

## Manager

- 用户编辑：永久次数绝对值 +「增加次数」增量（`addTurnBalance`）
- 系统变量：`quota.signup_turn_grant`
- free 档种子改为无日历限额；已有环境首次种子对齐时会回填 free 用户空余额

## 上线注意

1. 先 migrate 出 `turn_balance` 列，再部署 Go / 重启 manager（种子会改 free 限额）
2. 客户端需识别新错误码 `QUOTA_TURNS`
3. 已有测试账号若在 migrate 前创建，依赖种子回填或管理端手动加次
