import request from '../utils/request'

// ========== TOTP相关（谷歌验证码） ==========

// 检查是否已绑定TOTP
export function checkTotpBound() {
  return request({
    url: '/totp/check',
    method: 'get'
  })
}

// 生成TOTP密钥和二维码
export function generateTotpSecret() {
  return request({
    url: '/totp/generate',
    method: 'get'
  })
}

// 验证并绑定TOTP
export function bindTotp(secret: string, code: string) {
  return request({
    url: '/totp/bind',
    method: 'post',
    data: {
      secret,
      code
    }
  })
}

// 验证TOTP码
export function verifyTotp(code: string) {
  return request({
    url: '/totp/verify',
    method: 'post',
    data: {
      code
    }
  })
}

// 解绑TOTP
export function unbindTotp(code: string) {
  return request({
    url: '/totp/unbind',
    method: 'post',
    data: {
      code
    }
  })
}

// ========== 数据库管理相关 ==========

// 获取所有表
export function getAllTables() {
  return request({
    url: '/admin/db/tables',
    method: 'get'
  })
}

// 获取表结构
export function getTableStructure(tableName: string) {
  return request({
    url: `/admin/db/tables/${tableName}/structure`,
    method: 'get'
  })
}

// 获取表数据
export function getTableData(tableName: string, params: any) {
  return request({
    url: `/admin/db/tables/${tableName}/data`,
    method: 'get',
    params
  })
}

// 插入数据
export function insertTableRow(tableName: string, data: any, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/insert`,
    method: 'post',
    data: {
      data,
      recaptcha_token: recaptchaToken
    }
  })
}

// 更新数据
export function updateTableRow(tableName: string, where: any, data: any, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/update`,
    method: 'put',
    data: {
      where,
      data,
      recaptcha_token: recaptchaToken
    }
  })
}

// 删除数据
export function deleteTableRows(tableName: string, where: any, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/delete`,
    method: 'delete',
    data: {
      where,
      recaptcha_token: recaptchaToken
    }
  })
}

// 批量更新
export function batchUpdateTableRows(tableName: string, items: any[], primaryKey: string, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/batch-update`,
    method: 'put',
    data: {
      items,
      primary_key: primaryKey,
      recaptcha_token: recaptchaToken
    }
  })
}

// 批量删除
export function batchDeleteTableRows(tableName: string, ids: any[], primaryKey: string, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/batch-delete`,
    method: 'delete',
    data: {
      ids,
      primary_key: primaryKey,
      recaptcha_token: recaptchaToken
    }
  })
}

// 获取表备注
export function getTableComment(tableName: string) {
  return request({
    url: `/admin/db/tables/${tableName}/comment`,
    method: 'get'
  })
}

// 设置表备注
export function setTableComment(tableName: string, comment: string, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/comment`,
    method: 'post',
    data: {
      comment,
      recaptcha_token: recaptchaToken
    }
  })
}

// 获取字段备注
export function getColumnComment(tableName: string, columnName: string) {
  return request({
    url: `/admin/db/tables/${tableName}/columns/${columnName}/comment`,
    method: 'get'
  })
}

// 设置字段备注
export function setColumnComment(tableName: string, columnName: string, comment: string, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/columns/${columnName}/comment`,
    method: 'post',
    data: {
      comment,
      recaptcha_token: recaptchaToken
    }
  })
}

// 获取所有表备注
export function getAllTableComments() {
  return request({
    url: `/admin/db/table-comments`,
    method: 'get'
  })
}

// 获取所有字段备注
export function getAllColumnComments() {
  return request({
    url: `/admin/db/column-comments`,
    method: 'get'
  })
}

// 添加字段
export function addColumn(tableName: string, columnName: string, columnType: string, defaultValue: string, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/columns`,
    method: 'post',
    data: {
      column_name: columnName,
      column_type: columnType,
      default_value: defaultValue,
      recaptcha_token: recaptchaToken
    }
  })
}

// 删除字段
export function dropColumn(tableName: string, columnName: string, recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/columns/${columnName}`,
    method: 'delete',
    data: {
      recaptcha_token: recaptchaToken
    }
  })
}

// 获取字段排序
export function getColumnOrders(tableName: string) {
  return request({
    url: `/admin/db/tables/${tableName}/column-orders`,
    method: 'get'
  })
}

// 保存字段排序
export function saveColumnOrders(tableName: string, orders: string[], recaptchaToken: string) {
  return request({
    url: `/admin/db/tables/${tableName}/column-orders`,
    method: 'post',
    data: {
      orders,
      recaptcha_token: recaptchaToken
    }
  })
}
