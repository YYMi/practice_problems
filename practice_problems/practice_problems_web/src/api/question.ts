import axios from 'axios';
// 假设你有一个通用的类型定义文件，如果没有，可以删掉这行导入，直接用 any 或在下面定义
import type { ApiResponse } from '../types'; 

const API_URL = 'http://localhost:8080/api/v1/questions';

// --- 类型定义 (建议放在 src/types/index.ts 中，这里为了方便先写在这里) ---

// 创建题目的参数结构 (对应后端 CreateQuestionRequest)
export interface CreateQuestionParams {
  knowledgePointId: number;
  questionText: string;
  option1: string;
  option2: string;
  option3: string;
  option4: string;
  correctAnswer: number; // 1, 2, 3, 4
  explanation: string;
  note?: string;
}

// 题目列表项结构 (对应后端 Question)
export interface QuestionItem {
  id: number;
  knowledgePointId: number;
  questionText: string;
  option1: string;
  option2: string;
  option3: string;
  option4: string;
  correctAnswer: number;
  explanation: string;
  note: string;
  createTime: string;
}

// --- 接口函数 ---

/**
 * 获取某知识点下的题目列表
 * GET /api/v1/questions?point_id=1
 */
export const getQuestions = (pointId: number) => {
    return axios.get<ApiResponse<QuestionItem[]>>(`${API_URL}?point_id=${pointId}`);
}

/**
 * 创建新题目
 * POST /api/v1/questions
 */
export const createQuestion = (data: CreateQuestionParams) => {
    return axios.post<ApiResponse<{id: number}>>(API_URL, data);
}

/**
 * 更新题目 (预留)
 * PUT /api/v1/questions/:id
 */
export const updateQuestion = (id: number, data: Partial<CreateQuestionParams>) => {
    return axios.put<ApiResponse<null>>(`${API_URL}/${id}`, data);
}

/**
 * 删除题目
 * DELETE /api/v1/questions/:id
 */
export const deleteQuestion = (id: number) => {
    return axios.delete<ApiResponse<null>>(`${API_URL}/${id}`);
}
export const getQuestionsByCategory = (categoryId: number) => {
  // 假设后端有这个接口，URL 类似 /api/v1/questions?category_id=xxx
  // 如果后端没有，你可能需要前端循环调用 getQuestions(pointId) 来拼接，但建议后端加接口
  return axios.get<ApiResponse<QuestionItem[]>>(`${API_URL}?category_id=${categoryId}`);
}
