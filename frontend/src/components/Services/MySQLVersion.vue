<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue';
import ServicesAppVersion from './ServicesAppVersion.vue';
import type { ConfigVersionMySQL } from '@/types';
import { GetMySqlConfig } from '../../../wailsjs/go/config/configs';
import { GetTempDirectory, GetUserOS } from '../../../wailsjs/go/utils/utils';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';

const osMap = {
  windows: 'Windows',
  linux: 'linux',
  darwin: 'macOS',
};

const items = ref<{ version: string; downloadUrl: string }[]>([]);
const isLoading = ref(true);

onMounted(async () => {
  try {
    const OS = await GetUserOS();
    const dir = await GetTempDirectory();
    console.log(dir);
    const data = (await GetMySqlConfig()) as ConfigVersionMySQL;
    const currentOs = osMap[OS as keyof typeof osMap];

    if (data && data.mysql) {
      const osData = data.mysql.find((v) => v.os === currentOs);
      if (osData?.data) {
        // Transform the data to match ServicesAppVersion props
        items.value = osData.data.map((item) => ({
          version: item.version,
          downloadUrl: item.link, // Handle different property names
        }));
      }
    }
  } catch (error) {
    console.error('Error fetching MySQL config:', error);
  } finally {
    isLoading.value = false;
  }
});

// onMounted(() => {
//   EventsOn('download-file', (name, totalBytes, downloadedBytes) => {
//     const progress = (100 * downloadedBytes) / totalBytes;

//     console.log({
//       name,
//       totalBytes,
//       downloadedBytes,
//       progress,
//     });
//   });
// });

// onUnmounted(() => {
//   EventsOff('download-file');
// });
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
