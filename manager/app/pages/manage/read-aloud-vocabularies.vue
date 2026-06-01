<script setup lang="ts">
const route = useRoute();
const query = route.query.categoryId
  ? { categoryId: String(route.query.categoryId) }
  : {};
await navigateTo({ path: "/manage/read-aloud", query }, { replace: true });
</script>
