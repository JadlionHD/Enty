<script setup lang="ts">
import MainLayout from '@/layouts/MainLayout.vue';
import TerminalCLI from '@/components/TerminalCLI.vue';
</script>

<template>
  <div class="h-screen">
    <MainLayout>
      <template #center>
        <div class="w-full h-full relative">
          <TerminalCLI />
        </div>
      </template>

      <template #right> </template>
    </MainLayout>
  </div>
</template>
