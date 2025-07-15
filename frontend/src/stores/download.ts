import { ref, computed } from 'vue';
import { defineStore } from 'pinia';
import { EventsOn, EventsOff, EventsEmit } from '../../wailsjs/runtime/runtime';
import { DownloadFile } from '../../wailsjs/go/utils/utils';

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
      clean();
    });

    // Clean up file when download is cancelled
    EventsOn('download-cancelled', (name: string) => {
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
    return await DownloadFile(name, fileName, url, buffer ?? 32);
  };

  const cancel = () => {
    EventsEmit('cancel-download-file');
  };

  const clean = () => {
    files.value = [];
  };
  return { files, isDownloading, on, off, download, cancel, clean };
});
