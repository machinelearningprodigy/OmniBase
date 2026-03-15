<svelte:head><title>Image Pipeline — OmniBase</title></svelte:head>
<div class="page-header">
  <div><h1 class="page-title">Image Pipeline</h1><p class="page-subtitle">On-the-fly image transformation — resize, crop, format conversion, blur, watermark</p></div>
</div>
<div class="page-content">
  <div class="grid-cols-2 gap-4" style="margin-bottom:24px">
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Transform via URL Parameters</h3>
      <p style="font-size:12px;color:var(--text-secondary);margin-bottom:12px">Append transform params to any storage URL:</p>
      <code style="display:block;font-size:11px;background:rgba(255,255,255,0.04);padding:12px;border-radius:8px;line-height:1.8">
        /storage/v1/object/public/avatars/user.jpg<br/>
        ?width=200&height=200<br/>
        &fit=cover<br/>
        &format=webp<br/>
        &quality=85
      </code>
    </div>
    <div class="card">
      <h3 style="font-size:13px;font-weight:600;margin-bottom:12px">Supported Transforms</h3>
      {#each [
        {p:'width, height','desc':'Resize in pixels'},
        {p:'fit','desc':'cover, contain, fill, inside, outside'},
        {p:'format','desc':'webp, avif, jpeg, png, heic'},
        {p:'quality','desc':'1-100 (default: 80)'},
        {p:'blur','desc':'Gaussian blur 0-100'},
        {p:'rotate','desc':'90, 180, 270 degrees'},
      ] as t}
        <div style="display:flex;gap:12px;padding:8px 0;border-bottom:1px solid var(--border-subtle)">
          <code style="flex-shrink:0;width:120px;font-size:11px">{t.p}</code>
          <span style="font-size:12px;color:var(--text-secondary)">{t.desc}</span>
        </div>
      {/each}
    </div>
  </div>
</div>
