import request from '../utils/request';
import type { ApiResponse, Category, CategoryForm } from '../types';

const API_PATH = '/categories';

// 获取分类列表
// 获取分类列表 (必须传 subjectId)
export const getCategories = (subjectId: number) => {
    // 将 subjectId 拼接到 URL 参数中
    return request.get<any, { data: ApiResponse<Category[]> }>(`${API_PATH}?subject_id=${subjectId}`);
};

// ★★★ 之前可能缺失的方法：创建分类 ★★★
export const createCategory = (data: CategoryForm) => {
    return request.post<any, { data: ApiResponse<{id: number}> }>(API_PATH, data);
};

// ★★★ 之前可能缺失的方法：更新分类 ★★★
export const updateCategory = (id: number, data: CategoryForm) => {
    return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`, data);
};

// ★★★ 之前可能缺失的方法：删除分类 ★★★
export const deleteCategory = (id: number) => {
    return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`);
};

// 分类排序
export const sortCategories = (id: number, direction: string) => {
    return request.post<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/sort`, { direction });
};

// ★★★ 修复重点：添加 updateCategorySort 方法 ★★★
export const updateCategorySort = (id: number, direction: string) => {
    // ★★★ 修改点：把 key 从 direction 改成 action ★★★
    return request.post<any, { data: ApiResponse<null> }>(
        `${API_PATH}/${id}/sort`, 
        { action: direction } // 这里！
    );
};