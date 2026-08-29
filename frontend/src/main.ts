/**
 * OnBober application entry point.
 * Bootstraps the Vue 3 app with Vue Router and global styles.
 */
import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(router)
app.mount('#app')
