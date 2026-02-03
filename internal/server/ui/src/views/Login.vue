<template>
  <div class="flex min-h-full flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <h2 class="mt-6 text-center text-2xl/9 font-bold tracking-tight text-gray-900">Sign in to your account</h2>
    </div>

    <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-[480px]">
      <div class="bg-white px-6 py-12 shadow sm:rounded-lg sm:px-12">
        <form @submit="onSubmit" class="space-y-6">
          <div>
            <label for="email" class="block text-sm/6 font-medium text-gray-900">Email address</label>
            <div class="mt-2">
              <input
                id="email"
                type="email"
                name="email"
                required
                autocomplete="email"
                v-model="email"
                class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-stone-600 sm:text-sm/6"
              />
            </div>
          </div>

          <div>
            <label for="password" class="block text-sm/6 font-medium text-gray-900">Password</label>
            <div class="mt-2">
              <input
                id="password"
                type="password"
                name="password"
                v-model="password"
                required
                autocomplete="current-password"
                class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-stone-600 sm:text-sm/6"
              />
            </div>
          </div>

          <div>
            <button
              type="submit"
              class="flex w-full justify-center rounded-md bg-stone-600 px-3 py-1.5 text-sm/6 font-semibold text-white shadow-sm hover:bg-stone-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-stone-600"
            >
              Sign in
            </button>
          </div>
        </form>
      </div>

      <p class="mt-10 text-center text-sm/6 text-gray-500">
        <a href="/account/create" class="font-semibold text-stone-600 hover:text-stone-500">Create an account</a>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AuthClient } from "@/lib/auth";
import { useRouter } from "vue-router";

let email = "";
let password = "";

const router = useRouter();
const client = new AuthClient();

const onSubmit = async (e: Event) => {
  e.preventDefault();

  try {
    await client.login(email, password);
  } catch (e) {
    console.log(e);
  }

  await router.push({ name: "login-list" });
};
</script>
