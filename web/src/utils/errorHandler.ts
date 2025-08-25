/**
 * 网络错误处理工具
 * 
 * 提供统一的API错误处理和用户友好的错误消息
 * 这个工具可以帮助前端应用更好地处理各种网络和API错误
 * 
 * 前端学习要点：
 * 1. 错误分类 - 区分不同类型的错误（网络、HTTP、业务逻辑）
 * 2. 用户体验 - 提供用户友好的错误消息
 * 3. 错误恢复 - 某些错误可以自动重试
 * 4. 日志记录 - 记录详细错误信息用于调试
 */

import { AxiosError } from 'axios'

// 错误类型枚举
export enum ErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',           // 网络连接错误
  SERVER_ERROR = 'SERVER_ERROR',             // 服务器错误 (5xx)
  CLIENT_ERROR = 'CLIENT_ERROR',             // 客户端错误 (4xx)
  API_ERROR = 'API_ERROR',                   // API业务逻辑错误
  VALIDATION_ERROR = 'VALIDATION_ERROR',     // 参数验证错误
  TIMEOUT_ERROR = 'TIMEOUT_ERROR',           // 请求超时
  UNKNOWN_ERROR = 'UNKNOWN_ERROR'            // 未知错误
}

// 处理后的错误信息接口
export interface ProcessedError {
  type: ErrorType
  message: string          // 用户友好的错误消息
  originalError?: any      // 原始错误对象
  code?: string | number   // 错误代码
  canRetry?: boolean       // 是否可以重试
  details?: string         // 详细错误信息（调试用）
}

/**
 * 网络错误处理器类
 */
export class NetworkErrorHandler {
  /**
   * 处理API错误的主要方法
   * 
   * @param error 原始错误对象
   * @returns 处理后的错误信息
   */
  static handleError(error: any): ProcessedError {
    console.error('🚨 API Error occurred:', error)

    // 处理Axios错误
    if (error.isAxiosError || error.response) {
      return this.handleAxiosError(error as AxiosError)
    }

    // 处理网络连接错误
    if (error.code === 'NETWORK_ERR' || error.message?.includes('Network Error')) {
      return {
        type: ErrorType.NETWORK_ERROR,
        message: '网络连接失败，请检查您的网络连接',
        originalError: error,
        canRetry: true,
        details: '无法连接到服务器，可能是网络问题或服务器暂时不可用'
      }
    }

    // 处理超时错误
    if (error.code === 'ECONNABORTED' || error.message?.includes('timeout')) {
      return {
        type: ErrorType.TIMEOUT_ERROR,
        message: '请求超时，请稍后重试',
        originalError: error,
        canRetry: true,
        details: '请求处理时间过长，可能是服务器负载较高'
      }
    }

    // 处理其他JavaScript错误
    if (error instanceof Error) {
      return {
        type: ErrorType.UNKNOWN_ERROR,
        message: '操作失败，请稍后重试',
        originalError: error,
        canRetry: false,
        details: error.message
      }
    }

    // 兜底错误处理
    return {
      type: ErrorType.UNKNOWN_ERROR,
      message: '发生未知错误，请刷新页面重试',
      originalError: error,
      canRetry: false,
      details: String(error)
    }
  }

  /**
   * 处理Axios HTTP错误
   */
  private static handleAxiosError(error: AxiosError): ProcessedError {
    const { response, request } = error

    // 有响应但状态码不是2xx
    if (response) {
      const status = response.status
      const data = response.data as any

      // 处理服务器返回的业务错误
      if (data && typeof data === 'object') {
        // 检查是否是标准的API错误格式
        if (data.code && data.msg) {
          return {
            type: ErrorType.API_ERROR,
            message: this.getChineseErrorMessage(data.msg, status),
            originalError: error,
            code: data.code,
            canRetry: this.canRetryByStatus(status),
            details: `API错误: ${data.msg} (Code: ${data.code})`
          }
        }

        // 处理其他格式的错误响应
        if (data.message || data.error) {
          return {
            type: ErrorType.API_ERROR,
            message: this.getChineseErrorMessage(data.message || data.error, status),
            originalError: error,
            canRetry: this.canRetryByStatus(status),
            details: data.message || data.error
          }
        }
      }

      // 根据HTTP状态码处理错误
      return this.handleHttpStatusError(status, error)
    }

    // 请求发送了但没收到响应
    if (request) {
      return {
        type: ErrorType.NETWORK_ERROR,
        message: '网络请求失败，请检查网络连接',
        originalError: error,
        canRetry: true,
        details: '请求已发送但未收到服务器响应'
      }
    }

    // 请求配置错误
    return {
      type: ErrorType.CLIENT_ERROR,
      message: '请求配置错误',
      originalError: error,
      canRetry: false,
      details: error.message
    }
  }

  /**
   * 根据HTTP状态码处理错误
   */
  private static handleHttpStatusError(status: number, error: AxiosError): ProcessedError {
    const statusErrorMap: Record<number, Partial<ProcessedError>> = {
      400: {
        type: ErrorType.VALIDATION_ERROR,
        message: '请求参数错误，请检查输入信息',
        canRetry: false
      },
      401: {
        type: ErrorType.CLIENT_ERROR,
        message: '未授权访问，请重新登录',
        canRetry: false
      },
      403: {
        type: ErrorType.CLIENT_ERROR,
        message: '没有访问权限',
        canRetry: false
      },
      404: {
        type: ErrorType.CLIENT_ERROR,
        message: '请求的资源不存在',
        canRetry: false
      },
      408: {
        type: ErrorType.TIMEOUT_ERROR,
        message: '请求超时，请稍后重试',
        canRetry: true
      },
      429: {
        type: ErrorType.CLIENT_ERROR,
        message: '请求过于频繁，请稍后重试',
        canRetry: true
      },
      500: {
        type: ErrorType.SERVER_ERROR,
        message: '服务器内部错误，请稍后重试',
        canRetry: true
      },
      502: {
        type: ErrorType.SERVER_ERROR,
        message: '网关错误，请稍后重试',
        canRetry: true
      },
      503: {
        type: ErrorType.SERVER_ERROR,
        message: '服务暂时不可用，请稍后重试',
        canRetry: true
      },
      504: {
        type: ErrorType.SERVER_ERROR,
        message: '网关超时，请稍后重试',
        canRetry: true
      }
    }

    const errorInfo = statusErrorMap[status]
    
    if (errorInfo) {
      return {
        ...errorInfo,
        originalError: error,
        code: status,
        details: `HTTP ${status} 错误`
      } as ProcessedError
    }

    // 其他状态码
    if (status >= 500) {
      return {
        type: ErrorType.SERVER_ERROR,
        message: '服务器错误，请稍后重试',
        originalError: error,
        code: status,
        canRetry: true,
        details: `HTTP ${status} 服务器错误`
      }
    }

    return {
      type: ErrorType.CLIENT_ERROR,
      message: '请求失败，请检查请求参数',
      originalError: error,
      code: status,
      canRetry: false,
      details: `HTTP ${status} 客户端错误`
    }
  }

  /**
   * 转换为中文错误消息
   */
  private static getChineseErrorMessage(message: string, status?: number): string {
    // 常见英文错误消息的中文映射
    const messageMap: Record<string, string> = {
      'Network Error': '网络连接失败',
      'Request failed': '请求失败',
      'Invalid mnemonic phrase': '无效的助记词',
      'Invalid address': '无效的地址',
      'Insufficient balance': '余额不足',
      'Transaction failed': '交易失败',
      'Wallet not found': '钱包未找到',
      'Invalid parameters': '参数无效',
      'Internal server error': '服务器内部错误',
      'Bad request': '请求参数错误',
      'Unauthorized': '未授权访问',
      'Forbidden': '禁止访问',
      'Not found': '资源未找到',
      'Method not allowed': '请求方法不允许',
      'Conflict': '资源冲突',
      'Too many requests': '请求过于频繁'
    }

    // 尝试直接映射
    const mapped = messageMap[message]
    if (mapped) {
      return mapped
    }

    // 如果消息已经是中文，直接返回
    if (/[\u4e00-\u9fa5]/.test(message)) {
      return message
    }

    // 根据状态码提供通用消息
    if (status) {
      if (status >= 500) {
        return '服务器繁忙，请稍后重试'
      } else if (status >= 400) {
        return '请求处理失败，请检查输入信息'
      }
    }

    // 兜底返回原消息或通用错误
    return message || '操作失败，请重试'
  }

  /**
   * 判断错误是否可以重试
   */
  private static canRetryByStatus(status: number): boolean {
    // 5xx错误和某些4xx错误可以重试
    return status >= 500 || status === 408 || status === 429
  }

  /**
   * 格式化错误消息用于显示给用户
   */
  static formatErrorForUser(error: ProcessedError): string {
    return error.message
  }

  /**
   * 格式化错误消息用于开发调试
   */
  static formatErrorForDev(error: ProcessedError): string {
    return `${error.message}${error.details ? ` (${error.details})` : ''}`
  }

  /**
   * 判断是否应该显示重试按钮
   */
  static shouldShowRetry(error: ProcessedError): boolean {
    return error.canRetry === true
  }
}

/**
 * 便捷的错误处理函数
 * 用于在组件中快速处理API错误
 */
export function handleApiError(error: any): ProcessedError {
  return NetworkErrorHandler.handleError(error)
}

export default NetworkErrorHandler