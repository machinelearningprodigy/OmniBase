<script lang="ts">
  import { onMount } from 'svelte'

  interface Bucket {
    id: string
    name: string
    public: boolean
  }

  let buckets: Bucket[] = []
  let loading = true
  const OMNIBASE_URL = 'http://localhost:8000'

  onMount(async () => {
    try {
      const session = typeof localStorage !== 'undefined' ? localStorage.getItem('omnibase.session') : null
      const token = session ? JSON.parse(session).access_token : null

      const resp = await fetch(`${OMNIBASE_URL}/storage/v1/bucket`, {
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        }
      })
      if (resp.ok) {
        buckets = await resp.json() || []
      }
    } catch {
      buckets = [{ id: 'public-assets', name: 'public-assets', public: true }]
    } finally {
      loading = false
    }
  })
</script>

<svelte:head>
  <title>Storage — OmniBase</title>
</svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Storage Buckets</h1>
    <p class="page-subtitle">Manage files and folders using S3-compatible endpoints</p>
  </div>
  <button class="btn btn-primary">+ New Bucket</button>
</div>

<div class="page-content">
  {#if loading}
    <div class="skeleton" style="height: 100px; border-radius: 8px;"></div>
  {:else if buckets.length === 0}
    <div class="empty-state">
      <p style="font-size: 32px; margin-bottom: 12px;">📁</p>
      <h3 style="font-size: 16px; font-weight: 600;">No buckets found</h3>
      <p style="color: var(--text-muted); font-size: 13px;">Create a bucket to start uploading files.</p>
    </div>
  {:else}
    <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px;">
      {#each buckets as bucket}
        <div class="card" style="display: flex; flex-direction: column; gap: 12px; cursor: pointer;">
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <h3 style="font-size: 14px; font-weight: 600; display: flex; align-items: center; gap: 8px;">
              <span style="color: var(--brand-primary);">📁</span> {bucket.name}
            </h3>
            <span class="badge {bucket.public ? 'badge-success' : 'badge-neutral'}">{bucket.public ? 'Public' : 'Private'}</span>
          </div>
          <div style="font-size: 12px; color: var(--text-secondary); font-family: var(--font-mono);">
            ID: {bucket.id}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
