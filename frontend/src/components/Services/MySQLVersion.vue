<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import ServicesAppVersion from './ServicesAppVersion.vue';
import type { ConfigVersionMySQL } from '@/types';
import { GetMySqlConfig } from '../../../wailsjs/go/config/configs';
import { GetTempDirectory, GetUserOS } from '../../../wailsjs/go/utils/utils';
import { useDownloadStore } from '@/stores/download';

const osMap = {
  windows: 'Windows',
  linux: 'linux',
  darwin: 'macOS',
};

const items = ref<{ version: string; downloadUrl: string }[]>([]);
const isLoading = ref(true);
const download = useDownloadStore();

const refreshItems = async () => {
  try {
    const OS = await GetUserOS();
    const tempFiles = await GetTempDirectory();
    const data = (await GetMySqlConfig()) as ConfigVersionMySQL;
    const currentOs = osMap[OS as keyof typeof osMap];

    if (data && data.mysql) {
      const osData = data.mysql.find((v) => v.os === currentOs);
      if (osData?.data) {
        items.value = osData.data.map((item) => {
          const fileName = `mysql-${item.version}.zip`;
          const isInstalled = tempFiles.some((path) => path.endsWith(fileName));

          return {
            version: item.version,
            downloadUrl: item.link,
            installed: isInstalled,
          };
        });
      }
    }
  } catch (error) {
    console.error('Error fetching MySQL config:', error);
  }
};

const progress = computed(() => download.files[0]?.progress);

watch(progress, (newValue) => {
  if (newValue === 100) {
    refreshItems();
  }
});

onMounted(async () => {
  try {
    await refreshItems();
  } finally {
    isLoading.value = false;
  }
});
</script>

<template>
  <div>
    <div v-if="isLoading" class="text-center font-bold">Loading...</div>
    <ServicesAppVersion
      v-else-if="items.length > 0"
      :items="items"
      name="mysql"
    ></ServicesAppVersion>
    <div v-else>No versions available</div>
  </div>
</template>
