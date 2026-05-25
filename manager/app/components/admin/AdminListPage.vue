<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4">
    <div v-if="$slots.header" class="shrink-0 space-y-4">
      <slot name="header" />
    </div>
    <div class="flex min-h-0 flex-1 flex-col">
      <slot />
    </div>
  </div>
</template>
