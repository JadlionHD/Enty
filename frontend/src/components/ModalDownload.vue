<script setup lang="ts">
import { useDownloadStore } from '@/stores/download';
import { ref } from 'vue';

const value = ref(100);

const download = useDownloadStore();
</script>

<template>
  <div class="z-20">
    <UModal>
      <UButton
        class="absolute bottom-4 right-4"
        variant="outline"
        icon="i-lucide-download"
        color="secondary"
      ></UButton>

      <template #header>
        <div class="text-xl font-bold">Downloads</div>
      </template>

      <template #body>
        <div class="flex flex-col gap-y-4">
          <UCard v-for="i in 10">
            <ModalDownloadProgress :value="Math.floor(Math.random() * 100)" />
          </UCard>
        </div>
      </template>
    </UModal>
  </div>
</template>
