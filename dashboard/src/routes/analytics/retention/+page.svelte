<svelte:head><title>Retention — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Retention Cohorts</h1><p class="page-subtitle">D1/D7/D30 user retention — see how many users come back after their first visit</p></div></div>
<div class="page-content">
  <div class="grid-cols-3 gap-4" style="margin-bottom:24px">
    {#each [{l:'D1 Retention',v:'42%',d:'+3%',pos:true},{l:'D7 Retention',v:'18%',d:'-1%',pos:false},{l:'D30 Retention',v:'9%',d:'+2%',pos:true}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div><span class="stat-delta {s.pos?'positive':'negative'}">{s.d}</span></div>
    {/each}
  </div>
  <div class="card">
    <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Cohort Table (Weekly)</h3>
    <div class="table-wrapper">
      <table>
        <thead><tr><th>Cohort</th><th>Users</th><th>Week 0</th><th>Week 1</th><th>Week 2</th><th>Week 3</th><th>Week 4</th></tr></thead>
        <tbody>
          {#each [
            {cohort:'Mar W1',users:340,w0:100,w1:42,w2:28,w3:18,w4:12},
            {cohort:'Feb W4',users:420,w0:100,w1:38,w2:24,w3:16,w4:10},
            {cohort:'Feb W3',users:380,w0:100,w1:44,w2:30,w3:20,w4:14},
          ] as r}
            <tr>
              <td>{r.cohort}</td>
              <td class="cell-mono">{r.users}</td>
              {#each [r.w0,r.w1,r.w2,r.w3,r.w4] as v}
                <td>
                  <span style="padding:2px 8px;border-radius:4px;font-size:12px;font-weight:600;background:rgba(108,71,255,{v/100 * 0.5});color:{v>30?'#a78bfa':'var(--text-secondary)'}">{v}%</span>
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
