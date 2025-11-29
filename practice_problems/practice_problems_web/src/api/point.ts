import axios from 'axios';
import type { ApiResponse, PointSummary, PointDetail, CreatePointParams, UpdatePointParams } from '../types';

const API_URL = 'http://localhost:8080/api/v1/points';

// 获取列表
export const getPoints = (categoryId: number) => {
    return axios.get<ApiResponse<PointSummary[]>>(`${API_URL}?category_id=${categoryId}`);
};

// 获取详情
export const getPointDetail = (id: number) => {
    return axios.get<ApiResponse<PointDetail>>(`${API_URL}/${id}`);
};

// 创建
export const createPoint = (data: CreatePointParams) => {
    return axios.post<ApiResponse<{id: number}>>(API_URL, data);
};

// 更新
export const updatePoint = (id: number, data: UpdatePointParams) => {
    return axios.put<ApiResponse<null>>(`${API_URL}/${id}`, data);
};

// 删除知识点
export const deletePoint = (id: number) => {
    return axios.delete<ApiResponse<null>>(`${API_URL}/${id}`);
};

// 删除知识点下的某张图片
export const deletePointImage = (id: number, filePath: string) => {
    return axios.delete<ApiResponse<null>>(`${API_URL}/${id}/image`, {
        data: { filePath }
    });
};

// 上传图片 (通用)
export const uploadImage = (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', 'point'); // 告诉后端存到 point 目录
    return axios.post<ApiResponse<{url: string, path: string}>>('http://localhost:8080/api/v1/upload', formData);
};

