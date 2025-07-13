<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
  value: number;
}>();
</script>

<template>
  <div>
    <div>Hello worlds</div>
    <div class="flex justify-between items-center w-full gap-x-2">
      <div class="text-muted">{{ props.value }}%</div>
      <div class="w-full">
        <UProgress v-model="props.value" :max="100" />
      </div>
      <div>
        <UButton variant="outline" icon="i-lucide-x"></UButton>
      </div>
    </div>
  </div>
</template>
