import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import pinia, { initStores } from './store' // 👈 导入 initStores
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

const app = createApp(App)
app.use(router)
app.use(pinia)
app.use(ElementPlus)

// ✅ 新增：在 app.use(pinia) 之后初始化
initStores()

app.mount('#app')