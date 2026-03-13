<script lang="ts">
  import { onMount } from 'svelte'

  interface Bucket {
    id: string
    name: string
    public: boolean
    created_at: string
  }

  let buckets = $state<Bucket[]>([])
  let selectedBucket = $state<string | null>(null)
  let objects = $state<any[]>([])
  let loading = $state(true)
  let objLoading = $state(false)
  let error = $state<string | null>(null)
  let showCreateBucket = $state(false)
  let newBucketName = $state('')
  let isPublic = $state(false)

  const OMNIBASE_URL = 'http://localhost:8000'

  function getHeaders() {
    const session = typeof localStorage !== 'undefined' ? localStorage.getItem('omnibase.session') : null
    const token = session ? JSON.parse(session).access_token : null
    return {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  }

  async function loadBuckets() {
    try {
      loading = true
      const resp = await fetch(`${OMNIBASE_URL}/storage/v1/bucket`, { headers: getHeaders() })
      if (resp.ok) {
        buckets = await resp.json()
      } else {
        error = 'Failed to load buckets'
      }
    } catch {
      error = 'Connection error'
    } finally {
      loading = false
    }
  }

  async function createBucket() {
    if (!newBucketName) return
    try {
      const resp = await fetch(`${OMNIBASE_URL}/storage/v1/bucket`, {
        method: 'POST',
        headers: { ...getHeaders(), 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newBucketName, public: isPublic })
      })
      if (resp.ok) {
        showCreateBucket = false
        newBucketName = ''
        await loadBuckets()
      } else {
        const d = await resp.json()
        alert(d.error)
      }
    } catch {
      alert('Error creating bucket')
    }
  }

  async function selectBucket(name: string) {
    selectedBucket = name
    // Future: Load objects in bucket
    objects = []
  }

  onMount(loadBuckets)
</script>

<svelte:head>
  <title>Storage — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Storage</h1>
    <p class="page-subtitle">Manage files and buckets</p>
  </div>
  <button class="btn btn-primary btn-sm" onclick={() => showCreateBucket = true}>
    Create Bucket
  </button>
</div>

<div class="page-content">
  {#if loading}
    <div class="skeleton" style="height: 200px; border-radius: 12px;"></div>
  {:else if buckets.length === 0}
    <div class="empty-state card">
      <div style="font-size: 40px; margin-bottom: 16px;">📦</div>
      <h3>No buckets found</h3>
      <p>Create a bucket to start uploading files</p>
      <button class="btn btn-primary btn-sm" style="margin-top: 16px;" onclick={() => showCreateBucket = true}>
        Create Bucket
      </button>
    </div>
  {:else}
    <div style="display: grid; grid-template-columns: 280px 1fr; gap: 24px;">
      <!-- Bucket Sidebar -->
      <div class="card" style="padding: 12px; height: fit-content;">
        <h3 style="font-size: 13px; margin: 0 0 12px 0; color: var(--text-muted); text-transform: uppercase;">Buckets</h3>
        <div style="display: flex; flex-direction: column; gap: 4px;">
          {#each buckets as bucket}
            <button 
              class="nav-item {selectedBucket === bucket.name ? 'active' : ''}"
              onclick={() => selectBucket(bucket.name)}
              style="width: 100%; border: none; background: none; padding: 10px 12px; border-radius: 6px; text-align: left; cursor: pointer; display: flex; align-items: center; justify-content: space-between;"
            >
              <span style="font-size: 14px;">{bucket.name}</span>
              {#if bucket.public}
                <span class="badge badge-success" style="font-size: 10px;">Public</span>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Object View -->
      <div class="card">
        {#if !selectedBucket}
          <div class="empty-state" style="padding: 60px 0;">
            <p>Select a bucket to view files</p>
          </div>
        {:else}
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
            <h2 style="font-size: 16px; margin: 0;">{selectedBucket}</h2>
            <div class="flex gap-2">
              <button class="btn btn-secondary btn-sm">Upload File</button>
            </div>
          </div>
          <div class="empty-state" style="padding: 40px 0;">
             <div style="font-size: 30px; margin-bottom: 12px;">📂</div>
             <p>No files found in this bucket</p>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

{#if showCreateBucket}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 400px;">
      <div class="modal-header">
        <h3>Create Bucket</h3>
        <button class="btn-close" onclick={() => showCreateBucket = false}>×</button>
      </div>
      <div class="modal-body">
        <div class="form-group mb-4">
          <label for="bName">Bucket Name</label>
          <input type="text" id="bName" bind:value={newBucketName} class="input" placeholder="e.g. avatars" />
        </div>
        <div class="flex items-center gap-2">
          <input type="checkbox" id="isPublic" bind:checked={isPublic} />
          <label for="isPublic" style="font-size: 13px;">Public Bucket</label>
        </div>
        <p style="font-size: 11px; color: var(--text-muted); margin-top: 8px;">
          Public buckets allow anyone with the URL to download files.
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateBucket = false}>Cancel</button>
        <button class="btn btn-primary" onclick={createBucket}>Create</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-content { padding: 0 24px 24px; }
  .nav-item:hover { background: rgba(255,255,255,0.04); }
  .nav-item.active { background: rgba(var(--brand-primary-rgb), 0.1); color: var(--brand-primary); }
</style>
