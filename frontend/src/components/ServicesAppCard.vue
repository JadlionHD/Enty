<script lang="ts" setup>
import type { ServiceInstanceState } from '@/types';
import { computed, ref } from 'vue';
import { RouterLink } from 'vue-router';

const props = defineProps<{
  name: string;
  label: string;
  description: string;
  coloredIcon: string;
}>();

const to = `/services/${props.name}`;
const state = ref<ServiceInstanceState>('running');
const version = ref('1.0.0');
</script>

<template>
  <RouterLink :to="to">
    <UCard class="hover:scale-105 transition-all">
      <div class="flex items-center gap-x-8 justify-center">
        <div>
          <UIcon :name="coloredIcon" class="size-10" />
        </div>

        <div class="flex flex-col gap-y-1">
          <h1 class="font-bold text-xl">{{ label }}</h1>
          <p class="text-muted text-base">{{ description }}</p>
          <p class="text-muted text-base">
            Version: <span class="font-bold">{{ version }}</span>
          </p>
          <div class="text-muted text-base font-bold">
            <div v-if="state === 'running'" class="flex items-center gap-x-2">
              <UChip standalone inset />
              <div>Running</div>
            </div>
            <div v-else-if="state === 'stopped'" class="flex items-center gap-x-2">
              <UChip standalone inset color="error" />
              <div>Stopped</div>
            </div>
          </div>
        </div>
      </div>
    </UCard>
  </RouterLink>
</template>
