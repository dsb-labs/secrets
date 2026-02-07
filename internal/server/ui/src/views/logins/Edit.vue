<template>WIP</template>

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
