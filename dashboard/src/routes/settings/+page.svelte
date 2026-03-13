<script lang="ts">
  let url = 'http://localhost:8000'
  let anonKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9...'
  let serviceKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIn0...'
  
  function copy(text: string) {
    navigator.clipboard.writeText(text)
  }
</script>

<svelte:head>
  <title>API Settings — OmniBase</title>
</svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">API Settings</h1>
    <p class="page-subtitle">Project URL and API keys for the OmniBase client SDKs</p>
  </div>
</div>

<div class="page-content" style="max-width: 800px;">
  <div class="card" style="margin-bottom: 24px;">
    <h2 style="font-size: 15px; font-weight: 600; margin-bottom: 8px;">Project URL</h2>
    <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 16px;">This is your RESTful endpoint for querying and managing your database.</p>
    
    <div style="display: flex; gap: 8px;">
      <input type="text" class="input" value={url} readonly style="flex: 1; font-family: var(--font-mono); color: var(--text-primary);" />
      <button class="btn btn-secondary" on:click={() => copy(url)}>Copy</button>
    </div>
  </div>

  <div class="card">
    <h2 style="font-size: 15px; font-weight: 600; margin-bottom: 8px;">Project API Keys</h2>
    <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 24px;">
      Your API keys allow you to authenticate with the OmniBase backend from your client apps or server.
    </p>

    <!-- Anon Key -->
    <div style="margin-bottom: 24px;">
      <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
        <span class="badge badge-success">anon</span>
        <span class="badge badge-neutral">public</span>
      </div>
      <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 8px;">
        This key is safe to use in a browser if you have enabled Row Level Security (RLS) for your tables and configured policies.
      </p>
      <div style="display: flex; gap: 8px;">
        <input type="password" class="input" value={anonKey} readonly style="flex: 1; font-family: var(--font-mono);" />
        <button class="btn btn-secondary" on:click={() => copy(anonKey)}>Copy</button>
      </div>
    </div>

    <hr style="border: 0; border-top: 1px solid var(--border-subtle); margin: 24px 0;" />

    <!-- Service Role Key -->
    <div>
      <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
        <span class="badge badge-error" style="background: rgba(255,82,82,0.1); color: #ff5252;">service_role</span>
        <span class="badge badge-neutral">secret</span>
      </div>
      <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 8px;">
        This key has the ability to bypass Row Level Security. <strong>Never</strong> share it publicly or use it in the browser.
      </p>
      <div style="display: flex; gap: 8px;">
        <input type="password" class="input" value={serviceKey} readonly style="flex: 1; font-family: var(--font-mono);" />
        <button class="btn btn-secondary" on:click={() => copy(serviceKey)}>Copy</button>
      </div>
    </div>
  </div>
</div>
