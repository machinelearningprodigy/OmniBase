export function getOmniBaseUrl(): string {
  if (typeof window !== 'undefined') {
    const fromStorage = localStorage.getItem('omnibase.api_url')
    if (fromStorage) return fromStorage
    if ((window as any).__omnibase_url__) return (window as any).__omnibase_url__ as string
  }
  return (typeof import.meta !== 'undefined' && (import.meta as any).env?.PUBLIC_OMNIBASE_URL) || 'http://localhost:8000'
}

export function getSession() {
  if (typeof localStorage === 'undefined') {
    return null
  }

  try {
    const raw = localStorage.getItem('omnibase.session')
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

export function getAccessToken() {
  return getSession()?.access_token ?? null
}

export function getHeaders(includeContentType = true) {
  const token = getAccessToken()
  let projectId = ''
  if (typeof localStorage !== 'undefined') {
    const activeProject = localStorage.getItem('omnibase.active_project')
    if (activeProject) {
      try {
        projectId = JSON.parse(activeProject).id
      } catch {}
    }
  }

  return {
    ...(includeContentType ? { 'Content-Type': 'application/json' } : {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(projectId ? { 'X-OmniBase-Project-ID': projectId } : {}),
  }
}

export function getRealtimeUrl() {
  const httpUrl = getOmniBaseUrl()
  const wsBase = httpUrl.startsWith('https://')
    ? httpUrl.replace('https://', 'wss://')
    : httpUrl.replace('http://', 'ws://')
  const token = getAccessToken()
  const suffix = token ? `?apikey=${encodeURIComponent(token)}` : ''
  return `${wsBase}/realtime/v1/websocket${suffix}`
}

async function refreshStoredSession() {
  if (typeof localStorage === 'undefined') {
    return null
  }

  const session = getSession()
  if (!session?.refresh_token) {
    return null
  }

  const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      grant_type: 'refresh_token',
      refresh_token: session.refresh_token,
    }),
  })

  if (!resp.ok) {
    localStorage.removeItem('omnibase.session')
    return null
  }

  const data = await resp.json()
  const nextSession = {
    access_token: data.access_token,
    refresh_token: data.refresh_token ?? session.refresh_token,
  }

  localStorage.setItem('omnibase.session', JSON.stringify(nextSession))
  window.dispatchEvent(new CustomEvent('omnibase:session-updated', { detail: nextSession }))
  return nextSession.access_token as string
}

export async function apiFetch(input: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers ?? {})
  const defaultHeaders = getHeaders(false)
  Object.entries(defaultHeaders).forEach(([k, v]) => {
    if (v && !headers.has(k)) {
      headers.set(k, v)
    }
  })

  let response = await fetch(input, { ...init, headers })
  if (response.status !== 401) {
    return response
  }

  const refreshedToken = await refreshStoredSession()
  if (!refreshedToken) {
    return response
  }

  const retryHeaders = new Headers(init.headers ?? {})
  retryHeaders.set('Authorization', `Bearer ${refreshedToken}`)

  response = await fetch(input, { ...init, headers: retryHeaders })
  return response
}

export async function getErrorMessage(response: Response, fallback: string) {
  if (response.status === 401) {
    return 'Your session is missing or expired. Sign in again.'
  }

  try {
    const text = await response.text()
    if (!text) return fallback
    try {
      const data = JSON.parse(text)
      return data.error || data.message || data.details || fallback
    } catch {
      // If JSON parsing fails, return the raw text if it's not too long, otherwise the fallback with status.
      return text.length < 100 ? text : `${fallback} (HTTP ${response.status})`
    }
  } catch {
    return fallback
  }
}
