/**
 * 与 `server/api/admin/<resource>/` 下独立路由对应，路径不含末尾斜杠。
 * 例：`basePath = '/api/admin/languages'`
 */
export function useAdminResourceApi(basePath: string) {
  async function list(query: Record<string, string | number | boolean>) {
    return await $fetch<{ items: Record<string, unknown>[]; total: number }>(basePath, {
      query,
    });
  }
  async function create(body: object) {
    return await $fetch(basePath, { method: "POST", body });
  }
  async function update(id: string, body: object) {
    return await $fetch(`${basePath}/${id}`, { method: "PUT", body });
  }
  async function remove(id: string) {
    return await $fetch(`${basePath}/${id}`, { method: "DELETE" });
  }
  return { list, create, update, remove };
}
