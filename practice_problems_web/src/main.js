import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import '@wangeditor/editor/dist/css/style.css'
import App from './App.vue'

// 🔥🔥🔥 1. 必须引入 router！
import router from './router' 

const app = createApp(App)

// 🔥🔥🔥 2. 必须注册 router！没有这行代码，页面就是白的！
app.use(router)

// 3. 注册图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 4. 使用 ElementPlus
app.use(ElementPlus)

// 5. 全局阻止回车刷新
window.addEventListener('submit', (e) => { e.preventDefault(); }, false);

// =======================================================
// ★★★ 修复版：全局防重复提交指令 v-reclick ★★★
// =======================================================
app.directive('reclick', {
  mounted(el, binding) {
    el.addEventListener('click', async (e) => {
      // 如果按钮是禁用状态，直接忽略
      if (el.disabled || el.classList.contains('is-disabled') || el.classList.contains('is-loading')) {
        e.stopImmediatePropagation(); // 阻止后续事件
        return;
      }

      // 1. 设置 loading 和 禁用
      el.classList.add('is-loading');
      // 动态添加 loading 图标
      const icon = document.createElement('i');
      icon.className = 'el-icon is-loading';
      icon.innerHTML = '<svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="1em" height="1em"><path fill="currentColor" d="M512 64a32 32 0 0 1 32 32v192a32 32 0 0 1-64 0V96a32 32 0 0 1 32-32zm0 640a32 32 0 0 1 32 32v192a32 32 0 0 1-64 0V736a32 32 0 0 1 32-32zm448-192a32 32 0 0 1-32 32H736a32 32 0 1 1 0-64h192a32 32 0 0 1 32 32zm-640 0a32 32 0 0 1-32 32H96a32 32 0 0 1 0-64h192a32 32 0 0 1 32 32zM195.2 195.2a32 32 0 0 1 45.248 0L376.32 331.008a32 32 0 0 1-45.248 45.248L195.2 240.448a32 32 0 0 1 0-45.248zm452.544 452.544a32 32 0 0 1 45.248 0L828.8 783.68a32 32 0 0 1-45.248 45.248L647.744 692.992a32 32 0 0 1 0-45.248zM828.8 195.264a32 32 0 0 1 0 45.184L692.992 376.32a32 32 0 0 1-45.248-45.248l135.808-135.808a32 32 0 0 1 45.248 0zm-452.544 452.48a32 32 0 0 1 0 45.248L240.448 828.8a32 32 0 0 1-45.248-45.248l135.808-135.808a32 32 0 0 1 45.248 0z"></path></svg>';
      el.prepend(icon);
      
      // 禁用点击
      el.style.pointerEvents = 'none';

      try {
        // 2. 执行绑定的函数
        if (typeof binding.value === 'function') {
          const result = binding.value();
          // 如果是 Promise (async函数)，等待它完成
          if (result instanceof Promise) {
            await result;
          } else {
            // 如果不是 Promise，人为延迟 500ms 以展示反馈
            await new Promise(resolve => setTimeout(resolve, 500));
          }
        }
      } catch (err) {
        console.error(err);
      } finally {
        // 3. 恢复状态
        el.classList.remove('is-loading');
        if (icon) icon.remove();
        el.style.pointerEvents = 'auto';
      }
    }, true); // useCapture = true 确保先执行指令逻辑
  }
});

app.mount('#app')
