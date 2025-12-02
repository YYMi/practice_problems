import request from '../utils/request';
import type { ApiResponse, PointSummary, PointDetail, CreatePointParams, UpdatePointParams } from '../types';

// 定义基础路径
const API_PATH = '/points';

// 获取列表
export const getPoints = (categoryId: number) => {
    return request.get<any, { data: ApiResponse<PointSummary[]> }>(`${API_PATH}?category_id=${categoryId}`);
};

// 获取详情
export const getPointDetail = (id: number) => {
    return request.get<any, { data: ApiResponse<PointDetail> }>(`${API_PATH}/${id}`);
};

// 创建
export const createPoint = (data: CreatePointParams) => {
    return request.post<any, { data: ApiResponse<{id: number}> }>(API_PATH, data);
};

// 更新
export const updatePoint = (id: number, data: UpdatePointParams) => {
    return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`, data);
};

// 删除知识点
export const deletePoint = (id: number) => {
    return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`);
};

// 删除知识点下的某张图片
export const deletePointImage = (id: number, filePath: string) => {
    return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/image`, {
        data: { filePath }
    });
};

// 知识点排序
export const updatePointSort = (id: number, action: 'top' | 'up' | 'down') => {
    // 注意：这里使用了 put 方法，保持了您原有的逻辑。
    // 如果新版后台规范统一改为 post，您可以将 request.put 改为 request.post
    return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/sort`, { action });
};

// 上传图片 (通用)
// 注意：上传通常是一个独立的接口，可能不完全遵循 API_PATH，这里假设上传接口路径为 /upload
export const uploadImage = (file: File,pointId:number) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', 'point'); // 告诉后端存到 point 目录
    formData.append('pointId', pointId.toString());
    
    // 使用 request 实例发送，以确保携带 token
    return request.post<any, { data: ApiResponse<{url: string, path: string}> }>( `upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data' // 明确指定上传类型，虽然 request 可能会自动处理，但显式声明更安全
        }
    });
};