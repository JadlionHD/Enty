<script lang="ts" setup>
import { useColorMode } from '@vueuse/core';
import { WindowToggleMaximise, WindowMinimise, Quit } from '../../wailsjs/runtime/runtime';

const mode = useColorMode();
</script>

<template>
  <div
    class="w-full h-8 bg-muted py-2 px-3 flex items-center justify-between fixed top-0 z-10"
    style="--wails-draggable: drag"
  >
    <div class="font-bold text-xl">Enty</div>

    <div class="flex items-center gap-x-4">
      <UButton
        :icon="mode === 'dark' ? 'i-lucide-moon' : 'i-lucide-sun'"
        color="info"
        variant="ghost"
        @click="mode = mode === 'dark' ? 'light' : 'dark'"
      />
      <UButton
        size="sm"
        variant="ghost"
        color="info"
        icon="i-lucide-minus"
        @click="WindowMinimise()"
      ></UButton>
      <UButton
        size="sm"
        variant="ghost"
        color="info"
        icon="i-lucide-square"
        @click="WindowToggleMaximise()"
      ></UButton>
      <UButton size="sm" variant="ghost" color="info" icon="i-lucide-x" @click="Quit()"></UButton>
    </div>
  </div>
</template>
