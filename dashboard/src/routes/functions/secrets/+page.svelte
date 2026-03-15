<svelte:head><title>Secrets Vault — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Secrets Vault</h1><p class="page-subtitle">Encrypted environment variables injected into your edge functions at runtime</p></div><button class="btn btn-primary">＋ Add Secret</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Name</th><th>Scope</th><th>Last Updated</th><th>Used By</th><th>Actions</th></tr></thead>
      <tbody>
        {#each [
          {name:'STRIPE_SECRET_KEY',scope:'project',updated:'2 days ago',fns:'process-payment, subscription-webhook'},
          {name:'OPENAI_API_KEY',scope:'project',updated:'5 days ago',fns:'ai-summary, embedding-gen'},
          {name:'SMTP_PASSWORD',scope:'project',updated:'10 days ago',fns:'send-email'},
          {name:'TWILIO_AUTH_TOKEN',scope:'project',updated:'3 days ago',fns:'send-sms'},
        ] as s}
          <tr>
            <td class="cell-mono">{s.name}</td>
            <td><span class="badge badge-neutral">{s.scope}</span></td>
            <td style="font-size:12px;color:var(--text-secondary)">{s.updated}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{s.fns}</td>
            <td style="display:flex;gap:6px">
              <button class="btn btn-secondary btn-sm">Update</button>
              <button class="btn btn-danger btn-sm">Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
