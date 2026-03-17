<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getHeaders, getErrorMessage } from '$lib/api'

  interface EnumMeta {
    name: string
    schema: string
    values: string[]
  }

  let enums = $state<EnumMeta[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  // Create Mode
  let showCreateModal = $state(false)
  let newName = $state('')
  let newValuesString = $state('')
  let isCreating = $state(false)

  // Edit/Add value Mode
  let showEditModal = $state(false)
  let selectedEnum = $state<EnumMeta | null>(null)
  let newValue = $state('')
  let isEditing = $state(false)

  async function loadEnums() {
    try {
      loading = true
      error = null
      const currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/enums?schema=${currentSchema}`, { headers: getHeaders() })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to load enums'))
      enums = await resp.json()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading = false
    }
  }

  async function createEnum() {
    if (!newName) return alert('Name is required')
    const values = newValuesString.split(',').map((v) => v.trim()).filter(Boolean)
    if (values.length === 0) return alert('At least one value is required')

    isCreating = true
    try {
      const currentSchema = localStorage.getItem('omnibase.current_schema') || 'public'
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/enums`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ name: newName, schema: currentSchema, values })
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to create enum'))
      
      showCreateModal = false
      newName = ''
      newValuesString = ''
      await loadEnums()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    } finally {
      isCreating = false
    }
  }

  async function deleteEnum(e: EnumMeta) {
    if (!confirm(`Are you sure you want to delete enum type "${e.name}"? This might break columns that use it.`)) return
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/enums?schema=${e.schema}&name=${e.name}`, {
        method: 'DELETE',
        headers: getHeaders(false)
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to delete enum'))
      await loadEnums()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Unknown error')
    }
  }

  async function addValue() {
    if (!selectedEnum || !newValue.trim()) return
    isEditing = true
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/enums`, {
        method: 'PATCH',
        headers: getHeaders(),
        body: JSON.stringify({ name: selectedEnum.name, schema: selectedEnum.schema, add_values: [newValue.trim()] })
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to add value'))
      
      showEditModal = false
      newValue = ''
      selectedEnum = null
      await loadEnums()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    } finally {
      isEditing = false
    }
  }

  onMount(() => {
    loadEnums()
    const handleSchemaChange = () => loadEnums()
    window.addEventListener('omnibase:schema-change', handleSchemaChange)
    return () => window.removeEventListener('omnibase:schema-change', handleSchemaChange)
  })
</script>

<svelte:head><title>Enum Types — OmniBase</title></svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Enum Types</h1>
    <p class="page-subtitle">Manage PostgreSQL enumerated types used across your database schema</p>
  </div>
  <button class="btn btn-primary" onclick={() => showCreateModal = true}>＋ New Enum</button>
</div>

<div class="page-content">
  {#if loading}
    <div class="skeleton" style="height: 100px; width: 300px; border-radius: var(--radius-md);"></div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else if enums.length === 0}
    <div class="empty-state">
      <div class="empty-icon">⋮</div>
      <h3>No Enum Types</h3>
      <p>Enumerations are custom data types that comprise a static, ordered set of values.</p>
      <button class="btn btn-primary btn-sm" style="margin-top:16px" onclick={() => showCreateModal = true}>Create Enum</button>
    </div>
  {:else}
    <div class="grid-cols-3 gap-4">
      {#each enums as e}
        <div class="card">
          <div style="font-size:12px;font-weight:700;color:#a78bfa;margin-bottom:8px;font-family:var(--font-mono)">{e.name}</div>
          <div style="display:flex;flex-wrap:wrap;gap:4px">
            {#each e.values as v}
              <span class="badge badge-neutral">{v}</span>
            {/each}
          </div>
          <div style="margin-top:12px;display:flex;gap:8px">
            <button class="btn btn-secondary btn-sm" onclick={() => { selectedEnum = e; showEditModal = true }}>Add Value</button>
            <button class="btn btn-danger btn-sm" onclick={() => deleteEnum(e)}>Delete</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if showCreateModal}
  <div class="modal-backdrop" onclick={() => !isCreating && (showCreateModal = false)}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2 class="modal-title">Create Enum Type</h2>
        <button class="modal-close" onclick={() => showCreateModal = false} disabled={isCreating}>×</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label class="form-label" for="enum-name">Name</label>
          <input id="enum-name" class="input" type="text" placeholder="e.g. user_role" bind:value={newName} disabled={isCreating} />
        </div>
        <div class="form-group">
          <label class="form-label" for="enum-values">Values (comma separated)</label>
          <input id="enum-values" class="input" type="text" placeholder="e.g. admin, member, guest" bind:value={newValuesString} disabled={isCreating} />
          <p class="form-hint">Enter the allowed string values separated by commas.</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateModal = false} disabled={isCreating}>Cancel</button>
        <button class="btn btn-primary" onclick={createEnum} disabled={isCreating || !newName || !newValuesString}>
          {isCreating ? 'Creating...' : 'Create Enum'}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showEditModal && selectedEnum}
  <div class="modal-backdrop" onclick={() => !isEditing && (showEditModal = false)}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2 class="modal-title">Add Value to {selectedEnum.name}</h2>
        <button class="modal-close" onclick={() => showEditModal = false} disabled={isEditing}>×</button>
      </div>
      <div class="modal-body">
        <p class="form-hint" style="margin-bottom:16px;color:var(--text-warning)">Note: In PostgreSQL, you can easily add new values to an enum, but removing existing values usually requires recreating the type.</p>
        <div class="form-group">
          <label class="form-label" for="new-enum-val">New Value</label>
          <input id="new-enum-val" class="input" type="text" placeholder="e.g. superadmin" bind:value={newValue} disabled={isEditing} />
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showEditModal = false} disabled={isEditing}>Cancel</button>
        <button class="btn btn-primary" onclick={addValue} disabled={isEditing || !newValue}>
          {isEditing ? 'Adding...' : 'Add Value'}
        </button>
      </div>
    </div>
  </div>
{/if}
