<template>
  <div class="db-admin-container">
    <!-- Google验证码绑定弹窗 -->
    <el-dialog
      v-model="recaptchaBindVisible"
      title="🔒 Google身份验证器绑定"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      center
    >
      <el-alert
        title="为了数据库安全,管理员必须绑定Google Authenticator"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 20px;"
      />
    
      <!-- 第一步:显示二维码 -->
      <div v-if="!totpVerifying" style="text-align: center; padding: 20px;">
        <div v-if="totpQrcode">
          <h3>步骤1: 扫描二维码</h3>
          <p style="color: #606266; margin: 15px 0;">
            请使用Google Authenticator或类似APP扫描下方二维码
          </p>
              
          <!-- 二维码显示 -->
          <div style="margin: 20px 0;">
            <img :src="qrcodeDataUrl" alt="TOTP QR Code" style="width: 200px; height: 200px;" />
          </div>
    
          <el-alert
            title="密钥(手动输入)"
            :description="totpSecret"
            type="info"
            :closable="false"
            show-icon
            style="margin: 20px 0;"
          />
    
          <el-button type="primary" @click="totpVerifying = true" size="large">
            下一步: 验证
          </el-button>
        </div>
    
        <!-- 加载中 -->
        <div v-else style="padding: 40px;">
          <el-icon :size="40" class="is-loading">
            <Loading />
          </el-icon>
          <p style="margin-top: 15px; color: #909399;">正在生成密钥...</p>
        </div>
      </div>
    
      <!-- 第二步:输入验证码 -->
      <div v-else style="padding: 20px;">
        <h3 style="text-align: center; margin-bottom: 20px;">步骤2: 输入6位验证码</h3>
        <el-form label-width="120px">
          <el-form-item label="验证码">
            <el-input
              v-model="totpCode"
              placeholder="请输入APP中显示的6位数字"
              maxlength="6"
              clearable
              size="large"
              style="width: 300px;"
            />
          </el-form-item>
        </el-form>
    
        <div style="text-align: center; margin-top: 20px;">
          <el-button @click="totpVerifying = false">返回</el-button>
          <el-button type="primary" @click="handleBindTotp" :loading="binding">
            确认绑定
          </el-button>
        </div>
      </div>
    </el-dialog>

    <el-card class="header-card">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <h2>数据库管理系统</h2>
        <el-button type="danger" size="small" @click="handleExit" plain>退出管理</el-button>
      </div>
      <el-alert
        title="⚠️ 警告：所有修改操作需要通过Google验证码验证！请谨慎操作！"
        type="warning"
        :closable="false"
        show-icon
      />
    </el-card>

    <el-row :gutter="20" style="margin-top: 20px; height: calc(100vh - 150px); overflow: hidden;">
      <!-- 左侧：表列表 -->
      <el-col :span="6" style="height: 100%;">
        <el-card class="table-list-card" style="height: 100%;">
          <template #header>
            <div class="card-header">
              <span>表列表</span>
              <el-button type="primary" size="small" @click="loadTables">刷新</el-button>
            </div>
          </template>
          <div style="height: calc(100% - 50px); overflow-y: auto;">
            <el-menu
              :default-active="currentTable"
              @select="handleTableSelect"
            >
              <el-menu-item
                v-for="table in tables"
                :key="table.name"
                :index="table.name"
                style="height: auto; min-height: 40px; line-height: normal; padding: 8px 20px;"
              >
                <div style="width: 100%;">
                  <div style="display: flex; justify-content: space-between; align-items: center;">
                    <span style="font-size: 14px;">{{ table.name }}</span>
                    <el-tag size="small">{{ table.count }}</el-tag>
                  </div>
                  <div 
                    v-if="tableComments[table.name] && tableComments[table.name].trim() !== ''" 
                    style="font-size: 12px; color: #909399; margin-top: 4px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                    :title="tableComments[table.name]"
                  >
                    {{ tableComments[table.name] }}
                  </div>
                </div>
              </el-menu-item>
            </el-menu>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：数据操作区 -->
      <el-col :span="18" style="height: 100%;">
        <div style="height: 100%; overflow: hidden;">
          <el-scrollbar style="height: 100%;">
            <el-card v-if="currentTable" style="height: 100%; display: flex; flex-direction: column;">
              <template #header>
                <div class="card-header">
                  <div style="display: flex; flex-direction: column;">
                    <div style="display: flex; align-items: center;">
                      <span>{{ currentTable }} - 数据管理</span>
                    </div>
                    <div 
                      v-if="tableComment && tableComment.trim() !== ''" 
                      style="font-size: 12px; color: #909399; margin-top: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                      :title="tableComment"
                    >
                      {{ tableComment }}
                    </div>
                  </div>
                  <div>
                    <el-button type="success" size="small" @click="showAddDialog">新增</el-button>
                    <el-button type="danger" size="small" @click="handleBatchDelete" :disabled="!selectedRows.length">批量删除</el-button>
                    <el-button type="warning" size="small" @click="showFieldManageDialog">字段管理</el-button>
                    <el-button type="primary" size="small" @click="showTableCommentDialog">表备注</el-button>
                  </div>
                </div>
              </template>

              <div style="flex: 1; overflow: hidden; display: flex; flex-direction: column;">
                <!-- 字段选择器 -->
                <div class="field-selector">
                  <el-text>选择字段：</el-text>
                  <el-checkbox-group v-model="selectedFields" @change="handleFieldChange">
                    <el-checkbox
                      v-for="col in tableStructure"
                      :key="col.name"
                      :label="col.name"
                    >
                      {{ col.name }}
                    </el-checkbox>
                  </el-checkbox-group>
                  <el-button type="primary" size="small" style="margin-left: 10px;" @click="loadTableData">查询</el-button>
                </div>

                <!-- 条件查询 -->
                <el-collapse style="margin-top: 15px;">
                  <div style="display: flex; align-items: center; justify-content: space-between;">
                    <el-collapse-item title="高级查询" name="1" style="flex: 1;">
                      <div style="max-height: 150px; overflow-y: auto; padding: 8px; background-color: #f0f0f0; border-radius: 4px;">
                        <el-form :inline="true">
                          <el-row :gutter="4">
                            <el-col 
                              v-for="col in tableStructure" 
                              :key="col.name" 
                              :span="24" 
                              style="margin-bottom: 4px;"
                            >
                              <div style="display: flex; align-items: center; background-color: #fff; padding: 2px; border-radius: 2px; border: 1px solid #eee;">
                                <span style="width: 60px; font-weight: bold; font-size: 10px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" :title="col.name">{{ col.name }}:</span>
                                <el-select 
                                  v-model="queryConditions[col.name].operator" 
                                  size="small" 
                                  style="width: 50px; margin-right: 2px;"
                                >
                                  <el-option label="=" value="eq"></el-option>
                                  <el-option label="≠" value="ne"></el-option>
                                  <el-option label="含" value="like"></el-option>
                                  <el-option label="空" value="null"></el-option>
                                </el-select>
                                <el-input
                                  v-model="queryConditions[col.name].value"
                                  placeholder="值"
                                  clearable
                                  size="small"
                                  style="flex: 1; font-size: 12px;"
                                  :disabled="queryConditions[col.name].operator === 'null'"
                                />
                              </div>
                            </el-col>
                          </el-row>
                        </el-form>
                      </div>
                    </el-collapse-item>
                    <div style="margin-left: 10px;">
                      <el-button type="primary" size="small" @click="loadTableData" style="padding: 6px 12px;">
                        查询
                      </el-button>
                      <el-button size="small" @click="resetQueryConditions" style="padding: 6px 12px; margin-left: 5px;">
                        重置
                      </el-button>
                    </div>
                  </div>
                </el-collapse>

                <!-- 数据表格 -->
                <div class="table-wrapper" style="flex: 1; overflow: hidden; margin-top: 15px; display: flex; flex-direction: column;">
                  <div style="flex: 1; overflow: hidden;">
                    <el-table
                      :data="tableData"
                      style="width: 100%; height: 100%;"
                      border
                      @selection-change="handleSelectionChange"
                      v-loading="loading"
                      :row-style="{ height: '40px' }"
                      :cell-style="{ padding: '4px', maxHeight: '40px', overflow: 'hidden' }"
                      max-height="100%"
                    >
                      <el-table-column type="selection" width="55" />
                      <el-table-column
                        v-for="field in sortedSelectedFields"
                        :key="field"
                        :prop="field"
                        :label="field"
                        :min-width="120"
                      >
                        <template #header>
                          <div style="display: flex; flex-direction: column; line-height: 1.2;">
                            <div style="display: flex; align-items: center; justify-content: space-between;">
                              <span style="font-weight: bold;">{{ field }}</span>
                              <el-button 
                                type="primary" 
                                size="small" 
                                link 
                                @click.stop="showColumnCommentDialog(field)"
                                title="编辑字段备注"
                              >
                                <el-icon><Edit /></el-icon>
                              </el-button>
                            </div>
                            <div 
                              v-if="columnComments[field] && columnComments[field].trim() !== ''" 
                              style="font-size: 11px; color: #909399; font-weight: normal; margin-top: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                              :title="columnComments[field]"
                            >
                              {{ columnComments[field] }}
                            </div>
                            <div 
                              v-else 
                              style="font-size: 11px; color: #c0c4cc; font-weight: normal; margin-top: 2px;"
                            >
                              暂无备注
                            </div>
                          </div>
                        </template>
                        <template #default="{ row }">
                          <div 
                            :class="['cell-wrapper', getCellClass(row[field])]"
                            @dblclick="handleCellDblClick(row, field)"
                            style="position: relative; padding-left: 20px;"
                          >
                            <el-tooltip content="复制" placement="top">
                              <el-button 
                                type="primary" 
                                size="small" 
                                link
                                :icon="CopyDocument"
                                @click.stop="copyCellValue(row[field])"
                                class="copy-btn"
                              />
                            </el-tooltip>
                            <span class="cell-text">
                              {{ formatCellValue(row[field]) || '\u00A0' }}
                            </span>
                            <span 
                              v-if="isTextOverflow(row[field])" 
                              class="ellipsis-btn"
                              @click.stop="handleEllipsisClick(row, field)"
                              title="点击查看完整内容"
                            >
                              ...
                            </span>
                          </div>
                        </template>
                      </el-table-column>
                      <el-table-column label="操作" width="150" fixed="right">
                        <template #default="{ row }">
                          <el-button type="primary" size="small" @click="showEditDialog(row)" style="padding: 4px 8px; font-size: 12px;">编辑</el-button>
                          <el-button type="danger" size="small" @click="handleDelete(row)" style="padding: 4px 8px; font-size: 12px;">删除</el-button>
                        </template>
                      </el-table-column>
                    </el-table>
                  </div>
                </div>

                <!-- 分页 -->
                <el-pagination
                  v-model:current-page="pagination.page"
                  v-model:page-size="pagination.pageSize"
                  :page-sizes="[10, 20, 50, 100]"
                  :total="pagination.total"
                  layout="total, sizes, prev, pager, next, jumper"
                  @size-change="loadTableData"
                  @current-change="loadTableData"
                  style="margin-top: 20px; justify-content: center;"
                />
              </div>
            </el-card>

            <el-empty v-else description="请选择一个表" />
          </el-scrollbar>
        </div>
      </el-col>
    </el-row>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? '新增数据' : '编辑数据'"
      width="600px"
    >
      <el-form :model="formData" label-width="120px">
        <el-form-item
          v-for="col in editableColumns"
          :key="col.name"
          :label="col.name"
          :required="col.not_null && !col.default"
        >
          <el-input
            v-model="formData[col.name]"
            :placeholder="getCurrentValueHint(col.name)"
            :disabled="col.pk && dialogMode === 'edit'"
          />
          <div style="margin-top: 5px; display: flex; gap: 5px;">
            <el-button 
              v-if="!col.not_null"
              size="small" 
              @click="formData[col.name] = '__NULL__'"
            >
              设置为 NULL
            </el-button>
            <el-button 
              size="small" 
              @click="formData[col.name] = '__EMPTY_STRING__'"
            >
              设置为空字符串
            </el-button>
          </div>
          <el-text size="small" type="info">{{ col.type }}</el-text>
        </el-form-item>
        <el-form-item label="Google验证码" required>
          <el-input
            v-model="totpCodeForEdit"
            placeholder="请输入6位验证码"
            maxlength="6"
            clearable
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ dialogMode === 'add' ? '新增' : '保存' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 查看完整内容对话框 -->
    <el-dialog
      v-model="viewContentVisible"
      :title="viewContentTitle"
      width="600px"
    >
      <el-input
        v-model="viewContentData"
        type="textarea"
        :rows="10"
        readonly
        style="font-family: monospace;"
      />
      <template #footer>
        <el-button @click="viewContentVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 双击快速编辑对话框 -->
    <el-dialog
      v-model="quickEditVisible"
      title="快速编辑"
      width="450px"
    >
      <el-form label-width="120px">
        <el-form-item :label="quickEditField">
          <el-input
            v-model="quickEditValue"
            :placeholder="getQuickEditPlaceholder()"
            clearable
          />
          <div style="margin-top: 5px; display: flex; gap: 5px;">
            <el-button 
              size="small" 
              @click="quickEditValue = ''"
            >
              设置为 NULL
            </el-button>
            <el-button 
              size="small" 
              @click="quickEditValue = '__EMPTY_STRING__'"
            >
              设置为空字符串
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="Google验证码">
          <el-input
            v-model="totpCodeForEdit"
            placeholder="请输入6位验证码"
            maxlength="6"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="quickEditVisible = false">取消</el-button>
        <el-button type="primary" @click="handleQuickEdit" :loading="submitting">保存</el-button>
      </template>
    </el-dialog>

    <!-- 字段管理对话框 -->
    <el-dialog
      v-model="fieldManageVisible"
      :title="`${currentTable} - 字段管理`"
      width="800px"
      top="5vh"
    >
      <el-alert
        title="注意：删除字段将导致数据丢失，请谨慎操作！可通过上移/下移按钮调整字段顺序。"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 15px;"
      />
      
      <!-- 字段卡片列表 - 默认4行，超出滚动 -->
      <div style="max-height: 520px; overflow-y: auto; overflow-x: hidden;">
        <el-row :gutter="10">
          <el-col :span="8" v-for="(col, index) in sortedTableStructure" :key="col.name" style="margin-bottom: 10px;">
            <el-card shadow="hover" :body-style="{ padding: '10px' }">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px;">
                <span style="font-weight: bold; font-size: 13px;">{{ index + 1 }}. {{ col.name }}</span>
                <div>
                  <el-tag v-if="col.pk" type="danger" size="small">PK</el-tag>
                  <el-tag v-if="col.not_null" type="warning" size="small">NN</el-tag>
                </div>
              </div>
              <div style="font-size: 12px; color: #606266; margin-bottom: 5px;">
                <span>类型: {{ col.type }}</span>
                <span v-if="col.default" style="margin-left: 10px;">默认: {{ col.default }}</span>
              </div>
              <div 
                v-if="columnComments[col.name] && columnComments[col.name].trim() !== ''" 
                style="font-size: 11px; color: #909399; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 8px;"
                :title="columnComments[col.name]"
              >
                备注: {{ columnComments[col.name] }}
              </div>
              <div v-else style="font-size: 11px; color: #c0c4cc; margin-bottom: 8px;">
                暂无备注
              </div>
              <div style="display: flex; gap: 5px; flex-wrap: wrap;">
                <el-button size="small" :disabled="index === 0" @click="moveFieldUp(index)" style="padding: 3px 6px; font-size: 12px;">↑</el-button>
                <el-button size="small" :disabled="index === sortedTableStructure.length - 1" @click="moveFieldDown(index)" style="padding: 3px 6px; font-size: 12px;">↓</el-button>
                <el-button type="primary" size="small" @click="showColumnCommentDialog(col.name)" style="padding: 3px 8px; font-size: 12px;">备注</el-button>
                <el-button type="danger" size="small" @click="handleDeleteField(col)" :disabled="col.pk" style="padding: 3px 8px; font-size: 12px;">删除</el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <el-divider />

      <h4 style="margin-bottom: 10px;">添加新字段</h4>
      <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
        <div style="display: flex; align-items: center;">
          <span style="width: 50px; font-size: 13px;">字段名:</span>
          <el-input v-model="newFieldForm.name" placeholder="字段名" size="small" style="width: 100px;" />
        </div>
        <div style="display: flex; align-items: center;">
          <span style="width: 40px; font-size: 13px;">类型:</span>
          <el-select v-model="newFieldForm.type" placeholder="类型" size="small" style="width: 100px;">
            <el-option label="TEXT" value="TEXT" />
            <el-option label="INTEGER" value="INTEGER" />
            <el-option label="REAL" value="REAL" />
            <el-option label="DATETIME" value="DATETIME" />
            <el-option label="BLOB" value="BLOB" />
          </el-select>
        </div>
        <div style="display: flex; align-items: center;">
          <span style="width: 50px; font-size: 13px;">默认值:</span>
          <el-input v-model="newFieldForm.default" placeholder="选填" size="small" style="width: 80px;" />
        </div>
        <div style="display: flex; align-items: center;">
          <span style="width: 50px; font-size: 13px; color: #f56c6c;">*验证码:</span>
          <el-input v-model="totpCodeForEdit" placeholder="6位" maxlength="6" size="small" style="width: 70px;" clearable />
        </div>
        <el-button type="primary" size="small" @click="handleAddField">添加字段</el-button>
      </div>

      <template #footer>
        <el-button type="success" @click="handleSaveColumnOrders">保存排序</el-button>
        <el-button @click="fieldManageVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 表备注对话框 -->
    <el-dialog
      v-model="tableCommentVisible"
      :title="`${currentTable} - 表备注`"
      width="600px"
    >
      <el-form label-width="80px">
        <el-form-item label="备注内容">
          <el-input
            v-model="tableComment"
            type="textarea"
            :rows="4"
            placeholder="请输入表备注信息"
          />
        </el-form-item>
        <el-form-item label="验证码" required>
          <el-input
            v-model="totpCodeForEdit"
            placeholder="请输入6位Google验证码"
            maxlength="6"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tableCommentVisible = false">取消</el-button>
        <el-button type="primary" @click="saveTableComment" :loading="tableCommentLoading">保存</el-button>
      </template>
    </el-dialog>

    <!-- 字段备注对话框 -->
    <el-dialog
      v-model="columnCommentVisible"
      :title="`${currentTable}.${columnCommentForm.columnName} - 字段备注`"
      width="600px"
    >
      <el-form label-width="80px">
        <el-form-item label="字段名">
          <el-input
            v-model="columnCommentForm.columnName"
            disabled
          />
        </el-form-item>
        <el-form-item label="备注内容">
          <el-input
            v-model="columnCommentForm.comment"
            type="textarea"
            :rows="4"
            placeholder="请输入字段备注信息"
          />
        </el-form-item>
        <el-form-item label="验证码" required>
          <el-input
            v-model="totpCodeForEdit"
            placeholder="请输入6位Google验证码"
            maxlength="6"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="columnCommentVisible = false">取消</el-button>
        <el-button type="primary" @click="saveColumnComment" :loading="columnCommentLoading">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, CopyDocument, InfoFilled, Edit } from '@element-plus/icons-vue'
import {
  getAllTables,
  getTableStructure,
  getTableData,
  insertTableRow,
  updateTableRow,
  deleteTableRows,
  batchDeleteTableRows,
  checkTotpBound,
  generateTotpSecret,
  bindTotp,
  getTableComment,
  setTableComment,
  getColumnComment,
  setColumnComment,
  getAllTableComments,
  getAllColumnComments,
  addColumn,
  dropColumn,
  getColumnOrders,
  saveColumnOrders
} from '../../api/dbAdmin'

// 状态
const tables = ref<any[]>([])
const currentTable = ref('')
const tableStructure = ref<any[]>([])
const tableData = ref<any[]>([])
const selectedFields = ref<string[]>([])
const queryConditions = ref<any>({})

// 初始化查询条件
const initQueryConditions = () => {
  const newConditions: any = {}
  tableStructure.value.forEach(col => {
    newConditions[col.name] = {
      operator: 'eq',
      value: ''
    }
  })
  queryConditions.value = newConditions
}
const selectedRows = ref<any[]>([])
const loading = ref(false)

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

// 对话框
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const formData = ref<any>({})
const currentRow = ref<any>(null)
const submitting = ref(false)

// Google验证码绑定状态
const recaptchaBindVisible = ref(false)
const recaptchaBound = ref(false)
const totpSecret = ref('') // TOTP密钥
const totpQrcode = ref('') // TOTP二维码URL
const qrcodeDataUrl = ref('') // 二维码图片base64
const totpVerifying = ref(false) // 是否进入验证步骤
const totpCode = ref('') // 用户输入的验证码
const binding = ref(false) // 绑定加载状态

// 双击编辑
const quickEditVisible = ref(false)
const quickEditField = ref('')
const quickEditValue = ref('')
const quickEditRow = ref<any>(null)

// 查看完整内容
const viewContentVisible = ref(false)
const viewContentTitle = ref('')
const viewContentData = ref('')

// Google验证码（用于增删改操作）
const totpCodeForEdit = ref('')

// 字段管理
const fieldManageVisible = ref(false)
const newFieldForm = ref({
  name: '',
  type: 'TEXT',
  default: ''
})

// 表备注管理
const tableComment = ref('')
const tableComments = ref<Record<string, string>>({})
const tableCommentVisible = ref(false)
const tableCommentLoading = ref(false)

// 字段备注管理
const columnComments = ref<Record<string, string>>({})
const columnCommentForm = ref({
  columnName: '',
  comment: ''
})
const columnCommentVisible = ref(false)
const columnCommentLoading = ref(false)

// 字段排序管理
const columnOrders = ref<Record<string, number>>({})
const sortedTableStructure = computed(() => {
  // 如果有排序配置，按排序配置来；否则按原始顺序
  return [...tableStructure.value].sort((a, b) => {
    const orderA = columnOrders.value[a.name] ?? 999
    const orderB = columnOrders.value[b.name] ?? 999
    return orderA - orderB
  })
})

// 排序后的已选字段（用于表格展示）
const sortedSelectedFields = computed(() => {
  return selectedFields.value.slice().sort((a, b) => {
    const orderA = columnOrders.value[a] ?? 999
    const orderB = columnOrders.value[b] ?? 999
    return orderA - orderB
  })
})

// 计算属性
const editableColumns = computed(() => {
  return tableStructure.value.filter(col => {
    // 编辑模式下可以编辑所有字段，新增模式下跳过自增主键
    if (dialogMode.value === 'edit') return true
    return !(col.pk && col.type.includes('INTEGER'))
  })
})

const primaryKey = computed(() => {
  const pk = tableStructure.value.find(col => col.pk)
  return pk ? pk.name : 'id'
})

// 重置查询条件
const resetQueryConditions = () => {
  const newConditions: any = {}
  tableStructure.value.forEach(col => {
    newConditions[col.name] = {
      operator: 'eq',
      value: ''
    }
  })
  queryConditions.value = newConditions
}

// 方法
const loadTables = async () => {
  try {
    const res = await getAllTables()
    if (res.data.code === 200) {
      tables.value = res.data.data
      
      // 加载所有表备注
      await loadAllTableComments()
    }
  } catch (error) {
    ElMessage.error('加载表列表失败')
  }
}

const handleTableSelect = async (tableName: string) => {
  currentTable.value = tableName
  selectedFields.value = []
  selectedRows.value = []
  pagination.value.page = 1

  // 加载表结构
  try {
    const res = await getTableStructure(tableName)
    if (res.data.code === 200) {
      tableStructure.value = res.data.data
      // 默认全选
      selectedFields.value = tableStructure.value.map(col => col.name)
      
      // 初始化查询条件
      initQueryConditions()
      
      // 自动加载表数据
      await loadTableData()
      
      // 加载字段备注
      await loadAllColumnComments()
      
      // 加载表备注
      await loadTableComment()
      
      // 加载字段排序
      await loadColumnOrders()
    }
  } catch (error) {
    ElMessage.error('加载表结构失败')
  }
}

const handleFieldChange = () => {
  // 字段变化后不自动查询，需要用户手动点击查询按钮
}

const loadTableData = async () => {
  if (!currentTable.value) {
    ElMessage.warning('请先选择一个表')
    return
  }

  // 如果没有选择字段，则默认选择所有字段
  let fieldsToQuery = selectedFields.value
  if (!fieldsToQuery.length) {
    fieldsToQuery = tableStructure.value.map(col => col.name)
  }

  loading.value = true
  try {
    // 构建查询参数
    const where: any = {}
    Object.keys(queryConditions.value).forEach(key => {
      const condition = queryConditions.value[key]
      // 只添加有效的条件
      if (condition && condition.operator) {
        // 对于需要值的操作符，检查值是否非空
        if ((condition.operator === 'eq' || condition.operator === 'ne' || 
             condition.operator === 'like' || condition.operator === 'starts' || 
             condition.operator === 'ends') && 
            condition.value !== '') {
          where[key] = condition
        } 
        // 对于不需要值的操作符，直接添加
        else if (condition.operator === 'null' || condition.operator === 'notnull') {
          where[key] = condition
        }
      }
    })

    const params = {
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      fields: fieldsToQuery.join(','),
      where: Object.keys(where).length > 0 ? JSON.stringify(where) : ''
    }

    const res = await getTableData(currentTable.value, params)
    if (res.data.code === 200) {
      tableData.value = res.data.data.list || []
      pagination.value.total = res.data.data.total || 0
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleSelectionChange = (rows: any[]) => {
  selectedRows.value = rows
}

const formatCellValue = (value: any) => {
  if (value === null || value === undefined) return 'NULL'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

// 判断文本是否超出（简单判断：超过30个字符）
const isTextOverflow = (value: any) => {
  const formatted = formatCellValue(value)
  return formatted && formatted !== 'NULL' && formatted.length > 30
}

// 获取当前值提示
const getCurrentValueHint = (fieldName: string) => {
  if (dialogMode.value === 'edit' && currentRow.value) {
    const currentValue = currentRow.value[fieldName]
    if (currentValue === null || currentValue === undefined) {
      return `当前为NULL，输入新值、留空设为NULL、或设置为空字符串""`
    } else if (currentValue === '') {
      return `当前为空字符串""，可修改值、设为NULL或其他值`
    } else {
      return `当前值: ${String(currentValue).substring(0, 20)}${String(currentValue).length > 20 ? '...' : ''}`
    }
  }
  return `请输入${fieldName}`
}

// 获取快速编辑占位符
const getQuickEditPlaceholder = () => {
  if (quickEditRow.value && quickEditField.value) {
    const currentValue = quickEditRow.value[quickEditField.value]
    if (currentValue === null || currentValue === undefined) {
      return `当前为NULL，输入新值、留空设为NULL、或设置为空字符串""`
    } else if (currentValue === '') {
      return `当前为空字符串""，可修改值、设为NULL或其他值`
    } else {
      return `当前值: ${String(currentValue).substring(0, 20)}${String(currentValue).length > 20 ? '...' : ''}，输入新值或留空设为NULL`
    }
  }
  return `输入新值，留空设置NULL`
}

const showAddDialog = () => {
  dialogMode.value = 'add'
  formData.value = {}
  currentRow.value = null
  totpCodeForEdit.value = ''
  dialogVisible.value = true
}

const showEditDialog = (row: any) => {
  dialogMode.value = 'edit'
  formData.value = { ...row }
  currentRow.value = row
  totpCodeForEdit.value = ''
  dialogVisible.value = true
}

// Google reCAPTCHA Token 获取（使用用户输入的TOTP验证码）
const getRecaptchaToken = async (): Promise<string> => {
  if (!totpCodeForEdit.value || totpCodeForEdit.value.length !== 6) {
    throw new Error('请输入6位Google验证码')
  }
  // 返回用户输入的TOTP验证码
  return totpCodeForEdit.value
}

const handleSubmit = async () => {
  try {
    // 验证Google验证码
    if (!totpCodeForEdit.value || totpCodeForEdit.value.length !== 6) {
      ElMessage.warning('请输入6位Google验证码')
      return
    }

    await ElMessageBox.confirm(
      '此操作将修改数据库，是否继续？',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    submitting.value = true
    const recaptchaToken = await getRecaptchaToken()

    if (dialogMode.value === 'add') {
      // 新增
      // 处理特殊标记
      const processedData: any = {}
      Object.keys(formData.value).forEach(key => {
        let value = formData.value[key]
        // 处理特殊标记
        if (value === '__NULL__') {
          // 特殊标记表示NULL
          value = null
        } else if (value === '__EMPTY_STRING__') {
          // 特殊标记表示空字符串
          value = ''
        }
        processedData[key] = value
      })
      
      const res = await insertTableRow(currentTable.value, processedData, recaptchaToken)
      if (res.data.code === 200) {
        ElMessage.success('新增成功')
        dialogVisible.value = false
        totpCodeForEdit.value = ''
        loadTableData()
      }
    } else {
      // 编辑
      const where: any = {}
      where[primaryKey.value] = currentRow.value[primaryKey.value]
      
      // 处理特殊标记
      const processedData: any = {}
      Object.keys(formData.value).forEach(key => {
        let value = formData.value[key]
        // 处理特殊标记
        if (value === '__NULL__') {
          // 特殊标记表示NULL
          value = null
        } else if (value === '__EMPTY_STRING__') {
          // 特殊标记表示空字符串
          value = ''
        }
        processedData[key] = value
      })
      
      const res = await updateTableRow(currentTable.value, where, processedData, recaptchaToken)
      if (res.data.code === 200) {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        totpCodeForEdit.value = ''
        loadTableData()
      }
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || error.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row: any) => {
  try {
    // 先让用户输入Google验证码
    const { value: totpCode } = await ElMessageBox.prompt(
      '请输入Google验证码以确认删除操作',
      '警告',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        inputPattern: /^\d{6}$/,
        inputErrorMessage: '请输入6位数字验证码',
        inputPlaceholder: '请输入6位验证码'
      }
    )

    await ElMessageBox.confirm(
      '确定要删除这条数据吗？此操作不可恢复！',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const where: any = {}
    where[primaryKey.value] = row[primaryKey.value]

    const res = await deleteTableRows(currentTable.value, where, totpCode)
    if (res.data.code === 200) {
      ElMessage.success('删除成功')
      loadTableData()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '删除失败')
    }
  }
}

const handleBatchDelete = async () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请至少选择一条数据')
    return
  }

  try {
    // 先让用户输入Google验证码
    const { value: totpCode } = await ElMessageBox.prompt(
      '请输入Google验证码以确认批量删除操作',
      '警告',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        inputPattern: /^\d{6}$/,
        inputErrorMessage: '请输入6位数字验证码',
        inputPlaceholder: '请输入6位验证码'
      }
    )

    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedRows.value.length} 条数据吗？此操作不可恢复！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const ids = selectedRows.value.map(row => row[primaryKey.value])

    const res = await batchDeleteTableRows(currentTable.value, ids, primaryKey.value, totpCode)
    if (res.data.code === 200) {
      ElMessage.success(`成功删除 ${res.data.data.affected} 条数据`)
      selectedRows.value = []
      loadTableData()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '批量删除失败')
    }
  }
}

// ========== 新增功能：Google验证码绑定 =========
// 生成TOTP二维码
const generateTotpQrcode = async () => {
  try {
    const res = await generateTotpSecret()
    if (res.data.code === 200) {
      const data = res.data.data
      
      // 如果已绑定，关闭弹窗
      if (data.bound) {
        recaptchaBound.value = true
        recaptchaBindVisible.value = false
        ElMessage.success('已绑定Google身份验证器')
        return
      }
      
      // 保存密钥和二维码URL
      totpSecret.value = data.secret
      totpQrcode.value = data.qrcode
      
      // 生成二维码图片
      const QRCode = (await import('qrcode')).default
      qrcodeDataUrl.value = await QRCode.toDataURL(data.qrcode)
    }
  } catch (error: any) {
    ElMessage.error('生成密钥失败')
  }
}

// 绑定TOTP
const handleBindTotp = async () => {
  if (!totpCode.value || totpCode.value.length !== 6) {
    ElMessage.warning('请输入6位验证码')
    return
  }
  
  try {
    binding.value = true
    const res = await bindTotp(totpSecret.value, totpCode.value)
    if (res.data.code === 200) {
      ElMessage.success('绑定成功！')
      recaptchaBound.value = true
      recaptchaBindVisible.value = false
      totpCode.value = ''
      totpVerifying.value = false
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '绑定失败,请检查验证码是否正确')
  } finally {
    binding.value = false
  }
}

// ========== 新增功能：NULL值灰色显示 =========
const getCellClass = (value: any) => {
  if (value === null || value === undefined || value === 'NULL') {
    return 'cell-null'
  }
  return ''
}

// ========== 新增功能：单击省略号查看完整内容 =========
const handleEllipsisClick = (row: any, field: string) => {
  const value = row[field]
  const formattedValue = formatCellValue(value)
  
  viewContentTitle.value = `${currentTable.value}.${field}`
  viewContentData.value = formattedValue
  viewContentVisible.value = true
}

// ========== 新增功能：复制字段值 =========
const copyCellValue = (value: any) => {
  const formattedValue = formatCellValue(value)
  navigator.clipboard.writeText(formattedValue).then(() => {
    ElMessage.success('复制成功')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

// ========== 新增功能：双击编辑 =========
const handleCellDblClick = (row: any, field: string) => {
  // 不能编辑主键
  if (field === primaryKey.value) {
    ElMessage.warning('不能编辑主键字段')
    return
  }
  
  quickEditRow.value = row
  quickEditField.value = field
  // 如果原值是NULL，显示特殊标识
  quickEditValue.value = row[field] === null || row[field] === undefined ? '' : String(row[field])
  totpCodeForEdit.value = ''
  quickEditVisible.value = true
}

const handleQuickEdit = async () => {
  try {
    // 验证Google验证码
    if (!totpCodeForEdit.value || totpCodeForEdit.value.length !== 6) {
      ElMessage.warning('请输入6位Google验证码')
      return
    }

    await ElMessageBox.confirm(
      '此操作将修改数据库，是否继续？',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    submitting.value = true
    const recaptchaToken = totpCodeForEdit.value

    const where: any = {}
    where[primaryKey.value] = quickEditRow.value[primaryKey.value]

    const data: any = {}
    // 处理特殊标记
    let finalValue: string | null = quickEditValue.value
    if (finalValue === '__EMPTY_STRING__') {
      // 特殊标记表示空字符串
      finalValue = ''
    } else if (finalValue === '') {
      // 真正的空值设为NULL
      finalValue = null
    }
    data[quickEditField.value] = finalValue

    const res = await updateTableRow(currentTable.value, where, data, recaptchaToken)
    if (res.data.code === 200) {
      ElMessage.success('修改成功')
      quickEditVisible.value = false
      totpCodeForEdit.value = ''
      loadTableData()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '修改失败')
    }
  } finally {
    submitting.value = false
  }
}

// ========== 新增功能：字段管理 =========
const showFieldManageDialog = () => {
  fieldManageVisible.value = true
  newFieldForm.value = { name: '', type: 'TEXT', default: '' }
}

const handleAddField = async () => {
  if (!newFieldForm.value.name) {
    ElMessage.warning('请输入字段名')
    return
  }

  // 验证Google验证码
  if (!totpCodeForEdit.value || totpCodeForEdit.value.length !== 6) {
    ElMessage.warning('请输入6位Google验证码')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要向表 ${currentTable.value} 添加字段 ${newFieldForm.value.name} (类型: ${newFieldForm.value.type})？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const recaptchaToken = totpCodeForEdit.value
    
    const res = await addColumn(
      currentTable.value,
      newFieldForm.value.name,
      newFieldForm.value.type,
      newFieldForm.value.default,
      recaptchaToken
    )
    
    if (res.data.code === 200) {
      ElMessage.success('添加字段成功')
      // 重新加载表结构
      await handleTableSelect(currentTable.value)
      // 清空表单
      newFieldForm.value.name = ''
      newFieldForm.value.type = 'TEXT'
      newFieldForm.value.default = ''
      totpCodeForEdit.value = ''
    } else {
      ElMessage.error(res.data.msg || '添加字段失败')
    }
    
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '添加字段失败')
    }
  }
}

const handleDeleteField = async (row: any) => {
  if (row.pk) {
    ElMessage.warning('不能删除主键字段')
    return
  }

  try {
    // 弹出确认对话框，并输入验证码
    const { value: recaptchaToken } = await ElMessageBox.prompt(
      `确定要删除字段 ${row.name}？此操作将导致数据丢失！\n\n请输入6位Google验证码：`,
      '警告',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'error',
        inputPattern: /^\d{6}$/,
        inputErrorMessage: '请输入正确的6位验证码'
      }
    )
    
    const res = await dropColumn(
      currentTable.value,
      row.name,
      recaptchaToken
    )
    
    if (res.data.code === 200) {
      ElMessage.success('删除字段成功')
      // 重新加载表结构
      await handleTableSelect(currentTable.value)
    } else {
      ElMessage.error(res.data.msg || '删除字段失败')
    }
    
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '删除字段失败')
    }
  }
}

onMounted(async () => {
  // 检查是否已绑定Google验证码
  try {
    const res = await checkTotpBound()
    if (res.data.code === 200) {
      const { bound, is_admin } = res.data.data
      
      // 如果是管理员且未绑定,显示绑定弹窗
      if (is_admin === 1 && !bound) {
        recaptchaBindVisible.value = true
        // 生成TOTP二维码
        await generateTotpQrcode()
      } else if (bound) {
        recaptchaBound.value = true
      }
    }
  } catch (error) {
    console.error('检查TOTP绑定状态失败', error)
  }
  
  loadTables()
})

// ========== 新增功能：退出管理界面 =========
const handleExit = () => {
  // 清除可能存在的会话数据
  localStorage.removeItem('token')
  
  // 跳转到首页或其他指定页面
  window.location.href = '/'
}

// ========== 新增功能：表备注管理 =========
// 显示表备注对话框
const showTableCommentDialog = async () => {
  if (!currentTable.value) {
    ElMessage.warning('请先选择一个表')
    return
  }
  
  tableCommentLoading.value = true
  try {
    const res = await getTableComment(currentTable.value)
    if (res.data.code === 200) {
      tableComment.value = res.data.data || ''
      tableCommentVisible.value = true
    }
  } catch (error) {
    ElMessage.error('获取表备注失败')
  } finally {
    tableCommentLoading.value = false
  }
}

// 保存表备注
const saveTableComment = async () => {
  if (!currentTable.value) {
    ElMessage.warning('请先选择一个表')
    return
  }
    
  try {
    // 验证Google验证码
    if (!totpCodeForEdit.value || totpCodeForEdit.value.length !== 6) {
      ElMessage.warning('请输入6位Google验证码')
      return
    }
      
    tableCommentLoading.value = true
    const recaptchaToken = totpCodeForEdit.value
      
    const res = await setTableComment(currentTable.value, tableComment.value, recaptchaToken)
    if (res.data.code === 200) {
      ElMessage.success('表备注保存成功')
      tableCommentVisible.value = false
      totpCodeForEdit.value = ''
        
      // 更新本地缓存
      tableComment.value = tableComment.value
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '保存表备注失败')
  } finally {
    tableCommentLoading.value = false
  }
}

// ========== 新增功能：字段备注管理 =========
// 显示字段备注对话框
const showColumnCommentDialog = (columnName: string) => {
  columnCommentForm.value.columnName = columnName
  columnCommentForm.value.comment = columnComments.value[columnName] || ''
  columnCommentVisible.value = true
}

// 保存字段备注
const saveColumnComment = async () => {
  if (!currentTable.value || !columnCommentForm.value.columnName) {
    ElMessage.warning('缺少必要的参数')
    return
  }
  
  try {
    // 验证Google验证码
    if (!totpCodeForEdit.value || totpCodeForEdit.value.length !== 6) {
      ElMessage.warning('请输入6位Google验证码')
      return
    }
    
    columnCommentLoading.value = true
    const recaptchaToken = totpCodeForEdit.value
    
    const res = await setColumnComment(
      currentTable.value, 
      columnCommentForm.value.columnName, 
      columnCommentForm.value.comment, 
      recaptchaToken
    )
    
    if (res.data.code === 200) {
      ElMessage.success('字段备注保存成功')
      columnCommentVisible.value = false
      totpCodeForEdit.value = ''
      
      // 更新本地缓存
      columnComments.value[columnCommentForm.value.columnName] = columnCommentForm.value.comment
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.msg || '保存字段备注失败')
  } finally {
    columnCommentLoading.value = false
  }
}

// 加载所有字段备注
const loadAllColumnComments = async () => {
  if (!currentTable.value) return
  
  try {
    const res = await getAllColumnComments()
    if (res.data.code === 200) {
      // 只获取当前表的字段备注
      const tableComments = res.data.data[currentTable.value] || {}
      columnComments.value = tableComments
    }
  } catch (error) {
    console.error('加载字段备注失败:', error)
  }
}

// 加载所有表备注
const loadAllTableComments = async () => {
  try {
    const res = await getAllTableComments()
    if (res.data.code === 200) {
      tableComments.value = res.data.data || {}
    }
  } catch (error) {
    console.error('加载表备注失败:', error)
  }
}

// 加载表备注
const loadTableComment = async () => {
  if (!currentTable.value) return
  
  try {
    const res = await getTableComment(currentTable.value)
    if (res.data.code === 200) {
      tableComment.value = res.data.data || ''
    }
  } catch (error) {
    console.error('加载表备注失败:', error)
  }
}

// ========== 新增功能：字段排序管理 =========
// 加载字段排序
const loadColumnOrders = async () => {
  if (!currentTable.value) return
  
  try {
    const res = await getColumnOrders(currentTable.value)
    if (res.data.code === 200) {
      columnOrders.value = res.data.data || {}
    }
  } catch (error) {
    console.error('加载字段排序失败:', error)
  }
}

// 上移字段
const moveFieldUp = (index: number) => {
  if (index <= 0) return
  const list = sortedTableStructure.value
  const temp = list[index]
  list[index] = list[index - 1]
  list[index - 1] = temp
  // 更新排序到本地状态
  updateLocalOrders(list)
}

// 下移字段
const moveFieldDown = (index: number) => {
  const list = sortedTableStructure.value
  if (index >= list.length - 1) return
  const temp = list[index]
  list[index] = list[index + 1]
  list[index + 1] = temp
  // 更新排序到本地状态
  updateLocalOrders(list)
}

// 更新本地排序状态
const updateLocalOrders = (list: any[]) => {
  const newOrders: Record<string, number> = {}
  list.forEach((col, idx) => {
    newOrders[col.name] = idx
  })
  columnOrders.value = newOrders
}

// 保存字段排序
const handleSaveColumnOrders = async () => {
  if (!currentTable.value) {
    ElMessage.warning('请先选择一个表')
    return
  }

  try {
    const { value: recaptchaToken } = await ElMessageBox.prompt(
      '确定要保存当前字段排序吗？\n\n请输入6位Google验证码：',
      '保存排序',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info',
        inputPattern: /^\d{6}$/,
        inputErrorMessage: '请输入正确的6位验证码'
      }
    )

    const orders = sortedTableStructure.value.map(col => col.name)
    const res = await saveColumnOrders(currentTable.value, orders, recaptchaToken)
    
    if (res.data.code === 200) {
      ElMessage.success('字段排序保存成功')
    } else {
      ElMessage.error(res.data.msg || '保存失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.msg || '保存字段排序失败')
    }
  }
}
</script>

<style scoped>
.db-admin-container {
  padding: 15px;
  background-color: #f5f5f5;
  min-height: 100vh;
  max-height: 100vh;
  overflow: hidden;
  box-sizing: border-box;
}

.header-card {
  margin-bottom: 20px;
}

.header-card h2 {
  margin: 0 0 15px 0;
  color: #333;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-list-card {
  height: calc(100vh - 180px);
  overflow-y: auto;
}

.field-selector {
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.field-selector .el-checkbox {
  margin-right: 15px;
}

/* 表格容器，添加滚动 */
.table-wrapper {
  margin-top: 20px;
  overflow-x: auto;
  overflow-y: auto;
}

/* 单元格包裹器 */
.cell-wrapper {
  display: flex;
  align-items: center;
  min-height: 18px;
  height: 100%;
  width: 100%;
  line-height: 18px;
  position: relative;
  cursor: pointer;
  padding: 0 2px;
}

.cell-wrapper:hover {
  background-color: #f5f7fa;
}

/* 单元格文本 */
.cell-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: break-all;
  font-size: 12px;
}

/* 省略号按钮 */
.ellipsis-btn {
  flex-shrink: 0;
  margin-left: 4px;
  color: #409eff;
  cursor: pointer;
  font-weight: bold;
  padding: 0 4px;
  user-select: none;
}

.ellipsis-btn:hover {
  color: #66b1ff;
  text-decoration: underline;
}

/* NULL值灰色显示 */
.cell-null {
  color: #999;
  font-style: italic;
  opacity: 0.6;
}

/* 表格双击提示 */
.el-table__body td {
  cursor: pointer;
  transition: background-color 0.2s;
}

.el-table__body td:hover {
  background-color: #f5f7fa;
}

/* 固定表格行高 */
:deep(.el-table__row) {
  height: 40px !important;
}

:deep(.el-table__cell) {
  padding: 4px !important;
  height: 40px !important;
}

/* 查询条件卡片 */
.query-condition-card {
  margin-bottom: 10px;
}

/* 查询操作符选择器 */
.operator-select {
  width: 100px;
  margin-right: 5px;
}

/* 查询输入框 */
.query-input {
  flex: 1;
}

/* 复制按钮 */
.copy-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  left: 0;
  z-index: 10;
  padding: 2px;
  opacity: 0.6;
}

.copy-btn:hover {
  opacity: 1;
  background-color: rgba(64, 158, 255, 0.1);
  border-radius: 2px;
}
</style>
