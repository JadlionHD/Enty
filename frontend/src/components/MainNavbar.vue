<script lang="ts" setup>
import { useColorMode } from '@vueuse/core';
import { WindowToggleMaximise, WindowMinimise, Quit } from '../../wailsjs/runtime/runtime';
import { computed, nextTick } from 'vue';

const mode = useColorMode();
const isDark = computed(() => mode.value === 'dark');
const toggleTheme = () => (mode.value = mode.value === 'dark' ? 'light' : 'dark');

const isAppearanceTransition =
  // @ts-expect-error: Transition API
  document.startViewTransition && !window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function transitDark(event?: MouseEvent) {
  if (!isAppearanceTransition || !event) {
    toggleTheme();
    return;
  }
  const x = event.clientX;
  const y = event.clientY;
  const endRadius = Math.hypot(Math.max(x, innerWidth - x), Math.max(y, innerHeight - y));

  const transition = document.startViewTransition(async () => {
    toggleTheme();
    await nextTick();
  });
  transition.ready.then(() => {
    const clipPath = [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`];
    document.documentElement.animate(
      {
        clipPath: isDark.value ? [...clipPath].reverse() : clipPath,
      },
      {
        duration: 600,
        easing: 'ease-in-out',
        pseudoElement: isDark.value ? '::view-transition-old(root)' : '::view-transition-new(root)',
      },
    );
  });
}
</script>

<template>
  <div
    class="w-full h-8 bg-muted transition-all py-2 px-3 flex items-center justify-between fixed top-0 z-10"
    style="--wails-draggable: drag"
  >
    <div class="font-bold text-xl">Enty</div>

    <div class="flex items-center gap-x-4">
      <UButton
        :icon="
          isDark
            ? 'line-md:sunny-outline-to-moon-loop-transition'
            : 'line-md:moon-to-sunny-outline-loop-transition'
        "
        color="info"
        size="md"
        variant="ghost"
        :key="
          isDark
            ? 'toggler-theme-line-md:sunny-outline-to-moon-loop-transition'
            : 'toggler-theme-line-md:moon-to-sunny-outline-loop-transition'
        "
        @click="transitDark"
      />

      <UButton
        size="md"
        variant="ghost"
        color="info"
        icon="i-lucide-minus"
        @click="WindowMinimise()"
      ></UButton>
      <UButton
        size="md"
        variant="ghost"
        color="info"
        icon="i-lucide-square"
        @click="WindowToggleMaximise()"
      ></UButton>
      <UButton size="md" variant="ghost" color="info" icon="i-lucide-x" @click="Quit()"></UButton>
    </div>
  </div>
</template>

<style>
::view-transition-old(root),
::view-transition-new(root) {
  animation: none;
  mix-blend-mode: normal;
}

::view-transition-old(root) {
  z-index: 1;
}
::view-transition-new(root) {
  z-index: 2147483646;
}
.dark::view-transition-old(root) {
  z-index: 2147483646;
}
.dark::view-transition-new(root) {
  z-index: 1;
}
</style>
