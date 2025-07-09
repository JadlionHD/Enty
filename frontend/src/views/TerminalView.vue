<script setup lang="ts">
import MainSidebar from '@/components/MainSidebar.vue';
import MainLayout from '@/layouts/MainLayout.vue';

import { onMounted, ref } from 'vue';
</script>

<template>
  <div>
    <MainLayout>
      <template #center>
        <div class="w-full h-full">
          <TerminalCLI />
        </div>
      </template>

      <template #right> </template>
    </MainLayout>
  </div>
</template>
