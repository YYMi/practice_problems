// src/api/subject.ts
import request from '../utils/request';
import type { ApiResponse, Subject, SubjectForm } from '../types';

const API_PATH = '/subjects';

// 1. 这里使用 axios 的 Response 类型定义
// request.get 返回的是 Promise<AxiosResponse<ApiResponse<Subject[]>>>
// 但我们在拦截器里可能直接返回了 data，或者保留了 axios 结构
// 最稳妥的写法是不加太复杂的泛型，让 TS 自动推断，或者显式指定
export const getSubjects = () => {
    // 注意：request.get 返回的是一个 Promise
    return request.get<any, ApiResponse<Subject[]>>(API_PATH);
};

export const createSubject = (data: SubjectForm) => {
    return request.post<any, ApiResponse<{id: number}>>(API_PATH, data);
};

export const updateSubject = (id: number, data: SubjectForm) => {
    return request.put<any, ApiResponse<null>>(`${API_PATH}/${id}`, data);
};

export const deleteSubject = (id: number) => {
    return request.delete<any, ApiResponse<null>>(`${API_PATH}/${id}`);
};
