<svelte:head><title>Web3 Auth — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Web3 Auth</h1><p class="page-subtitle">Sign in with MetaMask, WalletConnect, or any EVM-compatible wallet signature</p></div>
  <span class="badge badge-success">NEW</span>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4">
    {#each [
      {name:'MetaMask',desc:'Browser extension wallet — Ethereum + EVM chains',icon:'🦊',enabled:true},
      {name:'WalletConnect',desc:'QR-code-based connection for 300+ mobile wallets',icon:'🔗',enabled:false},
      {name:'Coinbase Wallet',desc:'Coinbase-based self-custody wallet',icon:'🔵',enabled:false},
      {name:'Rainbow',desc:'Consumer-focused Ethereum wallet',icon:'🌈',enabled:false},
    ] as w}
      <div class="card" style="display:flex;align-items:center;gap:16px">
        <div style="font-size:32px">{w.icon}</div>
        <div style="flex:1">
          <div style="font-weight:700;font-size:14px;margin-bottom:2px">{w.name}</div>
          <p style="font-size:12px;color:var(--text-secondary)">{w.desc}</p>
        </div>
        <button class="btn btn-sm {w.enabled ? 'btn-danger' : 'btn-primary'}">{w.enabled ? 'Disable' : 'Enable'}</button>
      </div>
    {/each}
  </div>
</div>
