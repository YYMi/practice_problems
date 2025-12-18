import request from '../utils/request';
import type { ApiResponse } from '../types';

const API_PATH = '/collections';

// 集合相关类型定义
export interface Collection {
  id: number;
  name: string;
  createTime: string;
  updateTime: string;
  isPublic?: boolean;      // 是否公有
  isOwner?: boolean;       // 是否是所有者
  ownerUserCode?: string;  // 所有者用户代码
}

export interface CollectionForm {
  name: string;
}

export interface CollectionPoint {
  id: number;
  pointId: number;
  subjectId: number;
  categoryId: number;
  title: string;
  subjectName: string;
  categoryName: string;
  pointDifficulty: number;
  categoryDifficulty: number;
  sortOrder: number;
  createTime: string;
}

export interface CollectionQuestionItem {
  id: number;
  knowledgePointId: number;
  questionText: string;
  option1: string;
  option1Img: string;
  option2: string;
  option2Img: string;
  option3: string;
  option3Img: string;
  option4: string;
  option4Img: string;
  correctAnswer: number;
  explanation: string;
  note: string;
  createTime: string;
}

export interface CollectionPointsResponse {
  list: CollectionPoint[];
  total: number;
  page: number;
  pageSize: number;
}

// 集合权限相关类型定义
export interface CollectionPermission {
  userCode: string;
  nickname: string;
  email: string;
  expireTime: string;
  createTime: string;
}

export interface CollectionPermissionsResponse {
  list: CollectionPermission[];
  total: number;
  page: number;
  pageSize: number;
}

// 获取集合列表
export const getCollections = () => {
  return request.get<any, { data: ApiResponse<Collection[]> }>(API_PATH);
};

// 创建集合
export const createCollection = (data: CollectionForm) => {
  return request.post<any, { data: ApiResponse<{ id: number; name: string }> }>(API_PATH, data);
};

// 更新集合
export const updateCollection = (id: number, data: CollectionForm) => {
  return request.put<any, { data: ApiResponse<{ id: number; name: string }> }>(`${API_PATH}/${id}`, data);
};

// 删除集合
export const deleteCollection = (id: number) => {
  return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`);
};

// 添加知识点到集合
export const addPointToCollection = (collectionId: number, pointId: number) => {
  return request.post<any, { data: ApiResponse<null> }>(`${API_PATH}/points`, {
    collection_id: collectionId,
    point_id: pointId,
  });
};

// 批量添加知识点到集合（科目/分类级别）
export const batchAddPointsToCollection = (collectionId: number, options: { subjectId?: number; categoryId?: number }) => {
  return request.post<any, { data: ApiResponse<{ added: number }> }>(`${API_PATH}/points/batch`, {
    collection_id: collectionId,
    subject_id: options.subjectId || 0,
    category_id: options.categoryId || 0,
  });
};

// 获取集合的知识点列表（分页）
export const getCollectionPoints = (collectionId: number, page: number = 1, pageSize: number = 20) => {
  return request.get<any, { data: ApiResponse<CollectionPointsResponse> }>(
    `${API_PATH}/${collectionId}/points`,
    {
      params: { page, pageSize }
    }
  );
};

// 获取集合中知识点的详情（专用于集合页面）
export const getCollectionPointDetail = (collectionId: number, pointId: number) => {
  return request.get<any, { data: ApiResponse<any> }>(
    `${API_PATH}/${collectionId}/points/${pointId}`
  );
};

// 从集合中移除知识点（按itemId）
export const removePointFromCollection = (itemId: number) => {
  return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/items/${itemId}`);
};

// 从集合中移除知识点（按collectionId和pointId）
export const removePointByIds = (collectionId: number, pointId: number) => {
  return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${collectionId}/points/${pointId}`);
};

// 更新集合项排序
export const updateCollectionItemsOrder = (collectionId: number, items: { id: number; sort_order: number }[]) => {
  return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/items/order`, {
    collection_id: collectionId,
    items: items,
  });
};

// 获取集合中所有题目（用于综合刷题）
export const getCollectionQuestions = (collectionId: number, limit: number = 20) => {
  return request.get<any, { data: ApiResponse<CollectionQuestionItem[]> }>(
    `${API_PATH}/${collectionId}/questions`,
    {
      params: { limit }
    }
  );
};

// 设置集合权限（公有/私有）
export const setCollectionPermission = (id: number, isPublic: boolean) => {
  return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/permission`, {
    isPublic
  });
};

// 添加集合授权
export const addCollectionPermission = (id: number, userCode: string, expireTime?: string) => {
  return request.post<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/permissions`, {
    userCode,
    expireTime
  });
};

// 获取集合授权列表（分页）
export const getCollectionPermissions = (id: number, page: number = 1, pageSize: number = 10, search: string = '') => {
  return request.get<any, { data: ApiResponse<CollectionPermissionsResponse> }>(`${API_PATH}/${id}/permissions`, {
    params: { page, pageSize, search }
  });
};

// 更新集合授权时间
export const updateCollectionPermission = (id: number, userCode: string, expireTime?: string) => {
  return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/permissions`, {
    userCode,
    expireTime
  });
};

// 删除集合授权
export const deleteCollectionPermission = (id: number, userCode: string) => {
  return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/permissions`, {
    params: { userCode }
  });
};

// 获取知识点已绑定的集合列表
export const getPointCollections = (pointId: number) => {
  return request.get<any, { data: ApiResponse<Collection[]> }>(`${API_PATH}/point-collections`, {
    params: { point_id: pointId }
  });
};

// 查找知识点在哪个集合中（用于绑定跳转）
export const findPointInCollections = (pointId: number, currentCollectionId?: number) => {
  return request.get<any, { data: ApiResponse<{
    found: boolean;
    collectionId?: number;
    page?: number;
    points?: CollectionPoint[];
    total?: number;
    message?: string;
  }> }>(`${API_PATH}/find-point`, {
    params: { 
      point_id: pointId,
      current_collection_id: currentCollectionId || ''
    }
  });
};
