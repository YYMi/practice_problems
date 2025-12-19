import request from '../utils/request';
import type { ApiResponse, PointSummary, PointDetail, CreatePointParams, UpdatePointParams } from '../types';

// 定义基础路径
const API_PATH = '/points';

// 获取列表（支持分页）
export const getPoints = (categoryId: number, page: number = 1, pageSize: number = 11) => {
    return request.get<any, { data: ApiResponse<{
        list: PointSummary[];
        total: number;
        page: number;
        pageSize: number;
    }> }>(`${API_PATH}?category_id=${categoryId}&page=${page}&pageSize=${pageSize}`);
};

// 获取详情
export const getPointDetail = (id: number) => {
    return request.get<any, { data: ApiResponse<PointDetail> }>(`${API_PATH}/${id}`);
};

// 创建
export const createPoint = (data: CreatePointParams) => {
    return request.post<any, { data: ApiResponse<{id: number}> }>(API_PATH, data);
};

// 更新
export const updatePoint = (id: number, data: UpdatePointParams) => {
    return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`, data);
};

// 删除知识点
export const deletePoint = (id: number) => {
    return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}`);
};

// 删除知识点下的某张图片
export const deletePointImage = (id: number, filePath: string) => {
    return request.delete<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/image`, {
        data: { filePath }
    });
};

// 知识点排序
export const updatePointSort = (id: number, action: 'top' | 'up' | 'down') => {
    // 注意：这里使用了 put 方法，保持了您原有的逻辑。
    // 如果新版后台规范统一改为 post，您可以将 request.put 改为 request.post
    return request.put<any, { data: ApiResponse<null> }>(`${API_PATH}/${id}/sort`, { action });
};

// 知识点模糊搜索
export interface SearchPointResult {
    pointId: number;
    pointTitle: string;
    categoryId: number;
    categoryName: string;
    subjectId: number;
    subjectName: string;
}

export const searchPoints = (keyword: string) => {
    return request.get<any, { data: ApiResponse<SearchPointResult[]> }>(`${API_PATH}/search?keyword=${encodeURIComponent(keyword)}`);
};

// 计算文件的 SHA-256 值
const calculateFileSHA256 = async (file: File): Promise<string> => {
    const arrayBuffer = await file.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
};

// 格式化哈希字符串 helper 函数
// 输入: 64位长字符串
// 输出: 4-6-10-12 格式的 32位字符串 (后端存储路径: 4/6/10-12)
const formatHash = (fullHash: string): string => {
    const shortHash = fullHash.substring(0, 32);
    const part1 = shortHash.substring(0, 4);
    const part2 = shortHash.substring(4, 10);
    const part3 = shortHash.substring(10, 20);
    const part4 = shortHash.substring(20, 32);
    return `${part1}-${part2}-${part3}-${part4}`;
};

// 检查文件是否已存在（通过hash判断）
const checkFileExists = (hash: string) => {
    return request.get<any, { data: ApiResponse<{ exists: boolean; path?: string; url?: string }> }>(
        `upload/check?hash=${hash}`
    );
};

// 上传图片 (通用) - 支持秒传
export const uploadImage = async (file: File, pointId: number) => {
    // 1. 计算文件 SHA-256
    const sha256 = await calculateFileSHA256(file);
    const formattedHash = formatHash(sha256);
    
    // 2. 先检查文件是否已存在
    const checkRes = await checkFileExists(formattedHash);
    if (checkRes.data.code === 200 && checkRes.data.data?.exists) {
        // 文件已存在，直接返回地址（秒传）
        return {
            data: {
                code: 200,
                data: {
                    url: checkRes.data.data.url!,
                    path: checkRes.data.data.path!
                },
                msg: '文件已存在'
            }
        };
    }
    
    // 3. 文件不存在，执行上传
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', 'point');
    formData.append('pointId', pointId.toString());
    formData.append('hash', formattedHash); // 传格式化后的hash，后端用于存储路径 4/6/10-12
    
    return request.post<any, { data: ApiResponse<{url: string, path: string}> }>(`upload`, formData, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    });
};