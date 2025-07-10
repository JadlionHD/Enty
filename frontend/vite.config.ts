import { fileURLToPath, URL } from 'node:url';

import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import ui from '@nuxt/ui/vite';
import tailwindcss from '@tailwindcss/vite';
import Components from 'unplugin-vue-components/vite';
import MotionResolver from 'motion-v/resolver';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    ui(),
    tailwindcss(),
    // TODO: duplicate auto import components
    Components({
      dts: true,
      resolvers: [MotionResolver()],
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
});
