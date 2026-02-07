<template>
  <div class="flex">
    <ul class="h-screen w-120 border-r border-gray-200 dark:border-white/10 dark:bg-black/10">
      <RouterLink
        v-for="login in logins"
        :key="login.id"
        :to="`/secrets/logins/${login.id}/view`"
        class="flex h-20 w-full items-center justify-between border-b border-gray-200 bg-white px-6 py-6 text-black dark:border-white/10 dark:bg-black/10 dark:text-white"
        active-class="bg-gray-50 dark:bg-white/5 text-indigo-600 dark:text-white"
      >
        <div class="flex min-w-0 gap-x-4">
          <div class="min-w-0 flex-auto">
            <p class="text-sm/6 font-semibold">{{ login.username }}</p>
          </div>
        </div>
        <div class="hidden shrink-0 sm:flex sm:flex-col sm:items-end" v-if="login.domains.length">
          <p class="text-sm/6 text-gray-900 dark:text-white">{{ login.domains[0] }}</p>
          <p class="mt-1 text-xs/5 text-gray-500 dark:text-gray-400" v-if="login.domains.length > 1">
            And {{ login.domains.length - 1 }} more
          </p>
        </div>
      </RouterLink>
    </ul>
    <Suspense>
      <RouterView :key="route.fullPath" />
    </Suspense>
  </div>
</template>

<script setup lang="ts">
import { LoginClient } from "@/lib/login";
import { useRoute } from "vue-router";

const route = useRoute();
const client = new LoginClient();

const logins = await client.list();
</script>
