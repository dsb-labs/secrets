<script setup lang="ts"></script>
<template>
  <nav class="h-screen w-72 border-r border-gray-200 bg-white dark:border-white/10 dark:bg-black/10">
    <!-- Sidebar component, swap this element with another sidebar if you like -->
    <div
      class="flex h-screen grow flex-col gap-y-5 overflow-y-auto border-r border-gray-200 bg-white px-6 dark:border-white/10 dark:bg-black/10"
    >
      <div class="flex h-16 shrink-0 items-center">
        <img
          src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=indigo&shade=600"
          alt="Your Company"
          class="h-8 w-auto dark:hidden"
        />
        <img
          src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=indigo&shade=500"
          alt="Your Company"
          class="hidden h-8 w-auto dark:block"
        />
      </div>
      <nav class="flex flex-1 flex-col">
        <ul role="list" class="flex flex-1 flex-col gap-y-7">
          <li>
            <ul role="list" class="flex flex-1 flex-col gap-y-7">
              <li>
                <ul role="list" class="-mx-2 space-y-1">
                  <li>
                    <RouterLink
                      to="/secrets/logins"
                      class="group flex gap-x-3 rounded-md p-2 text-sm/6 font-semibold text-gray-700 hover:bg-gray-50 hover:text-indigo-600 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-white"
                    >
                      Logins
                    </RouterLink>
                  </li>
                  <li>
                    <RouterLink
                      to="/secrets/notes"
                      class="group flex gap-x-3 rounded-md p-2 text-sm/6 font-semibold text-gray-700 hover:bg-gray-50 hover:text-indigo-600 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-white"
                    >
                      Notes
                    </RouterLink>
                  </li>
                </ul>
              </li>
            </ul>
          </li>
        </ul>
      </nav>
    </div>
  </nav>
</template>
