<script setup lang="ts">
import { ElConfigProvider } from 'element-plus'
// 引入中文包
import zhCn from 'element-plus/es/locale/lang/zh-cn'
</script>

<template>
  <!-- 这里包裹一层配置组件，把语言设为中文 -->
  <el-config-provider :locale="zhCn">
    <div>
      <router-view />
    </div>
  </el-config-provider>
</template>

<style>
body { margin: 0; }
</style>