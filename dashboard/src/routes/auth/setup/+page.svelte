<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import { getOmniBaseUrl } from '$lib/api'

  let apiUrl = $state(getOmniBaseUrl())
  let checking = $state(true)
  let status = $state<'healthy' | 'down'>('down')
  let message = $state('Checking API gateway...')
  let initialized = $state(false)

  async function checkApi() {
    checking = true
    message = 'Checking API gateway...'
    status = 'down'

    try {
      const response = await fetch(`${apiUrl.replace(/\/$/, '')}/health`, {
        signal: AbortSignal.timeout(5000),
      })

      if (!response.ok) {
        throw new Error('Gateway responded but is not healthy yet.')
      }

      status = 'healthy'
      message = 'Local project detected. You can create an account and start using the APIs.'
    } catch {
      status = 'down'
      message = 'OmniBase API is not reachable. Start the stack or update the project URL first.'
    } finally {
      checking = false
    }
  }

  function saveApiUrl() {
    const normalized = apiUrl.trim().replace(/\/$/, '')
    if (!normalized) {
      return
    }

    localStorage.setItem('omnibase.api_url', normalized)
    apiUrl = normalized
    checkApi()
  }

  function openAuth(path: '/auth/login' | '/auth/signup') {
    const redirect = $page.url.searchParams.get('redirect')
    const suffix = redirect ? `?redirect=${encodeURIComponent(redirect)}` : ''
    goto(`${path}${suffix}`)
  }

  $effect(() => {
    if (initialized) {
      return
    }

    initialized = true
    checkApi()
  })
</script>

<svelte:head>
  <title>Set Up OmniBase</title>
</svelte:head>

<div class="setup-page">
  <div class="setup-shell">
    <section class="hero">
      <div class="hero-copy">
        <div class="eyebrow">Self-hosted backend</div>
        <h1>Your OmniBase project is ready. Connect, create an account, and start shipping.</h1>
        <p>
          This local stack already provisions a usable project. Sign up once, then use the dashboard, REST API,
          GraphQL, storage, realtime, and functions from your own app.
        </p>
      </div>

      <div class="status-card">
        <div class="status-head">
          <span class:ok={status === 'healthy'} class="status-pill">
            {checking ? 'Checking' : status === 'healthy' ? 'Online' : 'Offline'}
          </span>
          <span class="status-url">{apiUrl}</span>
        </div>
        <p>{message}</p>
        <div class="actions">
          <button class="btn btn-primary" type="button" onclick={() => openAuth('/auth/signup')} disabled={status !== 'healthy'}>
            Create account
          </button>
          <button class="btn btn-secondary" type="button" onclick={() => openAuth('/auth/login')}>
            Sign in
          </button>
        </div>
      </div>
    </section>

    <section class="grid">
      <div class="card">
        <h2>Project connection</h2>
        <p class="muted">Use the API gateway URL your apps will call.</p>
        <div class="input-row">
          <input class="input input-mono" bind:value={apiUrl} placeholder="http://localhost:8000" />
          <button class="btn btn-secondary" type="button" onclick={saveApiUrl}>Save</button>
        </div>
        <p class="hint">Default local stack: <code>http://localhost:8000</code></p>
      </div>

      <div class="card">
        <h2>What happens next</h2>
        <ul>
          <li>Create your first user account.</li>
          <li>Open the dashboard and manage tables, auth users, files, and policies.</li>
          <li>Copy the project URL and anon key from Settings for your app.</li>
        </ul>
      </div>

      <div class="card">
        <h2>Project endpoints</h2>
        <ul>
          <li><code>{apiUrl}/rest/v1</code> for REST</li>
          <li><code>{apiUrl}/graphql/v1</code> for GraphQL</li>
          <li><code>{apiUrl}/storage/v1</code> for file storage</li>
          <li><code>{apiUrl}/realtime/v1/websocket</code> for realtime</li>
        </ul>
      </div>
    </section>
  </div>
</div>

<style>
  .setup-page {
    min-height: 100vh;
    padding: 32px;
    background:
      radial-gradient(circle at top left, rgba(0, 212, 255, 0.16), transparent 28%),
      radial-gradient(circle at top right, rgba(108, 71, 255, 0.28), transparent 30%),
      linear-gradient(180deg, #09090f 0%, #111122 100%);
  }

  .setup-shell {
    max-width: 1120px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .hero {
    display: grid;
    grid-template-columns: minmax(0, 1.3fr) minmax(320px, 0.9fr);
    gap: 24px;
    align-items: stretch;
  }

  .hero-copy,
  .status-card,
  .card {
    border: 1px solid var(--border-default);
    background: rgba(17, 17, 24, 0.86);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-lg);
  }

  .hero-copy {
    padding: 36px;
  }

  .eyebrow {
    display: inline-flex;
    padding: 6px 10px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.06);
    color: var(--text-secondary);
    font-size: 12px;
    margin-bottom: 18px;
  }

  h1 {
    font-size: clamp(32px, 5vw, 56px);
    line-height: 1.05;
    margin: 0 0 16px;
  }

  h2 {
    font-size: 18px;
    margin: 0 0 10px;
  }

  p {
    margin: 0;
    color: var(--text-secondary);
    line-height: 1.6;
  }

  .status-card {
    padding: 28px;
    display: flex;
    flex-direction: column;
    gap: 18px;
    justify-content: space-between;
  }

  .status-head {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .status-pill {
    display: inline-flex;
    width: fit-content;
    padding: 6px 10px;
    border-radius: 999px;
    background: rgba(255, 82, 82, 0.14);
    color: var(--status-error);
    font-size: 12px;
    font-weight: 700;
  }

  .status-pill.ok {
    background: rgba(0, 230, 118, 0.14);
    color: var(--status-success);
  }

  .status-url {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-code);
  }

  .actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 20px;
  }

  .card {
    padding: 24px;
  }

  .muted,
  .hint {
    font-size: 13px;
  }

  .input-row {
    display: flex;
    gap: 10px;
    margin: 16px 0 10px;
  }

  ul {
    margin: 0;
    padding-left: 18px;
    color: var(--text-secondary);
    line-height: 1.8;
  }

  @media (max-width: 900px) {
    .setup-page {
      padding: 20px;
    }

    .hero,
    .grid {
      grid-template-columns: 1fr;
    }

    .hero-copy {
      padding: 28px;
    }

    h1 {
      font-size: 34px;
    }

    .input-row {
      flex-direction: column;
    }
  }
</style>
