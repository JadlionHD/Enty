<script setup lang="ts">
import MainSidebar from '@/components/MainSidebar.vue'
import { SERVICE_APPS } from '@/const'
import MainLayout from '@/layouts/MainLayout.vue'
import { computed } from 'vue'

const items = computed(() => {
  const apps = SERVICE_APPS.map((v) => {
    return {
      name: v.name,
      label: v.label,
      description: v.description,
      coloredIcon: v.coloredIcon,
    }
  })
  return apps
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
          <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-4 p-4">
            <div v-for="item in items" class="">
              <ServicesAppCard v-bind="item"></ServicesAppCard>
            </div>
          </div>
        </div>
      </template>

      <template #right> </template>
    </MainLayout>
  </div>
</template>
