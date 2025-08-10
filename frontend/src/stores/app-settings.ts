import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

export const useAppSettings = defineStore('app-service-settings', () => {
  const items = ref([{}]);

  return { items };
});
