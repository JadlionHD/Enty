<script setup lang="ts">
import MainSidebar from '@/components/MainSidebar.vue'
import MainLayout from '@/layouts/MainLayout.vue'
import { SERVICE_APPS } from '@/const'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import type { TabsItem } from '@nuxt/ui'

const tabs: TabsItem[] = [
  {
    label: 'Details',
    slot: 'details' as const,
    icon: 'i-lucide-book',
  },
  {
    label: 'Options',
    slot: 'options' as const,
    icon: 'i-lucide-settings',
  },
  {
    label: 'Terminal',
    slot: 'terminal' as const,
    icon: 'i-lucide-terminal',
  },
]
const route = useRoute()

const item = computed(() => {
  return SERVICE_APPS.find((v) => v.name === route.params.app)
})
</script>

<template>
  <div>
    <MainLayout>
      <template #left>
        <MainSidebar />
      </template>

      <template #center>
        <div class="w-full">
          <div v-if="!!item" class="p-4">
            <UCard>
              <div class="flex items-center gap-x-12 justify-center mb-4">
                <div>
                  <UIcon :name="item.coloredIcon" class="size-18" />
                </div>

                <div class="flex flex-col gap-y-1">
                  <h1 class="font-bold text-3xl">{{ item.label }}</h1>
                  <p class="text-muted text-base">{{ item.description }}</p>
                  <div class="flex gap-x-2">
                    <UButton size="sm" variant="solid">Install</UButton>
                    <UButton size="sm" variant="ghost">Version 1.0.0</UButton>
                  </div>
                </div>
              </div>

              <UTabs :items="tabs" class="w-full" variant="link" :ui="{ trigger: 'grow' }">
                <template #details="{ item }">
                  <p>This is the details tab.</p>
                </template>

                <template #options="{ item }">
                  <p>This is the config tab.</p>
                </template>

                <template #terminal="{ item }">
                  <p>This is the terminal tab.</p>
                </template>
              </UTabs>
            </UCard>
          </div>
          <div v-else>Service not found D;</div>
        </div>
      </template>
    </MainLayout>
  </div>
</template>
