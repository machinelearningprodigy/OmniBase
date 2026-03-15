<script lang="ts">
  import { onMount } from 'svelte'
  import { getHeaders, getOmniBaseUrl, getErrorMessage } from '$lib/api'

  interface OmniFunction {
    id: string
    name: string
    slug: string
    runtime: 'static-json' | 'webhook' | 'javascript' | 'python'
    verify_jwt: boolean
    created_at: string
    updated_at: string
    source: Record<string, any>
  }

  const runtimeTemplates: Record<OmniFunction['runtime'], any> = {
    'static-json': { status: 200, body: { ok: true, message: "Hello from OmniBase" } },
    'webhook': { url: "https://echo.free.beeceptor.com", method: "POST", headers: { "X-Custom-Header": "OmniBase" } },
    'javascript': { 
      timeout_seconds: 15, 
      handler: "main",
      code: `// Process incoming requests with JavaScript\nasync function main(payload) {\n  console.log("Receiving payload:", payload);\n  return {\n    message: "Hello from OmniJS!",\n    timestamp: new Date().toISOString(),\n    received: payload\n  };\n}` 
    },
    'python': { 
      timeout_seconds: 15, 
      handler: "handler",
      code: `import datetime\n\ndef handler(payload):\n    print(f"Python processing: {payload}")\n    return {\n        "message": "Hello from OmniPy!",\n        "time": str(datetime.datetime.now()),\n        "input": payload\n    }` 
    },
  }

  let functions = $state<OmniFunction[]>([])
  let selectedFunction = $state<OmniFunction | null>(null)
  let loading = $state(true)
  let error = $state<string | null>(null)
  let showCreateModal = $state(false)
  let saving = $state(false)
  let activeTab = $state<'editor' | 'config' | 'test' | 'logs'>('editor')

  // Form for creation
  let createForm = $state({
    name: '',
    slug: '',
    runtime: 'javascript' as OmniFunction['runtime'],
  })

  // Edit draft
  let editDraft = $state<Partial<OmniFunction> | null>(null)
  let testPayload = $state('{\n  "name": "OmniBase Tester"\n}')
  let testResult = $state<any>(null)
  let testing = $state(false)
  let logs = $state<any[]>([])
  let loadingLogs = $state(false)

  async function loadLogs(slug: string) {
    try {
      loadingLogs = true
      const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/${slug}/logs`, { headers: getHeaders() })
      const data = await resp.json()
      if (resp.ok) logs = data
    } catch (err) {} finally {
      loadingLogs = false
    }
  }

  async function loadFunctions() {
    try {
      loading = true
      const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/`, { headers: getHeaders(false) })
      const data = await resp.json()
      if (!resp.ok) throw new Error(data.error || 'Failed to load functions')
      functions = data
      error = null
    } catch (err: any) {
      error = err.message || 'Connection error'
    } finally {
      loading = false
    }
  }

  function selectFunction(fn: OmniFunction) {
    selectedFunction = fn
    editDraft = JSON.parse(JSON.stringify(fn))
    activeTab = 'editor'
    testResult = null
    logs = []
  }

  $effect(() => {
    if (activeTab === 'logs' && selectedFunction) {
      loadLogs(selectedFunction.slug)
    }
  })

  function applyRuntimeTemplate(runtime: OmniFunction['runtime']) {
    createForm.runtime = runtime
  }

  async function createFunction() {
    try {
      saving = true
      const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({
          ...createForm,
          verify_jwt: true,
          source: runtimeTemplates[createForm.runtime],
        }),
      })
      const data = await resp.json()
      if (!resp.ok) throw new Error(data.error || 'Failed to create function')
      
      showCreateModal = false
      await loadFunctions()
      // Automatically select the new function
      const newFn = functions.find(f => f.slug === createForm.slug)
      if (newFn) selectFunction(newFn)
      
      createForm = { name: '', slug: '', runtime: 'javascript' }
    } catch (err: any) {
      error = err.message || 'Failed to create function'
    } finally {
      saving = false
    }
  }

  async function saveChanges() {
    if (!selectedFunction || !editDraft) return
    try {
      saving = true
      const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/`, {
        method: 'POST', // Backend uses upsert (POST to /)
        headers: getHeaders(),
        body: JSON.stringify({
          name: editDraft.name,
          slug: editDraft.slug,
          runtime: editDraft.runtime,
          verify_jwt: editDraft.verify_jwt,
          source: typeof editDraft.source === 'string' ? JSON.parse(editDraft.source) : editDraft.source,
        }),
      })
      if (!resp.ok) {
        const data = await resp.json()
        throw new Error(data.error || 'Failed to save changes')
      }
      await loadFunctions()
      // Refresh selected
      const updated = functions.find(f => f.slug === editDraft!.slug)
      if (updated) {
        selectedFunction = updated
        editDraft = JSON.parse(JSON.stringify(updated))
      }
    } catch (err: any) {
      error = err.message || 'Failed to save changes'
    } finally {
      saving = false
    }
  }

  async function deleteFn(slug: string) {
    if (!confirm(`Are you sure you want to delete function ${slug}?`)) return
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/${slug}`, {
        method: 'DELETE',
        headers: getHeaders(false),
      })
      if (resp.ok || resp.status === 204) {
        selectedFunction = null
        editDraft = null
        await loadFunctions()
      }
    } catch (err: any) {
      error = 'Failed to delete function'
    }
  }

  async function runTest() {
    if (!selectedFunction) return
    try {
      testing = true
      testResult = null
      let payload = {}
      try {
        payload = JSON.parse(testPayload)
      } catch {
        error = "Invalid JSON payload for testing"
        return
      }

      const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/${selectedFunction.slug}/invoke`, {
        method: 'POST',
        headers: getHeaders(false), // Most functions use JWT, but admin can bypass? actually backend expects it if verify_jwt is true
        body: JSON.stringify(payload)
      })
      
      const data = await resp.json().catch(() => null)
      testResult = {
        status: resp.status,
        headers: Object.fromEntries(resp.headers.entries()),
        body: data || 'No response body'
      }
    } catch (err: any) {
      testResult = { error: err.message || 'Failed to invoke function' }
    } finally {
      testing = false
    }
  }

  onMount(loadFunctions)
</script>

<svelte:head>
  <title>Edge Functions — OmniBase</title>
</svelte:head>

<div class="functions-layout">
  <!-- Sidebar -->
  <aside class="sidebar-list">
    <div class="sidebar-header">
      <h3>Edge Functions</h3>
      <button class="btn btn-primary btn-xs" onclick={() => showCreateModal = true}>+</button>
    </div>
    <div class="sidebar-search">
       <input type="text" placeholder="Filter functions..." class="input" />
    </div>
    <div class="list-container">
      {#each functions as fn}
        <button 
          class="fn-item {selectedFunction?.id === fn.id ? 'active' : ''}"
          onclick={() => selectFunction(fn)}
        >
          <div class="fn-item-icon {fn.runtime}">
             {fn.runtime === 'javascript' ? 'js' : fn.runtime === 'python' ? 'py' : 'json'}
          </div>
          <div class="fn-item-meta">
            <span class="fn-name">{fn.name}</span>
            <span class="fn-slug">/{fn.slug}</span>
          </div>
        </button>
      {/each}
      {#if functions.length === 0 && !loading}
        <div class="empty-list">No functions found</div>
      {/if}
    </div>
  </aside>

  <!-- Main View -->
  <main class="function-editor">
    {#if selectedFunction && editDraft}
      <div class="editor-header">
        <div class="header-info">
          <div class="badge-row">
            <span class="badge badge-info">{selectedFunction.runtime}</span>
            <span class="badge {selectedFunction.verify_jwt ? 'badge-success' : 'badge-neutral'}">
              {selectedFunction.verify_jwt ? 'JWT Restricted' : 'Public'}
            </span>
          </div>
          <h2>{selectedFunction.name}</h2>
          <code class="endpoint-url">{getOmniBaseUrl()}/functions/v1/{selectedFunction.slug}/invoke</code>
        </div>
        <div class="header-actions">
           <button class="btn btn-secondary btn-sm" onclick={() => deleteFn(selectedFunction!.slug)}>Delete</button>
           <button class="btn btn-primary btn-sm" onclick={saveChanges} disabled={saving}>
             {saving ? 'Saving...' : 'Save Changes'}
           </button>
        </div>
      </div>

      <div class="tabs-nav">
        <button class="tab-link {activeTab === 'editor' ? 'active' : ''}" onclick={() => activeTab = 'editor'}>Source Code</button>
        <button class="tab-link {activeTab === 'config' ? 'active' : ''}" onclick={() => activeTab = 'config'}>Configuration</button>
        <button class="tab-link {activeTab === 'test' ? 'active' : ''}" onclick={() => activeTab = 'test'}>Test / Invoke</button>
        <button class="tab-link {activeTab === 'logs' ? 'active' : ''}" onclick={() => activeTab = 'logs'}>Invocation Logs</button>
      </div>

      <div class="tab-content">
        {#if activeTab === 'editor'}
          <div class="code-editor-container">
             {#if editDraft.runtime === 'javascript' || editDraft.runtime === 'python'}
               <textarea 
                 class="code-textarea" 
                 spellcheck="false" 
                 bind:value={editDraft.source.code}
               ></textarea>
             {:else if editDraft.runtime === 'static-json'}
                <div class="json-editor-grid">
                  <div class="field">
                    <label>HTTP Status</label>
                    <input type="number" class="input" bind:value={editDraft.source.status} />
                  </div>
                  <div class="field">
                    <label>JSON Response Body</label>
                    <textarea class="input" style="font-family: monospace;" rows="15" 
                      value={JSON.stringify(editDraft.source.body, null, 2)}
                      onchange={(e) => {
                        try { editDraft!.source!.body = JSON.parse((e.currentTarget as HTMLTextAreaElement).value); } catch(err) {}
                      }}
                    ></textarea>
                  </div>
                </div>
             {:else if editDraft.runtime === 'webhook'}
                <div class="webhook-form" style="display: grid; gap: 16px;">
                   <div class="field">
                     <label>Target URL</label>
                     <input type="text" class="input" bind:value={editDraft.source.url} />
                   </div>
                   <div class="field">
                     <label>HTTP Method</label>
                     <select class="input" bind:value={editDraft.source.method}>
                       <option value="GET">GET</option>
                       <option value="POST">POST</option>
                       <option value="PUT">PUT</option>
                       <option value="DELETE">DELETE</option>
                     </select>
                   </div>
                </div>
             {/if}
          </div>
        {:else if activeTab === 'config'}
          <div class="config-grid" style="display: grid; gap: 20px; max-width: 600px;">
            <div class="field">
              <label>Function Name</label>
              <input type="text" class="input" bind:value={editDraft.name} />
            </div>
            <div class="field">
              <label>Slug (URL segment)</label>
              <input type="text" class="input" bind:value={editDraft.slug} />
            </div>
            <div class="field">
              <label>Handler Function</label>
              <input type="text" class="input" bind:value={editDraft.source.handler} placeholder="main" />
            </div>
            <div class="field">
              <label>Timeout (seconds)</label>
              <input type="number" class="input" bind:value={editDraft.source.timeout_seconds} />
            </div>
            <label class="checkbox-label">
              <input type="checkbox" bind:checked={editDraft.verify_jwt} />
              <span>Verify client JWT before execution</span>
            </label>
          </div>
        {:else if activeTab === 'test'}
          <div class="tester-layout">
            <div class="payload-panel">
               <h4>Invocation Payload (JSON)</h4>
               <textarea class="input" style="font-family: monospace; height: 160px;" bind:value={testPayload}></textarea>
               <button class="btn btn-primary" style="margin-top: 12px; width: 100%;" onclick={runTest} disabled={testing}>
                 {testing ? 'Invoking...' : 'Invoke Function'}
               </button>
            </div>
            <div class="result-panel">
               <h4>Result</h4>
               {#if testResult}
                 <div class="result-display">
                    <div class="result-meta">
                      <span class="badge {testResult.status < 400 ? 'badge-success' : 'badge-error'}">
                        Status {testResult.status}
                      </span>
                    </div>
                    <pre class="result-json"><code>{JSON.stringify(testResult.body, null, 2)}</code></pre>
                 </div>
               {:else}
                 <div class="empty-result">Set payload and invoke to see output</div>
               {/if}
            </div>
          </div>
        {:else if activeTab === 'logs'}
          <div class="logs-container">
            <div class="logs-header" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
               <h4 style="margin: 0;">Recent Invocations</h4>
               <button class="btn btn-secondary btn-xs" onclick={() => loadLogs(selectedFunction!.slug)} disabled={loadingLogs}>
                 {loadingLogs ? 'Refreshing...' : 'Refresh Logs'}
               </button>
            </div>
            {#if logs.length === 0}
              <div class="empty-logs" style="text-align: center; padding: 48px; color: var(--text-muted); border: 1px dashed var(--border-subtle); border-radius: 8px;">
                No execution logs found for this function yet.
              </div>
            {:else}
              <div class="logs-table-wrapper" style="border: 1px solid var(--border-subtle); border-radius: 8px; overflow: hidden;">
                <table class="logs-table" style="width: 100%; border-collapse: collapse;">
                  <thead style="background: #0d0d12;">
                    <tr>
                      <th style="padding: 12px; text-align: left; font-size: 11px; color: #888;">Status</th>
                      <th style="padding: 12px; text-align: left; font-size: 11px; color: #888;">Duration</th>
                      <th style="padding: 12px; text-align: left; font-size: 11px; color: #888;">Output Preview</th>
                      <th style="padding: 12px; text-align: left; font-size: 11px; color: #888;">Time</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each logs as log}
                      <tr style="border-top: 1px solid var(--border-subtle);">
                        <td style="padding: 12px;">
                           <span class="badge {log.status < 400 ? 'badge-success' : 'badge-error'}">{log.status}</span>
                        </td>
                        <td style="padding: 12px; font-family: monospace; font-size: 13px;">{log.duration_ms}ms</td>
                        <td style="padding: 12px; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; color: #eee; font-family: monospace;">
                          {log.error || log.output || '-'}
                        </td>
                        <td style="padding: 12px; font-size: 12px; color: #888;">{new Date(log.executed_at).toLocaleString()}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {:else if loading}
       <div class="centered-loading">
         <div class="skeleton" style="width: 100%; height: 300px; border-radius: 12px;"></div>
       </div>
    {:else}
       <div class="welcome-view">
         <div class="welcome-icon">λ</div>
         <h2>Select a function to begin</h2>
         <p>Create and deploy backend logic that executes in isolated runtimes.</p>
         <button class="btn btn-primary" onclick={() => showCreateModal = true}>Create New Function</button>
       </div>
    {/if}
  </main>
</div>

{#if showCreateModal}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 480px;">
      <div class="modal-header">
        <h3>Deploy New Function</h3>
        <button class="btn-close" onclick={() => showCreateModal = false}>x</button>
      </div>
      <div class="modal-body" style="display: grid; gap: 16px;">
        <div class="field">
          <label>Display Name</label>
          <input class="input" bind:value={createForm.name} placeholder="e.g. Charge Customer" />
        </div>
        <div class="field">
          <label>Slug</label>
          <input class="input" bind:value={createForm.slug} placeholder="charge-customer" />
        </div>
        <div class="field">
          <label>Runtime</label>
          <div class="runtime-selector">
            {#each ['javascript', 'python', 'webhook', 'static-json'] as r}
              <button 
                class="runtime-opt {createForm.runtime === r ? 'active' : ''}"
                onclick={() => applyRuntimeTemplate(r as OmniFunction['runtime'])}
              >
                {r}
              </button>
            {/each}
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateModal = false}>Cancel</button>
        <button class="btn btn-primary" onclick={createFunction} disabled={saving}>
          {saving ? 'Creating...' : 'Deploy Function'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .functions-layout {
    height: calc(100vh - var(--topbar-height));
    display: grid;
    grid-template-columns: 320px 1fr;
    background: var(--bg-surface);
  }

  /* Sidebar */
  .sidebar-list {
    border-right: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    background: #0d0d12;
  }
  .sidebar-header {
    padding: 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .sidebar-header h3 { margin: 0; font-size: 16px; font-weight: 700; }
  .sidebar-search { padding: 0 20px 16px; }
  .list-container { flex: 1; overflow-y: auto; padding: 0 12px 12px; }
  
  .fn-item {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: transparent;
    border: none;
    border-radius: 8px;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: 4px;
  }
  .fn-item:hover { background: rgba(255,255,255,0.03); }
  .fn-item.active { background: rgba(108, 71, 255, 0.1); }
  
  .fn-item-icon {
    width: 32px; height: 32px; border-radius: 6px;
    display: grid; place-items: center;
    font-size: 10px; font-weight: 800; text-transform: uppercase;
    background: rgba(255,255,255,0.05); color: #888;
  }
  .fn-item-icon.javascript { color: #f7df1e; background: rgba(247, 223, 30, 0.1); }
  .fn-item-icon.python { color: #3776ab; background: rgba(55, 118, 171, 0.1); }
  .fn-item-icon.webhook { color: #4ade80; background: rgba(74, 222, 128, 0.1); }
  
  .fn-item-meta { display: flex; flex-direction: column; }
  .fn-name { font-size: 13px; font-weight: 600; color: #eee; }
  .fn-slug { font-size: 11px; color: #666; font-family: monospace; }
  .fn-item.active .fn-name { color: var(--brand-primary); }

  /* Editor Main */
  .function-editor { display: flex; flex-direction: column; height: 100%; overflow: hidden; background: #08080c; }
  .editor-header { padding: 24px 32px; border-bottom: 1px solid var(--border-subtle); display: flex; justify-content: space-between; align-items: flex-start; }
  .header-info h2 { margin: 8px 0 4px; font-size: 24px; }
  .endpoint-url { font-size: 12px; color: var(--text-muted); background: rgba(0,0,0,0.3); padding: 4px 8px; border-radius: 4px; }
  .badge-row { display: flex; gap: 8px; }
  .header-actions { display: flex; gap: 12px; }

  .tabs-nav { display: flex; padding: 0 32px; border-bottom: 1px solid var(--border-subtle); background: #0d0d12; }
  .tab-link {
    padding: 14px 20px; background: transparent; border: none; border-bottom: 2px solid transparent;
    color: var(--text-muted); font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s;
  }
  .tab-link:hover { color: #fff; }
  .tab-link.active { color: var(--brand-primary); border-bottom-color: var(--brand-primary); }

  .tab-content { flex: 1; overflow-y: auto; padding: 32px; }
  
  .code-editor-container { height: 100%; display: flex; flex-direction: column; }
  .code-textarea {
    flex: 1; min-height: 480px; background: #050508; border: 1px solid var(--border-subtle);
    border-radius: 8px; padding: 20px; color: #86efac; font-family: 'JetBrains Mono', monospace; font-size: 14px;
    line-height: 1.6; outline: none; resize: none;
  }

  .field label { display: block; font-size: 12px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; margin-bottom: 8px; letter-spacing: 0.05em; }
  .checkbox-label { display: flex; align-items: center; gap: 10px; font-size: 14px; color: #ccc; cursor: pointer; }

  /* Tester */
  .tester-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 32px; height: 100%; }
  .tester-layout h4 { margin: 0 0 12px; font-size: 14px; color: var(--text-muted); }
  .result-display { background: #050508; border: 1px solid var(--border-subtle); border-radius: 8px; display: flex; flex-direction: column; height: 400px; }
  .result-meta { padding: 12px; border-bottom: 1px solid var(--border-subtle); }
  .result-json { flex: 1; margin: 0; padding: 16px; overflow: auto; font-size: 13px; color: #fff; }
  .empty-result { height: 300px; display: grid; place-items: center; color: var(--text-muted); font-style: italic; border: 1px dashed var(--border-subtle); border-radius: 8px; }

  .welcome-view { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; text-align: center; padding: 40px; }
  .welcome-icon { font-size: 64px; margin-bottom: 24px; color: var(--brand-primary); opacity: 0.5; }
  .welcome-view h2 { margin-bottom: 12px; }
  .welcome-view p { color: var(--text-muted); max-width: 400px; margin-bottom: 24px; }

  .runtime-selector { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
  .runtime-opt {
    padding: 10px; background: rgba(255,255,255,0.03); border: 1px solid var(--border-subtle);
    border-radius: 6px; color: #aaa; font-size: 12px; font-weight: 600; cursor: pointer; transition: all 0.2s;
  }
  .runtime-opt:hover { background: rgba(255,255,255,0.06); color: #fff; }
  .runtime-opt.active { background: rgba(108, 71, 255, 0.1); border-color: var(--brand-primary); color: #fff; }

  .centered-loading { padding: 40px; }
</style>
