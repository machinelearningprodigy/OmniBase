<svelte:head><title>Sessions — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Sessions</h1><p class="page-subtitle">View and manage active user sessions across all devices</p></div>
  <button class="btn btn-danger">Revoke All Sessions</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>User</th><th>Device</th><th>IP Address</th><th>Location</th><th>Created</th><th>Expires</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {user:'alice@example.com',device:'Chrome / Windows',ip:'203.0.113.42',loc:'Mumbai, IN',created:'2h ago',expires:'29d 22h'},
          {user:'bob@example.com',device:'Safari / iPhone',ip:'198.51.100.7',loc:'New York, US',created:'1d ago',expires:'29d'},
          {user:'carol@example.com',device:'Firefox / macOS',ip:'192.0.2.89',loc:'London, UK',created:'3d ago',expires:'27d'},
        ] as s}
          <tr>
            <td>{s.user}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{s.device}</td>
            <td class="cell-mono">{s.ip}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{s.loc}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{s.created}</td>
            <td class="cell-mono" style="font-size:12px">{s.expires}</td>
            <td><button class="btn btn-danger btn-sm">Revoke</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
