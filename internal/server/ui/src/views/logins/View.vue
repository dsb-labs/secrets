<template>
  <section class="flex-1 bg-white px-6 py-6 dark:bg-black/10">
    <div class="px-4 sm:px-0">
      <h3 class="text-base/7 font-semibold text-gray-900 dark:text-white">{{ login.username }}</h3>
      <p class="mt-1 max-w-2xl text-sm/6 text-gray-500 dark:text-gray-400" v-if="login.domains.length">
        {{ login.domains[0] }}
      </p>
    </div>
    <div class="mt-6 border-t border-gray-100 dark:border-white/10">
      <dl class="divide-y divide-gray-100 dark:divide-white/10">
        <div class="px-4 py-6 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
          <dt class="text-sm/6 font-medium text-gray-900 dark:text-gray-100">Username</dt>
          <dd class="mt-1 text-sm/6 text-gray-700 sm:col-span-2 sm:mt-0 dark:text-gray-400">{{ login.username }}</dd>
        </div>
        <div class="px-4 py-6 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
          <dt class="text-sm/6 font-medium text-gray-900 dark:text-gray-100">Password</dt>
          <dd class="mt-1 text-sm/6 text-gray-700 sm:col-span-2 sm:mt-0 dark:text-gray-400">
            {{ "*".repeat(login.password.length) }}
          </dd>
        </div>
        <div class="px-4 py-6 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0">
          <dt class="text-sm/6 font-medium text-gray-900 dark:text-gray-100">Domains</dt>
          <dd class="mt-2 text-sm text-gray-900 sm:col-span-2 sm:mt-0 dark:text-white">
            <ul
              role="list"
              class="divide-y divide-gray-100 rounded-md border border-gray-200 dark:divide-white/5 dark:border-white/10"
            >
              <li
                class="flex items-center justify-between py-4 pr-5 pl-4 text-sm/6"
                v-for="domain in login.domains"
                :key="domain"
              >
                <div class="flex w-0 flex-1 items-center">
                  <div class="ml-4 flex min-w-0 flex-1 gap-2">
                    <img alt="favicon" :src="`${domain}/favicon.ico`" class="w-6" />
                    <a :href="domain" target="_blank" class="truncate font-medium text-gray-900 dark:text-white">{{
                      domain
                    }}</a>
                  </div>
                </div>
              </li>
            </ul>
          </dd>
        </div>
      </dl>
    </div>
  </section>
</template>

<script setup lang="ts">
import { LoginClient } from "@/lib/login";

const { id } = defineProps({
  id: {
    type: String,
    required: true,
  },
});

const client = new LoginClient();
const login = await client.find(id);
</script>
