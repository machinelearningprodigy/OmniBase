<script lang="ts">
  import { page } from '$app/stores'

  const navItems = [
    {
      section: 'Project',
      items: [
        { label: 'Overview', href: '/', icon: 'home' },
        { label: 'API Logs', href: '/logs', icon: 'zap' },
        { label: 'API Settings', href: '/settings', icon: 'settings' },
      ]
    },
    {
      section: 'Database',
      items: [
        { label: 'Table Editor', href: '/database', icon: 'table' },
        { label: 'SQL Editor', href: '/database/editor', icon: 'code' },
        { label: 'DB Functions', href: '/database/functions', icon: 'zap' },
        { label: 'Migrations', href: '/database/migrations', icon: 'git-branch' },
        { label: 'RLS Policies', href: '/database/rls', icon: 'shield' },
      ]
    },
    {
      section: 'Logic',
      items: [
        { label: 'Edge Functions', href: '/functions', icon: 'zap' },
      ]
    },
    {
      section: 'Auth',
      items: [
        { label: 'Users', href: '/auth/users', icon: 'users' },
        { label: 'Providers', href: '/auth/providers', icon: 'key' },
        { label: 'Policies', href: '/auth/policies', icon: 'lock' },
      ]
    },
    {
      section: 'Storage',
      items: [
        { label: 'Buckets', href: '/storage', icon: 'database' },
      ]
    },
    {
      section: 'Realtime',
      items: [
        { label: 'Inspector', href: '/realtime', icon: 'activity' },
      ]
    },
    {
      section: 'API',
      items: [
        { label: 'GraphQL', href: '/graphql', icon: 'zap' },
      ]
    },
  ]

  // Simple icon renderer (using Unicode + CSS)
  function getIcon(name: string) {
    const icons: Record<string, string> = {
      'home': '⌂', 'settings': '⚙', 'table': '◫', 'code': '‹›',
      'git-branch': '⑂', 'shield': '⊕', 'users': '◎', 'key': '⚷',
      'lock': '⊗', 'database': '⊞', 'activity': '⍁', 'zap': '⚡',
    }
    return icons[name] ?? '•'
  }

  let currentPath = $derived($page.url.pathname)
</script>

<aside class="sidebar">
  <!-- Logo -->
  <a href="/" class="logo" style="margin-bottom: 20px; padding: 0 8px; text-decoration: none;">
    <div class="logo-icon">Ω</div>
    <span>OmniBase</span>
  </a>

  {#each navItems as group}
    <div class="nav-section">
      <div class="nav-section-label">{group.section}</div>
      {#each group.items as item}
        <a
          href={item.href}
          class="nav-item {currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href)) ? 'active' : ''}"
        >
          <span class="icon" style="font-style: normal; font-size: 14px;">{getIcon(item.icon)}</span>
          {item.label}
        </a>
      {/each}
    </div>
  {/each}

  <div style="margin-top: auto; padding: 12px 8px; border-top: 1px solid var(--border-subtle);">
    <div style="font-size: 11px; color: var(--text-muted); margin-bottom: 8px;">OmniBase v0.1.0 · Phase 1</div>
    <a href="https://github.com/machinelearningprodigy/OmniBase" target="_blank" rel="noopener" class="nav-item" style="font-size: 12px;">
      <span>★</span> Star on GitHub
    </a>
  </div>
</aside>
