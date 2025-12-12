<template>
  <div class="collection-container">
    <!-- 头部区域：标题 + 集合标签在同一行 -->
    <div class="header-section">
      <div class="header-top" style="display: flex; align-items: center; justify-content: space-between; gap: 20px;">
        <!-- 左侧：标题 + 创建集合 + 集合标签 -->
        <div style="display: flex; align-items: center; gap: 20px; flex: 1; min-width: 0;">
          <div style="display: flex; align-items: center; gap: 12px; flex-shrink: 0;">
            <div class="logo-box">
              <el-icon><Collection /></el-icon>
            </div>
            <h2 style="margin: 0; color: #fff;">
              我的集合
              <span class="wordbook-entry" @click="goToWordbook">词</span>
            </h2>
          </div>
          
          <!-- 创建集合按钮 -->
          <el-button 
            type="primary" 
            size="small" 
            icon="Plus"
            @click="handleCreateCollection" 
            plain
            style="flex-shrink: 0;"
          >
            集合
          </el-button>
          
          <!-- 集合标签导航栏（横向） -->
          <div class="collection-tabs" style="flex: 1; min-width: 0; margin: 0;">
            <div 
              v-for="collection in collections" 
              :key="collection.id"
              class="collection-tab"
              :class="{ 'active': currentCollectionId === collection.id }"
            >
              <span @click="selectCollection(collection.id)">{{ collection.name }}</span>
              <!-- 只有所有者才显示三个点菜单 -->
              <el-dropdown 
                v-if="collection.isOwner"
                trigger="click" 
                @command="(cmd: string) => handleCollectionMenu(cmd, collection)"
              >
                <span class="more-icon" @click.stop>
                  <span class="dots-vertical">⋮</span>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">
                      <el-icon><Edit /></el-icon>
                      修改名称
                    </el-dropdown-item>
                    <el-dropdown-item command="manage" divided>
                      <el-icon><Setting /></el-icon>
                      管理
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" divided>
                      <el-icon><Delete /></el-icon>
                      删除集合
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </div>
        
        <!-- 右侧：返回按钮 -->
        <el-button type="primary" size="small" @click="handleBack" plain style="flex-shrink: 0;">返回主页</el-button>
      </div>
    </div>

    <el-row :gutter="20" style="margin-top: 12px; height: calc(100vh - 100px); overflow: hidden; padding: 0 20px;">
      <!-- 左侧：知识点列表 -->
      <el-col :span="6" style="height: 100%;">
        <el-card class="category-list-card" style="height: 100%; box-shadow: 0 4px 20px rgba(118, 75, 162, 0.15); display: flex; flex-direction: column;">
          <template #header>
            <div class="card-header" style="display: flex; align-items: center; justify-content: space-between;">
              <div style="display: flex; align-items: center; gap: 8px; flex-shrink: 0;">
                <el-icon style="font-size: 18px; color: #764ba2;"><List /></el-icon>
                <span style="font-size: 16px; font-weight: 600;">知识列表</span>
              </div>
              <div style="display: flex; gap: 6px; align-items: center; flex-shrink: 0;">
                <!-- 二级按钮组（随机模式下的辅助操作） -->
                <template v-if="isRandomMode">
                  <el-tooltip content="重新打乱" placement="top">
                    <el-button 
                      size="small" 
                      @click="refreshRandomOrder"
                      :icon="Refresh"
                      text
                    />
                  </el-tooltip>
                  <el-tooltip content="重新加载" placement="top">
                    <el-button 
                      size="small" 
                      @click="reloadPoints"
                      :icon="Download"
                      text
                    />
                  </el-tooltip>
                  <el-divider direction="vertical" style="height: 18px; margin: 0 4px;" />
                </template>
                
                <!-- 一级按钮组 -->
                <!-- 模式切换按钮 -->
                <el-button 
                  size="small" 
                  :type="isRandomMode ? 'warning' : ''"
                  @click="toggleMode"
                  :plain="!isRandomMode"
                  style="min-width: 60px;"
                >
                  <el-icon style="margin-right: 4px;"><Refresh v-if="isRandomMode" /><List v-else /></el-icon>
                  {{ isRandomMode ? '随机' : '普通' }}
                </el-button>
                <el-button 
                  :icon="Trophy" 
                  size="small" 
                  type="primary"
                  @click="handleStartPractice"
                  :disabled="collectionPoints.length === 0"
                  style="min-width: 60px;"
                >
                  刷题
                </el-button>
                <!-- 只有所有者才能看到编辑按钮 -->
                <el-button 
                  v-if="isCollectionOwner"
                  :icon="Edit" 
                  size="small" 
                  :type="isEditMode ? 'success' : ''"
                  @click="isEditMode = !isEditMode"
                  :plain="!isEditMode"
                  style="min-width: 60px;"
                >
                  {{ isEditMode ? '完成' : '编辑' }}
                </el-button>
              </div>
            </div>
          </template>
          <div style="flex: 1; display: flex; flex-direction: column; min-height: 0; position: relative;">
            <!-- 排序标签页（悬浮样式） -->
            <div style="position: absolute; top: 0; right: 0; z-index: 10; display: flex; gap: 3px;">
              <div 
                v-if="categorySortOrder"
                @click="togglePointSort"
                :class="['sort-tab-compact', { 'active': pointSortOrder, 'disabled': !categorySortOrder }]"
              >
                <el-icon v-if="pointSortOrder === 'asc'" style="font-size: 9px;"><SortUp /></el-icon>
                <el-icon v-else-if="pointSortOrder === 'desc'" style="font-size: 9px;"><SortDown /></el-icon>
                <span>知识</span>
              </div>
              <div 
                @click="toggleCategorySort"
                :class="['sort-tab-compact', { 'active': categorySortOrder, 'disabled': false }]"
              >
                <el-icon v-if="categorySortOrder === 'asc'" style="font-size: 9px;"><SortUp /></el-icon>
                <el-icon v-else-if="categorySortOrder === 'desc'" style="font-size: 9px;"><SortDown /></el-icon>
                <span>分类</span>
              </div>
            </div>
            <div style="flex: 1; overflow-y: auto; overflow-x: hidden;">
            <div v-if="pointsLoading" style="text-align: center; padding: 40px;">
              <el-icon class="is-loading" style="font-size: 32px;"><Loading /></el-icon>
            </div>
            <template v-else>
              <div v-if="collectionPoints.length === 0">
                <el-empty description="暂无知识点" :image-size="80" />
              </div>
              <div v-else class="points-list">
                <div 
                  v-for="(point, index) in collectionPoints" 
                  :key="point.id"
                  class="point-item"
                  :class="{ 'active': selectedPointId === point.pointId }"
                >
                  <div :class="['point-card', getDifficultyClass(point.pointDifficulty)]" @click="selectPoint(point.pointId)">
                    <div class="point-card-body">
                      <div class="point-title">{{ point.title }}</div>
                      <div class="point-meta">
                        <span class="meta-item" :style="{ backgroundColor: getSubjectColor(point.subjectName), padding: '2px 6px', borderRadius: '3px', color: '#333', fontWeight: '500' }">
                          <el-icon style="font-size: 14px; filter: drop-shadow(0 0 1px white) drop-shadow(0 0 1px white);"><FolderOpened /></el-icon>
                          {{ point.subjectName }}
                        </span>
                        <span class="meta-item" :style="{ backgroundColor: getCategoryColor(point.categoryName), padding: '2px 6px', borderRadius: '3px', color: '#333', fontWeight: '500' }">
                          <el-icon style="font-size: 14px; filter: drop-shadow(0 0 1px white) drop-shadow(0 0 1px white);"><CollectionIcon /></el-icon>
                          {{ point.categoryName }}
                        </span>
                      </div>
                    </div>
                  </div>
                  <!-- 操作按钮组（仅在编辑模式显示） -->
                  <div v-if="isEditMode" class="point-actions-expanded">
                    <el-button 
                      :icon="Top" 
                      size="small" 
                      @click.stop="movePointToTop(index)"
                      :disabled="index === 0"
                    >
                      置顶
                    </el-button>
                    <el-button 
                      :icon="ArrowUp" 
                      size="small" 
                      @click.stop="movePointUp(index)"
                      :disabled="index === 0"
                    >
                      上移
                    </el-button>
                    <el-button 
                      :icon="ArrowDown" 
                      size="small" 
                      @click.stop="movePointDown(index)"
                      :disabled="index === collectionPoints.length - 1"
                    >
                      下移
                    </el-button>
                    <el-button 
                      :icon="Delete" 
                      size="small" 
                      type="danger"
                      @click.stop="removePoint(point)"
                    >
                      移除
                    </el-button>
                  </div>
                </div>
              </div>
            </template>
            </div>
            <!-- 分页（仅普通模式显示） -->
            <div v-if="!isRandomMode && pointsTotal > 0" style="padding: 16px; background: #fff; border-top: 1px solid #ebeef5;">
              <el-pagination
                v-model:current-page="pointsPage"
                :page-size="pointsPageSize"
                :total="pointsTotal"
                layout="prev, pager, next"
                small
                @current-change="fetchCollectionPoints"
                style="justify-content: center;"
              />
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：知识点内容 -->
      <el-col :span="18" style="height: 100%;">
        <div style="height: 100%; box-shadow: 0 4px 20px rgba(118, 75, 162, 0.15); border-radius: 12px; overflow: hidden; background: rgba(255, 255, 255, 0.85); display: flex; flex-direction: column;">
          <DetailPanel 
            :currentPoint="selectedPointDetail"
            :canGoBack="navigationStack.length > 0"
            :hasPermission="false"
            viewMode="view"
            :currentSubject="null"
            :currentPointBindings="currentPointBindings"
            :pointsInfoMap="new Map()"
            :userInfo="null"
            :isPointOwner="false"
            :drawerVisible="pointPracticeDrawerVisible"
            :parsedLinks="[]"
            :editTitleDialog="null"
            @update:drawerVisible="(val) => pointPracticeDrawerVisible = val"
            @open-practice="pointPracticeDrawerVisible = true"
            @navigate-to-point="(data: any) => navigateToPointFromBinding(data.pointId)"
            @navigate-back="goBackToPreviousPoint"
            style="flex: 1; min-height: 0;"
          />
        </div>
      </el-col>
    </el-row>

    <!-- 创建集合弹框 -->
    <el-dialog 
      v-model="showCreateDialog" 
      title="创建集合" 
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="集合名称">
          <el-input 
            v-model="createForm.name" 
            placeholder="请输入集合名称"
            maxlength="20"
            show-word-limit
            clearable
          />
          <div style="color: #909399; font-size: 12px; margin-top: 8px;">
            ℹ️ 系统将自动添加序号（如：1. 集合名称）
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmCreate" :loading="createLoading">创建</el-button>
        </span>
      </template>
    </el-dialog>
    <!-- 编辑集合弹框 -->
    <el-dialog 
      v-model="showEditDialog" 
      title="修改集合名称" 
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="集合名称">
          <el-input 
            v-model="editForm.name" 
            placeholder="请输入集合名称"
            maxlength="20"
            show-word-limit
            clearable
          />
          <div style="color: #909399; font-size: 12px; margin-top: 8px;">
            ℹ️ 序号不可修改，只更新名称部分
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showEditDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmEdit" :loading="editLoading">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 集合综合刷题抽屉 -->
    <CollectionPracticeDrawer 
      v-model:visible="practiceDrawerVisible"
      :collection-id="currentCollectionId"
      :collection-name="currentCollectionName"
    />
    
    <!-- 单词本 -->
    <WordBook v-model:visible="wordbookVisible" />

    <!-- 集合管理对话框 -->
    <el-dialog 
      v-model="showManageDialog" 
      :title="`管理集合：${currentManageCollection?.name || ''}`" 
      width="1000px"
      height="600px"
      :close-on-click-modal="false"
      class="collection-manage-dialog"
      @closed="handleManageDialogClosed"
    >
      <el-tabs v-model="manageTabActive" class="manage-tabs">
        <!-- 设置访问权限标签页 -->
        <el-tab-pane label="设置访问权限" name="permission">
          <div class="tab-content">
            <!-- 集合基本信息 -->
            <el-card v-if="currentManageCollection" class="collection-info-card" shadow="never">
              <div class="collection-info-grid">
                <div class="info-item">
                  <div class="info-label">
                    <el-icon><CollectionIcon /></el-icon>
                    集合名称
                  </div>
                  <div class="info-value">{{ currentManageCollection.name }}</div>
                </div>
                <div class="info-item">
                  <div class="info-label">
                    <el-icon><List /></el-icon>
                    知识点数量
                  </div>
                  <div class="info-value">{{ collectionPoints.length }} 个</div>
                </div>
                <div class="info-item">
                  <div class="info-label">
                    <el-icon><Lock v-if="!permissionForm.isPublic" /><Unlock v-else /></el-icon>
                    访问权限
                  </div>
                  <div class="info-value">
                    <el-tag :type="permissionForm.isPublic ? 'success' : 'warning'">
                      {{ permissionForm.isPublic ? '公有' : '私有' }}
                    </el-tag>
                  </div>
                </div>
                <div class="info-item">
                  <div class="info-label">
                    <el-icon><User /></el-icon>
                    所有者
                  </div>
                  <div class="info-value">{{ currentManageCollection.ownerUserCode }}</div>
                </div>
              </div>
            </el-card>
            
            <el-form :model="permissionForm" label-width="100px" :disabled="!currentManageCollection?.isOwner">
              <el-form-item label="访问权限">
                <el-radio-group v-model="permissionForm.isPublic">
                  <el-radio :label="true">公有（所有人可见）</el-radio>
                  <el-radio :label="false">私有（仅自己和授权用户可见）</el-radio>
                </el-radio-group>
              </el-form-item>
              
              <!-- 权限说明 -->
              <el-alert 
                v-if="permissionForm.isPublic"
                title="公有集合：所有用户都可以查看和使用此集合" 
                type="info" 
                show-icon
                class="permission-info"
              />
              <el-alert 
                v-else
                title="私有集合：仅您和授权用户可以在有效期内查看和使用此集合" 
                type="info" 
                show-icon
                class="permission-info"
              />
            </el-form>
            
            <!-- 权限对比卡片 -->
            <el-card class="permission-comparison-card" shadow="never" v-if="currentManageCollection?.isOwner">
              <div slot="header" class="clearfix">
                <span>权限设置对比</span>
              </div>
              <el-row :gutter="20">
                <el-col :span="12">
                  <div class="permission-type public">
                    <div class="permission-title">
                      <el-icon><Unlock /></el-icon>
                      公有集合
                    </div>
                    <ul class="permission-list">
                      <li>✓ 所有用户可见</li>
                      <li>✓ 无需授权即可访问</li>
                      <li>✗ 无法控制具体用户</li>
                    </ul>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="permission-type private">
                    <div class="permission-title">
                      <el-icon><Lock /></el-icon>
                      私有集合
                    </div>
                    <ul class="permission-list">
                      <li>✓ 精确控制访问用户</li>
                      <li>✓ 可设置过期时间</li>
                      <li>✗ 需要手动授权</li>
                    </ul>
                  </div>
                </el-col>
              </el-row>
            </el-card>
            
            <div class="form-actions">
              <el-button 
                v-if="currentManageCollection?.isOwner"
                type="primary" 
                @click="savePermission"
                :loading="manageLoading"
              >
                保存设置
              </el-button>
            </div>
            
            <!-- 添加授权（只在私有集合时显示） -->
            <div v-if="!permissionForm.isPublic && currentManageCollection?.isOwner" class="auth-form-container">
              <div class="section-title">添加授权</div>
              <div class="simple-auth-form">
                <div class="form-row">
                  <div class="form-group">
                    <label class="form-label">用户Code *</label>
                    <el-input 
                      v-model="authForm.userCode" 
                      placeholder="请输入用户Code"
                      class="form-input"
                    />
                  </div>
                  <div class="form-group">
                    <label class="form-label">过期时间</label>
                    <el-date-picker
                      v-model="authForm.expireTime"
                      type="datetime"
                      placeholder="请选择过期时间"
                      format="YYYY-MM-DD HH:mm:ss"
                      value-format="YYYY-MM-DD HH:mm:ss"
                      class="form-datepicker"
                    />
                  </div>
                  <div class="form-group form-button-group">
                    <el-button 
                      type="primary" 
                      @click="addAuthorization"
                      :loading="manageLoading"
                      class="add-auth-btn"
                    >
                      添加
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
        
        <!-- 授权管理标签页 -->
        <el-tab-pane label="授权管理" name="authorization" v-if="currentManageCollection?.isOwner && !permissionForm.isPublic">
          <div class="tab-content">
            <!-- 授权列表 -->
            <div class="auth-list-container">
              <div class="section-title">已授权用户</div>
              <!-- 授权列表工具栏 -->
              <div class="auth-list-toolbar">
                <div class="auth-stats">
                  共 {{ authTotal }} 个授权用户
                </div>
                <el-input
                  v-model="authSearchKeyword"
                  placeholder="搜索用户Code"
                  clearable
                  style="width: 200px;"
                  size="small"
                  @input="handleAuthSearch"
                >
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                </el-input>
              </div>
              
              <el-table 
                :data="authorizations" 
                style="width: 100%"
                v-loading="manageLoading"
                empty-text="暂无授权用户"
                class="auth-table"
                :show-header="true"
                :header-cell-style="{ 
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', 
                  color: '#ffffff', 
                  fontWeight: '600',
                  fontSize: '13px'
                }"
              >
                <el-table-column prop="userCode" label="用户Code" width="170" align="center">
                  <template #default="{ row }">
                    <div style="display: flex; align-items: center; justify-content: center; gap: 8px;">
                      <el-icon style="color: #667eea; font-size: 16px;"><User /></el-icon>
                      <span style="font-weight: 500;">{{ row.userCode }}</span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column prop="nickname" label="昵称" width="140" align="center">
                  <template #default="{ row }">
                    <el-tag size="small" effect="plain" style="border-radius: 12px;">
                      {{ row.nickname }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="email" label="邮箱" width="180">
                  <template #default="{ row }">
                    <div style="display: flex; align-items: center; gap: 6px;">
                      <el-icon style="color: #909399; font-size: 14px;"><Message /></el-icon>
                      <span style="color: #606266;">{{ row.email }}</span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="过期时间" width="260" align="center">
                  <template #default="{ row }">
                    <el-tag 
                      v-if="row.expireTime" 
                      type="warning" 
                      size="small" 
                      effect="light"
                      style="border-radius: 8px; font-weight: 500;"
                    >
                      <el-icon style="margin-right: 4px;"><Clock /></el-icon>
                      {{ row.expireTime }}
                    </el-tag>
                    <el-tag 
                      v-else 
                      type="success" 
                      size="small" 
                      effect="light"
                      style="border-radius: 8px; font-weight: 500;"
                    >
                      <el-icon style="margin-right: 4px;"><Check /></el-icon>
                      永久
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="170" align="center" fixed="right">
                  <template #default="{ row }">
                    <div style="display: flex; gap: 8px; justify-content: center;">
                      <el-button 
                        link 
                        type="primary" 
                        size="small" 
                        @click="editAuthorization(row)"
                        style="font-weight: 500;"
                      >
                        <el-icon style="margin-right: 2px;"><Edit /></el-icon>
                        修改
                      </el-button>
                      <el-button 
                        link 
                        type="danger" 
                        size="small" 
                        @click="deleteAuthorization(row)"
                        style="font-weight: 500;"
                      >
                        <el-icon style="margin-right: 2px;"><Delete /></el-icon>
                        删除
                      </el-button>
                    </div>
                  </template>
                </el-table-column>
              </el-table>
              
              <!-- 分页 -->
              <div v-if="authTotal > 0" style="padding: 16px 0; display: flex; justify-content: center;">
                <el-pagination
                  v-model:current-page="authPage"
                  v-model:page-size="authPageSize"
                  :total="authTotal"
                  :page-sizes="[10, 20, 50]"
                  layout="total, sizes, prev, pager, next, jumper"
                  @size-change="loadAuthorizations"
                  @current-change="loadAuthorizations"
                  small
                />
              </div>
            </div>
          </div>
        </el-tab-pane>
        
        <!-- 授权管理标签页（删除） -->
      </el-tabs>
    </el-dialog>

    <!-- 修改授权时间对话框 -->
    <el-dialog 
      v-model="showEditAuthDialog" 
      title="修改授权时间" 
      width="450px"
      :close-on-click-modal="false"
      class="edit-auth-dialog"
    >
      <el-form :model="editAuthForm" label-width="80px">
        <el-form-item label="用户Code">
          <el-input 
            v-model="editAuthForm.userCode" 
            disabled
          />
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker
            v-model="editAuthForm.expireTime"
            type="datetime"
            placeholder="请选择过期时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%;"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showEditAuthDialog = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="saveEditAuthorization"
            :loading="manageLoading"
          >
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { 
  Plus, 
  More, 
  Edit, 
  Delete, 
  Loading, 
  List, 
  Document, 
  Download, 
  FolderOpened, 
  Collection as CollectionIcon, 
  Top, 
  ArrowUp, 
  ArrowDown, 
  Trophy,
  Setting,
  Lock,
  Unlock,
  User,
  Message,
  Clock,
  Check,
  Refresh,
  Sort,
  SortUp,
  SortDown
} from '@element-plus/icons-vue';
import DetailPanel from '../Home/components/DetailPanel.vue';
import CollectionPracticeDrawer from '../../components/CollectionPracticeDrawer.vue';
import WordBook from '../../components/WordBook.vue';
import { useCollectionLogic } from './logic';

// 使用逻辑层
const {
  // 状态
  collections,
  currentCollectionId,
  currentCollectionName,
  loading,
  wordbookVisible,
  isCollectionOwner,
  collectionPoints,
  pointsLoading,
  pointsPage,
  pointsPageSize,
  pointsTotal,
  isEditMode,
  isRandomMode,
  selectedPointDetail,
  selectedPointId,
  currentPointBindings,
  practiceDrawerVisible,
  pointPracticeDrawerVisible,
  showCreateDialog,
  createLoading,
  createForm,
  showEditDialog,
  editLoading,
  editForm,
  showManageDialog,
  manageTabActive,
  manageLoading,
  currentManageCollection,
  permissionForm,
  authForm,
  authorizations,
  authPage,
  authPageSize,
  authTotal,
  authSearchKeyword,
  showEditAuthDialog,
  editAuthForm,
  categorySortOrder,
  pointSortOrder,
  
  // 方法
  fetchCollections,
  selectCollection,
  fetchCollectionPoints,
  selectPoint,
  handleBack,
  goToWordbook,
  handleCreateCollection,
  confirmCreate,
  handleStartPractice,
  getDifficultyText,
  getDifficultyClass,
  handleCollectionMenu,
  handleManageCollection,
  handleManageDialogClosed,
  confirmEdit,
  handleDeleteCollection,
  movePointToTop,
  movePointUp,
  movePointDown,
  removePoint,
  handleAuthSearch,
  loadAuthorizations,
  savePermission,
  addAuthorization,
  editAuthorization,
  saveEditAuthorization,
  deleteAuthorization,
  getPointsByDifficulty,
  toggleMode,
  refreshRandomOrder,
  reloadPoints,
  toggleCategorySort,
  togglePointSort,
  getSubjectColor,
  getCategoryColor,
  navigateToPointFromBinding,
  goBackToPreviousPoint,
  navigationStack
} = useCollectionLogic();

// 页面加载时获取集合列表
onMounted(() => {
  fetchCollections();
});
</script>

<style src="./style.css" scoped></style>
