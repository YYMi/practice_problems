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
}