import { AuthClient } from './auth/AuthClient'
import { DatabaseClient } from './database/DatabaseClient'
import { StorageClient } from './storage/StorageClient'
import { RealtimeClient } from './realtime/RealtimeClient'

export interface OmniBaseClientOptions {
  /** The URL of your OmniBase project. Example: http://localhost:8000 */
  url: string
  /** Your project's anonymous key (safe to expose in client-side code) */
  anonKey: string
  /**
   * Global headers sent with every request.
   * Useful for X-OmniBase-Key or custom headers.
   */
  headers?: Record<string, string>
  /**
   * Authentication storage. Defaults to localStorage in browser, memory in Node.js.
   * @default 'localStorage'
   */
  auth?: {
    storage?: 'localStorage' | 'sessionStorage' | 'memory' | Storage
    autoRefreshToken?: boolean
    persistSession?: boolean
    detectSessionInUrl?: boolean
  }
}

/**
 * OmniBaseClient is the main entry point for all OmniBase services.
 *
 * @example
 * ```typescript
 * import { createClient } from '@omnibase/omnibase-js'
 *
 * const omni = createClient('http://localhost:8000', 'your-anon-key')
 *
 * // Auth
 * const { user, error } = await omni.auth.signUp({ email, password })
 *
 * // Database — auto-generated REST API with full TypeScript types
 * const { data, error } = await omni.from('posts').select('*').eq('user_id', user.id)
 *
 * // Storage
 * const { data, error } = await omni.storage.from('avatars').upload('me.jpg', file)
 *
 * // Realtime — reactive by default
 * omni.channel('posts').on('INSERT', (payload) => console.log(payload)).subscribe()
 * ```
 */
export class OmniBaseClient {
  readonly auth: AuthClient
  readonly storage: StorageClient
  readonly realtime: RealtimeClient

  private _db: DatabaseClient
  private options: OmniBaseClientOptions

  constructor(options: OmniBaseClientOptions) {
    this.options = options

    const baseHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-OmniBase-Key': options.anonKey,
      ...(options.headers ?? {}),
    }

    this.auth = new AuthClient({
      url: `${options.url}/auth/v1`,
      headers: baseHeaders,
      storageType: options.auth?.storage ?? 'localStorage',
      autoRefreshToken: options.auth?.autoRefreshToken ?? true,
      persistSession: options.auth?.persistSession ?? true,
    })

    this._db = new DatabaseClient({
      url: `${options.url}/rest/v1`,
      headers: baseHeaders,
      auth: this.auth,
    })

    this.storage = new StorageClient({
      url: `${options.url}/storage/v1`,
      headers: baseHeaders,
      auth: this.auth,
    })

    this.realtime = new RealtimeClient({
      url: options.url.replace('http', 'ws') + '/realtime/v1/websocket',
      headers: baseHeaders,
      auth: this.auth,
    })
  }

  /**
   * Start a database query from a table.
   * Returns a QueryBuilder for chaining filter, select, order, limit operations.
   */
  from<T = Record<string, unknown>>(table: string) {
    return this._db.from<T>(table)
  }

  /**
   * Get a realtime channel for subscriptions, presence, and broadcast.
   */
  channel(name: string) {
    return this.realtime.channel(name)
  }

  /**
   * Remove all realtime subscriptions
   */
  removeAllChannels() {
    return this.realtime.removeAllChannels()
  }
}

/**
 * Create an OmniBase client.
 *
 * @param url - Your OmniBase project URL
 * @param anonKey - Your project's anonymous key
 * @param options - Additional options
 */
export function createClient(url: string, anonKey: string, options?: Partial<OmniBaseClientOptions>): OmniBaseClient {
  return new OmniBaseClient({ url, anonKey, ...options })
}
