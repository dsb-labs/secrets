<template>
  <section class="flex-1 bg-white px-6 py-6 dark:bg-black/10">
    <div class="dark:border-white/10">
      <div class="flex flex-wrap items-center justify-between sm:flex-nowrap">
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ login.username }}</h3>
        <div class="mt-4 ml-4 flex shrink-0">
          <RouterLink
            type="button"
            class="relative inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-gray-300 ring-inset hover:bg-gray-50 dark:bg-white/10 dark:text-white dark:shadow-none dark:ring-white/5 dark:hover:bg-white/20"
            :to="`/secrets/logins/${login.id}/edit`"
          >
            <span>Edit</span>
          </RouterLink>
          <RouterLink
            type="button"
            class="relative ml-3 inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-gray-300 ring-inset hover:bg-gray-50 dark:bg-white/10 dark:text-white dark:shadow-none dark:ring-white/5 dark:hover:bg-white/20"
            :to="`/secrets/logins/${login.id}/delete`"
          >
            <span>Delete</span>
          </RouterLink>
        </div>
      </div>
    </div>
    <div class="mt-6 border-t border-gray-200 dark:border-white/10">
      <dl class="divide-y divide-gray-200 dark:divide-white/10">
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
              class="divide-y divide-gray-200 rounded-md border border-gray-200 dark:divide-white/5 dark:border-white/10"
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
