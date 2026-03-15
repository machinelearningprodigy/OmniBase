<svelte:head><title>GDPR Tools — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">GDPR Tools</h1><p class="page-subtitle">Right to erasure, data export, and consent management — GDPR/CCPA compliance tools</p></div></div>
<div class="page-content">
  <div class="grid-cols-3 gap-4">
    <div class="card">
      <div style="font-size:28px;margin-bottom:8px">🗑️</div>
      <h3 style="font-size:14px;font-weight:700;margin-bottom:6px">Right to Erasure</h3>
      <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">Delete all data for a specific user across all tables that reference their ID.</p>
      <div class="form-group"><label class="label">User ID or Email</label><input class="input" placeholder="user_id or email@example.com" /></div>
      <button class="btn btn-danger">Erase User Data</button>
    </div>
    <div class="card">
      <div style="font-size:28px;margin-bottom:8px">📦</div>
      <h3 style="font-size:14px;font-weight:700;margin-bottom:6px">Data Export</h3>
      <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">Export all data for a specific user as a JSON/CSV archive for data portability requests.</p>
      <div class="form-group"><label class="label">User ID or Email</label><input class="input" placeholder="user_id or email@example.com" /></div>
      <button class="btn btn-secondary">Export Data</button>
    </div>
    <div class="card">
      <div style="font-size:28px;margin-bottom:8px">✅</div>
      <h3 style="font-size:14px;font-weight:700;margin-bottom:6px">Consent Logs</h3>
      <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">Record and audit user consent events for marketing, analytics, and data processing.</p>
      <button class="btn btn-secondary">View Consent Records</button>
    </div>
  </div>
</div>
