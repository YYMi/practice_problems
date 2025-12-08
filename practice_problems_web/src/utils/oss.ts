/**
 * OSS 资源路径处理工具
 * 
 * 使用说明：
 * - 如果服务器配置了 OSS，图片等资源会从 OSS 加载
 * - 如果未配置 OSS，资源从本地服务器加载
 */

/**
 * 获取 OSS 基础地址
 * @returns OSS 地址，未配置时返回空字符串
 */
export function getOssUrl(): string {
  return localStorage.getItem('oss_url') || ''
}

/**
 * 判断是否启用了 OSS
 * @returns true 表示已配置 OSS
 */
export function isOssEnabled(): boolean {
  return !!getOssUrl()
}

/**
 * 获取资源的完整 URL
 * 如果配置了 OSS，返回 OSS 地址；否则返回本地服务器地址
 * 
 * @param path 资源相对路径（例如：/uploads/images/xxx.jpg）
 * @returns 完整的资源 URL
 * 
 * @example
 * // 未配置 OSS 时
 * getResourceUrl('/uploads/images/1.jpg') 
 * // 返回: '/uploads/images/1.jpg'
 * 
 * // 配置了 OSS (endpoint: https://bucket.oss.com, prefix: '') 时
 * getResourceUrl('/uploads/images/1.jpg')
 * // 返回: 'https://bucket.oss.com/uploads/images/1.jpg'
 */
export function getResourceUrl(path: string): string {
  if (!path) return ''
  
  const ossUrl = getOssUrl()
  
  // 如果未配置 OSS，直接返回原路径（从本地服务器获取）
  if (!ossUrl) {
    return path
  }
  
  // 处理路径拼接，避免双斜杠
  const cleanPath = path.startsWith('/') ? path : '/' + path
  const cleanOssUrl = ossUrl.endsWith('/') ? ossUrl.slice(0, -1) : ossUrl
  
  return cleanOssUrl + cleanPath
}

/**
 * 获取图片资源 URL
 * 这是 getResourceUrl 的别名，语义更清晰
 */
export const getImageUrl = getResourceUrl
