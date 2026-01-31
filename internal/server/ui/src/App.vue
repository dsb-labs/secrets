<template>
  <main class="h-full w-full">
    <RouterView :key="route.fullPath" />
  </main>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router"

const route = useRoute()
</script>

<style lang="css">
@import "tailwindcss";
</style>
