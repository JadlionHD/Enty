import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import ui from '@nuxt/ui/vue-plugin'
import router from './router'

const app = createApp(App)

app.use(ui)
app.use(createPinia())
app.use(router)

app.mount('#app')
