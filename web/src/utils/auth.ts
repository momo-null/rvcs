// JWT token 解析工具

export interface UserInfo {
  id: string
  username: string
  email: string
  role: string
  avatar?: string
}

export interface TokenPayload {
  user_id: string
  username?: string
  role?: string
  exp: number
  iat: number
}

const TOKEN_KEY = 'access_token'
const USER_KEY = 'user_info'

// 解析 token
export function parseJwt(token: string): TokenPayload | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload) as TokenPayload
  } catch (error) {
    console.error('Error log', error)
    return null
  }
}

// 检查 token 是否过期
export function isTokenExpired(token: string): boolean {
  const payload = parseJwt(token)
  if (!payload) return true
  
  const currentTime = Math.floor(Date.now() / 1000)
  return payload.exp < currentTime
}

// 获取 token
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

// 保存 token
export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

// 删除 token
export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

// 获取用户信息
export function getUserInfo(): UserInfo | null {
  const userInfoStr = localStorage.getItem(USER_KEY)
  if (!userInfoStr) return null
  
  try {
    return JSON.parse(userInfoStr) as UserInfo
  } catch (error) {
    console.error('Error log', error)
    return null
  }
}

// 保存用户信息
export function setUserInfo(userInfo: UserInfo): void {
  localStorage.setItem(USER_KEY, JSON.stringify(userInfo))
}

// 删除用户信息
export function removeUserInfo(): void {
  localStorage.removeItem(USER_KEY)
}

// 检查是否已认证
export function isAuthenticated(): boolean {
  const token = getToken()
  if (!token) return false
  
  return !isTokenExpired(token)
}

// 清除认证信息
export function clearAuth(): void {
  removeToken()
  removeUserInfo()
}

// 获取认证请求头
export function getAuthHeader(): Record<string, string> {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}
