<script lang="ts">
  import { onMount } from 'svelte'

  interface Bucket {
    id: string
    name: string
    public: boolean
    created_at: string
    file_count?: number
  }

  interface StorageObject {
    name: string
    id: string
    updated_at: string
    created_at: string
    metadata?: {
      size?: number
      mimetype?: string
      contentLength?: number
    }
  }

  let buckets = $state<Bucket[]>([])
  let selectedBucket = $state<string | null>(null)
  let objects = $state<StorageObject[]>([])
  let loading = $state(true)
  let objLoading = $state(false)
  let error = $state<string | null>(null)
  let showCreateBucket = $state(false)
  let newBucketName = $state('')
  let newBucketPublic = $state(false)
  let creating = $state(false)
  let uploading = $state(false)
  let uploadInput = $state<HTMLInputElement | undefined>(undefined)
  let currentPath = $state('')
  let confirmDeleteObj = $state<string | null>(null)
  let selectedObj = $state<StorageObject | null>(null)
  let signedUrl = $state<string | null>(null)
  let signedUrlLoading = $state(false)

  import { getOmniBaseUrl, getHeaders } from '$lib/api'

  async function loadBuckets() {
    try {
      loading = true
      error = null
      const resp = await fetch(`${getOmniBaseUrl()}/storage/v1/bucket`, { headers: getHeaders() })
      if (resp.ok) {
        buckets = await resp.json()
      } else {
        const d = await resp.json().catch(() => ({}))
        error = d.error || 'Failed to load buckets'
      }
    } catch {
      error = 'Connection error — is the storage service running?'
    } finally {
      loading = false
    }
  }

  async function createBucket() {
    if (!newBucketName.trim()) return
    creating = true
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/storage/v1/bucket`, {
        method: 'POST',
        headers: { ...getHeaders(), 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newBucketName, id: newBucketName, public: newBucketPublic })
      })
      if (resp.ok) {
        showCreateBucket = false
        newBucketName = ''
        newBucketPublic = false
        await loadBuckets()
      } else {
        const d = await resp.json().catch(() => ({}))
        alert(d.error || 'Failed to create bucket')
      }
    } catch {
      alert('Connection error')
    } finally {
      creating = false
    }
  }

  async function selectBucket(name: string) {
    selectedBucket = name
    currentPath = ''
    selectedObj = null
    signedUrl = null
    await loadObjects()
  }

  async function loadObjects() {
    if (!selectedBucket) return
    objLoading = true
    try {
      // List objects via PostgREST on storage.objects table
      const prefix = currentPath ? `${selectedBucket}/${currentPath}` : selectedBucket
      const resp = await fetch(
        `${getOmniBaseUrl()}/storage/v1/object/list/${selectedBucket}`,
        {
          method: 'POST',
          headers: { ...getHeaders(), 'Content-Type': 'application/json' },
          body: JSON.stringify({ prefix: currentPath, limit: 100 })
        }
      )
      if (resp.ok) {
        objects = await resp.json()
      } else {
        objects = []
      }
    } catch {
      objects = []
    } finally {
      objLoading = false
    }
  }

  async function uploadFile(files: FileList | null) {
    if (!files || files.length === 0 || !selectedBucket) return
    uploading = true
    try {
      for (const file of Array.from(files)) {
        const path = currentPath ? `${currentPath}/${file.name}` : file.name
        const resp = await fetch(`${getOmniBaseUrl()}/storage/v1/object/${selectedBucket}/${path}`, {
          method: 'POST',
          headers: {
            ...getHeaders(),
            'Content-Type': file.type || 'application/octet-stream',
          },
          body: file
        })
        if (!resp.ok) {
          const d = await resp.json().catch(() => ({}))
          alert(`Upload failed for ${file.name}: ${d.error || resp.statusText}`)
        }
      }
      await loadObjects()
    } catch {
      alert('Upload failed — connection error')
    } finally {
      uploading = false
    }
  }

  async function deleteObject(objName: string) {
    if (!selectedBucket) return
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/storage/v1/object/${selectedBucket}`, {
        method: 'DELETE',
        headers: { ...getHeaders(), 'Content-Type': 'application/json' },
        body: JSON.stringify({ prefixes: [objName] })
      })
      if (resp.ok) {
        confirmDeleteObj = null
        if (selectedObj?.name === objName) {
          selectedObj = null
          signedUrl = null
        }
        await loadObjects()
      } else {
        const d = await resp.json().catch(() => ({}))
        alert(d.error || 'Failed to delete object')
      }
    } catch {
      alert('Connection error')
    }
  }

  async function getSignedUrl(obj: StorageObject) {
    if (!selectedBucket) return
    signedUrlLoading = true
    signedUrl = null
    selectedObj = obj
    try {
      const resp = await fetch(
        `${getOmniBaseUrl()}/storage/v1/object/sign/${selectedBucket}/${obj.name}`,
        {
          method: 'POST',
          headers: { ...getHeaders(), 'Content-Type': 'application/json' },
          body: JSON.stringify({ expiresIn: 3600 })
        }
      )
      if (resp.ok) {
        const d = await resp.json()
        signedUrl = d.signedURL
      } else {
        // For public buckets, use public URL
        signedUrl = `${getOmniBaseUrl()}/storage/v1/object/public/${selectedBucket}/${obj.name}`
      }
    } catch {
      signedUrl = null
    } finally {
      signedUrlLoading = false
    }
  }

  function formatSize(bytes: number | undefined): string {
    if (!bytes) return '—'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }

  function formatDate(dateStr: string): string {
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
  }

  function getFileIcon(name: string, mimeType?: string): string {
    const ext = name.split('.').pop()?.toLowerCase() || ''
    if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'avif', 'svg'].includes(ext)) return '🖼️'
    if (['mp4', 'mov', 'avi', 'webm'].includes(ext)) return '🎬'
    if (['mp3', 'wav', 'ogg', 'flac'].includes(ext)) return '🎵'
    if (['pdf'].includes(ext)) return '📄'
    if (['zip', 'tar', 'gz', 'rar'].includes(ext)) return '📦'
    if (['js', 'ts', 'go', 'py', 'rs', 'json'].includes(ext)) return '💾'
    return '📁'
  }

  function isImage(name: string): boolean {
    const ext = name.split('.').pop()?.toLowerCase() || ''
    return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext)
  }

  // Drag and drop
  let isDragOver = $state(false)

  function handleDragOver(e: DragEvent) {
    e.preventDefault()
    isDragOver = true
  }
  function handleDragLeave() { isDragOver = false }
  function handleDrop(e: DragEvent) {
    e.preventDefault()
    isDragOver = false
    uploadFile(e.dataTransfer?.files || null)
  }

  onMount(loadBuckets)
</script>

<svelte:head>
  <title>Storage — OmniBase</title>
  <meta name="description" content="Manage files and buckets in your OmniBase project" />
</svelte:head>

<div class="storage-root">
  <!-- Bucket Sidebar -->
  <aside class="storage-sidebar">
    <div class="sidebar-header">
      <span class="sidebar-label">Buckets</span>
      <button class="btn btn-primary btn-sm" onclick={() => showCreateBucket = true} id="create-bucket-btn">
        + New
      </button>
    </div>

    <div class="sidebar-list">
      {#if loading}
        {#each Array(3) as _}
          <div class="skeleton" style="height: 44px; border-radius: 8px; margin-bottom: 4px;"></div>
        {/each}
      {:else if buckets.length === 0}
        <div style="padding: 20px; text-align: center; color: var(--text-muted); font-size: 12px;">
          No buckets yet
        </div>
      {:else}
        {#each buckets as bucket}
          <button
            class="bucket-item {selectedBucket === bucket.name ? 'active' : ''}"
            onclick={() => selectBucket(bucket.name)}
            id="bucket-{bucket.name}"
          >
            <span style="font-size: 18px;">🗂️</span>
            <div style="flex: 1; text-align: left; min-width: 0;">
              <div style="font-size: 13px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
                {bucket.name}
              </div>
            </div>
            {#if bucket.public}
              <span class="badge badge-success" style="font-size: 9px; padding: 1px 5px;">Public</span>
            {:else}
              <span class="badge badge-neutral" style="font-size: 9px; padding: 1px 5px;">Private</span>
            {/if}
          </button>
        {/each}
      {/if}
    </div>
  </aside>

  <!-- Main Object Browser -->
  <main class="storage-main">
    {#if !selectedBucket}
      <div class="empty-state" style="height: 100%;">
        <div style="font-size: 56px; margin-bottom: 16px; filter: drop-shadow(0 0 20px rgba(108,71,255,0.4));">📦</div>
        <h3>Select a bucket to browse files</h3>
        <p>Or create a new bucket to start uploading</p>
        <button class="btn btn-primary" style="margin-top: 16px;" onclick={() => showCreateBucket = true}>
          Create your first bucket
        </button>
      </div>
    {:else}
      <!-- Object browser header -->
      <div class="obj-header">
        <div>
          <h2 style="font-size: 16px; margin: 0; font-weight: 600;">{selectedBucket}</h2>
          <p style="font-size: 12px; color: var(--text-muted); margin: 2px 0 0;">
            {objects.length} object{objects.length !== 1 ? 's' : ''}
          </p>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary btn-sm" onclick={loadObjects}>Refresh</button>
          <button
            class="btn btn-primary btn-sm"
            onclick={() => uploadInput?.click()}
            disabled={uploading}
            id="upload-btn"
          >
            {uploading ? '⏳ Uploading...' : '⬆ Upload File'}
          </button>
          <input
            bind:this={uploadInput}
            type="file"
            multiple
            style="display: none;"
            onchange={(e) => uploadFile((e.target as HTMLInputElement).files)}
          />
        </div>
      </div>

      <!-- Drop zone + object list -->
      <div class="obj-content">
        {#if objLoading}
          {#each Array(5) as _}
            <div class="skeleton" style="height: 52px; border-radius: 8px; margin-bottom: 4px;"></div>
          {/each}
        {:else if objects.length === 0}
          <div
            class="drop-zone {isDragOver ? 'drag-over' : ''}"
            role="region"
            aria-label="File drop zone"
            ondragover={handleDragOver}
            ondragleave={handleDragLeave}
            ondrop={handleDrop}
          >
            <div style="font-size: 40px; margin-bottom: 12px;">☁️</div>
            <p style="font-size: 15px; font-weight: 500; margin: 0 0 8px;">Drop files here to upload</p>
            <p style="font-size: 12px; color: var(--text-muted); margin: 0 0 16px;">or</p>
            <button class="btn btn-primary btn-sm" onclick={() => uploadInput?.click()}>
              Choose Files
            </button>
          </div>
        {:else}
          <div
            class="obj-list {isDragOver ? 'drag-over-overlay' : ''}"
            ondragover={handleDragOver}
            ondragleave={handleDragLeave}
            ondrop={handleDrop}
            role="region"
            aria-label="Object list"
          >
            {#each objects as obj}
              <div
                class="obj-row {selectedObj?.name === obj.name ? 'selected' : ''}"
                onclick={() => getSignedUrl(obj)}
                role="button"
                tabindex="0"
                onkeydown={(e) => e.key === 'Enter' && getSignedUrl(obj)}
              >
                <span class="obj-icon">{getFileIcon(obj.name, obj.metadata?.mimetype)}</span>
                <div style="flex: 1; min-width: 0;">
                  <div style="font-size: 13px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
                    {obj.name}
                  </div>
                  <div style="font-size: 11px; color: var(--text-muted);">
                    {obj.metadata?.mimetype || 'unknown'} · {formatSize(obj.metadata?.size || obj.metadata?.contentLength)}
                  </div>
                </div>
                <div style="font-size: 11px; color: var(--text-muted); margin: 0 16px;">
                  {formatDate(obj.updated_at || obj.created_at).split(',')[0]}
                </div>
                <button
                  class="btn btn-secondary btn-sm"
                  style="color: var(--status-error); flex-shrink: 0;"
                  onclick={(e) => { e.stopPropagation(); confirmDeleteObj = obj.name }}
                  id="del-obj-{obj.name}"
                >
                  Delete
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </main>

  <!-- Object Detail Panel -->
  {#if selectedObj && selectedBucket}
    <aside class="obj-detail">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h3 style="font-size: 14px; margin: 0;">File Details</h3>
        <button onclick={() => { selectedObj = null; signedUrl = null }} style="background: none; border: none; color: var(--text-muted); cursor: pointer; font-size: 18px;">×</button>
      </div>

      <!-- Preview -->
      {#if signedUrl && isImage(selectedObj.name)}
        <div style="margin-bottom: 16px; border-radius: 8px; overflow: hidden; background: #0a0e17; aspect-ratio: 1; display: flex; align-items: center; justify-content: center;">
          <img src={signedUrl} alt={selectedObj.name} style="max-width: 100%; max-height: 100%; object-fit: contain;" />
        </div>
      {:else}
        <div style="font-size: 48px; text-align: center; margin-bottom: 16px; padding: 20px; background: var(--bg-elevated); border-radius: 8px;">
          {getFileIcon(selectedObj.name, selectedObj.metadata?.mimetype)}
        </div>
      {/if}

      <div style="display: flex; flex-direction: column; gap: 12px;">
        <div>
          <div style="font-size: 11px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; letter-spacing: 0.05em;">Name</div>
          <div style="font-size: 13px; word-break: break-all;">{selectedObj.name}</div>
        </div>
        {#if selectedObj.metadata?.size || selectedObj.metadata?.contentLength}
          <div>
            <div style="font-size: 11px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; letter-spacing: 0.05em;">Size</div>
            <div style="font-size: 13px;">{formatSize(selectedObj.metadata.size || selectedObj.metadata.contentLength)}</div>
          </div>
        {/if}
        {#if selectedObj.metadata?.mimetype}
          <div>
            <div style="font-size: 11px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; letter-spacing: 0.05em;">Type</div>
            <div style="font-size: 13px; font-family: var(--font-mono);">{selectedObj.metadata.mimetype}</div>
          </div>
        {/if}

        <!-- URL -->
        {#if signedUrl}
          <div>
            <div style="font-size: 11px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; letter-spacing: 0.05em;">URL</div>
            <div style="display: flex; gap: 6px; align-items: center;">
              <input
                type="text"
                value={signedUrl}
                readonly
                class="input"
                style="font-size: 10px; font-family: var(--font-mono);"
              />
              <button
                class="btn btn-secondary btn-sm"
                onclick={() => { if (signedUrl) navigator.clipboard.writeText(signedUrl) }}
              >Copy</button>
            </div>
          </div>
          <a
            href={signedUrl}
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary"
            style="text-align: center; text-decoration: none;"
          >Open in New Tab ↗</a>
        {:else if signedUrlLoading}
          <div class="skeleton" style="height: 36px; border-radius: 8px;"></div>
        {/if}
      </div>
    </aside>
  {/if}
</div>

<!-- Create Bucket Modal -->
{#if showCreateBucket}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 420px;">
      <div class="modal-header">
        <h3>Create Bucket</h3>
        <button class="btn-close" onclick={() => { showCreateBucket = false; newBucketName = '' }}>×</button>
      </div>
      <div class="modal-body">
        <div class="form-group" style="margin-bottom: 16px;">
          <label for="bName">Bucket Name</label>
          <input type="text" id="bName" bind:value={newBucketName} class="input" placeholder="e.g. avatars" />
        </div>
        <div class="flex items-center gap-2" style="margin-bottom: 8px;">
          <input type="checkbox" id="bPublic" bind:checked={newBucketPublic} />
          <label for="bPublic" style="font-size: 13px; cursor: pointer;">Make bucket public</label>
        </div>
        <p style="font-size: 12px; color: var(--text-muted);">
          {newBucketPublic
            ? 'Anyone with the file URL can access files (no auth required).'
            : 'Files require authentication or signed URLs to access.'}
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => { showCreateBucket = false; newBucketName = '' }}>Cancel</button>
        <button class="btn btn-primary" onclick={createBucket} disabled={creating || !newBucketName.trim()}>
          {creating ? 'Creating...' : 'Create Bucket'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Confirm Delete Object Modal -->
{#if confirmDeleteObj}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 420px;">
      <div class="modal-header">
        <h3>Delete Object</h3>
        <button class="btn-close" onclick={() => confirmDeleteObj = null}>×</button>
      </div>
      <div class="modal-body">
        <p style="font-size: 14px;">
          Permanently delete <strong>{confirmDeleteObj}</strong>? This cannot be undone.
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => confirmDeleteObj = null}>Cancel</button>
        <button
          class="btn btn-primary"
          style="background: var(--status-error); border-color: var(--status-error);"
          onclick={() => deleteObject(confirmDeleteObj!)}
        >
          Delete
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .storage-root {
    display: grid;
    grid-template-columns: 220px 1fr;
    height: calc(100vh - var(--topbar-height));
    overflow: hidden;
  }
  .storage-root:has(.obj-detail) {
    grid-template-columns: 220px 1fr 280px;
  }

  /* Sidebar */
  .storage-sidebar {
    background: var(--bg-surface);
    border-right: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .sidebar-header {
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .sidebar-label {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }
  .sidebar-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }
  .bucket-item {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 10px;
    background: none;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: background 0.15s;
    color: var(--text-primary);
  }
  .bucket-item:hover { background: var(--bg-elevated); }
  .bucket-item.active {
    background: rgba(108, 71, 255, 0.12);
    color: var(--brand-primary);
  }

  /* Main */
  .storage-main {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .obj-header {
    padding: 14px 20px;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--bg-surface);
  }
  .obj-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
  }

  /* Drop zone */
  .drop-zone {
    border: 2px dashed var(--border-subtle);
    border-radius: var(--radius-lg);
    padding: 60px 20px;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s;
  }
  .drop-zone.drag-over, .drag-over-overlay {
    border-color: var(--brand-primary);
    background: rgba(108, 71, 255, 0.06);
  }

  /* Object list */
  .obj-list { display: flex; flex-direction: column; gap: 4px; }
  .obj-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.15s;
    user-select: none;
  }
  .obj-row:hover { border-color: var(--border-brand); background: var(--bg-elevated); }
  .obj-row.selected { border-color: var(--brand-primary); background: rgba(108, 71, 255, 0.06); }
  .obj-icon { font-size: 20px; flex-shrink: 0; }

  /* Detail panel */
  .obj-detail {
    background: var(--bg-surface);
    border-left: 1px solid var(--border-subtle);
    padding: 20px;
    overflow-y: auto;
  }

  /* Modal */
  .modal-overlay {
    position: fixed; top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.7); backdrop-filter: blur(4px);
    display: flex; align-items: center; justify-content: center; z-index: 1000;
  }
  .modal-card {
    background: var(--bg-surface); border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg); box-shadow: var(--shadow-xl);
    display: flex; flex-direction: column;
  }
  .modal-header {
    padding: 16px 20px; border-bottom: 1px solid var(--border-subtle);
    display: flex; justify-content: space-between; align-items: center;
  }
  .modal-header h3 { font-size: 15px; margin: 0; }
  .modal-body { padding: 20px; }
  .modal-footer {
    padding: 16px 20px; border-top: 1px solid var(--border-subtle);
    display: flex; justify-content: flex-end; gap: 12px;
  }
  .btn-close {
    background: none; border: none; color: var(--text-muted);
    font-size: 22px; cursor: pointer; line-height: 1;
  }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .form-group label { font-size: 13px; font-weight: 500; }
</style>
