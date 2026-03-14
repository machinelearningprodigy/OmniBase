import { writable } from 'svelte/store'
import { getOmniBaseUrl } from '$lib/api'

interface AuthState {
  user: { id: string; email: string; role: string } | null
  session: { access_token: string; refresh_token: string } | null
  loading: boolean
}

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>({
    user: null,
    session: null,
    loading: true,
  })

  async function refreshSession() {
    const stored = localStorage.getItem('omnibase.session')
    if (!stored) return null
    const session = JSON.parse(stored)
    if (!session.refresh_token) return null
    const OMNIBASE_URL = getOmniBaseUrl()
    const resp = await fetch(`${OMNIBASE_URL}/auth/v1/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ grant_type: 'refresh_token', refresh_token: session.refresh_token })
    })
    const data = await resp.json()
    if (!resp.ok) return null
    const newSession = {
      access_token: data.access_token,
      refresh_token: data.refresh_token ?? session.refresh_token
    }
    localStorage.setItem('omnibase.session', JSON.stringify(newSession))
    update((s) => ({ ...s, session: newSession, user: data.user ?? s.user }))
    return newSession.access_token
  }

  async function init() {
    try {
      const OMNIBASE_URL = getOmniBaseUrl()
      const stored = localStorage.getItem('omnibase.session')
      if (stored) {
        const session = JSON.parse(stored)
        if (session.access_token) {
          const resp = await fetch(`${OMNIBASE_URL}/auth/v1/user`, {
            headers: { Authorization: `Bearer ${session.access_token}` }
          })
          if (resp.ok) {
            const user = await resp.json()
            set({ user, session, loading: false })
            return
          }
          if (resp.status === 401 && session.refresh_token) {
            const newToken = await refreshSession()
            if (newToken) {
              const retry = await fetch(`${OMNIBASE_URL}/auth/v1/user`, {
                headers: { Authorization: `Bearer ${newToken}` }
              })
              if (retry.ok) {
                const user = await retry.json()
                const newStored = localStorage.getItem('omnibase.session')
                set({ user, session: newStored ? JSON.parse(newStored) : null, loading: false })
                return
              }
            }
          }
        }
      }
    } catch {}
    set({ user: null, session: null, loading: false })
  }

  async function signIn(email: string, password: string) {
    const OMNIBASE_URL = getOmniBaseUrl()
    const resp = await fetch(`${OMNIBASE_URL}/auth/v1/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ grant_type: 'password', email, password })
    })
    const data = await resp.json()
    if (resp.ok) {
      localStorage.setItem('omnibase.session', JSON.stringify({
        access_token: data.access_token,
        refresh_token: data.refresh_token
      }))
      update(s => ({ ...s, user: data.user, session: data }))
      return { error: null }
    }
    return { error: data }
  }

  async function signOut() {
    localStorage.removeItem('omnibase.session')
    set({ user: null, session: null, loading: false })
  }

  async function setSessionFromHash(accessToken: string, refreshToken: string) {
    const session = { access_token: accessToken, refresh_token: refreshToken }
    localStorage.setItem('omnibase.session', JSON.stringify(session))
    const OMNIBASE_URL = getOmniBaseUrl()
    try {
      const resp = await fetch(`${OMNIBASE_URL}/auth/v1/user`, {
        headers: { Authorization: `Bearer ${accessToken}` }
      })
      const user = resp.ok ? await resp.json() : null
      set({ user, session, loading: false })
    } catch {
      set({ user: null, session, loading: false })
    }
  }

  return { subscribe, init, signIn, signOut, setSessionFromHash, refreshSession }
}

export const authStore = createAuthStore()
