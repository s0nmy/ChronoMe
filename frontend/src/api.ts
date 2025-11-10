export interface Tag {
  id: string
  name: string
  color: string
}

export interface ReportDay {
  date: string
  total_seconds: number
}

export interface ProjectBreakdown {
  project_id: string | null
  name: string
  color: string
  total_seconds: number
}

export interface TagBreakdown {
  tag_id: string
  name: string
  color: string
  total_seconds: number
}

export interface WeeklyReport {
  week_start: string
  total_seconds: number
  days: ReportDay[]
  projects: ProjectBreakdown[]
  tags: TagBreakdown[]
}

export interface MonthlyReport {
  month: string
  total_seconds: number
  days_in_month: number
  days: ReportDay[]
  weeks: { week_start: string; total_seconds: number }[]
  projects: ProjectBreakdown[]
  tags: TagBreakdown[]
}

const CSRF_COOKIE = 'chronome_csrf'
const CSRF_HEADER = 'X-CSRF-Token'
const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS'])

function getCSRFCookie(): string | null {
  if (typeof document === 'undefined') {
    return null
  }
  const match = document.cookie.match(
    new RegExp(`(?:^|; )${CSRF_COOKIE.replace(/[-[\]/{}()*+?.\\^$|]/g, '\\$&')}=([^;]*)`),
  )
  return match ? decodeURIComponent(match[1]) : null
}

function needsCsrfHeader(method?: string) {
  const normalized = (method ?? 'GET').toUpperCase()
  return !SAFE_METHODS.has(normalized)
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  const method = init?.method ?? 'GET'
  if (needsCsrfHeader(method)) {
    const token = getCSRFCookie()
    if (token) {
      headers.set(CSRF_HEADER, token)
    }
  }
  const res = await fetch(path, {
    credentials: 'include',
    ...init,
    headers,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  if (res.status === 204) {
    return {} as T
  }
  return res.json()
}

export async function fetchTags(): Promise<Tag[]> {
  const data = await request<{ tags: Tag[] }>('/api/tags')
  return data.tags ?? []
}

export async function fetchWeeklyReport(weekStart?: string): Promise<WeeklyReport> {
  const params = weekStart ? `?week_start=${encodeURIComponent(weekStart)}` : ''
  return request<WeeklyReport>(`/api/reports/weekly${params}`)
}

export async function fetchMonthlyReport(month: string): Promise<MonthlyReport> {
  const params = `?month=${encodeURIComponent(month)}`
  return request<MonthlyReport>(`/api/reports/monthly${params}`)
}
