<script lang="ts">
  import { onMount } from 'svelte'

  let loading = $state(true)
  let error = $state<string | null>(null)

  onMount(() => {
    // Just a placeholder for now
    setTimeout(() => {
      loading = false
      error = "GraphQL support is currently being integrated. Please use the REST API or SQL Editor as an alternative."
    }, 800)
  })
</script>

<svelte:head>
  <title>GraphQL Explorer — OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">GraphQL API</h1>
    <p class="page-subtitle">Query your database using GraphQL</p>
  </div>
  <div class="flex gap-2">
    <button class="btn btn-secondary btn-sm" disabled>Docs</button>
  </div>
</div>

<div class="page-content animate-fade-in" style="height: calc(100vh - 200px);">
  <div class="card" style="height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; background: #0d1117;">
    {#if loading}
      <div class="skeleton" style="width: 300px; height: 100px;"></div>
    {:else}
      <div style="text-align: center; max-width: 500px; padding: 40px;">
        <div style="font-size: 64px; margin-bottom: 24px; color: var(--brand-primary); opacity: 0.8;">⬡</div>
        <h2 style="color: #fff; margin-bottom: 16px;">GraphQL is coming soon</h2>
        <p style="color: var(--text-muted); line-height: 1.6;">
          An auto-generated GraphQL API is currently in development. It will automatically detect 
          your schema changes and respect all RLS policies.
        </p>
        <div style="margin-top: 32px; padding: 20px; background: rgba(255,255,255,0.03); border-radius: 8px; border: 1px solid rgba(255,255,255,0.05); font-family: var(--font-mono); font-size: 13px; color: #a5d6ff; text-align: left;">
          <span style="color: #ff7b72;">query</span> {'{'} <br>
          &nbsp;&nbsp;usersCollection {'{'} <br>
          &nbsp;&nbsp;&nbsp;&nbsp;edges {'{'} <br>
          &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;node {'{'} <br>
          &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;id <br>
          &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;email <br>
          &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{'}'} <br>
          &nbsp;&nbsp;&nbsp;&nbsp;{'}'} <br>
          &nbsp;&nbsp;{'}'} <br>
          {'}'}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .page-content { padding: 0 24px 24px; }
</style>
