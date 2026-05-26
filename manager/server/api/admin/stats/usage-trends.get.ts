import { fetchDashboardUsageTrends } from "../../../utils/dashboardUsageTrends";

export default defineEventHandler(async (event) => {
  const query = getQuery(event);
  const rawDays = Number(query.days ?? 7);
  const days: 7 | 30 = rawDays === 30 ? 30 : 7;
  return fetchDashboardUsageTrends(days);
});
