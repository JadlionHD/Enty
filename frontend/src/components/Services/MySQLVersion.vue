<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import ServicesAppVersion from './ServicesAppVersion.vue';
import type { ConfigVersionMySQL } from '@/types';
import { GetMySqlConfig } from '../../../wailsjs/go/main/App';
import { GetUserOS } from '../../../wailsjs/go/utils/utils';

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
    const data = (await GetMySqlConfig()) as ConfigVersionMySQL;
    const currentOs = osMap[OS as keyof typeof osMap];

    if (data && data.mysql) {
      const osData = data.mysql.find((v) => v.os === currentOs);
      if (osData?.data) {
        // Transform the data to match ServicesAppVersion props
        items.value = osData.data.map((item) => ({
          version: item.version,
          downloadUrl: (item as any).downloadUrl || (item as any).url || '#', // Handle different property names
        }));
      }
    }
  } catch (error) {
    console.error('Error fetching MySQL config:', error);
  } finally {
    isLoading.value = false;
  }
});
</script>

<template>
  <div>
    <div v-if="isLoading">Loading...</div>
    <ServicesAppVersion v-else-if="items.length > 0" :items="items"></ServicesAppVersion>
    <div v-else>No versions available</div>
  </div>
</template>
