<script lang="ts">
  let users = [
    { id: '1', email: 'admin@omnibase.dev', last_sign_in: '2 mins ago', created_at: '2026-03-01' },
    { id: '2', email: 'user@example.com', last_sign_in: '1 day ago', created_at: '2026-03-05' },
    { id: '3', email: 'test@test.com', last_sign_in: 'Never', created_at: '2026-03-12' },
  ]
</script>

<svelte:head>
  <title>Auth Users — OmniBase</title>
</svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Users</h1>
    <p class="page-subtitle">Manage your authenticated users</p>
  </div>
  <button class="btn btn-primary">+ Add User</button>
</div>

<div class="page-content">
  <div class="card" style="padding: 0; overflow:hidden;">
    <table style="width: 100%; border-collapse: collapse;">
      <thead style="background: var(--bg-surface); text-align: left; border-bottom: 1px solid var(--border-subtle);">
        <tr>
          <th style="padding: 12px 16px; font-weight: 500; font-size: 13px;">Email</th>
          <th style="padding: 12px 16px; font-weight: 500; font-size: 13px;">ID</th>
          <th style="padding: 12px 16px; font-weight: 500; font-size: 13px;">Last Sign In</th>
          <th style="padding: 12px 16px; font-weight: 500; font-size: 13px;">Created</th>
          <th style="padding: 12px 16px; font-weight: 500; font-size: 13px; text-align: right;">Action</th>
        </tr>
      </thead>
      <tbody>
        {#each users as user}
          <tr style="border-bottom: 1px solid var(--border-subtle);">
            <td style="padding: 12px 16px; font-size: 13px;">{user.email}</td>
            <td style="padding: 12px 16px; font-size: 12px; font-family: var(--font-mono); color: var(--text-muted);">{user.id}</td>
            <td style="padding: 12px 16px; font-size: 13px; color: var(--text-secondary);">{user.last_sign_in}</td>
            <td style="padding: 12px 16px; font-size: 13px; color: var(--text-secondary);">{user.created_at}</td>
            <td style="padding: 12px 16px; text-align: right;">
              <button class="btn btn-secondary btn-sm">Edit</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
