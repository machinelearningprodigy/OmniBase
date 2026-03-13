import type { AuthClient } from '../auth/AuthClient'
import type { OmniBaseError, StorageObject, SignedURL } from '../types'

interface StorageClientOptions {
  url: string
  headers: Record<string, string>
  auth: AuthClient
}

interface UploadOptions {
  contentType?: string
  upsert?: boolean
  metadata?: Record<string, unknown>
}

interface DownloadOptions {
  transform?: {
    width?: number
    height?: number
    quality?: number
    format?: 'webp' | 'avif' | 'jpeg' | 'png'
    resize?: 'cover' | 'contain' | 'fill'
  }
}

interface StorageResponse<T> {
  data: T | null
  error: OmniBaseError | null
}

/**
 * StorageClient provides file upload, download, and management.
 *
 * @example
 * ```typescript
 * // Upload a file
 * const { data, error } = await omni.storage.from('avatars').upload('user-1.jpg', file)
 *
 * // Get public URL
 * const { data } = omni.storage.from('avatars').getPublicUrl('user-1.jpg')
 *
 * // Get public URL with image transformation
 * const { data } = omni.storage.from('avatars').getPublicUrl('user-1.jpg', {
 *   transform: { width: 100, height: 100, format: 'webp' }
 * })
 *
 * // Create a signed URL for private files
 * const { data } = await omni.storage.from('private').createSignedUrl('doc.pdf', 3600)
 *
 * // List files in a bucket
 * const { data } = await omni.storage.from('avatars').list('', { limit: 50 })
 *
 * // Delete a file
 * const { error } = await omni.storage.from('avatars').remove(['user-1.jpg'])
 * ```
 */
export class StorageClient {
  private url: string
  private headers: Record<string, string>
  private auth: AuthClient

  constructor(options: StorageClientOptions) {
    this.url = options.url
    this.headers = options.headers
    this.auth = options.auth
  }

  /**
   * Access a storage bucket by name.
   */
  from(bucket: string): BucketClient {
    return new BucketClient(bucket, this.url, this._getHeaders.bind(this))
  }

  /**
   * List all storage buckets in the project.
   */
  async listBuckets(): Promise<StorageResponse<{ name: string; public: boolean }[]>> {
    return this._request('GET', '/bucket')
  }

  /**
   * Create a new storage bucket.
   */
  async createBucket(
    name: string,
    options?: { public?: boolean; allowedMimeTypes?: string[]; fileSizeLimit?: number }
  ): Promise<StorageResponse<{ name: string }>> {
    return this._request('POST', '/bucket', {
      id: name,
      name,
      public: options?.public ?? false,
      allowed_mime_types: options?.allowedMimeTypes,
      file_size_limit: options?.fileSizeLimit,
    })
  }

  /**
   * Delete a storage bucket and all its files.
   */
  async deleteBucket(name: string): Promise<StorageResponse<{ message: string }>> {
    return this._request('DELETE', `/bucket/${name}`)
  }

  private async _request<T>(method: string, path: string, body?: unknown): Promise<StorageResponse<T>> {
    const headers = this._getHeaders()
    try {
      const response = await fetch(`${this.url}${path}`, {
        method,
        headers: { ...headers, 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined,
      })
      const data = await response.json().catch(() => null)
      if (!response.ok) {
        return { data: null, error: { code: data?.statusCode ?? 'error', message: data?.message ?? `HTTP ${response.status}` } }
      }
      return { data, error: null }
    } catch (err) {
      return { data: null, error: { code: 'network_error', message: err instanceof Error ? err.message : 'Network error' } }
    }
  }

  private _getHeaders(): Record<string, string> {
    const headers = { ...this.headers }
    const token = this.auth.getAccessToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
    return headers
  }
}

/** BucketClient provides file operations within a specific bucket */
class BucketClient {
  private bucket: string
  private baseUrl: string
  private _getHeaders: () => Record<string, string>

  constructor(bucket: string, baseUrl: string, getHeaders: () => Record<string, string>) {
    this.bucket = bucket
    this.baseUrl = baseUrl
    this._getHeaders = getHeaders
  }

  /** Upload a file to this bucket */
  async upload(
    path: string,
    fileBody: File | Blob | ArrayBuffer | string,
    options?: UploadOptions
  ): Promise<StorageResponse<StorageObject>> {
    const method = options?.upsert ? 'PUT' : 'POST'
    const url = `${this.baseUrl}/object/${this.bucket}/${path}`

    const headers = this._getHeaders()
    if (options?.contentType) headers['Content-Type'] = options.contentType

    try {
      const response = await fetch(url, { method, headers, body: fileBody as BodyInit })
      const data = await response.json().catch(() => null)
      if (!response.ok) {
        return { data: null, error: { code: data?.statusCode ?? 'upload_error', message: data?.message ?? `Upload failed: HTTP ${response.status}` } }
      }
      return { data, error: null }
    } catch (err) {
      return { data: null, error: { code: 'network_error', message: err instanceof Error ? err.message : 'Network error' } }
    }
  }

  /** Download a file from this bucket */
  async download(path: string, options?: DownloadOptions): Promise<StorageResponse<Blob>> {
    const url = this._buildTransformUrl(path, options?.transform)
    const headers = this._getHeaders()

    try {
      const response = await fetch(url, { headers })
      if (!response.ok) {
        return { data: null, error: { code: 'download_error', message: `Download failed: HTTP ${response.status}` } }
      }
      return { data: await response.blob(), error: null }
    } catch (err) {
      return { data: null, error: { code: 'network_error', message: err instanceof Error ? err.message : 'Network error' } }
    }
  }

  /** Get the public URL for a file (only works for public buckets) */
  getPublicUrl(path: string, options?: DownloadOptions): { data: { publicUrl: string } } {
    const url = this._buildTransformUrl(path, options?.transform)
    return { data: { publicUrl: url } }
  }

  /** Create a signed URL for private file access */
  async createSignedUrl(path: string, expiresIn: number): Promise<StorageResponse<SignedURL>> {
    return this._request('POST', `/object/sign/${this.bucket}/${path}`, { expiresIn })
  }

  /** List files in a bucket folder */
  async list(
    prefix: string = '',
    options?: { limit?: number; offset?: number; sortBy?: { column: string; order: 'asc' | 'desc' } }
  ): Promise<StorageResponse<StorageObject[]>> {
    return this._request('POST', `/object/list/${this.bucket}`, {
      prefix,
      limit: options?.limit ?? 100,
      offset: options?.offset ?? 0,
      sortBy: options?.sortBy ?? { column: 'name', order: 'asc' },
    })
  }

  /** Move/rename a file */
  async move(fromPath: string, toPath: string): Promise<StorageResponse<{ message: string }>> {
    return this._request('POST', '/object/move', { bucketId: this.bucket, sourceKey: fromPath, destinationKey: toPath })
  }

  /** Copy a file */
  async copy(fromPath: string, toPath: string): Promise<StorageResponse<{ id: string }>> {
    return this._request('POST', '/object/copy', { bucketId: this.bucket, sourceKey: fromPath, destinationKey: toPath })
  }

  /** Delete one or more files */
  async remove(paths: string[]): Promise<StorageResponse<StorageObject[]>> {
    return this._request('DELETE', `/object/${this.bucket}`, { prefixes: paths })
  }

  private _buildTransformUrl(path: string, transform?: DownloadOptions['transform']): string {
    if (!transform) return `${this.baseUrl}/object/public/${this.bucket}/${path}`

    const params = new URLSearchParams()
    if (transform.width) params.set('width', String(transform.width))
    if (transform.height) params.set('height', String(transform.height))
    if (transform.quality) params.set('quality', String(transform.quality))
    if (transform.format) params.set('format', transform.format)
    if (transform.resize) params.set('resize', transform.resize)

    return `${this.baseUrl}/render/image/public/${this.bucket}/${path}?${params.toString()}`
  }

  private async _request<T>(method: string, path: string, body?: unknown): Promise<StorageResponse<T>> {
    const headers = { ...this._getHeaders(), 'Content-Type': 'application/json' }
    try {
      const response = await fetch(`${this.baseUrl}${path}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      })
      const data = await response.json().catch(() => null)
      if (!response.ok) return { data: null, error: { code: data?.statusCode ?? 'error', message: data?.message ?? `HTTP ${response.status}` } }
      return { data, error: null }
    } catch (err) {
      return { data: null, error: { code: 'network_error', message: err instanceof Error ? err.message : 'Network error' } }
    }
  }
}
