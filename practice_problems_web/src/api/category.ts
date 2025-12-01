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

export const updateCategory = (id: number, data: { categoryName: string }) => {
    return axios.put<ApiResponse<null>>(`${API_URL}/${id}`, data);
};

export const deleteCategory = (id: number) => {
    return axios.delete<ApiResponse<null>>(`${API_URL}/${id}`);
};
