// src/api/subject.ts
import request from '../utils/request';
import type { ApiResponse, Subject, SubjectForm } from '../types';

const API_PATH = '/subjects';

// 1. 这里使用 axios 的 Response 类型定义
// request.get 返回的是 Promise<AxiosResponse<ApiResponse<Subject[]>>>
// 但我们在拦截器里可能直接返回了 data，或者保留了 axios 结构
// 最稳妥的写法是不加太复杂的泛型，让 TS 自动推断，或者显式指定
export const getSubjects = () => {
    // 注意：request.get 返回的是一个 Promise，拦截器返回完整的 AxiosResponse
    // 所以实际类型是 AxiosResponse<ApiResponse<Subject[]>>
    return request.get(API_PATH);
};

export const createSubject = (data: SubjectForm) => {
    return request.post(API_PATH, data);
};

export const updateSubject = (id: number, data: SubjectForm) => {
    return request.put(`${API_PATH}/${id}`, data);
};

export const deleteSubject = (id: number) => {
    return request.delete(`${API_PATH}/${id}`);
};
