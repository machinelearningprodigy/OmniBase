<svelte:head><title>Event Tracking — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">Event Tracking</h1><p class="page-subtitle">Custom analytics events — track any user action with properties</p></div></div>
<div class="page-content">
  <div class="grid-cols-4 gap-4" style="margin-bottom:24px">
    {#each [{l:'Events Today',v:'48.2k'},{l:'Unique Events',v:'142'},{l:'DAU',v:'2,840'},{l:'MAU',v:'18,400'}] as s}
      <div class="stat-card"><div class="stat-value">{s.v}</div><div class="stat-label">{s.l}</div></div>
    {/each}
  </div>
  <div class="table-wrapper">
    <table>
      <thead><tr><th>Event</th><th>Count Today</th><th>Unique Users</th><th>Trend</th></tr></thead>
      <tbody>
        {#each [
          {event:'page_view',count:'18.4k',users:'2.8k',trend:'+12%'},
          {event:'button_click:signup',count:'480',users:'480',trend:'+8%'},
          {event:'feature_used:sql_editor',count:'1.2k',users:'340',trend:'+24%'},
          {event:'export_csv',count:'88',users:'72',trend:'-3%'},
          {event:'api_key_created',count:'24',users:'24',trend:'+50%'},
        ] as e}
          <tr>
            <td class="cell-mono">{e.event}</td>
            <td class="cell-mono">{e.count}</td>
            <td>{e.users}</td>
            <td style="color:{e.trend.startsWith('+')?'#00e676':'#ff5252'}">{e.trend}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
