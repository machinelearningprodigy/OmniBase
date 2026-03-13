// ─── Core Types ───────────────────────────────────────────────────────────────

export interface User {
  id: string
  email?: string
  phone?: string
  role: 'anon' | 'authenticated' | 'service_role'
  user_metadata: Record<string, unknown>
  app_metadata: Record<string, unknown>
  email_confirmed_at?: string
  phone_confirmed_at?: string
  last_sign_in_at?: string
  created_at: string
  updated_at: string
}

export interface Session {
  access_token: string
  refresh_token: string
  token_type: 'bearer'
  expires_in: number
  expires_at?: number
  user: User
}

export interface AuthResponse {
  data: {
    user: User | null
    session: Session | null
  }
  error: OmniBaseError | null
}

export interface OmniBaseError {
  code: string
  message: string
  details?: unknown
  status?: number
}

// ─── Database Types ────────────────────────────────────────────────────────────

export type FilterOperator =
  | 'eq'
  | 'neq'
  | 'gt'
  | 'gte'
  | 'lt'
  | 'lte'
  | 'like'
  | 'ilike'
  | 'is'
  | 'in'
  | 'cs'      // contains
  | 'cd'      // contained by
  | 'sl'      // strictly left
  | 'sr'      // strictly right
  | 'nxl'     // does not extend left
  | 'nxr'     // does not extend right

export type OrderDirection = 'asc' | 'desc'
export type NullsOrder = 'first' | 'last'

export interface PostgrestResponse<T> {
  data: T | null
  error: OmniBaseError | null
  count?: number
  status: number
  statusText: string
}

export interface PostgrestSingleResponse<T> {
  data: T
  error: OmniBaseError | null
  status: number
  statusText: string
}

// ─── Storage Types ─────────────────────────────────────────────────────────────

export interface StorageBucket {
  id: string
  name: string
  public: boolean
  allowed_mime_types?: string[]
  file_size_limit?: number
  created_at: string
  updated_at: string
}

export interface StorageObject {
  id: string
  bucket_id: string
  name: string
  owner?: string
  content_type?: string
  size: number
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface StorageResponse {
  data: StorageObject | null
  error: OmniBaseError | null
}

export interface SignedURL {
  signedUrl: string
  token: string
  path: string
}

// ─── Realtime Types ────────────────────────────────────────────────────────────

export type RealtimeEventType = 'INSERT' | 'UPDATE' | 'DELETE' | '*'

export interface RealtimePayload<T = Record<string, unknown>> {
  schema: string
  table: string
  eventType: RealtimeEventType
  old: Partial<T>
  new: T
  errors: string[] | null
}

export interface PresenceState {
  [key: string]: {
    user_id: string
    online_at: string
    [key: string]: unknown
  }[]
}

// ─── Database Query Builder Result Type ──────────────────────────────────────

export type DbResult<T> = PostgrestResponse<T[]>
export type DbSingleResult<T> = PostgrestSingleResponse<T>
