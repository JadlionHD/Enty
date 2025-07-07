<script setup lang="ts">
import { SERVICE_APPS } from '@/const'
import type { NavigationMenuItem } from '@nuxt/ui'
import { computed, ref } from 'vue'
import { useWindowSize } from '@vueuse/core'

const { width, height } = useWindowSize()

const items = computed<NavigationMenuItem[][]>(() => {
  const services = SERVICE_APPS.map((v) => {
    return {
      label: v.label,
      icon: v.plainIcon,
      to: `/services/${v.name}`,
    }
  })

  return [
    [
      {
        label: 'General',
        to: '/',
        icon: 'i-lucide-settings',
      },
      {
        label: 'Services',
        icon: 'i-lucide-database',
        to: '/services',
        defaultOpen: true,
        children: [...services],
      },
    ],
    [
      {
        label: 'Terminal',
        to: '/terminal',
        icon: 'i-lucide-terminal',
      },
    ],
    [
      {
        label: 'GitHub',
        icon: 'i-lucide-github',
        badge: '1',
        to: 'https://github.com/JadlionHD/Enty',
        target: '_blank',
      },
      {
        label: 'Help',
        icon: 'i-lucide-circle-help',
        disabled: true,
      },
    ],
  ]
})

const isCollapsed = computed<boolean>(() => width.value < 640)

console.log(items.value)
// const items = ref<NavigationMenuItem[][]>([])
</script>

<template>
  <UNavigationMenu
    orientation="vertical"
    :items="items"
    :collapsed="isCollapsed"
    :popover="isCollapsed"
    class="sm:data-[orientation=vertical]:w-48 w-8 font-bold"
  />
</template>
