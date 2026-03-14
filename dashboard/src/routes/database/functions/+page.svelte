<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  interface DBFunction {
    name: string
    schema: string
    return_type: string
    language: string
    arguments: string
    definition: string
  }

  let functions = $state<DBFunction[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let currentSchema = $state('public')
  let selectedFn = $state<DBFunction | null>(null)
  let searchQuery = $state('')
  let copyStatus = $state<string | null>(null)

  const filteredFunctions = $derived(
    functions.filter(f => f.name.toLowerCase().includes(searchQuery.toLowerCase()))
  )

  function showCopyToast() {
    copyStatus = 'Definition copied!'
    setTimeout(() => copyStatus = null, 2000)
  }

  async function loadFunctions() {
    try {
      loading = true
      error = null
      const resp = await fetch(`${getOmniBaseUrl()}/pg/functions?schema=${currentSchema}`, {
        headers: getHeaders()
      })
      if (resp.ok) {
        functions = await resp.json()
      } else {
        error = 'Failed to load functions'
      }
    } catch {
      error = 'Connection error'
    } finally {
      loading = false
    }
  }

  onMount(() => {
    currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
    loadFunctions()

    const handleSchemaChange = (e: any) => {
      currentSchema = e.detail
      selectedFn = null
      loadFunctions()
    }

    window.addEventListener('omnibase:schema-change', handleSchemaChange)
    return () => window.removeEventListener('omnibase:schema-change', handleSchemaChange)
  })
</script>

<svelte:head>
  <title>Database Functions — OmniBase</title>
</svelte:head>

<div class="functions-layout animate-fade-in">
  <!-- Sidebar -->
  <aside class="fn-sidebar">
    <div class="sidebar-search">
      <div class="search-input-wrapper">
         <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="search-icon"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
         <input type="text" class="search-input" placeholder="Filter functions..." bind:value={searchQuery} />
      </div>
    </div>

    <div class="fn-list">
      {#if loading}
        {#each Array(6) as _, i}
          <div class="fn-skeleton" style="animation-delay: {i * 0.05}s"></div>
        {/each}
      {:else if filteredFunctions.length === 0}
        <div class="empty-list">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
          <p>{searchQuery ? 'No matches found' : `No functions in ${currentSchema}`}</p>
        </div>
      {:else}
        {#each filteredFunctions as fn}
          <button 
            class="fn-item {selectedFn?.name === fn.name ? 'active' : ''}"
            onclick={() => selectedFn = fn}
          >
            <div class="fn-item-header">
               <span class="fn-name">{fn.name}</span>
               <span class="fn-lang-tag {fn.language === 'plpgsql' ? 'lang-special' : ''}">{fn.language}</span>
            </div>
            <div class="fn-item-meta">
               <span class="return-type">{fn.return_type}</span>
            </div>
          </button>
        {/each}
      {/if}
    </div>
  </aside>

  <!-- Content -->
  <main class="fn-main">
    {#if !selectedFn}
      <div class="fn-empty-state">
        <div class="hero-icon">
          <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"/></svg>
        </div>
        <h2>Select a Function</h2>
        <p>Explore your database routines, triggers, and procedures.</p>
        <button class="btn btn-primary" onclick={() => window.location.href = '/database/editor'}>Create New Function</button>
      </div>
    {:else}
      <header class="fn-detail-header">
        <div class="fn-title-area">
           <div class="fn-label">DATABASE FUNCTION</div>
           <h1 class="fn-display-name">{selectedFn.name}</h1>
           <div class="fn-identity">
              <span class="fn-badge">{selectedFn.language.toUpperCase()}</span>
              <span class="fn-badge return-badge">RETURNS {selectedFn.return_type.toUpperCase()}</span>
           </div>
        </div>
        <div class="fn-actions">
           <button class="action-pill" onclick={() => { navigator.clipboard.writeText(selectedFn?.definition || ''); showCopyToast(); }}>
             {#if copyStatus}
               <span style="color: var(--brand-green);">{copyStatus}</span>
             {:else}
               <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
               Copy Definition
             {/if}
           </button>
           <button class="action-pill highlight" onclick={() => window.location.href = `/database/editor?query=${encodeURIComponent(selectedFn?.definition || '')}`}>
             <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
             Edit in Editor
           </button>
        </div>
      </header>

      <section class="fn-params-sec">
         <div class="params-header">ARGUMENTS</div>
         <div class="params-content">
            {selectedFn.arguments || 'None'}
         </div>
      </section>

      <section class="fn-code-sec">
         <div class="code-wrapper">
            <div class="code-gutter">
               {#each (selectedFn.definition.trim().split('\n')) as _, i}
                 <div class="line-num">{i + 1}</div>
               {/each}
            </div>
            <pre class="code-pre"><code>{selectedFn.definition.trim()}</code></pre>
         </div>
      </section>
    {/if}
  </main>
</div>

<style>
  .functions-layout {
    display: grid;
    grid-template-columns: 320px 1fr;
    height: calc(100vh - var(--topbar-height));
    background: var(--bg-surface);
    overflow: hidden;
  }

  /* Sidebar */
  .fn-sidebar {
    background: #0d0d12;
    border-right: 1px solid var(--border-subtle);
    display: flex; flex-direction: column; overflow: hidden;
  }
  .sidebar-search { padding: 20px; border-bottom: 1px solid var(--border-subtle); }
  .search-input-wrapper { position: relative; display: flex; align-items: center; }
  .search-icon { position: absolute; left: 12px; color: #444; }
  .search-input { width: 100%; background: #08080a; border: 1px solid #222; padding: 10px 12px 10px 36px; border-radius: 8px; color: #eee; font-size: 13px; outline: none; transition: border-color 0.2s; }
  .search-input:focus { border-color: var(--brand-primary); }

  .fn-list { flex: 1; overflow-y: auto; padding: 12px; }
  .fn-item {
    width: 100%; background: transparent; border: none; padding: 14px 16px; border-radius: 10px; cursor: pointer; text-align: left; margin-bottom: 6px; transition: all 0.2s;
  }
  .fn-item:hover { background: rgba(255,255,255,0.03); }
  .fn-item.active { background: rgba(108, 71, 255, 0.1); }
  
  .fn-item-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
  .fn-name { font-family: var(--font-mono); font-size: 13px; font-weight: 600; color: #bbb; transition: color 0.2s; }
  .fn-item.active .fn-name { color: var(--brand-primary); }
  
  .fn-lang-tag { font-size: 9px; padding: 2px 6px; background: #222; color: #666; border-radius: 4px; font-weight: 700; text-transform: uppercase; }
  .fn-lang-tag.lang-special { color: var(--brand-green); background: rgba(0, 230, 118, 0.1); }
  
  .fn-item-meta { font-size: 11px; color: #555; font-family: var(--font-mono); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

  /* Content Area */
  .fn-main { background: #07070a; overflow-y: auto; display: flex; flex-direction: column; }
  .fn-empty-state { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px; color: var(--text-muted); text-align: center; }
  .hero-icon { opacity: 0.1; margin-bottom: 12px; color: var(--brand-primary); }
  .fn-empty-state h2 { color: #fff; margin: 0; font-weight: 600; font-size: 24px; }

  .fn-detail-header { padding: 48px; border-bottom: 1px solid var(--border-subtle); display: flex; justify-content: space-between; align-items: flex-end; }
  .fn-label { font-size: 10px; font-weight: 700; color: #444; letter-spacing: 0.12em; margin-bottom: 12px; }
  .fn-display-name { font-family: var(--font-mono); font-size: 32px; color: #fff; margin: 0 0 20px; letter-spacing: -0.02em; }
  .fn-identity { display: flex; gap: 8px; }
  .fn-badge { background: #1a1a24; color: #888; padding: 5px 12px; border-radius: 6px; font-size: 10px; font-weight: 700; font-family: var(--font-sans); border: 1px solid rgba(255,255,255,0.03); }
  .return-badge { color: var(--brand-green); background: rgba(0, 230, 118, 0.05); border-color: rgba(0, 230, 118, 0.1); }

  .fn-actions { display: flex; gap: 12px; }
  .action-pill { background: transparent; border: 1px solid #222; color: #888; font-size: 12px; font-weight: 600; padding: 10px 20px; border-radius: 10px; cursor: pointer; display: flex; align-items: center; gap: 10px; transition: all 0.2s; }
  .action-pill:hover { border-color: #444; color: #fff; background: rgba(255,255,255,0.02); }
  .action-pill.highlight { border-color: var(--brand-primary); color: var(--brand-primary); box-shadow: 0 0 20px rgba(108, 71, 255, 0.1); }
  .action-pill.highlight:hover { background: rgba(108, 71, 255, 0.1); }

  .fn-params-sec { padding: 32px 48px; background: rgba(255,255,255,0.01); border-bottom: 1px solid var(--border-subtle); }
  .params-header { font-size: 11px; font-weight: 700; color: #444; margin-bottom: 14px; letter-spacing: 0.05em; }
  .params-content { font-family: var(--font-mono); font-size: 13px; color: #777; line-height: 1.6; }

  .fn-code-sec { padding: 48px; }
  .code-wrapper { background: #020204; border-radius: 16px; display: flex; overflow: hidden; border: 1px solid rgba(255,255,255,0.03); box-shadow: 0 4px 30px rgba(0,0,0,0.5); }
  .code-gutter { background: #08080a; padding: 32px 14px; border-right: 1px solid #1a1a24; display: flex; flex-direction: column; align-items: flex-end; user-select: none; }
  .line-num { font-family: var(--font-mono); font-size: 11px; color: #2a2a35; height: 21px; line-height: 21px; }
  .code-pre { margin: 0; padding: 32px; flex: 1; overflow-x: auto; font-family: var(--font-mono); font-size: 14px; color: #bbb; line-height: 21px; }

  .fn-skeleton { height: 74px; background: rgba(255,255,255,0.02); margin-bottom: 10px; border-radius: 12px; animation: pulse 1.5s infinite; }
  @keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; } }

  .empty-list { padding: 80px 20px; text-align: center; color: #444; font-size: 13px; }
  .empty-list p { margin-top: 14px; }
</style>
