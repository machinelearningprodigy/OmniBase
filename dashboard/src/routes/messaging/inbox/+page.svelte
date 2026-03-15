<svelte:head><title>In-App Inbox — OmniBase</title></svelte:head>
<div class="page-header"><div><h1 class="page-title">In-App Inbox</h1><p class="page-subtitle">Real-time in-app notification inbox — store, deliver, and track read status</p></div><button class="btn btn-primary">＋ Send Message</button></div>
<div class="page-content">
  <div class="table-wrapper">
    <table>
      <thead><tr><th>ID</th><th>Recipient</th><th>Title</th><th>Sent At</th><th>Read</th></tr></thead>
      <tbody>
        {#each [
          {id:'msg_01j',to:'alice@example.com',title:'Your export is ready',sent:'2min ago',read:true},
          {id:'msg_01i',to:'bob@example.com',title:'New comment on your post',sent:'8min ago',read:false},
          {id:'msg_01h',to:'carol@example.com',title:'Invite accepted',sent:'22min ago',read:true},
        ] as m}
          <tr>
            <td class="cell-mono" style="font-size:11px">{m.id}</td>
            <td>{m.to}</td>
            <td>{m.title}</td>
            <td style="font-size:12px;color:var(--text-secondary)">{m.sent}</td>
            <td><span class="badge {m.read ? 'badge-success' : 'badge-warning'}">{m.read ? 'Read' : 'Unread'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
