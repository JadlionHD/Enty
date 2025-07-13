import { ref, computed } from 'vue';
import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

export const useDownloadStore = defineStore('download', () => {
  const files = ref([]);

  const on = (
    cb: (name: string, totalBytes: number, downloadedBytes: number, progress: number) => void,
  ) => {
    EventsOn('download-file', (name, totalBytes, downloadedBytes) => {
      const progress = (100 * downloadedBytes) / totalBytes;
      cb(name, totalBytes, downloadedBytes, progress);
    });
  };

  const off = () => {
    EventsOff('download-file');
  };

  return { files, on, off };
});
