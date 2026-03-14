<script lang="ts">
  import { onMount } from 'svelte'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  interface OmniFunction {
    id: string
    name: string
    slug: string
    runtime: 'static-json' | 'webhook' | 'javascript' | 'python'
    verify_jwt: boolean
    created_at: string
    updated_at: string
    source: Record<string, unknown>
  }

  const runtimeTemplates: Record<OmniFunction['runtime'], string> = {
    'static-json': '{\n  "status": 200,\n  "body": {\n    "ok": true\n  }\n}',
    webhook: '{\n  "url": "http://internal-service.local/endpoint",\n  "method": "POST",\n  "forward_headers": true\n}',
    javascript:
      '{\n  "timeout_seconds": 15,\n  "code": "const input = JSON.parse(require(\'fs\').readFileSync(0, \'utf8\'));\\nconsole.log(JSON.stringify({ status: 200, body: { runtime: \'javascript\', input } }));"\n}',
    python:
      '{\n  "timeout_seconds": 15,\n  "code": "import json, sys\\ninput = json.load(sys.stdin)\\nprint(json.dumps({\\"status\\": 200, \\"body\\": {\\"runtime\\": \\"python\\", \\"input\\": input}}))"\n}',
  }

  let functions = $state<OmniFunction[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let showCreateModal = $state(false)
  let saving = $state(false)

  let form = $state({
    name: '',
    slug: '',
    runtime: 'static-json' as OmniFunction['runtime'],
    verify_jwt: true,
    source: runtimeTemplates['static-json'],
  })

  function applyRuntimeTemplate(runtime: OmniFunction['runtime']) {
    form = {
      ...form,
      runtime,
      source: runtimeTemplates[runtime],
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
      functions = []
    } finally {
      loading = false
    }
  }

  async function createFunction() {
    try {
      saving = true
      const source = JSON.parse(form.source)
      const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({
          name: form.name,
          slug: form.slug,
          runtime: form.runtime,
          verify_jwt: form.verify_jwt,
          source,
        }),
      })
      const data = await resp.json()
      if (!resp.ok) throw new Error(data.error || 'Failed to save function')
      showCreateModal = false
      form = {
        name: '',
        slug: '',
        runtime: 'static-json',
        verify_jwt: true,
        source: runtimeTemplates['static-json'],
      }
      await loadFunctions()
    } catch (err: any) {
      error = err.message || 'Failed to save function'
    } finally {
      saving = false
    }
  }

  async function deleteFunction(slug: string) {
    const resp = await fetch(`${getOmniBaseUrl()}/functions/v1/${slug}`, {
      method: 'DELETE',
      headers: getHeaders(false),
    })
    if (resp.ok || resp.status === 204) {
      await loadFunctions()
      return
    }
    const data = await resp.json().catch(() => ({}))
    error = data.error || 'Failed to delete function'
  }

  onMount(loadFunctions)
</script>

<svelte:head>
  <title>Functions - OmniBase</title>
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Functions</h1>
    <p class="page-subtitle">Real persisted function registry with isolated `javascript`, `python`, `static-json`, and `webhook` runtimes</p>
  </div>
  <button class="btn btn-primary btn-sm" onclick={() => showCreateModal = true}>
    New Function
  </button>
</div>

<div class="page-content animate-fade-in">
  {#if error}
    <div class="card" style="margin-bottom: 16px; border-color: var(--status-error); color: var(--status-error);">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="skeleton" style="height: 200px; border-radius: 12px;"></div>
  {:else if functions.length === 0}
    <div class="card" style="padding: 48px; text-align: center;">
      <div style="font-size: 40px; margin-bottom: 16px;">lambda</div>
      <h2 style="margin-bottom: 12px;">No functions deployed</h2>
      <p style="color: var(--text-muted); max-width: 560px; margin: 0 auto;">
        Deploy deterministic JSON handlers, internal webhooks, or isolated `javascript` and `python` functions with per-invocation timeouts.
      </p>
    </div>
  {:else}
    <div class="card" style="padding: 0; overflow: hidden;">
      <div class="table-wrapper">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Slug</th>
              <th>Runtime</th>
              <th>Auth</th>
              <th>Updated</th>
              <th style="text-align: right;">Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each functions as fn}
              <tr>
                <td>{fn.name}</td>
                <td style="font-family: var(--font-mono);">{fn.slug}</td>
                <td><span class="badge badge-info">{fn.runtime}</span></td>
                <td><span class="badge {fn.verify_jwt ? 'badge-success' : 'badge-neutral'}">{fn.verify_jwt ? 'JWT' : 'Public'}</span></td>
                <td>{new Date(fn.updated_at).toLocaleString()}</td>
                <td style="text-align: right;">
                  <button class="btn btn-secondary btn-sm" onclick={() => deleteFunction(fn.slug)}>Delete</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

{#if showCreateModal}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 720px;">
      <div class="modal-header">
        <h3>Create Function</h3>
        <button class="btn-close" onclick={() => showCreateModal = false}>x</button>
      </div>
      <div class="modal-body" style="display: grid; gap: 12px;">
        <input class="input" bind:value={form.name} placeholder="Function name" />
        <input class="input" bind:value={form.slug} placeholder="function-slug" />
        <select class="input" bind:value={form.runtime} onchange={(event) => applyRuntimeTemplate((event.currentTarget as HTMLSelectElement).value as OmniFunction['runtime'])}>
          <option value="static-json">static-json</option>
          <option value="webhook">webhook</option>
          <option value="javascript">javascript</option>
          <option value="python">python</option>
        </select>
        <label style="display: flex; align-items: center; gap: 8px; font-size: 13px;">
          <input type="checkbox" bind:checked={form.verify_jwt} />
          Require JWT to invoke
        </label>
        <textarea class="input" bind:value={form.source} rows="14" style="font-family: var(--font-mono);"></textarea>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateModal = false}>Cancel</button>
        <button class="btn btn-primary" onclick={createFunction} disabled={saving}>
          {saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-content { padding: 0 24px 24px; }
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  .modal-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
  }
  .modal-header, .modal-footer {
    padding: 16px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-subtle);
  }
  .modal-footer {
    border-bottom: 0;
    border-top: 1px solid var(--border-subtle);
    justify-content: flex-end;
    gap: 12px;
  }
  .modal-body { padding: 20px; }
  .btn-close {
    border: 0;
    background: transparent;
    color: var(--text-muted);
    font-size: 20px;
    cursor: pointer;
  }
</style>
