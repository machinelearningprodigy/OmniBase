<script>
  import { onMount } from 'svelte'
  import { getOmniBaseUrl, getHeaders } from '$lib/api'
  // ⚡ FIX: Using lucide-svelte (NOT lucide-react) for Svelte 5 compatibility
  import { Shield, Key, Eye, EyeOff, Copy, Check, RefreshCw, Zap } from 'lucide-svelte'

  let keys = $state([])
  let loading = $state(true)
  let error = $state(null)
  let copiedType = $state(null)
  let showTypes = $state({ anon: false, service: false })

  async function fetchKeys() {
    loading = true
    error = null
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/admin/config/keys`, {
        headers: getHeaders()
      })
      if (resp.ok) {
        const data = await resp.json()
        keys = [
          { 
            name: 'Anon (Public) Key', 
            key: data.anon_key, 
            type: 'anon',
            desc: 'Safe to use in browser code. Row Level Security enforced. Never bypass RLS with this key.' 
          },
          { 
            name: 'Service Key (Private)', 
            key: data.service_key, 
            type: 'service',
            desc: 'Bypasses RLS. Never expose in client code.' 
          }
        ]
      } else {
        error = 'Failed to fetch API keys. Is the Auth service running?'
      }
    } catch (e) {
      error = 'Connection failed: ' + e.message
    } finally {
      loading = false
    }
  }

  function toggleShow(type) {
    showTypes[type] = !showTypes[type]
  }

  async function copyToClipboard(text, type) {
    try {
      await navigator.clipboard.writeText(text)
      copiedType = type
      setTimeout(() => { copiedType = null }, 2000)
    } catch (err) {
      console.error('Failed to copy!', err)
    }
  }

  onMount(fetchKeys)
</script>

<svelte:head>
  <title>API Keys — OmniBase</title>
</svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">API Keys</h1>
    <p class="page-subtitle">Manage project API keys — anon key, service key, and custom scoped keys</p>
  </div>
  <button class="btn btn-primary" onclick={fetchKeys}>
    <RefreshCw size={14} style="margin-right: 6px" /> 
    Refresh Keys
  </button>
</div>

<div class="page-content">
  <div style="margin-bottom: 2rem; padding: 1rem; background: rgba(176, 251, 92, 0.05); border-radius: 8px; border: 1px solid rgba(176, 251, 92, 0.1); display: flex; align-items: center; gap: 12px;">
    <Zap size={20} color="var(--brand-primary)" />
    <div>
      <p style="font-weight: 600; color: var(--brand-primary); font-size: 0.875rem;">Identity Provider Active</p>
      <p style="font-size: 0.75rem; opacity: 0.7;">These keys are used by the SDK to authenticate requests.</p>
    </div>
  </div>

  {#if loading}
    <div style="display:flex;justify-content:center;padding:40px">
      <div class="spinner"></div>
    </div>
  {:else if error}
    <div class="card" style="border-color:var(--error);background:rgba(255,82,82,0.05)">
      <p style="color:var(--error);font-weight:600">Error: {error}</p>
      <p style="font-size: 0.75rem; color: var(--text-secondary); margin-top: 0.5rem;">Ensure both the Gateway and Auth service are running.</p>
    </div>
  {:else}
    <div style="display:flex;flex-direction:column;gap:16px">
      {#each keys as k}
        <div class="card">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px">
            <div style="display:flex;align-items:center;gap:8px">
              {#if k.type === 'service'}
                <Shield size={16} color="#fbbf24" />
              {:else}
                <Key size={16} color="#34d399" />
              {/if}
              <span style="font-weight:700;font-size:14px">{k.name}</span>
            </div>
            <span class="badge {k.type==='anon'?'badge-success':'badge-warning'}">{k.type}</span>
          </div>
          <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">{k.desc}</p>
          
          <div style="display:flex;gap:8px">
            <div style="position:relative;flex:1">
              <input 
                class="input input-mono" 
                value={k.key} 
                type={showTypes[k.type] ? "text" : "password"} 
                style="width:100% !important; font-size:11px; padding-right:40px; background: rgba(0,0,0,0.2) !important;" 
                readonly 
              />
            </div>
            
            <button 
              class="btn btn-secondary btn-sm" 
              onclick={() => toggleShow(k.type)}
              title={showTypes[k.type] ? "Hide key" : "Show key"}
              style="min-width: 44px; justify-content: center;"
            >
              {#if showTypes[k.type]}
                <EyeOff size={14} />
              {:else}
                <Eye size={14} />
              {/if}
            </button>

            <button 
              class="btn btn-secondary btn-sm" 
              onclick={() => copyToClipboard(k.key, k.type)}
              style="min-width: 90px"
            >
              {#if copiedType === k.type}
                <Check size={14} style="color: #34d399" />
                <span style="margin-left:6px; color: #34d399">Copied</span>
              {:else}
                <Copy size={14} />
                <span style="margin-left:6px">Copy Key</span>
              {/if}
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .badge-success { background: rgba(52, 211, 153, 0.1); color: #34d399; border: 1px solid rgba(52, 211, 153, 0.2); }
  .badge-warning { background: rgba(251, 191, 36, 0.1); color: #fbbf24; border: 1px solid rgba(251, 191, 36, 0.2); }
  
  .spinner {
    width: 24px;
    height: 24px;
    border: 3px solid rgba(255,255,255,0.1);
    border-top-color: var(--brand-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
