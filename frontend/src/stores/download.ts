import { ref, computed } from 'vue';
import { defineStore } from 'pinia';
import { EventsOn, EventsOff, EventsEmit } from '../../wailsjs/runtime/runtime';
import { DownloadFile } from '../../wailsjs/go/utils/utils';
import { useToast } from '@nuxt/ui/runtime/composables/useToast.js';

interface DownloadFile {
  name: string;
  totalBytes: number;
  downloadedBytes: number;
  progress: number;
}

export const useDownloadStore = defineStore('download', () => {
  const files = ref<DownloadFile[]>([]);
  // TODO: next this
  const isDownloading = ref(false);
  const toast = useToast();

  const on = (
    cb: (name: string, totalBytes: number, downloadedBytes: number, progress: number) => void,
  ) => {
    // Listen for download start
    EventsOn('start-download-file', (name: string, totalBytes: number) => {
      files.value.push({
        name,
        totalBytes,
        downloadedBytes: 0,
        progress: 0,
      });
      // Call callback with initial state
      cb(name, totalBytes, 0, 0);
    });

    EventsOn('download-file', (name, totalBytes, downloadedBytes) => {
      const progress = (100 * downloadedBytes) / totalBytes;

      // Update or add file in the files array
      const fileIndex = files.value.findIndex((f) => f.name === name);
      if (fileIndex !== -1) {
        files.value[fileIndex] = { name, totalBytes, downloadedBytes, progress };
      } else {
        files.value.push({ name, totalBytes, downloadedBytes, progress });
      }

      cb(name, totalBytes, downloadedBytes, progress);
    });

    // Clean up file when download is finished
    EventsOn('finish-download-file', (name: string) => {
      toast.add({
        title: `Successfully installed ${name}`,
        icon: 'i-lucide-check',
      });
      clean();
    });

    // Clean up file when download is cancelled
    EventsOn('download-cancelled', (name: string) => {
      toast.add({
        title: `Cancelled installing ${name}`,
        icon: 'i-lucide-file-x-2',
      });
      clean();
    });
  };

  const off = () => {
    EventsOff('start-download-file');
    EventsOff('download-file');
    EventsOff('finish-download-file');
    EventsOff('download-cancelled');
  };

  const download = async (name: string, fileName: string, url: string, buffer?: number) => {
    toast.add({
      title: `Start downloading ${name}`,
      icon: 'i-lucide-file-down',
    });
    return await DownloadFile(name, fileName, url, buffer ?? 32);
  };

  const cancel = () => {
    EventsEmit('cancel-download-file');
  };

  const clean = () => {
    files.value = [];
  };
  return { files, toast, isDownloading, on, off, download, cancel, clean };
});
