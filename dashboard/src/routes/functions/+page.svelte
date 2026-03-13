<script lang="ts">
  import { onMount } from 'svelte'

  interface Function {
    id: string
    name: string
    slug: string
    status: string
    version: string
    created_at: string
  }

  let functions = $state<Function[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  const OMNIBASE_URL = 'http://localhost:8000'

  function getHeaders() {
    const session = typeof localStorage !== 'undefined' ? localStorage.getItem('omnibase.session') : null
    const token = session ? JSON.parse(session).access_token : null
    return {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  }

  async function loadFunctions() {
    try {
      loading = true
      // Hits the gateway which proxies to the functions service
      const resp = await fetch(`${OMNIBASE_URL}/functions/v1/`, { headers: getHeaders() })
      if (resp.ok) {
        // For now the service returns a placeholder message
        functions = [] 
      }
    } catch {
      error = 'Connection error'
    } finally {
      loading = false
    }
  }

  onMount(loadFunctions)
</script>

<svelte:head>
  <title>Edge Functions — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Edge Functions</h1>
    <p class="page-subtitle">Serverless TypeScript functions at the edge</p>
  </div>
  <button class="btn btn-primary btn-sm" disabled>
    New Function
  </button>
</div>

<div class="page-content animate-fade-in">
  {#if loading}
    <div class="skeleton" style="height: 200px; border-radius: 12px;"></div>
  {:else if functions.length === 0}
    <div class="card" style="padding: 60px; text-align: center;">
      <div style="font-size: 48px; margin-bottom: 20px;">λ</div>
      <h2 style="margin-bottom: 12px;">No functions found</h2>
      <p style="color: var(--text-muted); max-width: 480px; margin: 0 auto 24px;">
        Edge Functions are currently in preview. They allow you to run custom business logic 
        closer to your users with ultra-low latency.
      </p>
      <div style="display: flex; gap: 12px; justify-content: center;">
        <a href="https://github.com/machinelearningprodigy/OmniBase" target="_blank" class="btn btn-secondary btn-sm">Read Roadmap</a>
        <button class="btn btn-primary btn-sm" disabled>Create your first function</button>
      </div>
    </div>
  {:else}
    <!-- Grid of functions -->
  {/if}
</div>

<style>
  .page-content { padding: 0 24px 24px; }
</style>
