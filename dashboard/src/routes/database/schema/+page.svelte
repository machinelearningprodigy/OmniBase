<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getHeaders, getOmniBaseUrl, getErrorMessage } from '$lib/api'
  import { 
    Maximize2, Minimize2, ZoomIn, ZoomOut, Database, 
    Table as TableIcon, Key, Fingerprint, RefreshCcw, 
    MousePointer2, Layers, Settings2, Plus, Share2, 
    Download, LayoutGrid, Zap, ShieldCheck
  } from 'lucide-svelte'

  interface Column {
    name: string
    type: string
    is_primary: boolean
    is_nullable: boolean
  }

  interface Table {
    name: string
    schema: string
    columns: Column[]
    has_rls: boolean
    x: number
    y: number
    w: number
    h: number
  }

  interface ForeignKey {
    constraint_name: string
    from_schema: string
    from_table: string
    from_column: string
    to_schema: string
    to_table: string
    to_column: string
  }

  let tables = $state<Table[]>([])
  let foreignKeys = $state<ForeignKey[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let currentSchema = $state('public')
  
  // Canvas State constants
  const NODE_WIDTH = 280
  const ROW_HEIGHT = 32
  const HEADER_HEIGHT = 50
  const FOOTER_HEIGHT = 32

  let zoom = $state(0.9)
  let pan = $state({ x: 100, y: 100 })
  let isPanning = $state(false)
  let isDraggingNode = $state<string | null>(null)
  let dragOffset = $state({ x: 0, y: 0 })
  let startPanPos = { x: 0, y: 0 }

  async function loadSchema() {
    try {
      loading = true
      error = null
      const [fullSchemaResp, fksResp] = await Promise.all([
        apiFetch(`${getOmniBaseUrl()}/pg/full-schema?schema=${currentSchema}`),
        apiFetch(`${getOmniBaseUrl()}/pg/foreign-keys?schema=${currentSchema}`)
      ])

      if (!fullSchemaResp.ok) throw new Error(await getErrorMessage(fullSchemaResp, 'Failed to load tables'))
      if (!fksResp.ok) throw new Error(await getErrorMessage(fksResp, 'Failed to load relationships'))

      const schemaData = await fullSchemaResp.json()
      foreignKeys = await fksResp.json()

      console.log('Schema Architect Data:', { currentSchema, tablesCount: schemaData.length, fksCount: foreignKeys.length })
      console.log('Tables:', schemaData)
      console.log('Foreign Keys:', foreignKeys)

      // Calculate initial layout and estimated heights
      tables = schemaData.map((t: any, i: number) => {
        const height = HEADER_HEIGHT + (t.columns.length * ROW_HEIGHT) + (t.has_rls ? FOOTER_HEIGHT : 16)
        return {
          ...t,
          w: NODE_WIDTH,
          h: height,
          x: (i % 3) * 380 + 100,
          y: Math.floor(i / 3) * 450 + 100
        }
      })
    } catch (e: any) {
      error = e.message
    } finally {
      loading = false
    }
  }

  // --- Canvas Interaction ---
  function handleMouseDown(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains('canvas-area')) {
      isPanning = true
      startPanPos = { x: e.clientX - pan.x, y: e.clientY - pan.y }
    }
  }

  function startDraggingNode(e: MouseEvent, tableName: string) {
    e.stopPropagation()
    isDraggingNode = tableName
    const table = tables.find(t => t.name === tableName)
    if (table) {
      dragOffset = {
        x: e.clientX / zoom - table.x,
        y: e.clientY / zoom - table.y
      }
    }
  }

  function handleMouseMove(e: MouseEvent) {
    if (isDraggingNode) {
      const idx = tables.findIndex(t => t.name === isDraggingNode)
      if (idx !== -1) {
        tables[idx].x = e.clientX / zoom - dragOffset.x
        tables[idx].y = e.clientY / zoom - dragOffset.y
        tables = [...tables]
      }
    } else if (isPanning) {
      pan = { x: e.clientX - startPanPos.x, y: e.clientY - startPanPos.y }
    }
  }

  function stopDragging() {
    isDraggingNode = null
    isPanning = false
  }

  function handleWheel(e: WheelEvent) {
    e.preventDefault()
    const dZoom = e.deltaY > 0 ? -0.05 : 0.05
    const nextZoom = Math.min(2, Math.max(0.2, zoom + dZoom))
    zoom = nextZoom
  }

  // Robust Path Calculation
  function getPath(fk: ForeignKey) {
    const from = tables.find(t => t.name === fk.from_table)
    const to = tables.find(t => t.name === fk.to_table)
    if (!from || !to) return ''

    const fromColIdx = from.columns.findIndex(c => c.name === fk.from_column)
    const toColIdx = to.columns.findIndex(c => c.name === fk.to_column)

    // Calculate Y coordinates (centered on row)
    const y1 = from.y + HEADER_HEIGHT + (Math.max(0, fromColIdx) * ROW_HEIGHT) + (ROW_HEIGHT / 2)
    const y2 = to.y + HEADER_HEIGHT + (Math.max(0, toColIdx) * ROW_HEIGHT) + (ROW_HEIGHT / 2)

    // Horizontal logic
    const isFromLeft = from.x + NODE_WIDTH < to.x
    const isFromRight = from.x > to.x + NODE_WIDTH
    
    let x1, x2, cp1x, cp2x

    if (isFromLeft) {
      x1 = from.x + NODE_WIDTH
      x2 = to.x
      cp1x = x1 + 40
      cp2x = x2 - 40
    } else if (isFromRight) {
      x1 = from.x
      x2 = to.x + NODE_WIDTH
      cp1x = x1 - 40
      cp2x = x2 + 40
    } else {
      // Stacked vertically - connect left side
      x1 = from.x
      x2 = to.x
      cp1x = x1 - 40
      cp2x = x2 - 40
    }

    return `M ${x1} ${y1} C ${cp1x} ${y1}, ${cp2x} ${y2}, ${x2} ${y2}`
  }

  function autoLayout() {
    tables = tables.map((t, i) => ({
      ...t,
      x: (i % 3) * 380 + 100,
      y: Math.floor(i / 3) * 450 + 100
    }))
  }

  onMount(() => {
    loadSchema()
    window.addEventListener('mouseup', stopDragging)
    return () => window.removeEventListener('mouseup', stopDragging)
  })

  $effect(() => {
    const handleSchema = (e: any) => {
      currentSchema = e.detail
      loadSchema()
    }
    window.addEventListener('omnibase:schema-change', handleSchema as EventListener)
    return () => window.removeEventListener('omnibase:schema-change', handleSchema as EventListener)
  })
</script>

<svelte:head><title>Schema Architect — OmniBase</title></svelte:head>

<div class="architect-container">
  <!-- Toolbar -->
  <div class="canvas-toolbar">
    <div class="toolbar-left">
      <div class="schema-status">
        <div class="pulse-indicator"></div>
        <div class="status-info">
          <span class="schema-label">Active Schema</span>
          <span class="schema-name">{currentSchema}</span>
        </div>
      </div>
      <div class="v-divider"></div>
      <div class="stats-group">
        <div class="stat-pill"><TableIcon size={14} /> <span>{tables.length} Tables</span></div>
        <div class="stat-pill"><Layers size={14} /> <span>{foreignKeys.length} Links</span></div>
      </div>
    </div>

    <div class="toolbar-center">
      <div class="premium-nav">
        <button class="nav-btn" onclick={() => zoom = Math.max(0.2, zoom - 0.1)}><ZoomOut size={16} /></button>
        <div class="zoom-display">{Math.round(zoom * 100)}%</div>
        <button class="nav-btn" onclick={() => zoom = Math.min(2, zoom + 0.1)}><ZoomIn size={16} /></button>
        <div class="h-divider"></div>
        <button class="nav-btn action" onclick={autoLayout} title="Auto-align Tables"><LayoutGrid size={16} /></button>
        <button class="nav-btn action" onclick={loadSchema} title="Refresh Live Data"><RefreshCcw size={16} /></button>
      </div>
    </div>

    <div class="toolbar-right">
      <button class="btn-utility"><Share2 size={16} /></button>
      <button class="btn-utility"><Download size={16} /></button>
      <button class="btn btn-primary btn-sm"><Plus size={16} /> New Table</button>
    </div>
  </div>

  <!-- Canvas -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    class="canvas-area {isPanning ? 'panning' : ''}" 
    onmousedown={handleMouseDown}
    onmousemove={handleMouseMove}
    onwheel={handleWheel}
  >
    {#if loading}
      <div class="canvas-loading">
        <div class="loader-ring"></div>
        <p>Architecting live schema blueprint...</p>
      </div>
    {:else if error}
      <div class="canvas-error">
        <div class="error-shield"><ShieldCheck size={48} /></div>
        <h3>Architect Offline</h3>
        <p>{error}</p>
        <button class="btn btn-primary" onclick={loadSchema}>Reboot Engine</button>
      </div>
    {:else}
      <div 
        class="canvas-view"
        style="transform: scale({zoom}) translate({pan.x / zoom}px, {pan.y / zoom}px); transform-origin: 0 0;"
      >
        <!-- Relationship Lines (Dotted/Glowing) -->
        <svg class="blueprint-lines">
          <defs>
            <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="3" result="blur" />
              <feComposite in="SourceGraphic" in2="blur" operator="over" />
            </filter>
            <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" style="stop-color:var(--primary-color); stop-opacity:0.6" />
              <stop offset="100%" style="stop-color:#7f5af0; stop-opacity:0.6" />
            </linearGradient>
            <marker id="arrowhead" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="4" markerHeight="4" orient="auto">
              <path d="M 0 0 L 10 5 L 0 10 z" fill="var(--primary-color)" />
            </marker>
          </defs>
          
          {#each foreignKeys as fk}
            <g class="relationship-group">
              <path d={getPath(fk)} class="line-shadow" />
              <path d={getPath(fk)} class="line-main" />
            </g>
          {/each}
        </svg>

        <!-- Table Nodes (Premium UI) -->
        {#each tables as table}
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div 
            class="table-node {isDraggingNode === table.name ? 'dragging' : ''}"
            style="left: {table.x}px; top: {table.y}px;"
            onmousedown={(e) => startDraggingNode(e, table.name)}
          >
            <div class="node-glass"></div>
            
            <div class="node-header">
              <div class="header-left">
                <div class="table-icon-box"><TableIcon size={14} /></div>
                <span class="table-name">{table.name}</span>
              </div>
              <button class="node-menu"><Settings2 size={14} /></button>
            </div>
            
            <div class="node-columns">
              {#each table.columns as col}
                <div class="column-row {col.is_primary ? 'is-pk' : ''}">
                  <div class="col-left">
                    {#if col.is_primary}
                      <Key size={11} class="pk-key" />
                    {:else}
                      <div class="col-dot"></div>
                    {/if}
                    <span class="col-name">{col.name}</span>
                  </div>
                  <span class="col-type">{col.type}</span>
                </div>
              {/each}
            </div>

            {#if table.has_rls}
              <div class="node-footer">
                <Fingerprint size={12} />
                <span>RLS Active</span>
                <div class="rls-shield"><Zap size={10} /></div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- HUD -->
  <div class="architect-hud">
    <div class="hud-pill">
      <MousePointer2 size={14} />
      <span>Pan: Middle Mouse | Drag: LMB</span>
    </div>
  </div>
</div>

<style>
  :root {
    --architect-bg: #0b0e13;
    --node-bg: rgba(22, 27, 34, 0.8);
    --node-border: rgba(255, 255, 255, 0.1);
    --primary-glow: rgba(var(--primary-rgb), 0.3);
  }

  .architect-container {
    height: calc(100vh - 64px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--architect-bg);
    font-family: 'Inter', sans-serif;
  }

  /* --- Toolbar --- */
  .canvas-toolbar {
    height: 72px;
    background: rgba(13, 17, 23, 0.9);
    backdrop-filter: blur(20px);
    border-bottom: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 24px;
    z-index: 100;
  }

  .toolbar-left, .toolbar-right { display: flex; align-items: center; gap: 16px; }

  .schema-status { display: flex; align-items: center; gap: 12px; }
  .pulse-indicator { 
    width: 8px; height: 8px; background: #2ecc71; border-radius: 50%; 
    box-shadow: 0 0 8px #2ecc71; 
    animation: pulse 2s infinite;
  }
  .status-info { display: flex; flex-direction: column; }
  .schema-label { font-size: 10px; color: var(--text-dim); text-transform: uppercase; letter-spacing: 0.1em; }
  .schema-name { font-size: 14px; font-weight: 700; color: #fff; }

  .v-divider { width: 1px; height: 32px; background: rgba(255,255,255,0.1); margin: 0 8px; }

  .stats-group { display: flex; gap: 8px; }
  .stat-pill {
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(255,255,255,0.05);
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 12px;
    color: var(--text-dim);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .premium-nav {
    display: flex;
    align-items: center;
    background: #161b22;
    padding: 4px;
    border-radius: 10px;
    border: 1px solid var(--border-color);
  }

  .nav-btn {
    width: 32px; height: 32px;
    background: transparent; border: none;
    color: var(--text-dim);
    display: flex; align-items: center; justify-content: center;
    border-radius: 8px; cursor: pointer; transition: all 0.2s;
  }
  .nav-btn:hover { background: rgba(255,255,255,0.05); color: #fff; }
  .nav-btn.action { color: var(--primary-color); }
  
  .zoom-display { font-size: 12px; font-weight: 700; width: 50px; text-align: center; color: #fff; }
  .h-divider { width: 1px; height: 16px; background: rgba(255,255,255,0.1); margin: 0 8px; }

  .btn-utility {
    width: 36px; height: 36px;
    background: transparent; border: 1px solid var(--border-color);
    border-radius: 8px; color: var(--text-dim);
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; transition: all 0.2s;
  }
  .btn-utility:hover { border-color: #555; color: #fff; }

  /* --- Canvas --- */
  .canvas-area {
    flex: 1;
    position: relative;
    overflow: hidden;
    background-color: var(--architect-bg);
    background-image: 
      radial-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
      linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
    background-size: 40px 40px, 120px 120px, 120px 120px;
    cursor: default;
  }
  .canvas-area.panning { cursor: grabbing; }

  .canvas-view {
    position: absolute;
    width: 10000px;
    height: 10000px;
    top: 0; left: 0;
  }

  /* --- Relationships --- */
  .blueprint-lines {
    position: absolute;
    top: 0; left: 0;
    width: 100%; height: 100%;
    pointer-events: none;
    z-index: 10;
  }

  .line-main {
    fill: none;
    stroke: #6366f1;
    stroke-width: 2;
    opacity: 0.8;
    stroke-dasharray: 6 4;
  }

  .line-shadow {
    fill: none;
    stroke: #6366f1;
    stroke-width: 6;
    opacity: 0.1;
    filter: blur(2px);
  }

  .relationship-group:hover .line-main { opacity: 1; stroke-width: 2.5; stroke-dasharray: none; }

  /* --- Table Nodes --- */
  .table-node {
    position: absolute;
    width: 280px;
    z-index: 20;
    border-radius: 12px;
    border: 1px solid var(--node-border);
    transition: box-shadow 0.3s, transform 0.2s;
    user-select: none;
  }

  .node-glass {
    position: absolute;
    top: 0; left: 0; right: 0; bottom: 0;
    background: var(--node-bg);
    backdrop-filter: blur(20px);
    border-radius: 12px;
    z-index: -1;
  }

  .table-node:hover {
    box-shadow: 0 10px 40px rgba(0,0,0,0.5), 0 0 20px var(--primary-glow);
    border-color: rgba(var(--primary-rgb), 0.5);
    z-index: 50;
  }

  .table-node.dragging {
    z-index: 1000;
    transform: scale(1.02);
    box-shadow: 0 20px 60px rgba(0,0,0,0.6);
    border-color: var(--primary-color);
  }

  .node-header {
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-left { display: flex; align-items: center; gap: 10px; }
  .table-icon-box {
    width: 24px; height: 24px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary-color);
    border-radius: 6px;
    display: flex; align-items: center; justify-content: center;
  }

  .table-name { font-weight: 700; font-size: 14px; color: #fff; letter-spacing: -0.01em; }

  .node-menu { background: none; border: none; color: var(--text-dim); cursor: pointer; opacity: 0.5; }
  .node-menu:hover { opacity: 1; color: #fff; }

  .node-columns { padding: 8px 0; }

  .column-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 16px;
    font-size: 13px;
    transition: background 0.1s;
  }
  .column-row:hover { background: rgba(255,255,255,0.03); }
  .column-row.is-pk { color: #ffd76d; font-weight: 600; }

  .col-left { display: flex; align-items: center; gap: 8px; }
  .col-dot { width: 4px; height: 4px; border-radius: 50%; background: #454d58; }
  .pk-key { color: #ffd76d; filter: drop-shadow(0 0 4px rgba(255, 215, 109, 0.4)); }

  .col-type { font-size: 11px; color: var(--text-dim); font-family: 'JetBrains Mono', monospace; opacity: 0.6; }

  .node-footer {
    padding: 8px 16px;
    border-top: 1px dotted rgba(255,255,255,0.1);
    background: rgba(var(--primary-rgb), 0.03);
    border-radius: 0 0 12px 12px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 11px;
    font-weight: 600;
    color: var(--primary-color);
  }

  .rls-shield {
    margin-left: auto;
    width: 16px; height: 16px;
    background: var(--primary-color);
    color: #000;
    border-radius: 4px;
    display: flex; align-items: center; justify-content: center;
  }

  /* --- UI Controls --- */
  .architect-hud {
    position: absolute;
    bottom: 24px; left: 50%;
    transform: translateX(-50%);
    pointer-events: none;
    z-index: 100;
  }

  .hud-pill {
    background: rgba(13, 17, 23, 0.8);
    backdrop-filter: blur(12px);
    border: 1px solid var(--border-color);
    padding: 10px 20px;
    border-radius: 99px;
    display: flex; align-items: center; gap: 12px;
    font-size: 12px; font-weight: 500; color: var(--text-dim);
    box-shadow: 0 8px 32px rgba(0,0,0,0.3);
  }

  /* --- Utils --- */
  .canvas-loading {
    height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center;
    gap: 20px; color: var(--text-dim);
  }
  .loader-ring {
    width: 48px; height: 48px;
    border: 2px solid rgba(255,255,255,0.05);
    border-top-color: var(--primary-color);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes pulse {
    0% { box-shadow: 0 0 0 0 rgba(46, 204, 113, 0.4); }
    70% { box-shadow: 0 0 0 10px rgba(46, 204, 113, 0); }
    100% { box-shadow: 0 0 0 0 rgba(46, 204, 113, 0); }
  }

  @keyframes dash {
    from { stroke-dashoffset: 0; }
    to { stroke-dashoffset: 1000; }
  }

  @keyframes spin { to { transform: rotate(360deg); } }
</style>
