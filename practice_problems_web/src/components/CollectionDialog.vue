<template>
  <div class="collection-container">
    <div class="collection-header">
      <h2>我的集合</h2>
      <el-button 
        type="primary" 
        icon="Close" 
        circle 
        plain 
        @click="$emit('close')"
      />
    </div>
    
    <div class="collection-body">
      <el-empty description="集合功能开发中..." />
    </div>
  </div>
</template>

<script setup lang="ts">
defineEmits(['close']);
</script>

<style scoped>
.collection-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.collection-header {
  padding: 20px 24px;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.collection-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.collection-body {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
