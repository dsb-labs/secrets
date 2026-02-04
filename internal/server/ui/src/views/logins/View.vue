<template>
  <section class="flex-1 bg-white dark:bg-black/10 dark:text-white">{{ id }}</section>
</template>

<script setup lang="ts">
const props = defineProps({
  id: {
    type: String,
    required: true,
  },
});
</script>
