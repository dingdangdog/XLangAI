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
  const { t } = useI18n();
  const toast = useToast();
  const { confirm, setLoading } = useConfirm();
  const activatingId = ref<string | null>(null);

  async function activateRow(row: Record<string, unknown>) {
    if (String(row.status) === "active") {
      toast.info(t("activate.alreadyActive"));
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
        const names = others.map(rowLabel).join(t("usage.separator"));
        message = t("activate.singleActiveMessage", { label, names });
      } else {
        message = t("activate.simpleMessage", { label });
      }
    } else {
      message = t("activate.simpleMessage", { label });
    }

    const ok = await confirm({
      title: t("activate.title"),
      message,
      confirmLabel: t("common.enable"),
    });
    if (!ok) return;

    activatingId.value = id;
    setLoading(true);
    try {
      await options.api.update(id, { id, status: "active" });
      toast.success(t("activate.success", { label }));
      await options.reload();
    } catch (e) {
      toast.error(t("activate.failed"));
      console.error(e);
    } finally {
      activatingId.value = null;
      setLoading(false);
    }
  }

  return { activateRow, activatingId };
}
