import { writable } from 'svelte/store'

// Environment config from SvelteKit public env
const OMNIBASE_URL = typeof window !== 'undefined'
  ? (window as any).__omnibase_url__ || 'http://localhost:8000'
  : 'http://localhost:8000'

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

  async function init() {
    try {
      const stored = localStorage.getItem('omnibase.session')
      if (stored) {
        const session = JSON.parse(stored)
        if (session.access_token) {
          // Verify token with auth service
          const resp = await fetch(`${OMNIBASE_URL}/auth/v1/user`, {
            headers: { Authorization: `Bearer ${session.access_token}` }
          })
          if (resp.ok) {
            const user = await resp.json()
            set({ user, session, loading: false })
            return
          }
        }
      }
    } catch {}
    set({ user: null, session: null, loading: false })
  }

  async function signIn(email: string, password: string) {
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

  return { subscribe, init, signIn, signOut }
}

export const authStore = createAuthStore()
