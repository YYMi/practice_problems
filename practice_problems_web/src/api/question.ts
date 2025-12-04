import request from '../utils/request';
import type { ApiResponse } from '../types';

// 定义基础路径 (对应 /api/v1/questions)
const API_PATH = '/questions';

// --- 类型定义 ---
// (建议后续将这些移动到 src/types/index.ts 中统一管理，这里暂时保留以确保代码可运行)

// 创建题目的参数结构
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

// 题目列表项结构
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
    return request.get<any, { data: ApiResponse<QuestionItem[]> }>(`${API_PATH}?point_id=${pointId}`);
};

/**
 * 创建新题目
 * POST /api/v1/questions
 */
export const createQuestion = (data: CreateQuestionParams) => {
    return request.post<any, { data: ApiResponse<{id: number}> }>(API_PATH, data);
};

/**
 * 更新题目
 * PUT /api/v1/questions/:id
 */
export const updateQuestion = (id: number, data: Partial<CreateQuestionParams>) => {
    return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`, data);
};

/**
 * 删除题目
 * DELETE /api/v1/questions/:id
 */
export const deleteQuestion = (id: number) => {
    return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`);
};

/**
 * 根据分类获取题目
 * GET /api/v1/questions?category_id=xxx
 */
export const getQuestionsByCategory = (categoryId: number) => {
    // ★★★ 修复点：把 point_id 改为 category_id ★★★
    return request.get<any, { data: ApiResponse<QuestionItem[]> }>(`${API_PATH}?category_id=${categoryId}`);
};

// 2. 新增这个！这是修改备注（用户用）
export const updateUserNote = (data: { question_id: number; note: string }) => {
    return request.post<any, { data: ApiResponse<null> }>(`${API_PATH}/note`, data);
};

