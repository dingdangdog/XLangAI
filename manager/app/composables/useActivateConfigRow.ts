export type ActivateExclusivity = "single-active" | "multi-active";

function rowLabel(row: Record<string, unknown>): string {
  const name = String(row.name ?? "").trim();
  if (name) return name;
  const code = String(row.code ?? "").trim();
  if (code) return code;
  return String(row.id ?? "");
}

export function useActivateConfigRow(options: {
  api: { update: (id: string, body: object) => Promise<unknown> };
  getList: () => Record<string, unknown>[];
  exclusivity: ActivateExclusivity;
  reload: () => Promise<void>;
}) {
  const toast = useToast();
  const { confirm, setLoading } = useConfirm();
  const activatingId = ref<string | null>(null);

  async function activateRow(row: Record<string, unknown>) {
    if (String(row.status) === "active") {
      toast.info("当前已是启用状态");
      return;
    }
    const id = String(row.id ?? "");
    if (!id) return;
    const label = rowLabel(row);

    let message: string;
    if (options.exclusivity === "single-active") {
      const others = options.getList().filter(
        (r) => String(r.status) === "active" && String(r.id) !== id,
      );
      if (others.length > 0) {
        const names = others.map(rowLabel).join("、");
        message =
          `确认启用「${label}」？\n\n` +
          `当前已启用：${names}\n` +
          `启用后，上述项将自动设为 inactive（全局仅保留一条 active）。`;
      } else {
        message = `确认启用「${label}」？`;
      }
    } else {
      message = `确认启用「${label}」？`;
    }

    const ok = await confirm({
      title: "启用配置",
      message,
      confirmLabel: "启用",
    });
    if (!ok) return;

    activatingId.value = id;
    setLoading(true);
    try {
      await options.api.update(id, { id, status: "active" });
      toast.success(`已启用：${label}`);
      await options.reload();
    } catch (e) {
      toast.error("启用失败");
      console.error(e);
    } finally {
      activatingId.value = null;
      setLoading(false);
    }
  }

  return { activateRow, activatingId };
}
