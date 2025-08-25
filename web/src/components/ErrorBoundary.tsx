/**
 * React错误边界组件
 * 
 * 用于捕获React组件树中的JavaScript错误，记录错误并显示备用UI
 * 这是一个重要的前端错误处理模式，可以防止整个应用崩溃
 * 
 * 前端学习要点：
 * 1. 类组件 - 错误边界必须是类组件，不能用函数组件
 * 2. 错误边界生命周期 - getDerivedStateFromError 和 componentDidCatch
 * 3. 错误恢复 - 提供重试机制让用户恢复应用状态
 */
import React, { Component, ErrorInfo, ReactNode } from 'react'

// 错误边界组件的Props类型
interface Props {
  children: ReactNode
  fallback?: ReactNode  // 可选的自定义错误显示组件
}

// 错误边界组件的State类型
interface State {
  hasError: boolean
  error?: Error
  errorInfo?: ErrorInfo
}

/**
 * 错误边界类组件
 * 
 * 当子组件抛出错误时，这个组件会：
 * 1. 捕获错误并更新state
 * 2. 显示错误信息而不是让整个应用崩溃
 * 3. 提供重试功能
 */
class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false
  }

  /**
   * 静态方法：从错误中派生新的state
   * 当子组件抛出错误时被调用
   * 
   * @param error 捕获到的错误对象
   * @returns 新的state对象
   */
  public static getDerivedStateFromError(error: Error): State {
    // 更新state以显示错误UI
    return {
      hasError: true,
      error
    }
  }

  /**
   * 组件捕获错误时的生命周期方法
   * 用于记录错误信息，通常用于错误报告
   * 
   * @param error 错误对象
   * @param errorInfo 错误的组件堆栈信息
   */
  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // 记录错误到控制台
    console.error('🚨 React Error Boundary caught an error:', {
      error,
      errorInfo,
      componentStack: errorInfo.componentStack
    })

    // 在生产环境中，这里可以发送错误到日志服务
    // 例如：sendErrorToLoggingService(error, errorInfo)

    // 更新state以包含详细的错误信息
    this.setState({
      error,
      errorInfo
    })
  }

  /**
   * 重置错误状态的方法
   * 让用户能够重试并恢复应用
   */
  private handleReset = () => {
    console.log('🔄 重置错误边界状态')
    this.setState({
      hasError: false,
      error: undefined,
      errorInfo: undefined
    })
  }

  /**
   * 渲染方法
   * 根据是否有错误来决定渲染什么内容
   */
  public render() {
    // 如果有错误，显示错误UI
    if (this.state.hasError) {
      // 如果提供了自定义的fallback组件，使用它
      if (this.props.fallback) {
        return this.props.fallback
      }

      // 否则显示默认的错误UI
      return (
        <div style={{
          padding: '20px',
          margin: '20px',
          border: '2px solid #ff6b6b',
          borderRadius: '8px',
          backgroundColor: '#fff5f5',
          color: '#c92a2a',
          fontFamily: 'Arial, sans-serif'
        }}>
          <h2 style={{ margin: '0 0 16px 0' }}>
            🚨 应用出现错误
          </h2>
          
          <p style={{ margin: '0 0 16px 0' }}>
            很抱歉，应用遇到了一个意外错误。请尝试刷新页面或点击下面的重试按钮。
          </p>

          {/* 开发环境显示详细错误信息 */}
          {process.env.NODE_ENV === 'development' && this.state.error && (
            <details style={{ 
              margin: '16px 0',
              padding: '12px',
              backgroundColor: '#fff',
              border: '1px solid #ddd',
              borderRadius: '4px'
            }}>
              <summary style={{ cursor: 'pointer', fontWeight: 'bold' }}>
                查看错误详情
              </summary>
              <pre style={{ 
                margin: '8px 0 0 0',
                fontSize: '12px',
                overflow: 'auto'
              }}>
                {this.state.error.toString()}
                {this.state.errorInfo?.componentStack}
              </pre>
            </details>
          )}

          <div style={{ display: 'flex', gap: '12px', marginTop: '16px' }}>
            <button
              onClick={this.handleReset}
              style={{
                padding: '8px 16px',
                backgroundColor: '#4dabf7',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer'
              }}
            >
              🔄 重试
            </button>
            
            <button
              onClick={() => window.location.reload()}
              style={{
                padding: '8px 16px',
                backgroundColor: '#51cf66',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer'
              }}
            >
              🔃 刷新页面
            </button>
          </div>
        </div>
      )
    }

    // 如果没有错误，正常渲染子组件
    return this.props.children
  }
}

export default ErrorBoundary