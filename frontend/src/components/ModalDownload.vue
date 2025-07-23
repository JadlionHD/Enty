<script setup lang="ts">
import { useDownloadStore } from '@/stores/download';
import { onMounted, onUnmounted, ref } from 'vue';

const value = ref(100);

const download = useDownloadStore();

onMounted(() => {
  download.on(() => {});
});

onUnmounted(() => {
  download.off();
});
</script>

<template>
  <div class="z-20">
    <UModal title="Downloads">
      <UButton
        class="absolute bottom-4 right-4"
        variant="outline"
        icon="i-lucide-download"
        color="secondary"
      ></UButton>

      <template #body>
        <div class="flex flex-col gap-y-4">
          <UCard v-for="item in download.files" v-if="!!download.files.length">
            <ModalDownloadProgress :title="item.name" :value="item.progress" />
          </UCard>
          <div v-else>No downloads available, everything clear here</div>
        </div>
      </template>
    </UModal>
  </div>
</template>
