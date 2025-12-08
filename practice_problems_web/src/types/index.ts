// 通用响应结构 (对应 Gin 的 gin.H 返回结构)
export interface ApiResponse<T> {
    code: number;
    msg: string;
    data: T;
}

// 科目模型 (对应 Go 的 model.Subject)
export interface Subject {
    id: number;
    name: string;
    status: number; // 1: 启用, 0: 禁用
    createTime?: string;
    updateTime?: string;
}

// 分类模型 (对应 Go 的 model.KnowledgeCategory)
export interface Category {
    id: number;
    subjectId: number;
    categoryName: string;
    createTime?: string;
    updateTime?: string;
    difficulty?: number; // ✅ 新增：难度字段 (0-3)
}

// 创建/修改科目的表单类型
export interface SubjectForm {
    id?: number;
    name: string;
    status: number;
}

// 创建/修改分类的表单类型
export interface CategoryForm {
    id?: number;
    subjectId?: number;
    categoryName: string;
}


// 知识点 (列表用)
export interface PointSummary {
    id: number;
    title: string;
    createTime: string;
      sortOrder: number;   // 新增
  difficulty?: number; // 新增
}

// 知识点 (详情用)
export interface PointDetail {
    id: number;
    categoryId: number;
    title: string;
    content: string;
    referenceLinks: string; // 后端传过来是 JSON 字符串
    localImageNames: string; // 后端传过来是 JSON 字符串
    updateTime: string;
    videoUrl?: string;      // JSON string (存的是 ["url1", "url2"])
}

// 创建知识点参数
export interface CreatePointParams {
    categoryId: number;
    title: string;
}

// 更新知识点参数
export interface UpdatePointParams {
    title?: string;
    content?: string;
    referenceLinks?: string;
    localImageNames?: string;
    difficulty?: number; // ✅ 必须显式加上这一行！
    // ★★★ 新增这一行，允许更新分类ID ★★★
    categoryId?: number; 
    videoUrl?: string; // 新增视频链接字段
}

// ★★★ 关键：必须定义通用的 API 响应结构 ★★★
export interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
}