<script setup lang="ts">
import MainSidebar from '@/components/MainSidebar.vue';
import ModalDownload from '@/components/ModalDownload.vue';
</script>

<template>
  <ModalDownload></ModalDownload>
  <div class="flex pt-8 pl-64">
    <MainSidebar />
    <div class="flex-1">
      <slot name="center"></slot>
    </div>
    <slot name="right"></slot>
  </div>
</template>
