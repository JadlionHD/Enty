<script setup lang="ts">
import MainSidebar from '@/components/MainSidebar.vue';
import MainLayout from '@/layouts/MainLayout.vue';
</script>

<template>
  <div>
    <MainLayout>
      <template #center>
        <div class="w-full min-h-screen flex flex-col gap-y-4 justify-center items-center">
          <div class="text-muted">Nothing in here, lets have a picnic with Enty</div>
          <div class="w-1/2 opacity-60">
            <img src="/enterprise_eagle.png" alt="Enty Enterprise :3" />
          </div>
        </div>
      </template>

      <template #right> </template>
    </MainLayout>
  </div>
</template>
