import request from '../utils/request';
import type { ApiResponse } from '../types'; // 假设你的通用类型定义在这里

// 定义请求参数类型 (这样写代码会有提示，更规范)
export interface CreateShareParams {
    subject_ids: number[];
    duration: string;
    type: number;
    targets: string[];
}

export interface BindShareParams {
    code: string;
}

// 1. 创建分享 (生成码 或 指定用户)
export const createShare = (data: CreateShareParams) => {
    // 后端返回的数据可能是 string (分享码) 或者 null (指定用户时只返回msg)
    return request.post<any, { data: ApiResponse<string | null> }>('/share/create', data);
};

// 2. 绑定分享
export const bindShare = (data: BindShareParams) => {
    return request.post<any, { data: ApiResponse<null> }>('/share/bind', data);
};

// 获取我的分享码列表
export const getMyShareCodes = () => {
    return request.get<any, { data: ApiResponse<any[]> }>('/share/list');
};

// 删除分享码
export const deleteShareCode = (id: number) => {
    return request.delete<any, { data: ApiResponse<null> }>(`/share/${id}`);
};

// 更新分享码信息
export const updateShareCode = (id: number, newDate: string, newDuration: string) => {
    return request.put<any, { data: ApiResponse<null> }>(`/share/${id}`, { 
        new_expire_date: newDate,
        new_duration: newDuration 
    });
}

// 获取某科目的授权用户列表
export const getSubjectUsers = (subjectId: number, params: any) => {
    return request.get<any, { data: ApiResponse<{list: any[], total: number}> }>(`/subject/${subjectId}/users`, { params });
};

// 更新授权有效期
export const updateAuth = (id: number, newDate: string) => {
    return request.put<any, { data: ApiResponse<null> }>(`/auth/${id}`, { new_expire_date: newDate });
};

// 解除授权
export const removeAuth = (id: number) => {
    return request.delete<any, { data: ApiResponse<null> }>(`/auth/${id}`);
};

// 批量更新有效期
export const batchUpdateAuth = (ids: number[], newDate: string) => {
    return request.put<any, { data: ApiResponse<null> }>('/auth/batch/update', { 
        ids, 
        new_expire_date: newDate 
    });
};

// 批量移除
export const batchRemoveAuth = (ids: number[]) => {
    return request.put<any, { data: ApiResponse<null> }>('/auth/batch/remove', { ids });
};
