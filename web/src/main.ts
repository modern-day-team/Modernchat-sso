import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import i18n from './i18n'

// 样式全量引入（与主仓库同一姿势，体积仅数十 KB 且避免漏样式）；
// 组件一律在 SFC 内按需 import（不 app.use(TDesign)），否则整库无法 tree-shake。
import 'tdesign-vue-next/es/style/index.css'
import './assets/main.css'
// UnoCSS 工具类最后引入
import 'virtual:uno.css'

const app = createApp(App)

app.use(createPinia())
app.use(i18n)
app.use(router)

app.mount('#app')
