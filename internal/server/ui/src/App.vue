<template>
  <div class="flex min-h-screen bg-gray-100 dark:bg-gray-800">
    <Sidebar />
    <main class="flex-1 overflow-y-auto">
      <Suspense>
        <RouterView />
      </Suspense>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";
import Sidebar from "@/components/Sidebar.vue";

const route = useRoute();
</script>

<style lang="css">
@import "tailwindcss";
</style>
