<script lang="ts">
  import { page } from '$app/stores'
  import { goto } from '$app/navigation'

  // ── Sidebar state ─────────────────────────────────────────────────────────
  let pinned = $state(false)      // toggled by clicking the logo/pin icon
  let hovered = $state(false)     // true while mouse is inside sidebar
  let expanded = $derived(pinned || hovered)

  // ── Tree state (which sections are open) ──────────────────────────────────
  let openSections = $state<Record<string, boolean>>({
    database: true,
    auth: true,
    realtime: false,
    storage: false,
    functions: false,
    messaging: false,
    analytics: false,
    aiEngine: false,
    hosting: false,
    observability: false,
    settings: false,
  })

  function toggleSection(key: string) {
    openSections[key] = !openSections[key]
  }

  let currentPath = $derived($page.url.pathname)

  function isActive(href: string) {
    if (href === '/') return currentPath === '/'
    return currentPath.startsWith(href)
  }

  // ── Navigation tree ───────────────────────────────────────────────────────
  // OmniBase-branded names (not copied from anywhere specific)
  const nav = [
    {
      key: null,
      label: 'HOME',
      items: [
        { label: 'Overview',        href: '/',           icon: 'home',        badge: null },
        { label: 'API Gateway',     href: '/gateway',    icon: 'api',         badge: null },
      ]
    },
    {
      key: 'database',
      label: 'DATA ENGINE',
      items: [
        { label: 'Table Studio',    href: '/database',           icon: 'table',        badge: null },
        { label: 'SQL Lab',         href: '/database/editor',    icon: 'code',         badge: null },
        { label: 'Schema Architect',href: '/database/schema',    icon: 'schema',       badge: null },
        { label: 'Access Policies', href: '/database/rls',       icon: 'shield',       badge: null },
        { label: 'DB Functions',    href: '/database/functions', icon: 'function',     badge: null },
        { label: 'Triggers',        href: '/database/triggers',  icon: 'zap',          badge: null },
        { label: 'Enum Types',      href: '/database/enums',     icon: 'enum',         badge: null },
        { label: 'Extensions',      href: '/database/extensions',icon: 'puzzle',       badge: null },
        { label: 'Indexes',         href: '/database/indexes',   icon: 'index',        badge: null },
        { label: 'Cron Jobs',       href: '/database/cron',      icon: 'clock',        badge: null },
        { label: 'Publications',    href: '/database/publications', icon: 'broadcast', badge: null },
        { label: 'Roles & Access',  href: '/database/roles',     icon: 'role',         badge: null },
        { label: 'Graph Queries',   href: '/database/graph',     icon: 'graph',        badge: 'NEW' },
        { label: 'Migrations',      href: '/database/migrations',icon: 'migrate',      badge: null },
        { label: 'DB Backups',      href: '/database/backups',   icon: 'backup',       badge: null },
        { label: 'Replication',     href: '/database/replication',icon: 'replicate',   badge: null },
        { label: 'Wrappers',        href: '/database/wrappers',  icon: 'wrap',         badge: null },
        { label: 'DB Branching',    href: '/database/branches',  icon: 'branch',       badge: 'BETA' },
        { label: 'Webhooks',        href: '/database/webhooks',  icon: 'webhook',      badge: null },
      ]
    },
    {
      key: 'auth',
      label: 'IDENTITY',
      items: [
        { label: 'Users',           href: '/auth/users',         icon: 'users',        badge: null },
        { label: 'OAuth Apps',      href: '/auth/providers',     icon: 'oauth',        badge: null },
        { label: 'Email Flows',     href: '/auth/email',         icon: 'email',        badge: null },
        { label: 'Identity Rules',  href: '/auth/policies',      icon: 'lock',         badge: null },
        { label: 'Auth Server',     href: '/auth/server',        icon: 'server',       badge: 'BETA' },
        { label: 'Sessions',        href: '/auth/sessions',      icon: 'session',      badge: null },
        { label: 'Rate Limits',     href: '/auth/rate-limits',   icon: 'ratelimit',    badge: null },
        { label: 'Multi-Factor',    href: '/auth/mfa',           icon: 'mfa',          badge: null },
        { label: 'Passkeys',        href: '/auth/passkeys',      icon: 'passkey',      badge: 'NEW' },
        { label: 'Web3 Auth',       href: '/auth/web3',          icon: 'web3',         badge: 'NEW' },
        { label: 'URL Config',      href: '/auth/urls',          icon: 'link',         badge: null },
        { label: 'Attack Guard',    href: '/auth/protection',    icon: 'guard',        badge: null },
        { label: 'Auth Hooks',      href: '/auth/hooks',         icon: 'hook',         badge: 'BETA' },
        { label: 'Audit Logs',      href: '/auth/audit',         icon: 'audit',        badge: null },
        { label: 'SAML / SSO',      href: '/auth/saml',          icon: 'saml',         badge: null },
        { label: 'Performance',     href: '/auth/performance',   icon: 'perf',         badge: null },
      ]
    },
    {
      key: 'realtime',
      label: 'LIVE SYNC',
      items: [
        { label: 'Inspector',       href: '/realtime',           icon: 'activity',     badge: null },
        { label: 'Channels',        href: '/realtime/channels',  icon: 'channel',      badge: null },
        { label: 'Presence',        href: '/realtime/presence',  icon: 'presence',     badge: null },
        { label: 'Broadcast',       href: '/realtime/broadcast', icon: 'broadcast',    badge: null },
        { label: 'Sync Rules',      href: '/realtime/rules',     icon: 'shield',       badge: null },
        { label: 'Sync Settings',   href: '/realtime/settings',  icon: 'settings',     badge: null },
      ]
    },
    {
      key: 'storage',
      label: 'OBJECT STORE',
      items: [
        { label: 'Buckets',         href: '/storage',            icon: 'bucket',       badge: null },
        { label: 'Image Pipeline',  href: '/storage/transform',  icon: 'image',        badge: null },
        { label: 'CDN Config',      href: '/storage/cdn',        icon: 'cdn',          badge: null },
        { label: 'File Policies',   href: '/storage/policies',   icon: 'lock',         badge: null },
        { label: 'Storage Settings',href: '/storage/settings',   icon: 'settings',     badge: null },
      ]
    },
    {
      key: 'functions',
      label: 'COMPUTE',
      items: [
        { label: 'Edge Functions',  href: '/functions',          icon: 'function',     badge: null },
        { label: 'Function Logs',   href: '/functions/logs',     icon: 'log',          badge: null },
        { label: 'Cron Scheduler',  href: '/functions/cron',     icon: 'clock',        badge: null },
        { label: 'Job Queues',      href: '/functions/queues',   icon: 'queue',        badge: null },
        { label: 'Secrets Vault',   href: '/functions/secrets',  icon: 'lock',         badge: null },
        { label: 'Runtimes',        href: '/functions/runtimes', icon: 'runtime',      badge: null },
        { label: 'Deployments',     href: '/functions/deploys',  icon: 'deploy',       badge: null },
      ]
    },
    {
      key: 'messaging',
      label: 'MESSAGING',
      items: [
        { label: 'Push Notifications', href: '/messaging/push',   icon: 'bell',        badge: null },
        { label: 'In-App Inbox',    href: '/messaging/inbox',     icon: 'inbox',       badge: null },
        { label: 'Email Campaigns', href: '/messaging/email',     icon: 'email',       badge: null },
        { label: 'SMS',             href: '/messaging/sms',       icon: 'sms',         badge: null },
        { label: 'Topics',          href: '/messaging/topics',    icon: 'topic',       badge: null },
        { label: 'Templates',       href: '/messaging/templates', icon: 'template',    badge: null },
        { label: 'Delivery Logs',   href: '/messaging/logs',      icon: 'log',         badge: null },
        { label: 'Providers',       href: '/messaging/providers', icon: 'plug',        badge: null },
      ]
    },
    {
      key: 'analytics',
      label: 'ANALYTICS',
      items: [
        { label: 'Crash Reports',   href: '/analytics/crashes',  icon: 'crash',       badge: null },
        { label: 'Performance',     href: '/analytics/perf',     icon: 'perf',        badge: null },
        { label: 'Event Tracking',  href: '/analytics/events',   icon: 'event',       badge: null },
        { label: 'Funnels',         href: '/analytics/funnels',  icon: 'funnel',      badge: null },
        { label: 'Retention',       href: '/analytics/retention',icon: 'chart',       badge: null },
        { label: 'A/B Testing',     href: '/analytics/ab',       icon: 'ab',          badge: null },
        { label: 'Session Replay',  href: '/analytics/replay',   icon: 'replay',      badge: 'BETA' },
        { label: 'Remote Config',   href: '/analytics/config',   icon: 'tune',        badge: null },
      ]
    },
    {
      key: 'aiEngine',
      label: 'AI ENGINE',
      items: [
        { label: 'Vector Search',   href: '/ai/vector',          icon: 'vector',      badge: null },
        { label: 'Hybrid Search',   href: '/ai/hybrid',          icon: 'search',      badge: null },
        { label: 'Embeddings',      href: '/ai/embeddings',      icon: 'embed',       badge: null },
        { label: 'AI Query Builder',href: '/ai/query',           icon: 'ai',          badge: 'NEW' },
        { label: 'Model Adapters',  href: '/ai/adapters',        icon: 'plug',        badge: null },
        { label: 'Semantic Cache',  href: '/ai/cache',           icon: 'cache',       badge: 'NEW' },
        { label: 'Agent Workflows', href: '/ai/agents',          icon: 'agent',       badge: 'NEW' },
      ]
    },
    {
      key: 'hosting',
      label: 'HOSTING',
      items: [
        { label: 'Sites',           href: '/hosting',            icon: 'globe',       badge: null },
        { label: 'Deployments',     href: '/hosting/deploys',    icon: 'deploy',      badge: null },
        { label: 'Custom Domains',  href: '/hosting/domains',    icon: 'domain',      badge: null },
        { label: 'Build Pipeline',  href: '/hosting/builds',     icon: 'build',       badge: null },
        { label: 'Edge Config',     href: '/hosting/edge',       icon: 'cdn',         badge: null },
        { label: 'Env Variables',   href: '/hosting/env',        icon: 'env',         badge: null },
      ]
    },
    {
      key: 'observability',
      label: 'OBSERVABILITY',
      items: [
        { label: 'API Logs',        href: '/logs',               icon: 'log',         badge: null },
        { label: 'Metrics',         href: '/observability/metrics', icon: 'chart',    badge: null },
        { label: 'Traces',          href: '/observability/traces',  icon: 'trace',    badge: null },
        { label: 'Advisors',        href: '/observability/advisors',icon: 'advisor',  badge: null },
        { label: 'Alerts',          href: '/observability/alerts',  icon: 'alert',    badge: null },
      ]
    },
    {
      key: 'settings',
      label: 'PROJECT',
      items: [
        { label: 'API Settings',    href: '/settings',           icon: 'settings',    badge: null },
        { label: 'Integrations',    href: '/settings/integrations', icon: 'plug',    badge: null },
        { label: 'Team Members',    href: '/settings/team',      icon: 'users',       badge: null },
        { label: 'API Keys',        href: '/settings/keys',      icon: 'key',         badge: null },
        { label: 'Billing',         href: '/settings/billing',   icon: 'billing',     badge: null },
        { label: 'GDPR Tools',      href: '/settings/gdpr',      icon: 'shield',      badge: null },
        { label: 'Danger Zone',     href: '/settings/danger',    icon: 'alert',       badge: null },
      ]
    },
  ]

  // ── SVG icon paths (inline, no external deps) ─────────────────────────────
  const icons: Record<string, string> = {
    'home':       'M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10',
    'api':        'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8',
    'table':      'M3 3h18v18H3z M3 9h18 M3 15h18 M9 3v18',
    'code':       'M16 18l6-6-6-6 M8 6L2 12l6 6',
    'schema':     'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    'shield':     'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z',
    'function':   'M9 3H5a2 2 0 0 0-2 2v4m6-6h10a2 2 0 0 1 2 2v4M9 3v18m0 0h10a2 2 0 0 0 2-2V9M9 21H5a2 2 0 0 1-2-2V9m0 0h18',
    'zap':        'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
    'enum':       'M8 6h13 M8 12h13 M8 18h13 M3 6h.01 M3 12h.01 M3 18h.01',
    'puzzle':     'M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z',
    'index':      'M4 6h16 M4 12h16 M4 18h7',
    'clock':      'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 6v6l4 2',
    'broadcast':  'M2 20h20 M12 4v.01 M6.343 6.343l-.707.707 M17.657 6.343l.707.707 M3 12h.01 M21 12h-.01 M12 17.657l-4.243-4.243a6 6 0 1 1 8.486 0z',
    'role':       'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2 M9 7a4 4 0 1 0 0-8 4 4 0 0 0 0 8z M23 21v-2a4 4 0 0 0-3-3.87 M16 3.13a4 4 0 0 1 0 7.75',
    'graph':      'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M8 12h8 M12 8v8',
    'migrate':    'M19 9l-7 7-7-7',
    'backup':     'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4 M7 10l5 5 5-5 M12 15V3',
    'replicate':  'M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3',
    'wrap':       'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
    'branch':     'M6 3v12 M18 9a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M6 21a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M18 9a9 9 0 0 1-9 9',
    'webhook':    'M18 8h1a4 4 0 0 1 0 8h-1 M2 8h16v9a4 4 0 0 1-4 4H6a4 4 0 0 1-4-4V8z M6 1v3 M10 1v3 M14 1v3',
    'users':      'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2 M9 7a4 4 0 1 0 0-8 4 4 0 0 0 0 8z M23 21v-2a4 4 0 0 0-3-3.87 M16 3.13a4 4 0 0 1 0 7.75',
    'oauth':      'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M2 12h20 M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z',
    'email':      'M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z M22 6l-10 7L2 6',
    'lock':       'M19 11H5a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7a2 2 0 0 0-2-2z M7 11V7a5 5 0 0 1 10 0v4',
    'server':     'M20 17h2a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-2 M0 17h2a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2H0 M4 7L20 7 M4 17L20 17',
    'session':    'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 8v4l3 3',
    'ratelimit':  'M18 20V10 M12 20V4 M6 20v-6',
    'mfa':        'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z M9 12l2 2 4-4',
    'passkey':    'M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4',
    'web3':       'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    'link':       'M15 7h3a5 5 0 0 1 5 5 5 5 0 0 1-5 5h-3m-6 0H6a5 5 0 0 1-5-5 5 5 0 0 1 5-5h3 M8 12h8',
    'guard':      'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z M12 8v4 M12 16h.01',
    'hook':       'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71',
    'audit':      'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8',
    'saml':       'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M4.93 4.93l14.14 14.14',
    'perf':       'M22 12h-4l-3 9L9 3l-3 9H2',
    'activity':   'M22 12h-4l-3 9L9 3l-3 9H2',
    'channel':    'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
    'presence':   'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2 M9 7a4 4 0 1 0 0-8 4 4 0 0 0 0 8z M22 10a10 10 0 0 0-10-8',
    'bucket':     'M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z',
    'image':      'M21 15.999V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16v-.001z M3.27 6.96L12 12.01l8.73-5.05 M12 22.08V12',
    'cdn':        'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    'settings':   'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z',
    'log':        'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M12 18v-6 M9 15l3-3 3 3',
    'queue':      'M8 6h13 M8 12h13 M8 18h13 M3 6h.01 M3 12h.01 M3 18h.01',
    'runtime':    'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    'deploy':     'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4 M17 8l-5-5-5 5 M12 3v12',
    'bell':       'M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9 M13.73 21a2 2 0 0 1-3.46 0',
    'inbox':      'M22 12h-6l-2 3h-4l-2-3H2 M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z',
    'sms':        'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z M8 10h.01 M12 10h.01 M16 10h.01',
    'topic':      'M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z M4 22v-7',
    'template':   'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8',
    'plug':       'M7 2v11m5-11v11m5-11v11m-15 3h20a1 1 0 0 1 1 1v6H1v-6a1 1 0 0 1 1-1z',
    'crash':      'M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10z M12 8v4 M12 16h.01',
    'event':      'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 6v6l4 2',
    'funnel':     'M22 3H2l8 9.46V19l4 2v-8.54L22 3z',
    'chart':      'M18 20V10 M12 20V4 M6 20v-6',
    'ab':         'M18 20V10 M12 20V4 M6 20v-6',
    'replay':     'M1 4v6h6 M3.51 15a9 9 0 1 0 .49-3.76',
    'tune':       'M4 21v-7 M4 10V3 M12 21v-9 M12 8V3 M20 21v-5 M20 12V3 M1 14h6 M9 8h6 M17 16h6',
    'vector':     'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    'search':     'M21 21l-6-6m2-5a7 7 0 1 1-14 0 7 7 0 0 1 14 0z',
    'embed':      'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    'ai':         'M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a7 7 0 0 1 7 7h1a1 1 0 0 1 0 2h-1v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-1H2a1 1 0 0 1 0-2h1a7 7 0 0 1 7-7h1V5.73A2 2 0 0 1 10 4a2 2 0 0 1 2-2z M9.5 11a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3z M14.5 11a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3z',
    'cache':      'M12 2L2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5',
    'agent':      'M12 8V4l-1-1 M8 8H4l-1 1 M16 8h4l1 1 M12 16v4l1 1 M4.5 15l-2 2 M19.5 15l2 2 M12 12m-3 0a3 3 0 1 0 6 0 3 3 0 0 0-6 0',
    'globe':      'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M2 12h20 M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z',
    'domain':     'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 8v4l3 3',
    'build':      'M2 20h.01 M7 20v-4 M12 20v-8 M17 20V8 M22 4v16',
    'env':        'M9 3H5a2 2 0 0 0-2 2v4m6-6h10a2 2 0 0 1 2 2v4M9 3v18m0 0h10a2 2 0 0 0 2-2V9M9 21H5a2 2 0 0 1-2-2V9m0 0h18',
    'metric':     'M22 12h-4l-3 9L9 3l-3 9H2',
    'trace':      'M22 12h-4l-3 9L9 3l-3 9H2',
    'advisor':    'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 16v-4 M12 8h.01',
    'alert':      'M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z M12 9v4 M12 17h.01',
    'key':        'M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4',
    'billing':    'M20 12V22H4V12 M22 7H2v5h20V7z M12 22V7 M12 7H7.5a2.5 2.5 0 0 1 0-5C11 2 12 7 12 7z M12 7h4.5a2.5 2.5 0 0 0 0-5C13 2 12 7 12 7z',
    'pin':        'M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z M12 13a3 3 0 1 0 0-6 3 3 0 0 0 0 6z',
    'metrics':    'M18 20V10 M12 20V4 M6 20v-6',
    'chevron':    'M9 18l6-6-6-6',
    'pin-icon':   'M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z M12 13a3 3 0 1 0 0-6 3 3 0 0 0 0 6z',
  }

  function icon(name: string) {
    return icons[name] ?? icons['settings']
  }
</script>

<!-- ═══════════════════════════════════════════════════════════════════════════
     SIDEBAR
     • collapsed = icon rail (64px) when mouse-out
     • expanded  = full panel (260px) when mouse-over OR pinned
     • pin/unpin = click the logo area
══════════════════════════════════════════════════════════════════════════════ -->
<aside
  class="sidebar-v2"
  class:expanded
  class:pinned
  onmouseenter={() => (hovered = true)}
  onmouseleave={() => (hovered = false)}
>
  <!-- ── Logo / Pin button ──────────────────────────────────────────────── -->
  <button class="sb-logo" onclick={() => (pinned = !pinned)} title={pinned ? 'Unpin sidebar' : 'Pin sidebar open'}>
    <div class="sb-logo-icon">Ω</div>
    {#if expanded}
      <span class="sb-logo-text">OmniBase</span>
      <svg class="pin-icon" class:pinned viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d={icon('pin-icon')} />
      </svg>
    {/if}
  </button>

  <!-- ── Nav sections ───────────────────────────────────────────────────── -->
  <nav class="sb-nav">
    {#each nav as group}
      {#if group.key === null}
        <!-- Non-collapsible section -->
        <div class="sb-section">
          {#if expanded}
            <div class="sb-section-label">{group.label}</div>
          {/if}
          {#each group.items as item}
            <a
              href={item.href}
              class="sb-item"
              class:active={isActive(item.href)}
              title={!expanded ? item.label : undefined}
            >
              <span class="sb-item-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d={icon(item.icon)} />
                </svg>
              </span>
              {#if expanded}
                <span class="sb-item-label">{item.label}</span>
                {#if item.badge}
                  <span class="sb-badge sb-badge-{item.badge.toLowerCase()}">{item.badge}</span>
                {/if}
              {/if}
            </a>
          {/each}
        </div>
      {:else}
        <!-- Collapsible tree section -->
        <div class="sb-section">
          <button
            class="sb-section-header"
            onclick={() => toggleSection(group.key!)}
            title={!expanded ? group.label : undefined}
          >
            {#if !expanded}
              <!-- collapsed: show first item icon as section icon -->
              <span class="sb-item-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d={icon(group.items[0].icon)} />
                </svg>
              </span>
            {:else}
              <span class="sb-section-label-inline">{group.label}</span>
              <svg
                class="sb-chevron"
                class:open={openSections[group.key!]}
                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              >
                <path d={icon('chevron')} />
              </svg>
            {/if}
          </button>

          {#if !expanded || openSections[group.key!]}
            {#each group.items as item}
              <a
                href={item.href}
                class="sb-item"
                class:active={isActive(item.href)}
                class:sb-item-child={expanded}
                title={!expanded ? item.label : undefined}
              >
                <span class="sb-item-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                    <path d={icon(item.icon)} />
                  </svg>
                </span>
                {#if expanded}
                  <span class="sb-item-label">{item.label}</span>
                  {#if item.badge}
                    <span class="sb-badge sb-badge-{item.badge.toLowerCase()}">{item.badge}</span>
                  {/if}
                {/if}
              </a>
            {/each}
          {/if}
        </div>
      {/if}
    {/each}
  </nav>

  <!-- ── Footer ────────────────────────────────────────────────────────── -->
  <div class="sb-footer">
    {#if expanded}
      <div class="sb-footer-text">OmniBase v0.1 · Phase 1</div>
    {/if}
    <a
      href="https://github.com/machinelearningprodigy/OmniBase"
      target="_blank" rel="noopener"
      class="sb-item"
      title="Star on GitHub"
    >
      <span class="sb-item-icon">
        <svg viewBox="0 0 24 24" fill="currentColor" stroke="none">
          <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2z" />
        </svg>
      </span>
      {#if expanded}
        <span class="sb-item-label">Star on GitHub</span>
      {/if}
    </a>
  </div>
</aside>

<style>
  /* ── Sidebar shell ──────────────────────────────────────────────────────── */
  .sidebar-v2 {
    position: relative;
    width: 64px;
    min-height: 0;
    height: 100%;
    background: linear-gradient(180deg, #0b0b14 0%, #0d0d1c 100%);
    border-right: 1px solid rgba(255,255,255,0.06);
    display: flex;
    flex-direction: column;
    transition: width 220ms cubic-bezier(0.4, 0, 0.2, 1);
    overflow: hidden;
    z-index: 200;
  }

  .sidebar-v2.expanded {
    width: 260px;
    box-shadow: 4px 0 32px rgba(0,0,0,0.5);
  }

  /* ── Logo / pin button ─────────────────────────────────────────────────── */
  .sb-logo {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 16px;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-primary);
    width: 100%;
    border-bottom: 1px solid rgba(255,255,255,0.05);
    flex-shrink: 0;
    white-space: nowrap;
    overflow: hidden;
    transition: background 150ms ease;
  }

  .sb-logo:hover {
    background: rgba(108,71,255,0.08);
  }

  .sb-logo-icon {
    width: 32px;
    height: 32px;
    background: linear-gradient(135deg, #6c47ff 0%, #00d4ff 100%);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 900;
    color: white;
    box-shadow: 0 0 18px rgba(108,71,255,0.4);
    flex-shrink: 0;
  }

  .sb-logo-text {
    font-size: 15px;
    font-weight: 700;
    background: linear-gradient(135deg, #c4b5fd, #7dd3fc);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    flex: 1;
  }

  .pin-icon {
    width: 14px;
    height: 14px;
    opacity: 0.4;
    transition: opacity 150ms, transform 150ms;
    flex-shrink: 0;
  }

  .pin-icon.pinned {
    opacity: 1;
    color: #6c47ff;
    -webkit-text-fill-color: #6c47ff;
  }

  /* ── Nav ──────────────────────────────────────────────────────────────── */
  .sb-nav {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 8px 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    scrollbar-width: thin;
    scrollbar-color: rgba(255,255,255,0.08) transparent;
  }

  .sb-nav::-webkit-scrollbar { width: 4px; }
  .sb-nav::-webkit-scrollbar-track { background: transparent; }
  .sb-nav::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.08); border-radius: 2px; }

  .sb-section {
    margin-bottom: 4px;
  }

  /* Section heading (collapsible) */
  .sb-section-header {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 8px;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted, #5050a0);
    border-radius: 6px;
    white-space: nowrap;
    overflow: hidden;
    transition: background 150ms, color 150ms;
    min-height: 32px;
  }

  .sb-section-header:hover {
    background: rgba(255,255,255,0.04);
    color: var(--text-secondary, #9090bb);
  }

  .sb-section-label {
    font-size: 9.5px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--text-muted, #5050a0);
    padding: 4px 8px 2px;
    white-space: nowrap;
  }

  .sb-section-label-inline {
    font-size: 9.5px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    flex: 1;
    text-align: left;
    color: inherit;
  }

  .sb-chevron {
    width: 12px;
    height: 12px;
    flex-shrink: 0;
    transition: transform 200ms ease;
    transform: rotate(0deg);
  }

  .sb-chevron.open {
    transform: rotate(90deg);
  }

  /* Nav items */
  .sb-item {
    display: flex;
    align-items: center;
    gap: 9px;
    padding: 7px 8px;
    border-radius: 7px;
    color: var(--text-secondary, #9090bb);
    text-decoration: none;
    font-size: 12.5px;
    font-weight: 500;
    transition: background 120ms ease, color 120ms ease;
    cursor: pointer;
    border: none;
    background: none;
    width: 100%;
    text-align: left;
    white-space: nowrap;
    overflow: hidden;
    position: relative;
  }

  .sb-item-child {
    padding-left: 12px;
  }

  .sb-item:hover {
    background: rgba(255,255,255,0.05);
    color: var(--text-primary, #f0f0ff);
  }

  .sb-item.active {
    background: rgba(108,71,255,0.16);
    color: #a78bfa;
    border: 1px solid rgba(108,71,255,0.22);
  }

  .sb-item.active .sb-item-icon svg {
    stroke: #a78bfa;
  }

  .sb-item-icon {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.75;
  }

  .sb-item.active .sb-item-icon { opacity: 1; }
  .sb-item:hover .sb-item-icon { opacity: 1; }

  .sb-item-icon svg {
    width: 16px;
    height: 16px;
  }

  .sb-item-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Badges */
  .sb-badge {
    font-size: 9.5px;
    font-weight: 700;
    padding: 1px 5px;
    border-radius: 4px;
    letter-spacing: 0.04em;
    flex-shrink: 0;
  }

  .sb-badge-beta {
    background: rgba(255,171,64,0.15);
    color: #ffab40;
    border: 1px solid rgba(255,171,64,0.3);
  }

  .sb-badge-new {
    background: rgba(0,230,118,0.15);
    color: #00e676;
    border: 1px solid rgba(0,230,118,0.3);
  }

  /* ── Footer ───────────────────────────────────────────────────────────── */
  .sb-footer {
    padding: 8px 8px 10px;
    border-top: 1px solid rgba(255,255,255,0.05);
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex-shrink: 0;
  }

  .sb-footer-text {
    font-size: 10px;
    color: #3a3a6a;
    padding: 2px 8px 4px;
    white-space: nowrap;
  }
</style>
