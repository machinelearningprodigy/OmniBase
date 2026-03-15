<svelte:head><title>Audit Logs — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Audit Logs</h1><p class="page-subtitle">Full audit trail of all authentication events — logins, signups, token revocations, admin actions</p></div>
  <button class="btn btn-secondary">⬇ Export CSV</button>
</div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Timestamp</th><th>Event</th><th>User</th><th>IP Address</th><th>Location</th><th>Result</th></tr></thead>
      <tbody>
        {#each [
          {ts:'07:46:32',event:'TOKEN_REFRESH',user:'alice@example.com',ip:'203.0.113.42',loc:'Mumbai, IN',ok:true},
          {ts:'07:44:11',event:'LOGIN',user:'bob@example.com',ip:'198.51.100.7',loc:'New York, US',ok:true},
          {ts:'07:42:03',event:'LOGIN_FAILED',user:'unknown@test.com',ip:'10.0.0.44',loc:'—',ok:false},
          {ts:'07:39:55',event:'SIGNUP',user:'carol@example.com',ip:'192.0.2.89',loc:'London, UK',ok:true},
          {ts:'07:35:22',event:'SESSION_REVOKE',user:'dave@example.com',ip:'172.16.0.1',loc:'Berlin, DE',ok:true},
          {ts:'07:31:08',event:'MFA_VERIFY',user:'alice@example.com',ip:'203.0.113.42',loc:'Mumbai, IN',ok:true},
        ] as l}
          <tr>
            <td class="cell-mono" style="font-size:11px">{l.ts}</td>
            <td><code>{l.event}</code></td>
            <td style="font-size:12px">{l.user}</td>
            <td class="cell-mono" style="font-size:11px">{l.ip}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{l.loc}</td>
            <td><span class="badge {l.ok ? 'badge-success' : 'badge-error'}">{l.ok ? 'OK' : 'FAIL'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
