<script lang="ts" setup>
import { h, resolveComponent, ref, onMounted, computed, watch } from 'vue';
import { useInfiniteScroll } from '@vueuse/core';
import type { TableColumn } from '@nuxt/ui';
import { DownloadFile } from '../../../wailsjs/go/utils/utils';
import { useDownloadStore } from '@/stores/download';

const UBadge = resolveComponent('UBadge');
const UButton = resolveComponent('UButton');

interface Props {
  name: string;
  items: {
    version: string;
    downloadUrl: string;
  }[];
}

const props = defineProps<Props>();

type AppVersion = {
  version: string;
  downloadUrl: string;
};

const data = computed(() => props.items);
const download = useDownloadStore();

const columns: TableColumn<AppVersion>[] = [
  {
    accessorKey: 'version',
    header: 'Version',
    cell: ({ row }) => h('div', { class: 'font-mono font-medium' }, `v${row.getValue('version')}`),
  },
  {
    accessorKey: 'downloadUrl',
    header: () => h('div', { class: 'text-center' }, 'Download'),
    cell: ({ row }) => {
      return h('div', { class: 'flex justify-center' }, [
        h(
          UButton,
          {
            size: 'sm',
            variant: 'outline',
            icon: 'i-heroicons-arrow-down-tray',
            onClick: async () => {
              // window.open(row.getValue('downloadUrl'), '_blank');

              const name = `${props.name}-${row.getValue('version')}`;
              const url = row.getValue('downloadUrl');
              console.log({ url, name }, row);

              try {
                const resultDownload = await download.download(
                  name,
                  `${name}.zip`,
                  url as string,
                  32,
                );
              } catch (e) {
                console.error('error while downloading', e);
                download.clean();
              }
            },
          },
          () => 'Download',
        ),
      ]);
    },
  },
];

const searchInput = ref('');
const globalFilter = ref('');

// Debounce function
const debounce = (func: Function, wait: number) => {
  let timeout: NodeJS.Timeout;
  return function executedFunction(...args: any[]) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};

// Debounced search function
const debouncedSearch = debounce((value: string) => {
  globalFilter.value = value;
}, 300);

// Watch for search input changes
watch(searchInput, (newValue) => {
  debouncedSearch(newValue);
});

// Client-side filtering for better performance
const filteredData = computed(() => {
  if (!globalFilter.value) return data.value;

  const searchTerm = globalFilter.value.toLowerCase();
  return data.value.filter((item) => item.version.toLowerCase().includes(searchTerm));
});
</script>

<template>
  <div class="flex flex-col flex-1 w-full">
    <div class="flex px-4 py-3.5">
      <UInput
        v-model="searchInput"
        class="max-w-sm"
        placeholder="Search versions..."
        icon="i-heroicons-magnifying-glass"
      />
    </div>

    <div class="overflow-auto max-h-64 z-0">
      <UTable
        ref="table"
        :data="filteredData"
        :columns="columns"
        empty="No versions found."
      ></UTable>
    </div>
  </div>
</template>
