import { writable } from 'svelte/store'
import { getOmniBaseUrl } from '$lib/api'

interface AuthState {
  user: { id: string; email: string; role: string; is_super_admin?: boolean } | null
  session: { access_token: string; refresh_token: string } | null
  loading: boolean
  activeProject: { id: string; name: string; anon_key: string; service_key: string } | null
}

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>({
    user: null,
    session: null,
    loading: true,
    activeProject: null,
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

  async function loadActiveProject(accessToken: string) {
    const OMNIBASE_URL = getOmniBaseUrl()
    try {
      // Check localStorage for last active project
      const storedProject = localStorage.getItem('omnibase.active_project')
      if (storedProject) {
        const proj = JSON.parse(storedProject)
        update(s => ({ ...s, activeProject: proj }))
        return proj
      }

      // Otherwise fetch first project from API
      const resp = await fetch(`${OMNIBASE_URL}/admin/v1/projects`, {
        headers: { Authorization: `Bearer ${accessToken}` }
      })
      if (resp.ok) {
        const projects = await resp.json()
        if (projects && projects.length > 0) {
          const proj = projects[0]
          const activeProject = {
            id: proj.id,
            name: proj.name,
            anon_key: proj.anon_key,
            service_key: proj.service_key,
          }
          localStorage.setItem('omnibase.active_project', JSON.stringify(activeProject))
          update(s => ({ ...s, activeProject }))
          return activeProject
        }
      }
    } catch {}
    return null
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
            set({ user, session, loading: false, activeProject: null })
            await loadActiveProject(session.access_token)
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
                set({ user, session: newStored ? JSON.parse(newStored) : null, loading: false, activeProject: null })
                await loadActiveProject(newToken)
                return
              }
            }
          }
        }
      }
    } catch {}
    set({ user: null, session: null, loading: false, activeProject: null })
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
      update(s => ({ ...s, user: data.user, session: data, activeProject: null }))
      // Load active project async
      if (data.access_token) {
        loadActiveProject(data.access_token)
      }
      return { error: null }
    }
    return { error: data }
  }

  async function signOut() {
    localStorage.removeItem('omnibase.session')
    localStorage.removeItem('omnibase.active_project')
    set({ user: null, session: null, loading: false, activeProject: null })
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
      set({ user, session, loading: false, activeProject: null })
      if (accessToken) {
        await loadActiveProject(accessToken)
      }
    } catch {
      set({ user: null, session, loading: false, activeProject: null })
    }
  }

  function setActiveProject(project: { id: string; name: string; anon_key: string; service_key: string }) {
    localStorage.setItem('omnibase.active_project', JSON.stringify(project))
    update(s => ({ ...s, activeProject: project }))
  }

  return { subscribe, init, signIn, signOut, setSessionFromHash, refreshSession, setActiveProject, loadActiveProject }
}

export const authStore = createAuthStore()
