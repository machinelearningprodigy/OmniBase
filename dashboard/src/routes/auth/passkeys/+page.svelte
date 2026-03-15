<svelte:head><title>Passkeys — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Passkeys</h1><p class="page-subtitle">WebAuthn / FIDO2 — hardware keys, biometrics, and device passkeys for passwordless auth</p></div>
  <span class="badge badge-success">NEW</span>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="card stat-card">
      <div class="stat-value">142</div>
      <div class="stat-label">Registered Passkeys</div>
    </div>
    <div class="card stat-card">
      <div class="stat-value">89%</div>
      <div class="stat-label">Login Success Rate</div>
    </div>
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Accepted Authenticator Types</h3>
    <div style="display:flex;flex-direction:column;gap:12px">
      {#each [
        {label:'Platform authenticators (Touch ID, Face ID, Windows Hello)',enabled:true},
        {label:'Roaming authenticators (YubiKey, FIDO hardware keys)',enabled:true},
        {label:'Hybrid (passkeys synced via iCloud Keychain / Google Password Manager)',enabled:false},
      ] as opt}
        <label style="display:flex;align-items:center;gap:12px;cursor:pointer">
          <input type="checkbox" checked={opt.enabled} style="accent-color:#6c47ff;width:16px;height:16px" />
          <span style="font-size:13px">{opt.label}</span>
        </label>
      {/each}
    </div>
  </div>
</div>
