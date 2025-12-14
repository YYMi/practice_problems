import request from '../utils/request'
import type { ApiResponse } from '../types'

// 定义基础路径
const API_PATH = '/points'

// 获取知识点笔记
export function getPointNote(pointId: number) {
  return request.get<any, { data: ApiResponse<{
    note: string
    create_time: string
    update_time: string
  }> }>(`${API_PATH}/${pointId}/note`)
}

// 保存知识点笔记
export function savePointNote(pointId: number, note: string) {
  return request.post<any, { data: ApiResponse<null> }>(`${API_PATH}/${pointId}/note`, { note })
}