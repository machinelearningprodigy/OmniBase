<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getHeaders, getErrorMessage } from '$lib/api'

  interface RoleMeta {
    rolename: string
    rolsuper: boolean
    rolinherit: boolean
    rolcreaterole: boolean
    rolcreatedb: boolean
    rolcanlogin: boolean
  }

  let roles = $state<RoleMeta[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  let showCreateModal = $state(false)
  let isCreating = $state(false)
  let newName = $state('')
  let newPassword = $state('')
  let canLogin = $state(false)
  let isSuper = $state(false)

  async function loadRoles() {
    try {
      loading = true
      error = null
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/roles`, { headers: getHeaders() })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to load roles'))
      roles = await resp.json()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading = false
    }
  }

  async function createRole() {
    if (!newName) return alert('Role name is required')
    
    isCreating = true
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/roles`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ rolename: newName, password: newPassword, canlogin: canLogin, issuper: isSuper })
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to create role'))
      
      showCreateModal = false
      newName = ''
      newPassword = ''
      canLogin = false
      isSuper = false
      await loadRoles()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    } finally {
      isCreating = false
    }
  }

  async function deleteRole(name: string) {
    if (!confirm(`Are you extremely sure you want to drop role "${name}"? This will drop/reassign all owned objects and cannot be undone.`)) return
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/roles?name=${name}`, {
        method: 'DELETE',
        headers: getHeaders(false)
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to delete role'))
      await loadRoles()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    }
  }

  onMount(() => loadRoles())
</script>

<svelte:head><title>Roles & Access — OmniBase</title></svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Roles & Access</h1>
    <p class="page-subtitle">Manage PostgreSQL database roles, service accounts, and privileges</p>
  </div>
  <button class="btn btn-primary" onclick={() => showCreateModal = true}>＋ New Role</button>
</div>

<div class="page-content">
  {#if loading}
    <div class="table-wrapper"><div class="skeleton" style="height: 300px; width: 100%; border-radius: var(--radius-md);"></div></div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else}
    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>Role Name</th>
            <th>Type</th>
            <th>Can Login</th>
            <th>Create DB/Role</th>
            <th style="text-align: right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each roles as r}
            <tr>
              <td>
                <span class="cell-mono font-bold {r.rolsuper ? 'text-warning' : ''}">{r.rolename}</span>
                {#if r.rolsuper}<span class="badge badge-warning ml-2" style="font-size:10px">SUPERUSER</span>{/if}
              </td>
              <td>
                <span class="badge badge-neutral">{r.rolename.startsWith('omnibase_') || r.rolename === 'postgres' || r.rolename === 'anon' || r.rolename === 'authenticated' || r.rolename === 'service_role' || r.rolename === 'authenticator' ? 'System' : 'Custom'}</span>
              </td>
              <td>
                {#if r.rolcanlogin}
                  <span class="badge badge-success">Yes</span>
                {:else}
                  <span class="badge">No</span>
                {/if}
              </td>
              <td>
                <div style="display:flex;gap:4px">
                  {#if r.rolcreatedb}<span class="badge badge-neutral" style="font-size:10px">DB</span>{/if}
                  {#if r.rolcreaterole}<span class="badge badge-neutral" style="font-size:10px">ROLE</span>{/if}
                  {#if !r.rolcreatedb && !r.rolcreaterole}<span style="color:var(--text-secondary)">-</span>{/if}
                </div>
              </td>
              <td style="text-align: right">
                {#if r.rolename !== 'postgres' && r.rolename !== 'omnibase' && r.rolsuper === false}
                  <button class="btn btn-danger btn-sm" onclick={() => deleteRole(r.rolename)}>Drop</button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

{#if showCreateModal}
  <div class="modal-backdrop" onclick={() => !isCreating && (showCreateModal = false)}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2 class="modal-title">Create Role</h2>
        <button class="modal-close" onclick={() => showCreateModal = false} disabled={isCreating}>×</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label class="form-label">Role Name</label>
          <input class="input" type="text" placeholder="e.g. readonly_user" bind:value={newName} disabled={isCreating} />
        </div>
        <div class="form-group">
          <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
            <input type="checkbox" bind:checked={canLogin} disabled={isCreating} />
            <span style="font-weight:600">Can Login</span>
          </label>
        </div>
        {#if canLogin}
          <div class="form-group">
            <label class="form-label">Password</label>
            <input class="input" type="password" placeholder="Leave blank for no password" bind:value={newPassword} disabled={isCreating} />
          </div>
        {/if}
        <div class="form-group" style="padding:12px;background:rgba(255,200,0,0.1);border-radius:var(--radius-md);border:1px solid rgba(255,200,0,0.2)">
          <div style="color:var(--text-warning);font-weight:600;margin-bottom:8px;font-size:13px">DANGER ZONE</div>
          <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
            <input type="checkbox" bind:checked={isSuper} disabled={isCreating} />
            <span style="font-weight:600;font-size:13px">Grant SUPERUSER privileges</span>
          </label>
          <p class="form-hint" style="margin-top:4px;color:rgba(255,255,255,0.6)">Superusers bypass all permission checks. Grant with extreme caution.</p>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateModal = false} disabled={isCreating}>Cancel</button>
        <button class="btn btn-primary" onclick={createRole} disabled={isCreating || !newName}>
          {isCreating ? 'Creating...' : 'Create Role'}
        </button>
      </div>
    </div>
  </div>
{/if}
