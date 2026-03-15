<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getErrorMessage } from '$lib/api'
  import { authStore } from '$lib/stores/auth'
  import { fade, slide, scale } from 'svelte/transition'

  interface Provider {
    id?: string; name: string; type: string
    client_id: string; client_secret: string; is_active: boolean
  }

  let providers        = $state<Provider[]>([])
  let loading          = $state(false)
  let saving           = $state(false)
  let errorMsg         = $state<string | null>(null)
  let successMsg       = $state<string | null>(null)
  let selectedProvider = $state<Provider | null>(null)
  let showModal        = $state(false)
  let canClose         = $state(false)
  let copySuccess      = $state(false)

  // ── Provider metadata ──────────────────────────────────────────────────────
  const providerMeta: Record<string, { clientIdLabel: string; clientSecretLabel: string; consoleUrl: string; note: string }> = {
    google:     { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://console.cloud.google.com/apis/credentials',  note: 'Create an OAuth 2.0 credential in Google Cloud Console' },
    github:     { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://github.com/settings/developers',             note: 'Register an OAuth App under GitHub Developer Settings' },
    discord:    { clientIdLabel: 'Application ID',            clientSecretLabel: 'Client Secret',               consoleUrl: 'https://discord.com/developers/applications',        note: 'Create an application at Discord Developer Portal' },
    facebook:   { clientIdLabel: 'App ID',                    clientSecretLabel: 'App Secret',                  consoleUrl: 'https://developers.facebook.com/apps',               note: 'Create a Facebook App and enable Facebook Login' },
    microsoft:  { clientIdLabel: 'Application (Client) ID',   clientSecretLabel: 'Client Secret Value',         consoleUrl: 'https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps', note: 'Register an app in Azure Active Directory' },
    apple:      { clientIdLabel: 'Service ID (Client ID)',     clientSecretLabel: 'Private Key (JWT Secret)',    consoleUrl: 'https://developer.apple.com/account/resources/identifiers/list', note: 'Create a Services ID and Sign in with Apple key' },
    twitter:    { clientIdLabel: 'API Key / Client ID',        clientSecretLabel: 'API Secret / Client Secret', consoleUrl: 'https://developer.twitter.com/en/portal/dashboard',   note: 'Create a Twitter Developer App with OAuth 2.0 enabled' },
    linkedin:   { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://www.linkedin.com/developers/apps',           note: 'Create an app and add Sign In with LinkedIn product' },
    spotify:    { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://developer.spotify.com/dashboard',            note: 'Create an app in Spotify Developer Dashboard' },
    twitch:     { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://dev.twitch.tv/console/apps',                 note: 'Register an application at Twitch Developer Console' },
    slack:      { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://api.slack.com/apps',                         note: 'Create a Slack App and add OAuth & Permissions scope' },
    gitlab:     { clientIdLabel: 'Application ID',            clientSecretLabel: 'Secret',                      consoleUrl: 'https://gitlab.com/-/profile/applications',          note: 'Create an OAuth Application at GitLab User Settings' },
    bitbucket:  { clientIdLabel: 'Key (Client ID)',            clientSecretLabel: 'Secret',                      consoleUrl: 'https://bitbucket.org/account/settings/app-passwords/', note: 'Create an OAuth Consumer in Bitbucket Workspace Settings' },
    reddit:     { clientIdLabel: 'App ID (Client ID)',         clientSecretLabel: 'App Secret',                  consoleUrl: 'https://www.reddit.com/prefs/apps',                  note: 'Create a "web app" type application at Reddit Preferences' },
    dropbox:    { clientIdLabel: 'App Key (Client ID)',        clientSecretLabel: 'App Secret',                  consoleUrl: 'https://www.dropbox.com/developers/apps',            note: 'Create an app at Dropbox App Console' },
    zoom:       { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://marketplace.zoom.us/develop/create',         note: 'Create an OAuth app in Zoom Marketplace' },
    notion:     { clientIdLabel: 'OAuth Client ID',            clientSecretLabel: 'OAuth Client Secret',         consoleUrl: 'https://www.notion.so/my-integrations',              note: 'Create a public integration at Notion My Integrations' },
    atlassian:  { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://developer.atlassian.com/console/myapps/',    note: 'Create an OAuth 2.0 app at Atlassian Developer Console' },
    salesforce: { clientIdLabel: 'Consumer Key (Client ID)',   clientSecretLabel: 'Consumer Secret',             consoleUrl: 'https://login.salesforce.com',                       note: 'Create a Connected App in Salesforce Setup → App Manager' },
    hubspot:    { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://app.hubspot.com/developer',                  note: 'Create a Public App in HubSpot Developer Account' },
    box:        { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://app.box.com/developers/console',             note: 'Create an OAuth 2.0 app in Box Developer Console' },
    instagram:  { clientIdLabel: 'App ID (Client ID)',         clientSecretLabel: 'App Secret',                  consoleUrl: 'https://developers.facebook.com/apps',               note: 'Instagram Login uses the Meta (Facebook) Developer platform' },
    line:       { clientIdLabel: 'Channel ID (Client ID)',     clientSecretLabel: 'Channel Secret',              consoleUrl: 'https://developers.line.biz/console/',               note: 'Create a LINE Login channel at LINE Developers Console' },
    paypal:     { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://developer.paypal.com/developer/applications', note: 'Create an app in PayPal Developer Dashboard and use Sandbox for testing' },
    amazon:     { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://developer.amazon.com/loginwithamazon/console/site/lwa/overview.html', note: 'Register an app at Login with Amazon Developer Console' },
    tiktok:     { clientIdLabel: 'Client Key (Client ID)',     clientSecretLabel: 'Client Secret',               consoleUrl: 'https://developers.tiktok.com/apps/',                note: 'Create an app at TikTok for Developers and enable Login Kit' },
    pinterest:  { clientIdLabel: 'App ID (Client ID)',         clientSecretLabel: 'App Secret Key',              consoleUrl: 'https://developers.pinterest.com/apps/',             note: 'Create an app at Pinterest Developer Platform' },
    snapchat:   { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://kit.snapchat.com/apps',                      note: 'Create an app at Snap Kit for Login Kit OAuth' },
    yahoo:      { clientIdLabel: 'Client ID (Consumer Key)',   clientSecretLabel: 'Client Secret (Consumer Secret)', consoleUrl: 'https://developer.yahoo.com/apps/',             note: 'Create a Yahoo App and enable social login with OpenID Connect' },
    okta:       { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://developer.okta.com/',                        note: 'Create an OIDC Web Application in your Okta Admin Console' },
    yandex:     { clientIdLabel: 'Application ID (Client ID)', clientSecretLabel: 'Password (Client Secret)',    consoleUrl: 'https://oauth.yandex.com/',                          note: 'Register an application at Yandex OAuth service' },
    wordpress:  { clientIdLabel: 'Client ID',                 clientSecretLabel: 'Client Secret',               consoleUrl: 'https://developer.wordpress.com/apps/',              note: 'Create a WordPress.com application at the Developer Console' },
    vk:         { clientIdLabel: 'App ID (Client ID)',         clientSecretLabel: 'Secure Key (Client Secret)',  consoleUrl: 'https://vk.com/apps?act=manage',                     note: 'Create a VK application at VK for Developers' },
  }

  // ── Providers list ─────────────────────────────────────────────────────────
  const availableOAuth = [
    { id: 'google',     name: 'Google',      color: '#4285F4' },
    { id: 'github',     name: 'GitHub',       color: '#6e40c9' },
    { id: 'discord',    name: 'Discord',      color: '#5865F2' },
    { id: 'facebook',   name: 'Facebook',     color: '#1877F2' },
    { id: 'microsoft',  name: 'Microsoft',    color: '#00A4EF' },
    { id: 'apple',      name: 'Apple',        color: '#888888' },
    { id: 'twitter',    name: 'Twitter / X',  color: '#1d9bf0' },
    { id: 'linkedin',   name: 'LinkedIn',     color: '#0A66C2' },
    { id: 'spotify',    name: 'Spotify',      color: '#1DB954' },
    { id: 'twitch',     name: 'Twitch',       color: '#9146FF' },
    { id: 'slack',      name: 'Slack',        color: '#E01E5A' },
    { id: 'gitlab',     name: 'GitLab',       color: '#FC6D26' },
    { id: 'bitbucket',  name: 'Bitbucket',    color: '#0052CC' },
    { id: 'reddit',     name: 'Reddit',       color: '#FF4500' },
    { id: 'dropbox',    name: 'Dropbox',      color: '#0061FF' },
    { id: 'zoom',       name: 'Zoom',         color: '#2D8CFF' },
    { id: 'notion',     name: 'Notion',       color: '#888888' },
    { id: 'atlassian',  name: 'Atlassian',    color: '#0052CC' },
    { id: 'salesforce', name: 'Salesforce',   color: '#00A1E0' },
    { id: 'hubspot',    name: 'HubSpot',      color: '#FF7A59' },
    { id: 'box',        name: 'Box',          color: '#0061D5' },
    { id: 'instagram',  name: 'Instagram',    color: '#E4405F' },
    { id: 'line',       name: 'Line',         color: '#00C300' },
    { id: 'paypal',     name: 'PayPal',       color: '#003087' },
    { id: 'amazon',     name: 'Amazon',       color: '#FF9900' },
    { id: 'tiktok',     name: 'TikTok',       color: '#ff0050' },
    { id: 'pinterest',  name: 'Pinterest',    color: '#E60023' },
    { id: 'snapchat',   name: 'Snapchat',     color: '#FFFC00' },
    { id: 'yahoo',      name: 'Yahoo',        color: '#6001D2' },
    { id: 'okta',       name: 'Okta',         color: '#007DC1' },
    { id: 'yandex',     name: 'Yandex',       color: '#FC3F1D' },
    { id: 'wordpress',  name: 'WordPress',    color: '#21759B' },
    { id: 'vk',         name: 'VK',           color: '#4680C2' },
  ]

  const iconMap: Record<string, string> = {
    google: 'google', github: 'github', discord: 'discord', facebook: 'facebook',
    microsoft: 'microsoft', apple: 'apple', twitter: 'x', linkedin: 'linkedin',
    spotify: 'spotify', twitch: 'twitch', slack: 'slack', gitlab: 'gitlab',
    bitbucket: 'bitbucket', reddit: 'reddit', dropbox: 'dropbox', zoom: 'zoom',
    notion: 'notion', atlassian: 'atlassian', salesforce: 'salesforce', hubspot: 'hubspot',
    box: 'box', instagram: 'instagram', line: 'line',
    paypal: 'paypal', amazon: 'amazon', tiktok: 'tiktok', pinterest: 'pinterest',
    snapchat: 'snapchat', yahoo: 'yahoo', okta: 'okta', yandex: 'yandex',
    wordpress: 'wordpress', vk: 'vk',
  }
  function getIconUrl(id: string) {
    return `https://unpkg.com/simple-icons@10.0.0/icons/${iconMap[id] ?? id}.svg`
  }

  // ── API helpers ────────────────────────────────────────────────────────────
  async function loadProviders() {
    loading = true; errorMsg = null
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/oauth/apps`)
      providers = resp.ok ? (Array.isArray(await resp.json()) ? await resp.json() : []) : []
    } catch { providers = [] }
    finally { loading = false }
  }

  async function saveProvider(e: Event) {
    e.preventDefault()
    if (!selectedProvider) return
    saving = true; errorMsg = null
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/oauth/apps`, {
        method: 'POST',
        body: JSON.stringify(selectedProvider),
      })
      if (resp.ok) {
        const label = availableOAuth.find(a => a.id === selectedProvider!.name)?.name ?? selectedProvider!.name
        successMsg = `${label} saved successfully`
        closeModal()
        await loadProviders()
      } else {
        errorMsg = await getErrorMessage(resp, 'Failed to save')
      }
    } catch { errorMsg = 'Network error — OmniBase gateway may be offline' }
    finally { saving = false }
  }

  // ── Modal helpers ──────────────────────────────────────────────────────────
  function openConfig(e: MouseEvent, id: string) {
    e.preventDefault(); e.stopPropagation()
    const existing = providers.find(p => p.name.toLowerCase() === id)
    selectedProvider = { name: id, type: 'oauth', client_id: existing?.client_id ?? '', client_secret: existing?.client_secret ?? '', is_active: existing?.is_active ?? true }
    showModal = true; canClose = false
    setTimeout(() => { canClose = true }, 300)
  }

  function closeModal() {
    if (!canClose) return
    showModal = false; selectedProvider = null; errorMsg = null
  }

  function handleOverlayClick(e: MouseEvent) {
    if (!canClose) return
    if ((e.target as HTMLElement) === (e.currentTarget as HTMLElement)) {
      showModal = false; selectedProvider = null; errorMsg = null
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && showModal && canClose) {
      showModal = false; selectedProvider = null; errorMsg = null
    }
  }

  function isConfigured(id: string) { return providers.some(p => p.name.toLowerCase() === id && p.client_id) }
  function isActive(id: string)     { return providers.find(p => p.name.toLowerCase() === id)?.is_active ?? false }

  async function copyToClipboard(text: string) {
    try { await navigator.clipboard.writeText(text) } catch { return }
    copySuccess = true; setTimeout(() => { copySuccess = false }, 2000)
  }

  onMount(() => { loadProviders() })

  let meta         = $derived(selectedProvider ? (providerMeta[selectedProvider.name] ?? providerMeta.google)  : null)
  let selectedInfo = $derived(selectedProvider ? availableOAuth.find(a => a.id === selectedProvider!.name) : null)
  let callbackUrl  = $derived(selectedProvider ? `${getOmniBaseUrl()}/auth/v1/callback?provider=${selectedProvider.name}` : '')

  // Configured count
  let configuredCount = $derived(availableOAuth.filter(a => isConfigured(a.id)).length)
  let activeCount     = $derived(availableOAuth.filter(a => isConfigured(a.id) && isActive(a.id)).length)
</script>

<svelte:window onkeydown={handleKeydown} />
<svelte:head><title>Identity Providers — OmniBase</title></svelte:head>

<!-- ── Page ──────────────────────────────────────────────────────────────────── -->
<div class="ip-page">

  <!-- Header -->
  <div class="ip-top">
    <div>
      <div class="ip-breadcrumb">Settings / <span>Authentication</span></div>
      <h1 class="ip-title">Identity Providers</h1>
      <p class="ip-sub">Enable social logins and manage OAuth for your project.</p>
    </div>
    <div class="ip-top-r">
      <!-- Stats -->
      <div class="ip-stat-row">
        <div class="ip-stat">
          <div class="ip-stat-n">{availableOAuth.length}</div>
          <div class="ip-stat-l">Providers</div>
        </div>
        <div class="ip-stat-div"></div>
        <div class="ip-stat">
          <div class="ip-stat-n">{configuredCount}</div>
          <div class="ip-stat-l">Configured</div>
        </div>
        <div class="ip-stat-div"></div>
        <div class="ip-stat">
          <div class="ip-stat-n st-on">{activeCount}</div>
          <div class="ip-stat-l">Active</div>
        </div>
      </div>
      <button type="button" class="ip-refresh" class:spinning={loading} onclick={(e) => { e.stopPropagation(); loadProviders() }} title="Refresh">
        <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M23 4v6h-6M1 20v-6h6"/>
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
        </svg>
      </button>
    </div>
  </div>

  <!-- Toast -->
  {#if successMsg}
    <div class="ip-toast ip-ok" transition:slide={{ duration: 180 }}>
      <span class="ip-tt"><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>{successMsg}</span>
      <button type="button" class="ip-tx" onclick={() => { successMsg = null }}>✕</button>
    </div>
  {/if}

  <!-- Section label + search hint -->
  <div class="ip-sec-row">
    <div class="ip-sec-label">External Providers — {availableOAuth.length} available</div>
    <div class="ip-sec-hint">Click any provider to configure it</div>
  </div>

  <!-- Grid -->
  <div class="ip-grid">
    {#each availableOAuth as item (item.id)}
      {@const configured = isConfigured(item.id)}
      {@const active     = isActive(item.id)}
      <button
        type="button"
        class="ip-card"
        class:ip-configured={configured}
        style="--brand: {item.color}"
        onclick={(e) => openConfig(e, item.id)}
      >
        <!-- Colorful icon -->
        <div class="ip-card-l">
          <div class="ip-icon-box" style="background: {item.color}20; border-color: {item.color}40;">
            <img src={getIconUrl(item.id)} alt={item.name} class="ip-icon" style="filter: brightness(0) saturate(100%) invert(1);" />
            {#if configured && active}
              <span class="ip-dot" style="border-color: var(--card-bg, #0d0d12)"></span>
            {/if}
          </div>
          <div>
            <div class="ip-card-name">{item.name}</div>
            <div class="ip-card-st" class:st-on={configured && active} class:st-pause={configured && !active} class:st-off={!configured}>
              {configured ? (active ? '● Active' : '● Paused') : 'Setup Needed'}
            </div>
          </div>
        </div>
        <!-- Right: status pill + chevron -->
        <div class="ip-card-r">
          {#if configured}
            <span class="ip-pill" class:ip-pill-on={active} class:ip-pill-off={!active}>{active ? 'On' : 'Off'}</span>
          {/if}
          <svg class="ip-chev" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 18 15 12 9 6"/></svg>
        </div>
      </button>
    {/each}
  </div>

  <!-- Guide box -->
  <div class="ip-guide">
    <div class="ip-guide-bar"></div>
    <div class="ip-guide-inner">
      <div class="ip-guide-left">
        <div class="ip-guide-ic">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
        </div>
        <div>
          <div class="ip-guide-title">Authorization Endpoint</div>
          <div class="ip-guide-sub">Use this URL in your app button to start the social login flow.</div>
        </div>
      </div>
    </div>
    <div class="ip-code-row">
      <code class="ip-code">{getOmniBaseUrl()}/auth/v1/authorize?provider=google&project_id={$authStore.activeProject?.id ?? 'PROJECT_ID'}&redirect_to=https://yourapp.com</code>
      <button type="button" class="ip-copy" class:ip-copy-ok={copySuccess}
        onclick={(e) => { e.stopPropagation(); copyToClipboard(`${getOmniBaseUrl()}/auth/v1/authorize?provider=google&project_id=${$authStore.activeProject?.id ?? 'YOUR_PROJECT_ID'}&redirect_to=https://yourapp.com`) }}>
        {#if copySuccess}
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
        {:else}
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
        {/if}
      </button>
    </div>
    <p class="ip-guide-note">Replace <code>google</code> with any provider ID. Whitelist your <code>redirect_to</code> in the provider's developer console.</p>
  </div>
</div>

<!-- ── Modal ─────────────────────────────────────────────────────────────────── -->
{#if showModal && selectedProvider && meta && selectedInfo}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="ip-overlay" onclick={handleOverlayClick} transition:fade={{ duration: 150 }}>
    <div class="ip-modal" onclick={(e) => e.stopPropagation()} in:scale={{ start: 0.96, duration: 180 }}>
      <!-- color accent strip at top -->
      <div class="ip-modal-strip" style="background: linear-gradient(90deg, {selectedInfo.color}, {selectedInfo.color}88)"></div>

      <div class="ip-mhead">
        <div class="ip-mhead-l">
          <div class="ip-micon" style="background: {selectedInfo.color}20; border-color: {selectedInfo.color}50;">
            <img src={getIconUrl(selectedProvider.name)} alt={selectedInfo.name} class="ip-icon" style="width:26px;height:26px;filter:brightness(0) invert(1);" />
          </div>
          <div>
            <div class="ip-mtitle">{selectedInfo.name}</div>
            <div class="ip-msub">OAuth 2.0 Provider Configuration</div>
          </div>
        </div>
        <button type="button" class="ip-mclose" onclick={() => { canClose = true; closeModal() }}>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      {#if meta.note}
        <div class="ip-setup-note">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" style="flex-shrink:0;margin-top:1px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          <span>{meta.note}.
            <a href={meta.consoleUrl} target="_blank" rel="noreferrer" class="ip-link">Open {selectedInfo.name} Console ↗</a>
          </span>
        </div>
      {/if}

      <form class="ip-mbody" onsubmit={saveProvider}>
        <!-- Callback URL -->
        <div class="ip-field">
          <label class="ip-label">Callback / Redirect URI
            <span class="ip-tag">➡ Add this URL to the provider's console</span>
          </label>
          <div class="ip-code-row" style="margin-bottom:0">
            <code class="ip-code">{callbackUrl}</code>
            <button type="button" class="ip-copy" onclick={(e) => { e.stopPropagation(); copyToClipboard(callbackUrl) }}>
              <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            </button>
          </div>
        </div>

        <div class="ip-field">
          <label for="ip-cid" class="ip-label">{meta.clientIdLabel}</label>
          <input id="ip-cid" type="text" class="ip-input" bind:value={selectedProvider.client_id} placeholder="Paste from {selectedInfo.name} developer console…" autocomplete="off" spellcheck="false" />
        </div>

        <div class="ip-field">
          <label for="ip-csec" class="ip-label">{meta.clientSecretLabel}</label>
          <input id="ip-csec" type="password" class="ip-input" bind:value={selectedProvider.client_secret} placeholder="Paste your secret…" autocomplete="new-password" />
          <p class="ip-hint">Stored encrypted server-side — never exposed to client apps.</p>
        </div>

        <div class="ip-divider"></div>

        <div class="ip-toggle-row">
          <div>
            <div class="ip-toggle-title">Enable Provider</div>
            <div class="ip-toggle-sub">Allow users to sign in with {selectedInfo.name}</div>
          </div>
          <!-- svelte-ignore a11y_label_has_associated_control -->
          <label class="ip-switch">
            <input type="checkbox" bind:checked={selectedProvider.is_active} />
            <span class="ip-slider" style="--sw-on:{selectedInfo.color}"></span>
          </label>
        </div>

        {#if errorMsg && showModal}
          <div class="ip-merr" transition:slide={{ duration: 130 }}>{errorMsg}</div>
        {/if}

        <div class="ip-mfoot">
          <button type="button" class="ip-btn-ghost" onclick={() => { canClose = true; closeModal() }}>Cancel</button>
          <button type="submit" class="ip-btn-save" disabled={saving} style="background:{selectedInfo.color};box-shadow:0 4px 14px {selectedInfo.color}44">
            {#if saving}<span class="ip-spin-sm"></span> Saving…{:else}Save Configuration{/if}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<style>
  /* ── Page fills content area fully ── */
  .ip-page { padding: 32px; width: 100%; font-family: 'Inter', system-ui, sans-serif; }

  /* ── Header ── */
  .ip-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 24px; margin-bottom: 32px; flex-wrap: wrap; }
  .ip-breadcrumb { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: .1em; color: #52525b; margin-bottom: 8px; }
  .ip-breadcrumb span { color: #a78bfa; }
  .ip-title { font-size: 28px; font-weight: 800; letter-spacing: -.03em; color: #f4f4f5; margin: 0 0 6px; }
  .ip-sub   { font-size: 14px; color: #71717a; margin: 0; }

  /* Top-right: stats + refresh */
  .ip-top-r { display: flex; align-items: center; gap: 16px; }
  .ip-stat-row { display: flex; align-items: center; gap: 16px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.08); border-radius: 12px; padding: 12px 20px; }
  .ip-stat { text-align: center; }
  .ip-stat-n { font-size: 20px; font-weight: 800; color: #f4f4f5; line-height: 1; }
  .ip-stat-l { font-size: 11px; color: #52525b; font-weight: 600; text-transform: uppercase; letter-spacing: .06em; margin-top: 3px; }
  .ip-stat-div { width: 1px; height: 32px; background: rgba(255,255,255,.08); }
  .st-on { color: #4ade80 !important; }

  /* Refresh */
  .ip-refresh { width: 38px; height: 38px; border-radius: 10px; background: rgba(255,255,255,.04); border: 1px solid rgba(255,255,255,.1); color: #71717a; cursor: pointer; display: grid; place-items: center; transition: .18s; }
  .ip-refresh:hover { background: rgba(255,255,255,.09); color: #fff; border-color: rgba(255,255,255,.2); }
  .ip-refresh.spinning svg { animation: ipspin .8s linear infinite; }
  @keyframes ipspin { to { transform: rotate(360deg); } }

  /* ── Toast ── */
  .ip-toast { display: flex; align-items: center; justify-content: space-between; padding: 12px 16px; border-radius: 12px; margin-bottom: 24px; font-size: 14px; font-weight: 500; border: 1px solid transparent; }
  .ip-ok  { background: rgba(34,197,94,.08); border-color: rgba(34,197,94,.2); color: #86efac; }
  .ip-tt  { display: flex; align-items: center; gap: 9px; }
  .ip-tx  { background: none; border: none; color: inherit; font-size: 15px; cursor: pointer; opacity: .6; }
  .ip-tx:hover { opacity: 1; }

  /* ── Section row ── */
  .ip-sec-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18px; padding-bottom: 10px; border-bottom: 1px solid rgba(255,255,255,.06); }
  .ip-sec-label { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: .1em; color: #3f3f46; }
  .ip-sec-hint  { font-size: 12px; color: #3f3f46; }

  /* ── Grid — 4 columns, fills space ── */
  .ip-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 10px; margin-bottom: 36px; }

  /* ── Provider card ── */
  .ip-card {
    display: flex; align-items: center; justify-content: space-between;
    padding: 13px 14px; border-radius: 12px; text-align: left; cursor: pointer;
    background: rgba(255,255,255,.03); border: 1px solid rgba(255,255,255,.07);
    transition: all .18s cubic-bezier(.4,0,.2,1); position: relative; overflow: hidden;
  }
  /* Left brand accent line */
  .ip-card::before { content: ''; position: absolute; left: 0; top: 0; bottom: 0; width: 3px; background: var(--brand); opacity: 0; transition: opacity .2s; border-radius: 12px 0 0 12px; }
  .ip-card:hover { background: rgba(255,255,255,.055); border-color: rgba(255,255,255,.18); transform: translateY(-2px); box-shadow: 0 6px 18px rgba(0,0,0,.3); }
  .ip-card:hover::before { opacity: 1; }
  .ip-configured { border-color: rgba(108,71,255,.35); background: rgba(108,71,255,.04); }
  .ip-configured::before { opacity: .7; }

  .ip-card-l { display: flex; align-items: center; gap: 11px; flex: 1; min-width: 0; }

  /* Colorful icon box */
  .ip-icon-box { width: 38px; height: 38px; border-radius: 10px; flex-shrink: 0; border: 1px solid transparent; display: grid; place-items: center; position: relative; }
  .ip-icon { width: 18px; height: 18px; object-fit: contain; display: block; }

  .ip-dot { position: absolute; top: -3px; right: -3px; width: 8px; height: 8px; border-radius: 50%; background: #4ade80; border: 2px solid; box-shadow: 0 0 6px rgba(74,222,128,.6); }

  .ip-card-name { font-size: 13px; font-weight: 600; color: #f4f4f5; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .ip-card-st   { font-size: 11px; font-weight: 500; margin-top: 2px; }
  .st-on    { color: #4ade80; }
  .st-pause { color: #facc15; }
  .st-off   { color: #52525b; }

  .ip-card-r { display: flex; align-items: center; gap: 6px; flex-shrink: 0; margin-left: 6px; }
  .ip-pill { font-size: 10px; font-weight: 700; padding: 2px 7px; border-radius: 8px; letter-spacing: .04em; text-transform: uppercase; }
  .ip-pill-on  { background: rgba(74,222,128,.12); color: #4ade80; border: 1px solid rgba(74,222,128,.25); }
  .ip-pill-off { background: rgba(250,204,21,.08); color: #facc15; border: 1px solid rgba(250,204,21,.2); }
  .ip-chev { color: #3f3f46; transition: .18s; transform: translateX(-3px); }
  .ip-card:hover .ip-chev { color: #a78bfa; transform: translateX(0); }

  /* ── Guide ── */
  .ip-guide { background: rgba(255,255,255,.03); border: 1px solid rgba(255,255,255,.08); border-radius: 16px; padding: 22px; position: relative; overflow: hidden; }
  .ip-guide-bar { position: absolute; top: 0; left: 0; right: 0; height: 3px; background: linear-gradient(90deg,#6c47ff,#a78bfa); }
  .ip-guide-inner { margin-bottom: 16px; }
  .ip-guide-left { display: flex; align-items: center; gap: 13px; }
  .ip-guide-ic { width: 38px; height: 38px; border-radius: 10px; flex-shrink: 0; background: rgba(108,71,255,.12); border: 1px solid rgba(108,71,255,.25); display: grid; place-items: center; color: #a78bfa; }
  .ip-guide-title { font-size: 15px; font-weight: 700; color: #f4f4f5; }
  .ip-guide-sub   { font-size: 13px; color: #71717a; margin-top: 2px; }
  .ip-guide-note  { font-size: 12px; color: #52525b; line-height: 1.7; }
  .ip-guide-note code { color: #f87171; background: rgba(248,113,113,.1); padding: 1px 5px; border-radius: 4px; }

  /* ── Code row ── */
  .ip-code-row { background: #09090c; border: 1px solid rgba(255,255,255,.08); border-radius: 10px; padding: 8px 8px 8px 13px; display: flex; align-items: center; gap: 10px; margin-bottom: 14px; }
  .ip-code { flex: 1; font-family: 'JetBrains Mono', monospace; font-size: 11.5px; color: #818cf8; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; background: none; padding: 0; }
  .ip-copy { flex-shrink: 0; width: 30px; height: 30px; border-radius: 7px; background: rgba(255,255,255,.05); border: 1px solid rgba(255,255,255,.09); color: #71717a; cursor: pointer; display: grid; place-items: center; transition: .18s; }
  .ip-copy:hover { background: rgba(255,255,255,.1); color: #fff; }
  .ip-copy-ok { background: rgba(34,197,94,.12) !important; color: #4ade80 !important; border-color: rgba(34,197,94,.3) !important; }

  /* ── Overlay ── */
  .ip-overlay { position: fixed; inset: 0; z-index: 9999; background: rgba(0,0,0,.8); backdrop-filter: blur(10px); display: flex; align-items: center; justify-content: center; padding: 20px; }

  /* ── Modal ── */
  .ip-modal { width: 100%; max-width: 500px; background: #0d0d12; border: 1px solid rgba(255,255,255,.14); border-radius: 20px; overflow: hidden; box-shadow: 0 24px 60px rgba(0,0,0,.7); max-height: 92vh; overflow-y: auto; }
  .ip-modal-strip { height: 4px; }

  /* Modal head */
  .ip-mhead { padding: 20px 22px 16px; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid rgba(255,255,255,.07); }
  .ip-mhead-l { display: flex; align-items: center; gap: 13px; }
  .ip-micon { width: 46px; height: 46px; border-radius: 12px; flex-shrink: 0; border: 1px solid transparent; display: grid; place-items: center; }
  .ip-mtitle { font-size: 18px; font-weight: 700; color: #f4f4f5; }
  .ip-msub   { font-size: 12px; color: #71717a; margin-top: 2px; }
  .ip-mclose { background: none; border: none; color: #52525b; cursor: pointer; width: 30px; height: 30px; display: grid; place-items: center; border-radius: 8px; transition: .15s; }
  .ip-mclose:hover { color: #fff; background: rgba(255,255,255,.08); }

  /* Setup note */
  .ip-setup-note { display: flex; align-items: flex-start; gap: 9px; background: rgba(108,71,255,.07); border: 1px solid rgba(108,71,255,.2); border-radius: 10px; padding: 11px 14px; margin: 14px 22px 0; font-size: 13px; color: #c4b5fd; line-height: 1.5; }
  .ip-link { color: #a78bfa; text-decoration: underline; font-weight: 600; }
  .ip-link:hover { color: #c4b5fd; }

  /* Modal body */
  .ip-mbody { padding: 18px 22px 0; }
  .ip-field { margin-bottom: 16px; }
  .ip-label { display: flex; align-items: center; gap: 8px; font-size: 11px; font-weight: 700; color: #71717a; text-transform: uppercase; letter-spacing: .07em; margin-bottom: 7px; }
  .ip-tag { text-transform: none; font-weight: 500; font-size: 10.5px; background: rgba(251,191,36,.1); color: #fbbf24; padding: 2px 7px; border-radius: 5px; letter-spacing: 0; }
  .ip-input { width: 100%; padding: 10px 13px; border-radius: 10px; background: #060608; border: 1px solid rgba(255,255,255,.12); color: #f4f4f5; font-size: 14px; font-family: inherit; outline: none; transition: border-color .18s, box-shadow .18s; }
  .ip-input:focus { border-color: #6c47ff; box-shadow: 0 0 0 3px rgba(108,71,255,.18); }
  .ip-input::placeholder { color: #3f3f46; }
  .ip-hint { font-size: 11.5px; color: #3f3f46; margin-top: 5px; }

  .ip-divider { height: 1px; background: rgba(255,255,255,.07); margin: 14px 0; }

  /* Toggle */
  .ip-toggle-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 12px; border-radius: 12px; background: rgba(255,255,255,.03); border: 1px solid rgba(255,255,255,.06); margin-bottom: 14px; }
  .ip-toggle-title { font-size: 13px; font-weight: 600; color: #e4e4e7; }
  .ip-toggle-sub   { font-size: 11.5px; color: #52525b; margin-top: 2px; }
  .ip-switch { position: relative; width: 44px; height: 24px; flex-shrink: 0; }
  .ip-switch input { opacity: 0; width: 0; height: 0; }
  .ip-slider { position: absolute; inset: 0; background: #3f3f46; border-radius: 24px; cursor: pointer; transition: .27s; }
  .ip-slider::before { content: ''; position: absolute; height: 18px; width: 18px; left: 3px; bottom: 3px; background: #fff; border-radius: 50%; transition: .27s; }
  .ip-switch input:checked + .ip-slider { background: var(--sw-on, #6c47ff); }
  .ip-switch input:checked + .ip-slider::before { transform: translateX(20px); }

  /* Modal error */
  .ip-merr { background: rgba(239,68,68,.08); border: 1px solid rgba(239,68,68,.2); color: #fca5a5; font-size: 13px; padding: 10px 14px; border-radius: 10px; margin-bottom: 12px; }

  /* Modal footer */
  .ip-mfoot { padding: 12px 0 18px; display: flex; justify-content: flex-end; gap: 10px; }
  .ip-btn-ghost { background: none; border: none; color: #71717a; font-weight: 600; font-size: 14px; cursor: pointer; padding: 9px 18px; border-radius: 9px; font-family: inherit; transition: .15s; }
  .ip-btn-ghost:hover { color: #f4f4f5; background: rgba(255,255,255,.06); }
  .ip-btn-save { display: flex; align-items: center; gap: 8px; color: #fff; border: none; font-weight: 700; font-size: 14px; padding: 9px 22px; border-radius: 9px; cursor: pointer; font-family: inherit; transition: .18s; }
  .ip-btn-save:hover:not(:disabled) { filter: brightness(1.15); transform: translateY(-1px); }
  .ip-btn-save:disabled { opacity: .6; cursor: not-allowed; transform: none; }
  .ip-spin-sm { display: inline-block; width: 13px; height: 13px; border: 2px solid rgba(255,255,255,.3); border-top-color: #fff; border-radius: 50%; animation: ipspin .7s linear infinite; }
</style>
