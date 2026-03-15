<svelte:head><title>Roles & Access — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Roles & Access</h1><p class="page-subtitle">Manage PostgreSQL database roles and their privileges</p></div>
  <button class="btn btn-primary">＋ New Role</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Role</th><th>Type</th><th>Login</th><th>Superuser</th><th>Owns</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'anon',type:'Service',login:false,super:false,owns:0},
          {name:'authenticated',type:'Service',login:false,super:false,owns:0},
          {name:'service_role',type:'Service',login:false,super:false,owns:0},
          {name:'postgres',type:'Owner',login:true,super:true,owns:42},
          {name:'omnibase_admin',type:'Custom',login:true,super:false,owns:5},
        ] as r}
          <tr>
            <td class="cell-mono">{r.name}</td>
            <td><span class="badge badge-neutral">{r.type}</span></td>
            <td><span class="badge {r.login ? 'badge-success' : 'badge-neutral'}">{r.login ? 'Yes' : 'No'}</span></td>
            <td><span class="badge {r.super ? 'badge-warning' : 'badge-neutral'}">{r.super ? 'Yes' : 'No'}</span></td>
            <td>{r.owns} objects</td>
            <td><button class="btn btn-secondary btn-sm">Edit</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
