import fs from "fs";

const files = fs
  .readdirSync("app/pages/manage")
  .filter((f) => f.endsWith(".vue"))
  .map((f) => `app/pages/manage/${f}`);

const scriptReplacements = [
  ['toast.warning("请填写语言代码与名称")', 'toast.warning(t("validation.fillLanguageCodeAndName"))'],
  ['toast.warning("请填写等级编码与名称")', 'toast.warning(t("validation.fillTierCodeAndName"))'],
  ['toast.warning("请填写模板编码与名称")', 'toast.warning(t("validation.fillTemplateCodeAndName"))'],
  ['toast.warning("请填写模板内容")', 'toast.warning(t("validation.fillTemplateContent"))'],
  ['toast.warning("请填写编码与名称")', 'toast.warning(t("validation.fillCodeAndName"))'],
  ['toast.warning("请填写名称")', 'toast.warning(t("validation.fillName"))'],
  ['toast.warning("请填写 key")', 'toast.warning(t("validation.fillKey"))'],
  ['toast.warning("请填写用户 ID")', 'toast.warning(t("validation.fillUserId"))'],
  ['toast.warning("请选择语言")', 'toast.warning(t("validation.selectTargetLanguage"))'],
  ['toast.warning("请填写会话 ID")', 'toast.warning(t("validation.fillConversationId"))'],
  ['toast.warning("请选择角色")', 'toast.warning(t("validation.selectRole"))'],
  ['toast.error("加载下拉数据失败")', 'toast.error(t("toast.loadRefsFailed"))'],
  ['toast.error("加载语言列表失败")', 'toast.error(t("toast.loadLanguagesFailed"))'],
  ['toast.error("加载下拉选项失败")', 'toast.error(t("toast.loadOptionsFailed"))'],
  ['toast.error("加载服务器商店配置失败")', 'toast.error(t("toast.loadServerStoreFailed"))'],
  ['toast.error("权益 features 须为合法 JSON")', 'toast.error(t("validation.invalidJson", { field: "features" }))'],
  ['toast.error("settings 须为合法 JSON")', 'toast.error(t("validation.invalidJson", { field: "settings" }))'],
  ['toast.error("metadata 须为合法 JSON")', 'toast.error(t("validation.invalidJson", { field: "metadata" }))'],
  ['toast.error("variables 须为合法 JSON（如数组）")', 'toast.error(t("validation.invalidJson", { field: "variables" }))'],
  ['toast.error("扩展配置须为合法 JSON")', 'toast.error(t("validation.invalidJson", { field: t("fields.extJson") }))'],
  ['message: "确认删除该语言？若仍被用户或会话引用可能导致数据不一致。"', 'message: t("confirm.deleteLanguage")'],
  ['message: "确认删除该会员等级？若用户仍引用请谨慎操作。"', 'message: t("confirm.deleteTier")'],
  ['message: "确认删除该提示词模板？"', 'message: t("confirm.deletePrompt")'],
  ['message: "确认软删除该消息？"', 'message: t("confirm.softDeleteMessage")'],
  ['message: "确认软删除该会话？"', 'message: t("confirm.softDeleteConversation")'],
  ['message: "确认软删除该用户？删除后无法登录，可在「含已删除」中查看。"', 'message: t("confirm.softDeleteUser")'],
  ['from "~/utils/usageDisplay"', 'from "~/composables/useUsageDisplay"'],
  ['serviceUsageMonthLine,\n  serviceUsageTodayLine,\n} from "~/composables/useUsageDisplay"', 'serviceUsageMonthLine,\n  serviceUsageTodayLine,\n} = useUsageDisplay()'],
  ['userUsageDetailLine,\n  userUsagePrimaryLine,\n} from "~/utils/usageDisplay"', 'userUsageDetailLine,\n  userUsagePrimaryLine,\n} = useUsageDisplay()'],
];

const templateReplacements = [
  ['title="语言"', ':title="$t(\'pages.languages.title\')"'],
  ['description="管理客户端可选语言列表；code 需唯一（如 zh、en）。"', ':description="$t(\'pages.languages.description\')"'],
  [':title="dialogMode === \'create\' ? \'新建语言\' : \'编辑语言\'"', ':title="dialogMode === \'create\' ? $t(\'pages.languages.createDialog\') : $t(\'pages.languages.editDialog\')"'],
  ['label="含已删除"', ':label="$t(\'common.includeDeleted\')"'],
  ['width="200px" align="right">操作</', 'width="200px" align="right">{{ $t("common.actions") }}</'],
  ['width="140px" align="right">操作</', 'width="140px" align="right">{{ $t("common.actions") }}</'],
  ['>代码</AdminTh>', '>{{ $t("common.code") }}</AdminTh>'],
  ['>名称</AdminTh>', '>{{ $t("common.name") }}</AdminTh>'],
  ['>排序</AdminTh>', '>{{ $t("common.sort") }}</AdminTh>'],
  ['>状态</AdminTh>', '>{{ $t("common.status") }}</AdminTh>'],
  ['>备注</AdminTh>', '>{{ $t("common.remark") }}</AdminTh>'],
  ['>更新时间</AdminTh>', '>{{ $t("common.updatedAt") }}</AdminTh>'],
  ['>本地名称</AdminTh>', '>{{ $t("fields.localName") }}</AdminTh>'],
  ['label="代码"', ':label="$t(\'common.code\')"'],
  ['label="名称"', ':label="$t(\'common.name\')"'],
  ['label="排序"', ':label="$t(\'common.sort\')"'],
  ['label="状态"', ':label="$t(\'common.status\')"'],
  ['label="备注"', ':label="$t(\'common.remark\')"'],
  ['label="ID"', ':label="$t(\'common.id\')"'],
  ['\n              启用\n', '\n              {{ $t("common.enable") }}\n'],
  ['\n              删除\n', '\n              {{ $t("common.delete") }}\n'],
];

for (const f of files) {
  let s = fs.readFileSync(f, "utf8");
  for (const [from, to] of scriptReplacements) {
    if (s.includes(from)) s = s.split(from).join(to);
  }
  for (const [from, to] of templateReplacements) {
    if (s.includes(from)) s = s.split(from).join(to);
  }
  fs.writeFileSync(f, s);
}
console.log("done", files.length);
