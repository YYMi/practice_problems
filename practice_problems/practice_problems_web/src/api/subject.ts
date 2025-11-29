import axios from 'axios';
import type { ApiResponse, Subject, SubjectForm } from '../types';

const API_URL = 'http://localhost:8080/api/v1/subjects';

// 注意：axios.get<T> 的 T 代表响应体的类型
export const getSubjects = () => {
    return axios.get<ApiResponse<Subject[]>>(API_URL);
};

export const createSubject = (data: SubjectForm) => {
    return axios.post<ApiResponse<{id: number}>>(API_URL, data);
};

export const updateSubject = (id: number, data: SubjectForm) => {
    return axios.put<ApiResponse<null>>(`${API_URL}/${id}`, data);
};

export const deleteSubject = (id: number) => {
    return axios.delete<ApiResponse<null>>(`${API_URL}/${id}`);
};
