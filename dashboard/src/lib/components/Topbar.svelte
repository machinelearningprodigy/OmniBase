<script lang="ts">
  import { authStore } from '$lib/stores/auth'
</script>

<header class="topbar">
  <!-- Left: breadcrumb/project switcher -->
  <div class="flex items-center gap-2" style="flex: 1;">
    <div style="font-size: 12px; color: var(--text-muted);">
      <span style="color: var(--text-secondary); font-weight: 500;">Default Project</span>
    </div>
    <span style="color: var(--border-default);">/</span>
    <div style="font-size: 12px; color: var(--text-secondary);">us-east-1</div>
  </div>

  <!-- Center: search (Phase 2) -->
  <div style="flex: 1; max-width: 400px;">
    <div style="background: rgba(255,255,255,0.04); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 6px 12px; display: flex; align-items: center; gap: 8px; color: var(--text-muted); font-size: 12px;">
      <span>⌕</span>
      <span>Search or jump to...</span>
      <kbd style="margin-left: auto; background: rgba(255,255,255,0.06); padding: 1px 5px; border-radius: 3px; font-size: 10px; font-family: var(--font-mono);">⌘K</kbd>
    </div>
  </div>

  <!-- Right: status + user -->
  <div class="flex items-center gap-2" style="flex: 1; justify-content: flex-end;">
    <!-- Health indicator -->
    <div class="flex items-center gap-2" style="font-size: 11px; color: var(--text-secondary); background: rgba(0,230,118,0.08); border: 1px solid rgba(0,230,118,0.15); padding: 4px 10px; border-radius: 20px;">
      <span class="status-dot online"></span>
      All systems operational
    </div>

    <!-- Docs link -->
    <a href="https://docs.omnibase.dev" target="_blank" class="btn btn-secondary btn-sm" style="text-decoration: none;">
      Docs
    </a>

    <!-- User avatar -->
    <div style="width: 30px; height: 30px; border-radius: 50%; background: var(--gradient-brand); display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; cursor: pointer;" data-tooltip="admin@omnibase.dev">
      A
    </div>
  </div>
</header>
