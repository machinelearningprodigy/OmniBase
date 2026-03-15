<svelte:head><title>Email Campaigns — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Email Campaigns</h1><p class="page-subtitle">Transactional and bulk email — via SMTP, SendGrid, Resend, Mailgun, or AWS SES</p></div><button class="btn btn-primary">＋ New Campaign</button></div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [{l:'Sent Today',v:'4,200'},{l:'Open Rate',v:'38%'},{l:'Click Rate',v:'8.4%'},{l:'Bounce Rate',v:'0.3%'}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div></div>
    {/each}
  </div>
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Campaign</th><th>Subject</th><th>Sent</th><th>Opens</th><th>Clicks</th><th>Status</th></tr></thead>
      <tbody>
        {#each [
          {name:'Welcome Series #1',subject:'Welcome to OmniBase!',sent:142,opens:'68%',clicks:'22%',status:'sent'},
          {name:'Feature Announcement',subject:'Introducing DB Branching',sent:2840,opens:'42%',clicks:'11%',status:'sent'},
          {name:'March Newsletter',subject:'OmniBase March Update',sent:0,opens:'—',clicks:'—',status:'draft'},
        ] as c}
          <tr>
            <td>{c.name}</td>
            <td style="color:var(--text-secondary);font-size:12px">{c.subject}</td>
            <td class="cell-mono">{c.sent > 0 ? c.sent.toLocaleString() : '—'}</td>
            <td class="cell-mono" style="color:#00e676">{c.opens}</td>
            <td class="cell-mono">{c.clicks}</td>
            <td><span class="badge {c.status==='sent'?'badge-success':'badge-neutral'}">{c.status}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
