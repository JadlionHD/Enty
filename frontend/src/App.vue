<script setup lang="ts">
import MainNavbar from '@/components/MainNavbar.vue';
import { RouterView } from 'vue-router';
</script>

<template>
  <UApp>
    <MainNavbar key="main-navbar"></MainNavbar>
    <RouterView />
  </UApp>
</template>
