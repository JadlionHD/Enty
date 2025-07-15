<script setup lang="ts">
import { useDownloadStore } from '@/stores/download';
import { ref } from 'vue';

const props = defineProps<{
  title: string;
  value: number;
}>();

const download = useDownloadStore();
</script>

<template>
  <div>
    <div class="flex justify-between items-center w-full mb-4">
      <div>{{ props.title }}</div>
      <div class="text-muted">
        <div>{{ Math.floor(props.value) }}%</div>
      </div>
    </div>
    <div class="flex justify-between items-center w-full gap-x-2">
      <div class="w-full">
        <UProgress v-model="props.value" :max="100" />
      </div>
      <div>
        <UButton variant="outline" icon="i-lucide-x" @click="download.cancel()"></UButton>
      </div>
    </div>
  </div>
</template>
