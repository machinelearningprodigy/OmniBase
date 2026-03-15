<svelte:head><title>Team Members — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Team Members</h1><p class="page-subtitle">Manage who has access to this OmniBase project and their roles</p></div><button class="btn btn-primary">＋ Invite Member</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Member</th><th>Role</th><th>Joined</th><th>Last Active</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'Alice Kumar',email:'alice@example.com',role:'Owner',joined:'Mar 1 2026',last:'2min ago'},
          {name:'Bob Lee',email:'bob@example.com',role:'Developer',joined:'Mar 5 2026',last:'1hr ago'},
          {name:'Carol Patel',email:'carol@example.com',role:'Viewer',joined:'Mar 10 2026',last:'2 days ago'},
        ] as m}
          <tr>
            <td>
              <div style="display:flex;align-items:center;gap:10px">
                <div style="width:32px;height:32px;border-radius:50%;background:linear-gradient(135deg,#6c47ff,#00d4ff);display:flex;align-items:center;justify-content:center;font-weight:700;color:white;font-size:13px">{m.name[0]}</div>
                <div>
                  <div style="font-weight:600;font-size:13px">{m.name}</div>
                  <div style="font-size:11px;color:var(--text-secondary)">{m.email}</div>
                </div>
              </div>
            </td>
            <td><span class="badge {m.role==='Owner'?'badge-success':m.role==='Developer'?'badge-info':'badge-neutral'}">{m.role}</span></td>
            <td style="font-size:12px;color:var(--text-secondary)">{m.joined}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{m.last}</td>
            <td>
              {#if m.role !== 'Owner'}
                <div style="display:flex;gap:6px">
                  <button class="btn btn-secondary btn-sm">Change Role</button>
                  <button class="btn btn-danger btn-sm">Remove</button>
                </div>
              {:else}
                <span style="font-size:11px;color:var(--text-muted)">—</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
