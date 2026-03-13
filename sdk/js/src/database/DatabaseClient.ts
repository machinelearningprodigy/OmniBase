import type { AuthClient } from '../auth/AuthClient'
import type { OmniBaseError, FilterOperator, OrderDirection, NullsOrder } from '../types'

interface DatabaseClientOptions {
  url: string
  headers: Record<string, string>
  auth: AuthClient
}

interface PostgrestResponse<T> {
  data: T | null
  error: OmniBaseError | null
  count?: number
  status: number
  statusText: string
}

/**
 * DatabaseClient wraps PostgREST to provide a fluent query builder.
 * Full TypeScript generics give you autocomplete on column names when
 * you pass the generated database types.
 *
 * @example
 * ```typescript
 * // With generated types
 * import type { Database } from './types/database'
 * const omni = createClient<Database>(url, key)
 *
 * // Simple select
 * const { data } = await omni.from('users').select('id, email, name')
 *
 * // Filtered select
 * const { data } = await omni.from('posts')
 *   .select('*, user:users(email)')
 *   .eq('published', true)
 *   .order('created_at', { ascending: false })
 *   .limit(10)
 *
 * // Insert
 * const { data } = await omni.from('posts').insert({ title: 'Hello', body: '...' })
 *
 * // Update
 * const { data } = await omni.from('posts').update({ title: 'New Title' }).eq('id', postId)
 *
 * // Delete
 * const { error } = await omni.from('posts').delete().eq('id', postId)
 * ```
 */
export class DatabaseClient {
  private url: string
  private headers: Record<string, string>
  private auth: AuthClient

  constructor(options: DatabaseClientOptions) {
    this.url = options.url
    this.headers = options.headers
    this.auth = options.auth
  }

  /**
   * Start a query on a table.
   */
  from<T = Record<string, unknown>>(table: string): QueryBuilder<T> {
    return new QueryBuilder<T>(this.url, table, this._getHeaders.bind(this))
  }

  private _getHeaders(): Record<string, string> {
    const headers = { ...this.headers }
    const token = this.auth.getAccessToken()
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    return headers
  }
}

/**
 * QueryBuilder provides a fluent interface for building PostgREST queries.
 * Chainable methods map directly to PostgREST query parameters.
 */
export class QueryBuilder<T> {
  private _table: string
  private _baseUrl: string
  private _getHeaders: () => Record<string, string>
  private _select: string = '*'
  private _filters: string[] = []
  private _order: string[] = []
  private _limit?: number
  private _offset?: number
  private _single = false
  private _count?: 'exact' | 'planned' | 'estimated'
  private _returning = 'representation'

  constructor(baseUrl: string, table: string, getHeaders: () => Record<string, string>) {
    this._baseUrl = baseUrl
    this._table = table
    this._getHeaders = getHeaders
  }

  // ─── Column Selection ────────────────────────────────────────────────────────

  /** Select specific columns. Supports relationships: 'id, name, posts(*)' */
  select(columns: string = '*', options?: { count?: 'exact' | 'planned' | 'estimated' }): this {
    this._select = columns
    if (options?.count) this._count = options.count
    return this
  }

  // ─── Filters ─────────────────────────────────────────────────────────────────

  /** Filter by equality: WHERE column = value */
  eq<K extends keyof T>(column: K, value: T[K]): this {
    this._filters.push(`${String(column)}=eq.${this._encodeValue(value)}`)
    return this
  }

  /** Filter by inequality: WHERE column != value */
  neq<K extends keyof T>(column: K, value: T[K]): this {
    this._filters.push(`${String(column)}=neq.${this._encodeValue(value)}`)
    return this
  }

  /** Filter by greater than: WHERE column > value */
  gt<K extends keyof T>(column: K, value: T[K]): this {
    this._filters.push(`${String(column)}=gt.${this._encodeValue(value)}`)
    return this
  }

  /** Filter by greater than or equal: WHERE column >= value */
  gte<K extends keyof T>(column: K, value: T[K]): this {
    this._filters.push(`${String(column)}=gte.${this._encodeValue(value)}`)
    return this
  }

  /** Filter by less than: WHERE column < value */
  lt<K extends keyof T>(column: K, value: T[K]): this {
    this._filters.push(`${String(column)}=lt.${this._encodeValue(value)}`)
    return this
  }

  /** Filter by less than or equal: WHERE column <= value */
  lte<K extends keyof T>(column: K, value: T[K]): this {
    this._filters.push(`${String(column)}=lte.${this._encodeValue(value)}`)
    return this
  }

  /** Filter using LIKE pattern: WHERE column LIKE '%value%' */
  like(column: keyof T, pattern: string): this {
    this._filters.push(`${String(column)}=like.${pattern}`)
    return this
  }

  /** Filter using case-insensitive ILIKE: WHERE column ILIKE '%value%' */
  ilike(column: keyof T, pattern: string): this {
    this._filters.push(`${String(column)}=ilike.${pattern}`)
    return this
  }

  /** Filter by IS NULL or IS TRUE/FALSE: WHERE column IS value */
  is(column: keyof T, value: 'null' | true | false): this {
    this._filters.push(`${String(column)}=is.${value}`)
    return this
  }

  /** Filter using IN: WHERE column IN (values) */
  in(column: keyof T, values: unknown[]): this {
    this._filters.push(`${String(column)}=in.(${values.map((v) => this._encodeValue(v)).join(',')})`)
    return this
  }

  /** Apply a raw PostgREST filter */
  filter(column: string, operator: FilterOperator, value: unknown): this {
    this._filters.push(`${column}=${operator}.${this._encodeValue(value)}`)
    return this
  }

  // ─── Ordering and Pagination ──────────────────────────────────────────────────

  /** Order results by a column */
  order(
    column: keyof T,
    options?: { ascending?: boolean; nullsFirst?: boolean }
  ): this {
    const dir: OrderDirection = options?.ascending !== false ? 'asc' : 'desc'
    const nulls: NullsOrder = options?.nullsFirst ? 'first' : 'last'
    this._order.push(`${String(column)}.${dir}.nulls${nulls}`)
    return this
  }

  /** Limit number of rows returned */
  limit(count: number): this {
    this._limit = count
    return this
  }

  /** Skip the first n rows (for pagination) */
  range(from: number, to: number): this {
    this._offset = from
    this._limit = to - from + 1
    return this
  }

  /** Return only a single row, error if 0 or >1 rows match */
  single(): SingleQueryBuilder<T> {
    this._single = true
    return this as unknown as SingleQueryBuilder<T>
  }

  // ─── Mutations ───────────────────────────────────────────────────────────────

  /** Insert one or more rows */
  insert(values: Partial<T> | Partial<T>[], options?: { onConflict?: string }): PostgrestMutationBuilder<T> {
    return new PostgrestMutationBuilder<T>(
      this._baseUrl,
      this._table,
      'POST',
      values,
      options,
      this._getHeaders
    )
  }

  /** Update rows matching the current filters */
  update(values: Partial<T>): PostgrestMutationBuilder<T> {
    return new PostgrestMutationBuilder<T>(
      this._baseUrl,
      this._table,
      'PATCH',
      values,
      undefined,
      this._getHeaders,
      this._filters
    )
  }

  /** Upsert rows (insert or update on conflict) */
  upsert(values: Partial<T> | Partial<T>[], options?: { onConflict?: string; ignoreDuplicates?: boolean }): PostgrestMutationBuilder<T> {
    return new PostgrestMutationBuilder<T>(
      this._baseUrl,
      this._table,
      'POST',
      values,
      { ...options, isUpsert: true },
      this._getHeaders
    )
  }

  /** Delete rows matching the current filters */
  delete(): PostgrestMutationBuilder<T> {
    return new PostgrestMutationBuilder<T>(
      this._baseUrl,
      this._table,
      'DELETE',
      undefined,
      undefined,
      this._getHeaders,
      this._filters
    )
  }

  // ─── Execute ─────────────────────────────────────────────────────────────────

  then<TResult1 = PostgrestResponse<T[]>, TResult2 = never>(
    onfulfilled?: ((value: PostgrestResponse<T[]>) => TResult1 | PromiseLike<TResult1>) | null,
    onrejected?: ((reason: unknown) => TResult2 | PromiseLike<TResult2>) | null
  ): Promise<TResult1 | TResult2> {
    return this._execute().then(onfulfilled, onrejected)
  }

  private async _execute(): Promise<PostgrestResponse<T[]>> {
    const url = new URL(`${this._baseUrl}/${this._table}`)
    url.searchParams.set('select', this._select)

    this._filters.forEach((f) => {
      const [key, val] = f.split('=')
      url.searchParams.append(key, val)
    })

    if (this._order.length) url.searchParams.set('order', this._order.join(','))
    if (this._limit !== undefined) url.searchParams.set('limit', String(this._limit))
    if (this._offset !== undefined) url.searchParams.set('offset', String(this._offset))

    const headers = this._getHeaders()

    if (this._count) headers['Prefer'] = `count=${this._count}`
    if (this._single) headers['Accept'] = 'application/vnd.pgrst.object+json'

    try {
      const response = await fetch(url.toString(), { headers })
      const data = await response.json().catch(() => null)

      if (!response.ok) {
        return {
          data: null,
          error: {
            code: data?.code ?? 'request_failed',
            message: data?.message ?? data?.hint ?? `HTTP ${response.status}`,
            status: response.status,
          },
          status: response.status,
          statusText: response.statusText,
        }
      }

      const count = response.headers.get('Content-Range')
        ? parseInt(response.headers.get('Content-Range')!.split('/')[1])
        : undefined

      return { data, error: null, count, status: response.status, statusText: response.statusText }
    } catch (err) {
      return {
        data: null,
        error: { code: 'network_error', message: err instanceof Error ? err.message : 'Network error' },
        status: 0,
        statusText: 'Network Error',
      }
    }
  }

  private _encodeValue(value: unknown): string {
    if (value === null) return 'null'
    if (typeof value === 'string') return value
    return String(value)
  }
}

// Type alias for single() queries
export type SingleQueryBuilder<T> = QueryBuilder<T>

/** Handles INSERT, UPDATE, DELETE operations */
class PostgrestMutationBuilder<T> {
  private _url: string
  private _table: string
  private _method: string
  private _body: unknown
  private _options?: Record<string, unknown>
  private _getHeaders: () => Record<string, string>
  private _filters: string[]

  constructor(
    baseUrl: string,
    table: string,
    method: string,
    body: unknown,
    options: Record<string, unknown> | undefined,
    getHeaders: () => Record<string, string>,
    filters: string[] = []
  ) {
    this._url = baseUrl
    this._table = table
    this._method = method
    this._body = body
    this._options = options
    this._getHeaders = getHeaders
    this._filters = filters
  }

  /** Add a filter for mutations (used with update/delete) */
  eq(column: string, value: unknown): this {
    this._filters.push(`${column}=eq.${value}`)
    return this
  }

  then<TResult1 = PostgrestResponse<T[]>, TResult2 = never>(
    onfulfilled?: ((value: PostgrestResponse<T[]>) => TResult1 | PromiseLike<TResult1>) | null,
    onrejected?: ((reason: unknown) => TResult2 | PromiseLike<TResult2>) | null
  ): Promise<TResult1 | TResult2> {
    return this._execute().then(onfulfilled, onrejected)
  }

  private async _execute(): Promise<PostgrestResponse<T[]>> {
    const url = new URL(`${this._url}/${this._table}`)
    this._filters.forEach((f) => {
      const [key, val] = f.split('=')
      url.searchParams.append(key, val)
    })

    const headers = {
      ...this._getHeaders(),
      'Prefer': 'return=representation',
    }

    if (this._options?.isUpsert) {
      headers['Prefer'] = `resolution=merge-duplicates,${headers['Prefer']}`
      headers['on-conflict'] = String(this._options.onConflict ?? 'id')
    }

    try {
      const response = await fetch(url.toString(), {
        method: this._method,
        headers,
        body: this._body ? JSON.stringify(this._body) : undefined,
      })
      const data = await response.json().catch(() => null)

      if (!response.ok) {
        return {
          data: null,
          error: {
            code: data?.code ?? 'request_failed',
            message: data?.message ?? `HTTP ${response.status}`,
            status: response.status,
          },
          status: response.status,
          statusText: response.statusText,
        }
      }

      return { data, error: null, status: response.status, statusText: response.statusText }
    } catch (err) {
      return {
        data: null,
        error: { code: 'network_error', message: err instanceof Error ? err.message : 'Network error' },
        status: 0,
        statusText: 'Network Error',
      }
    }
  }
}
