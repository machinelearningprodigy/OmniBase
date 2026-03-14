<script lang="ts">
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  let query = $state('query {\n  __schema {\n    types {\n      name\n    }\n  }\n}')
  let variables = $state('{}')
  let result = $state<any>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)

  async function runQuery() {
    if (!query.trim()) return
    loading = true
    error = null
    result = null

    try {
      let vars = {}
      if (variables.trim()) {
        try {
          vars = JSON.parse(variables)
        } catch {
          throw new Error('Variables must be valid JSON')
        }
      }

      const resp = await fetch(`${getOmniBaseUrl()}/graphql/v1`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ query, variables: vars })
      })

      const data = await resp.json()
      if (resp.ok) {
        result = data
      } else {
        error = data.error || data.message || 'Failed to execute GraphQL query'
      }
    } catch (e: any) {
      error = e.message || 'Connection error'
    } finally {
      loading = false
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      runQuery()
    }
  }
</script>

<svelte:head>
  <title>GraphQL API — OmniBase</title>
</svelte:head>

<div class="editor-container animate-fade-in">
  <div class="editor-header">
    <div class="flex items-center gap-4">
      <h1 class="page-title" style="font-size: 16px; margin: 0;">GraphQL Explorer</h1>
      <span style="font-size: 11px; color: var(--text-muted);">Ctrl + Enter to run</span>
    </div>
    <div class="flex gap-2">
      <button class="btn btn-secondary btn-sm" onclick={() => { query = ''; variables = '{}'; result = null }}>Clear</button>
      <button class="btn btn-primary btn-sm" onclick={runQuery} disabled={loading}>
        {loading ? 'Running...' : 'Run Query'}
      </button>
    </div>
  </div>

  <div class="editor-main">
    <div class="editor-panes">
      <!-- Input Pane -->
      <div class="input-pane">
        <div class="pane-label">Query</div>
        <textarea
          bind:value={query}
          onkeydown={handleKeydown}
          placeholder="Enter GraphQL query..."
          spellcheck="false"
        ></textarea>
        
        <div class="pane-label" style="border-top: 1px solid var(--border-subtle);">Variables (JSON)</div>
        <textarea
          bind:value={variables}
          placeholder="&#123;&#125;"
          spellcheck="false"
          style="height: 120px; flex: none;"
        ></textarea>
      </div>

      <!-- Result Pane -->
      <div class="result-pane">
        <div class="pane-label">Result</div>
        <div class="result-content">
          {#if error}
            <div class="error-msg">
              <span style="font-weight: 600;">Error:</span><br>
              {error}
            </div>
          {:else if loading}
            <div class="flex items-center justify-center" style="height: 100%;">
              <div class="skeleton" style="width: 100%; height: 100%; border-radius: 8px;"></div>
            </div>
          {:else if result === null}
            <div class="empty-results">
              Run a query to see results
            </div>
          {:else}
            <pre class="json-result">{JSON.stringify(result, null, 2)}</pre>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .editor-container {
    height: calc(100vh - var(--topbar-height));
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .editor-header {
    padding: 12px 24px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .editor-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .editor-panes {
    flex: 1;
    display: flex;
    overflow: hidden;
  }
  .input-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border-subtle);
    min-width: 0;
  }
  .result-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: var(--bg-elevated);
    min-width: 0;
  }
  .pane-label {
    padding: 8px 16px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
  }
  textarea {
    flex: 1;
    width: 100%;
    padding: 16px;
    background: #0d1117;
    color: #e6edf3;
    font-family: var(--font-mono);
    font-size: 13px;
    line-height: 1.5;
    border: none;
    resize: none;
    outline: none;
  }
  .result-content {
    flex: 1;
    overflow: auto;
    padding: 16px;
    background: #0a0e17;
  }
  .error-msg {
    padding: 16px;
    background: rgba(255, 82, 82, 0.05);
    border: 1px solid rgba(255, 82, 82, 0.2);
    border-radius: 8px;
    color: var(--status-error);
    font-size: 13px;
    font-family: var(--font-mono);
  }
  .empty-results {
    color: var(--text-muted);
    font-size: 13px;
    text-align: center;
    margin-top: 40px;
  }
  .json-result {
    margin: 0;
    color: #a5d6ff;
    font-family: var(--font-mono);
    font-size: 13px;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
