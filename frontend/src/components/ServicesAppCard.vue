<script lang="ts" setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps<{
  name: string
  label: string
  description: string
  coloredIcon: string
}>()

const to = `/services/${props.name}`
</script>

<template>
  <RouterLink :to="to">
    <UCard class="hover:scale-105 transition-all">
      <div class="flex items-center gap-x-8 justify-center">
        <div>
          <UIcon :name="coloredIcon" class="size-10" />
        </div>

        <div class="">
          <h1 class="font-bold text-xl">{{ label }}</h1>
          <p class="text-muted text-base">{{ description }}</p>
        </div>
      </div>
    </UCard>
  </RouterLink>
</template>
