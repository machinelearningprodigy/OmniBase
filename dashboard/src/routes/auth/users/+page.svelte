<script lang="ts">
  import { onMount } from 'svelte'
  import { getHeaders, getOmniBaseUrl } from '$lib/api'

  interface User {
    id: string
    email: string
    role: string
    is_banned: boolean
    email_confirmed_at: string | null
    last_sign_in_at: string | null
    created_at: string
    updated_at: string
  }

  let users = $state<User[]>([])
  let total = $state(0)
  let page = $state(1)
  let perPage = 20
  let loading = $state(true)
  let error = $state<string | null>(null)
  let searchQuery = $state('')
  let showInviteModal = $state(false)
  let inviteEmail = $state('')
  let inviting = $state(false)
  let actionLoading = $state<string | null>(null) // user id being actioned

  // Confirm delete dialog
  let confirmDelete = $state<User | null>(null)

  async function loadUsers() {
    try {
      loading = true
      error = null
      const resp = await fetch(
        `${getOmniBaseUrl()}/auth/v1/admin/users?page=${page}&per_page=${perPage}`,
        { headers: getHeaders() }
      )
      if (resp.ok) {
        const data = await resp.json()
        users = data.users || []
        total = data.total || 0
      } else {
        const d = await resp.json().catch(() => ({}))
        error = d.message || 'Failed to load users'
        users = []
      }
    } catch {
      error = 'Connection error — is the API gateway running?'
      users = []
    } finally {
      loading = false
    }
  }

  async function banUser(user: User) {
    actionLoading = user.id
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/admin/users/${user.id}/ban`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ banned: !user.is_banned })
      })
      if (resp.ok) {
        await loadUsers()
      } else {
        const d = await resp.json().catch(() => ({}))
        alert(d.message || 'Failed to update user')
      }
    } catch {
      alert('Connection error')
    } finally {
      actionLoading = null
    }
  }

  async function deleteUser(user: User) {
    actionLoading = user.id
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/admin/users/${user.id}`, {
        method: 'DELETE',
        headers: getHeaders()
      })
      if (resp.ok || resp.status === 204) {
        confirmDelete = null
        await loadUsers()
      } else {
        const d = await resp.json().catch(() => ({}))
        alert(d.message || 'Failed to delete user')
      }
    } catch {
      alert('Connection error')
    } finally {
      actionLoading = null
    }
  }

  async function inviteUser() {
    if (!inviteEmail.trim()) return
    inviting = true
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/auth/v1/admin/invite`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ email: inviteEmail })
      })
      if (resp.ok || resp.status === 201) {
        showInviteModal = false
        inviteEmail = ''
        await loadUsers()
      } else {
        const d = await resp.json().catch(() => ({}))
        alert(d.message || 'Failed to invite user')
      }
    } catch {
      alert('Connection error')
    } finally {
      inviting = false
    }
  }

  function formatDate(dateStr: string | null): string {
    if (!dateStr) return 'Never'
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
  }

  function timeAgo(dateStr: string | null): string {
    if (!dateStr) return 'Never'
    try {
      const diff = Date.now() - new Date(dateStr).getTime()
      const mins = Math.floor(diff / 60000)
      if (mins < 1) return 'Just now'
      if (mins < 60) return `${mins}m ago`
      const hrs = Math.floor(mins / 60)
      if (hrs < 24) return `${hrs}h ago`
      const days = Math.floor(hrs / 24)
      return `${days}d ago`
    } catch {
      return dateStr
    }
  }

  let filteredUsers = $derived(
    users.filter(u =>
      u.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      u.id.toLowerCase().includes(searchQuery.toLowerCase())
    )
  )

  onMount(loadUsers)
</script>

<svelte:head>
  <title>Auth Users — OmniBase</title>
  <meta name="description" content="Manage authenticated users in your OmniBase project" />
</svelte:head>

<div class="page-header animate-fade-in">
  <div>
    <h1 class="page-title">Users</h1>
    <p class="page-subtitle">
      {total} total users
    </p>
  </div>
  <div class="flex gap-2">
    <button class="btn btn-secondary btn-sm" onclick={loadUsers}>Refresh</button>
    <button class="btn btn-primary btn-sm" onclick={() => showInviteModal = true} id="invite-user-btn">
      + Invite User
    </button>
  </div>
</div>

<div class="page-content animate-fade-in">
  <!-- Search bar -->
  <div style="margin-bottom: 16px;">
    <input
      class="input"
      type="search"
      placeholder="Search by email or user ID..."
      bind:value={searchQuery}
      id="user-search"
      style="max-width: 400px;"
    />
  </div>

  {#if loading}
    <div class="card" style="padding: 0; overflow: hidden;">
      {#each Array(5) as _}
        <div class="skeleton" style="height: 60px; margin: 1px 0; border-radius: 0;"></div>
      {/each}
    </div>
  {:else if error}
    <div class="card" style="border-color: rgba(255,82,82,0.3); background: rgba(255,82,82,0.05); text-align: center; padding: 40px;">
      <p style="color: var(--status-error); margin-bottom: 12px;">⚠ {error}</p>
      <button class="btn btn-secondary btn-sm" onclick={loadUsers}>Retry</button>
    </div>
  {:else if filteredUsers.length === 0}
    <div class="empty-state card">
      <div style="font-size: 40px; margin-bottom: 16px;">👤</div>
      <h3>No users found</h3>
      <p>{searchQuery ? 'Try a different search query' : 'Invite your first user to get started'}</p>
      {#if !searchQuery}
        <button class="btn btn-primary btn-sm" style="margin-top: 16px;" onclick={() => showInviteModal = true}>
          Invite User
        </button>
      {/if}
    </div>
  {:else}
    <div class="card" style="padding: 0; overflow: hidden;">
      <div class="table-wrapper" style="margin: 0; border-radius: 0;">
        <table>
          <thead>
            <tr>
              <th>Email</th>
              <th>User ID</th>
              <th>Role</th>
              <th>Status</th>
              <th>Last Sign In</th>
              <th>Created</th>
              <th style="text-align: right;">Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each filteredUsers as user}
              <tr style="{user.is_banned ? 'opacity: 0.6;' : ''}">
                <td>
                  <div style="display: flex; align-items: center; gap: 8px;">
                    <div style="
                      width: 28px; height: 28px; border-radius: 50%;
                      background: linear-gradient(135deg, var(--brand-primary), var(--brand-secondary));
                      display: flex; align-items: center; justify-content: center;
                      font-size: 11px; font-weight: 700; color: white; flex-shrink: 0;
                    ">
                      {user.email[0].toUpperCase()}
                    </div>
                    <div>
                      <div style="font-size: 13px; font-weight: 500;">{user.email}</div>
                      {#if user.email_confirmed_at}
                        <div style="font-size: 10px; color: var(--status-success);">✓ Verified</div>
                      {:else}
                        <div style="font-size: 10px; color: var(--text-muted);">Unverified</div>
                      {/if}
                    </div>
                  </div>
                </td>
                <td style="font-family: var(--font-mono); font-size: 11px; color: var(--text-muted);">
                  {user.id.slice(0, 8)}…
                </td>
                <td>
                  <span class="badge {user.role === 'service_role' ? 'badge-warning' : 'badge-neutral'}">
                    {user.role}
                  </span>
                </td>
                <td>
                  {#if user.is_banned}
                    <span class="badge badge-error">Banned</span>
                  {:else}
                    <span class="badge badge-success">Active</span>
                  {/if}
                </td>
                <td style="font-size: 12px; color: var(--text-secondary);">
                  {timeAgo(user.last_sign_in_at)}
                </td>
                <td style="font-size: 12px; color: var(--text-secondary);">
                  {formatDate(user.created_at).split(',')[0]}
                </td>
                <td style="text-align: right;">
                  <div class="flex gap-1" style="justify-content: flex-end;">
                    <button
                      class="btn btn-secondary btn-sm"
                      onclick={() => banUser(user)}
                      disabled={actionLoading === user.id}
                      id="ban-{user.id}"
                    >
                      {actionLoading === user.id ? '...' : user.is_banned ? 'Unban' : 'Ban'}
                    </button>
                    <button
                      class="btn btn-secondary btn-sm"
                      style="color: var(--status-error);"
                      onclick={() => confirmDelete = user}
                      disabled={actionLoading === user.id}
                      id="delete-{user.id}"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      {#if total > perPage}
        <div style="padding: 12px 16px; border-top: 1px solid var(--border-subtle); display: flex; justify-content: space-between; align-items: center;">
          <span style="font-size: 12px; color: var(--text-muted);">
            Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} of {total}
          </span>
          <div class="flex gap-2">
            <button class="btn btn-secondary btn-sm" disabled={page === 1} onclick={() => { page--; loadUsers() }}>
              ← Prev
            </button>
            <button class="btn btn-secondary btn-sm" disabled={page * perPage >= total} onclick={() => { page++; loadUsers() }}>
              Next →
            </button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Invite Modal -->
{#if showInviteModal}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 420px;">
      <div class="modal-header">
        <h3>Invite User</h3>
        <button class="btn-close" onclick={() => { showInviteModal = false; inviteEmail = '' }}>×</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label for="inviteEmail">Email Address</label>
          <input
            type="email"
            id="inviteEmail"
            class="input"
            bind:value={inviteEmail}
            placeholder="user@example.com"
          />
        </div>
        <p style="font-size: 12px; color: var(--text-muted); margin-top: 8px;">
          An account will be created and the user will receive an invitation email.
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => { showInviteModal = false; inviteEmail = '' }}>Cancel</button>
        <button class="btn btn-primary" onclick={inviteUser} disabled={inviting || !inviteEmail.trim()}>
          {inviting ? 'Sending...' : 'Send Invite'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Confirm Delete Modal -->
{#if confirmDelete}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 420px;">
      <div class="modal-header">
        <h3>Delete User</h3>
        <button class="btn-close" onclick={() => confirmDelete = null}>×</button>
      </div>
      <div class="modal-body">
        <p style="font-size: 14px;">
          Are you sure you want to permanently delete
          <strong>{confirmDelete.email}</strong>?
          This action cannot be undone.
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => confirmDelete = null}>Cancel</button>
        <button
          class="btn btn-primary"
          style="background: var(--status-error); border-color: var(--status-error);"
          onclick={() => deleteUser(confirmDelete!)}
          disabled={actionLoading === confirmDelete?.id}
        >
          {actionLoading ? 'Deleting...' : 'Yes, Delete'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-content { padding: 0 24px 24px; }
  .modal-overlay {
    position: fixed; top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.7); backdrop-filter: blur(4px);
    display: flex; align-items: center; justify-content: center; z-index: 1000;
  }
  .modal-card {
    background: var(--bg-surface); border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg); box-shadow: var(--shadow-xl);
    display: flex; flex-direction: column; max-height: 90vh;
  }
  .modal-header {
    padding: 16px 20px; border-bottom: 1px solid var(--border-subtle);
    display: flex; justify-content: space-between; align-items: center;
  }
  .modal-header h3 { font-size: 16px; margin: 0; }
  .modal-body { padding: 20px; }
  .modal-footer {
    padding: 16px 20px; border-top: 1px solid var(--border-subtle);
    display: flex; justify-content: flex-end; gap: 12px;
  }
  .btn-close {
    background: none; border: none; color: var(--text-muted);
    font-size: 24px; cursor: pointer; line-height: 1;
  }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .form-group label { font-size: 13px; font-weight: 500; }
</style>
