import request from '../utils/request';

// 创建知识点绑定
export function createBinding(data: {
  sourceSubjectId: number;
  sourcePointId: number;
  targetSubjectId: number;
  targetPointId: number;
  bindText: string;
}) {
  return request({
    url: '/point-bindings',
    method: 'post',
    data
  });
}

// 获取知识点的所有绑定
export function getBindingsByPoint(pointId: number) {
  return request({
    url: `/point-bindings/${pointId}`,
    method: 'get'
  });
}

// 删除绑定
export function deleteBinding(id: number) {
  return request({
    url: `/point-bindings/${id}`,
    method: 'delete'
  });
}

// 获取科目下的所有分类（用于绑定选择）
export function getCategoriesBySubjectForBinding(subjectId: number) {
  return request({
    url: `/binding/subjects/${subjectId}/categories`,
    method: 'get'
  });
}

// 获取分类下的所有知识点（用于绑定选择）
export function getPointsByCategoryForBinding(categoryId: number) {
  return request({
    url: `/binding/categories/${categoryId}/points`,
    method: 'get'
  });
}
