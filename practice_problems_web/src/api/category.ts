import axios from 'axios';
import type { ApiResponse, Category, CategoryForm } from '../types';

const API_URL = 'http://localhost:8080/api/v1/categories';

export const getCategories = (subjectId?: number) => {
    let url = API_URL;
    if (subjectId) {
        url += `?subject_id=${subjectId}`;
    }
    return axios.get<ApiResponse<Category[]>>(url);
};

export const createCategory = (data: CategoryForm) => {
    return axios.post<ApiResponse<{id: number}>>(API_URL, data);
};

export const updateCategory = (id: number, data: { categoryName: string ,difficulty:number}) => {
    return axios.put<ApiResponse<null>>(`${API_URL}/${id}`, data);
};

export const deleteCategory = (id: number) => {
    return axios.delete<ApiResponse<null>>(`${API_URL}/${id}`);
};
// 新增：更新分类排序
export const updateCategorySort = (id: number, action: 'top' | 'up' | 'down') => {
    // 1. 保持和上面一致，使用 axios.put
    // 2. 使用 API_URL 拼接路径
    return axios.put<ApiResponse<null>>(`${API_URL}/${id}/sort`, { action });
};