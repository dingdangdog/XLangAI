import fs from "fs";

const files = [
  "app/pages/manage/languages.vue",
  "app/pages/manage/membership-tiers.vue",
  "app/pages/manage/prompt-templates.vue",
  "app/pages/manage/conversations.vue",
  "app/pages/manage/messages.vue",
  "app/pages/manage/users.vue",
  "app/pages/manage/voice-roles.vue",
  "app/pages/manage/server-store.vue",
  "app/pages/manage/system-settings.vue",
  "app/pages/manage/llm-service-configs.vue",
  "app/pages/manage/tts-service-configs.vue",
  "app/pages/manage/stt-service-configs.vue",
  "app/pages/manage/translate-service-configs.vue",
  "app/pages/manage/object-storage-configs.vue",
];

const replacements = [
  ['toast.error("加载失败")', 'toast.error(t("toast.loadFailed"))'],
  ['toast.error("保存失败")', 'toast.error(t("toast.saveFailed"))'],
  ['toast.error("删除失败")', 'toast.error(t("toast.deleteFailed"))'],
  ['toast.error("操作失败")', 'toast.error(t("toast.operationFailed"))'],
  ['toast.success("已创建")', 'toast.success(t("toast.created"))'],
  ['toast.success("已保存")', 'toast.success(t("toast.saved"))'],
  ['toast.success("已删除")', 'toast.success(t("toast.deleted"))'],
  ['toast.success("已标记删除")', 'toast.success(t("toast.markedDeleted"))'],
  ['confirmLabel: "删除"', 'confirmLabel: t("common.delete")'],
  ['confirmLabel: "软删"', 'confirmLabel: t("common.softDelete")'],
  ['>取消</', '>{{ $t("common.cancel") }}</'],
  ['>保存</', '>{{ $t("common.save") }}</'],
  ['>新建</', '>{{ $t("common.create") }}</'],
  ['>编辑</', '>{{ $t("common.edit") }}</'],
  ['>删除</', '>{{ $t("common.delete") }}</'],
  ['>启用</', '>{{ $t("common.enable") }}</'],
  ['>软删</', '>{{ $t("common.softDelete") }}</'],
  ['?? "—"', '?? t("common.emDash")'],
];

for (const f of files) {
  let s = fs.readFileSync(f, "utf8");
  if (!s.includes("useI18n")) {
    s = s.replace(
      '<script setup lang="ts">',
      '<script setup lang="ts">\nconst { t } = useI18n();',
    );
  }
  for (const [from, to] of replacements) {
    s = s.split(from).join(to);
  }
  fs.writeFileSync(f, s);
  console.log("updated", f);
}
