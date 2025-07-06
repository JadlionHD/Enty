<script setup lang="ts">
import MainSidebar from '@/components/MainSidebar.vue'
import MainLayout from '@/layouts/MainLayout.vue'
</script>

<template>
  <div>
    <MainLayout>
      <template #left>
        <MainSidebar />
      </template>

      <template #center>
        <div>Center</div>
      </template>

      <template #right> </template>
    </MainLayout>
  </div>
</template>
