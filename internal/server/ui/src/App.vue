<template>
  <Header />
  <main class="w-full">
    <RouterView :key="route.fullPath" />
  </main>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router"
import Header from "@/components/Header.vue"

const route = useRoute()
</script>

<style lang="css">
@import "tailwindcss";
</style>
