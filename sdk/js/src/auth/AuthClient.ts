import type { AuthResponse, OmniBaseError, Session, User } from '../types'

interface AuthClientOptions {
  url: string
  headers: Record<string, string>
  storageType: 'localStorage' | 'sessionStorage' | 'memory' | Storage
  autoRefreshToken: boolean
  persistSession: boolean
}

type AuthStateChangeCallback = (event: AuthChangeEvent, session: Session | null) => void

export type AuthChangeEvent =
  | 'SIGNED_IN'
  | 'SIGNED_OUT'
  | 'TOKEN_REFRESHED'
  | 'USER_UPDATED'
  | 'PASSWORD_RECOVERY'

/**
 * AuthClient handles all authentication operations.
 *
 * @example
 * ```typescript
 * // Sign up
 * const { data, error } = await omni.auth.signUp({ email, password })
 *
 * // Sign in
 * const { data, error } = await omni.auth.signInWithPassword({ email, password })
 *
 * // Get current user
 * const { data: { user } } = await omni.auth.getUser()
 *
 * // Subscribe to auth state changes
 * omni.auth.onAuthStateChange((event, session) => {
 *   console.log(event, session)
 * })
 * ```
 */
export class AuthClient {
  private url: string
  private headers: Record<string, string>
  private options: AuthClientOptions
  private _currentSession: Session | null = null
  private _listeners: Set<AuthStateChangeCallback> = new Set()
  private _refreshTimer: ReturnType<typeof setTimeout> | null = null

  constructor(options: AuthClientOptions) {
    this.url = options.url
    this.headers = options.headers
    this.options = options
    this._loadSession()
  }

  // ─── Core Auth Methods ───────────────────────────────────────────────────────

  /**
   * Creates a new user account with email and password.
   */
  async signUp(credentials: {
    email: string
    password: string
    options?: { data?: Record<string, unknown>; redirectTo?: string }
  }): Promise<AuthResponse> {
    const { data, error } = await this._request<{
      user: User
      access_token?: string
      refresh_token?: string
      expires_in?: number
    }>('POST', '/signup', {
      email: credentials.email,
      password: credentials.password,
      options: credentials.options,
    })

    if (error) return { data: { user: null, session: null }, error }

    let session: Session | null = null
    if (data?.access_token) {
      session = {
        access_token: data.access_token,
        refresh_token: data.refresh_token!,
        token_type: 'bearer',
        expires_in: data.expires_in!,
        expires_at: Math.floor(Date.now() / 1000) + data.expires_in!,
        user: data.user!,
      }
      this._setSession(session)
    }

    return { data: { user: data?.user ?? null, session }, error: null }
  }

  /**
   * Signs in with email and password.
   */
  async signInWithPassword(credentials: {
    email: string
    password: string
  }): Promise<AuthResponse> {
    const { data, error } = await this._request<{
      user: User
      access_token: string
      refresh_token: string
      expires_in: number
    }>('POST', '/token', {
      grant_type: 'password',
      email: credentials.email,
      password: credentials.password,
    })

    if (error) return { data: { user: null, session: null }, error }

    const session: Session = {
      access_token: data!.access_token,
      refresh_token: data!.refresh_token,
      token_type: 'bearer',
      expires_in: data!.expires_in,
      expires_at: Math.floor(Date.now() / 1000) + data!.expires_in,
      user: data!.user,
    }
    this._setSession(session)
    this._notifyListeners('SIGNED_IN', session)

    return { data: { user: data!.user, session }, error: null }
  }

  /**
   * Signs out the current user — invalidates the session.
   */
  async signOut(): Promise<{ error: OmniBaseError | null }> {
    if (this._currentSession) {
      await this._request('POST', '/logout', {})
    }
    this._clearSession()
    this._notifyListeners('SIGNED_OUT', null)
    return { error: null }
  }

  /**
   * Returns the current user.
   */
  async getUser(): Promise<{ data: { user: User | null }; error: OmniBaseError | null }> {
    const { data, error } = await this._request<User>('GET', '/user')
    return { data: { user: data ?? null }, error }
  }

  /**
   * Returns the current session.
   */
  async getSession(): Promise<{ data: { session: Session | null }; error: OmniBaseError | null }> {
    if (this._currentSession && this._isSessionExpired(this._currentSession)) {
      const refreshed = await this._refreshSession(this._currentSession.refresh_token)
      return { data: { session: refreshed }, error: null }
    }
    return { data: { session: this._currentSession }, error: null }
  }

  /**
   * Refreshes the access token using the refresh token.
   */
  async refreshSession(refreshToken?: string): Promise<AuthResponse> {
    const token = refreshToken ?? this._currentSession?.refresh_token
    if (!token) {
      return {
        data: { user: null, session: null },
        error: { code: 'no_refresh_token', message: 'No refresh token available' },
      }
    }

    const { data, error } = await this._request<{
      user: User
      access_token: string
      refresh_token: string
      expires_in: number
    }>('POST', '/token', { grant_type: 'refresh_token', refresh_token: token })

    if (error) return { data: { user: null, session: null }, error }

    const session: Session = {
      access_token: data!.access_token,
      refresh_token: data!.refresh_token,
      token_type: 'bearer',
      expires_in: data!.expires_in,
      expires_at: Math.floor(Date.now() / 1000) + data!.expires_in,
      user: data!.user,
    }
    this._setSession(session)
    this._notifyListeners('TOKEN_REFRESHED', session)

    return { data: { user: data!.user, session }, error: null }
  }

  /**
   * Sends a password reset email.
   */
  async resetPasswordForEmail(
    email: string,
    options?: { redirectTo?: string }
  ): Promise<{ error: OmniBaseError | null }> {
    const { error } = await this._request('POST', '/recover', { email, ...options })
    return { error }
  }

  /**
   * Subscribe to authentication state changes.
   * Returns a subscription object with an unsubscribe method.
   *
   * @example
   * ```typescript
   * const { data: { subscription } } = omni.auth.onAuthStateChange((event, session) => {
   *   if (event === 'SIGNED_IN') console.log('User signed in:', session?.user)
   *   if (event === 'SIGNED_OUT') console.log('User signed out')
   * })
   *
   * // Cleanup
   * subscription.unsubscribe()
   * ```
   */
  onAuthStateChange(callback: AuthStateChangeCallback): {
    data: { subscription: { unsubscribe: () => void } }
  } {
    this._listeners.add(callback)

    // Immediately call with current session
    if (this._currentSession) {
      setTimeout(() => callback('SIGNED_IN', this._currentSession), 0)
    }

    return {
      data: {
        subscription: {
          unsubscribe: () => this._listeners.delete(callback),
        },
      },
    }
  }

  /**
   * Get the current access token (for manual API calls)
   */
  getAccessToken(): string | null {
    return this._currentSession?.access_token ?? null
  }

  // ─── Private Methods ─────────────────────────────────────────────────────────

  private async _request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<{ data: T | null; error: OmniBaseError | null }> {
    const headers = { ...this.headers }
    if (this._currentSession?.access_token) {
      headers['Authorization'] = `Bearer ${this._currentSession.access_token}`
    }

    try {
      const response = await fetch(this.url + path, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      })

      const json = await response.json().catch(() => null)

      if (!response.ok) {
        return {
          data: null,
          error: {
            code: json?.code ?? 'request_failed',
            message: json?.message ?? `HTTP ${response.status}`,
            status: response.status,
          },
        }
      }

      return { data: json as T, error: null }
    } catch (err) {
      return {
        data: null,
        error: {
          code: 'network_error',
          message: err instanceof Error ? err.message : 'Network request failed',
        },
      }
    }
  }

  private async _refreshSession(refreshToken: string): Promise<Session | null> {
    const { data } = await this.refreshSession(refreshToken)
    return data?.session ?? null
  }

  private _setSession(session: Session) {
    this._currentSession = session
    this._persistSession(session)
    this._scheduleTokenRefresh(session)
  }

  private _clearSession() {
    this._currentSession = null
    if (this._refreshTimer) clearTimeout(this._refreshTimer)
    try {
      if (typeof localStorage !== 'undefined') {
        localStorage.removeItem('omnibase.session')
      }
    } catch {}
  }

  private _loadSession() {
    try {
      if (typeof localStorage !== 'undefined') {
        const stored = localStorage.getItem('omnibase.session')
        if (stored) {
          const session: Session = JSON.parse(stored)
          if (!this._isSessionExpired(session)) {
            this._currentSession = session
            this._scheduleTokenRefresh(session)
          }
        }
      }
    } catch {}
  }

  private _persistSession(session: Session) {
    if (!this.options.persistSession) return
    try {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('omnibase.session', JSON.stringify(session))
      }
    } catch {}
  }

  private _isSessionExpired(session: Session): boolean {
    if (!session.expires_at) return false
    const expiresAt = session.expires_at * 1000
    const bufferMs = 60_000 // Refresh 60 seconds before expiry
    return Date.now() > expiresAt - bufferMs
  }

  private _scheduleTokenRefresh(session: Session) {
    if (!this.options.autoRefreshToken || !session.expires_at) return
    if (this._refreshTimer) clearTimeout(this._refreshTimer)

    const expiresAt = session.expires_at * 1000
    const refreshIn = expiresAt - Date.now() - 60_000 // 60s before expiry

    if (refreshIn > 0) {
      this._refreshTimer = setTimeout(async () => {
        await this.refreshSession(session.refresh_token)
      }, refreshIn)
    }
  }

  private _notifyListeners(event: AuthChangeEvent, session: Session | null) {
    this._listeners.forEach((cb) => cb(event, session))
  }
}
