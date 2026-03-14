<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import { getOmniBaseUrl, getHeaders } from '$lib/api'
  import { authStore } from '$lib/stores/auth'
  import { onMount } from 'svelte'

  let projectName = $state('')
  let description = $state('')
  let region = $state('local')
  let error = $state<string | null>(null)
  let loading = $state(false)
  let createdProject = $state<{ id: string; name: string; anon_key: string; service_key: string; url: string } | null>(null)
  let copied = $state<string | null>(null)

  onMount(async () => {
    // If they already have a project from the URL params being skipped or came here manually
    // and have projects, skip to the dashboard
    const redirect = $page.url.searchParams.get('redirect')
    if (!redirect) {
      // Check if they have projects already from the store
      if ($authStore.activeProject) {
        await goto('/')
        return
      }
    }
  })

  function copy(text: string, key: string) {
    navigator.clipboard.writeText(text)
    copied = key
    setTimeout(() => { copied = null }, 2000)
  }

  async function handleCreate(e: Event) {
    e.preventDefault()
    error = null
    if (!projectName.trim()) {
      error = 'Project name is required'
      return
    }
    loading = true
    try {
      const resp = await fetch(`${getOmniBaseUrl()}/admin/v1/projects`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({
          name: projectName.trim(),
          description: description.trim(),
          region,
        })
      })
      const data = await resp.json()
      if (resp.ok) {
        createdProject = data
        authStore.setActiveProject({
          id: data.id,
          name: data.name,
          anon_key: data.anon_key,
          service_key: data.service_key,
        })
      } else {
        error = data.error || data.message || 'Failed to create project'
      }
    } catch (err) {
      error = 'Connection error — is the API running?'
    } finally {
      loading = false
    }
  }

  async function goToDashboard() {
    const redirect = $page.url.searchParams.get('redirect')
    await goto(redirect || '/')
  }
</script>

<svelte:head>
  <title>Create Project — OmniBase</title>
</svelte:head>

<div class="project-page">
  <div class="project-card">
    {#if !createdProject}
      <!-- Create Project Form -->
      <a href="/" class="logo-link">
        <span class="logo-icon">Ω</span>
        <span>OmniBase</span>
      </a>
      <h1>Create your first project</h1>
      <p class="subtitle">A project organizes your database, auth, storage and serverless functions. You'll get unique API keys to connect your app.</p>

      <form onsubmit={handleCreate} class="project-form">
        {#if error}
          <div class="error-banner" role="alert">{error}</div>
        {/if}
        <div class="form-group">
          <label for="project-name">Project name</label>
          <input
            id="project-name"
            type="text"
            bind:value={projectName}
            placeholder="my-awesome-app"
            autocomplete="off"
            required
            disabled={loading}
          />
        </div>
        <div class="form-group">
          <label for="project-desc">Description <span style="color: var(--text-muted);">(optional)</span></label>
          <input
            id="project-desc"
            type="text"
            bind:value={description}
            placeholder="What are you building?"
            disabled={loading}
          />
        </div>
        <div class="form-group">
          <label for="region">Region</label>
          <select id="region" bind:value={region} disabled={loading} class="input">
            <option value="local">Local (Self-hosted)</option>
            <option value="us-east-1">US East (Virginia)</option>
            <option value="eu-west-1">EU West (Ireland)</option>
            <option value="ap-south-1">Asia Pacific (Mumbai)</option>
          </select>
        </div>
        <button type="submit" class="btn btn-primary btn-block" disabled={loading}>
          {loading ? 'Creating project...' : '✨ Create project'}
        </button>
      </form>

      <p style="margin-top: 20px; font-size: 13px; color: var(--text-muted); text-align: center;">
        Already have a project? <a href="/" style="color: var(--brand-primary);">Go to dashboard →</a>
      </p>
    {:else}
      <!-- Project Created Successfully -->
      <div class="success-icon">✓</div>
      <h1>Project created!</h1>
      <p class="subtitle">Your project <strong>{createdProject.name}</strong> is ready. Copy your API keys below — you'll need them to connect your app.</p>

      <div class="keys-section">
        <div class="key-card">
          <div class="key-header">
            <span class="badge badge-success">anon</span>
            <span class="badge badge-neutral">public</span>
          </div>
          <p class="key-desc">Safe to use in browser with Row Level Security enabled.</p>
          <div class="key-group">
            <input type="password" class="input key-input" value={createdProject.anon_key} readonly />
            <button class="btn btn-secondary btn-sm" onclick={() => copy(createdProject!.anon_key, 'anon')}>
              {copied === 'anon' ? '✓ Copied' : 'Copy'}
            </button>
          </div>
        </div>

        <div class="key-card" style="border-color: rgba(255,82,82,0.2)">
          <div class="key-header">
            <span class="badge badge-error" style="background: rgba(255,82,82,0.1); color: #ff5252;">service_role</span>
            <span class="badge badge-neutral">secret</span>
          </div>
          <p class="key-desc">⚠️ Never expose this in the browser. Server/CLI use only.</p>
          <div class="key-group">
            <input type="password" class="input key-input" value={createdProject.service_key} readonly />
            <button class="btn btn-secondary btn-sm" onclick={() => copy(createdProject!.service_key, 'service')}>
              {copied === 'service' ? '✓ Copied' : 'Copy'}
            </button>
          </div>
        </div>

        <div class="key-card">
          <div class="key-header">
            <span class="badge badge-info">Project URL</span>
          </div>
          <p class="key-desc">Your OmniBase API gateway endpoint.</p>
          <div class="key-group">
            <input type="text" class="input key-input" value={createdProject.url} readonly style="font-family: var(--font-mono);" />
            <button class="btn btn-secondary btn-sm" onclick={() => copy(createdProject!.url, 'url')}>
              {copied === 'url' ? '✓ Copied' : 'Copy'}
            </button>
          </div>
        </div>

        <div class="sdk-snippet">
          <div style="font-size: 12px; color: var(--text-muted); margin-bottom: 8px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Quick Start</div>
          <pre class="code-block"><code><span style="color: #7dd3fc;">import</span> {'{ createClient }'} <span style="color: #7dd3fc;">from</span> <span style="color: #86efac;">'@omnibase/omnibase-js'</span>

<span style="color: #7dd3fc;">const</span> omni = createClient(
  <span style="color: #86efac;">'{createdProject.url}'</span>,
  <span style="color: #86efac;">'{createdProject.anon_key.slice(0, 24)}...'</span>
)</code></pre>
        </div>
      </div>

      <button class="btn btn-primary btn-block" onclick={goToDashboard} style="margin-top: 24px;">
        Go to Dashboard →
      </button>
    {/if}
  </div>
</div>

<style>
  .project-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background:
      radial-gradient(circle at 30% 20%, rgba(108, 71, 255, 0.15), transparent 40%),
      radial-gradient(circle at 70% 80%, rgba(0, 212, 255, 0.08), transparent 40%),
      var(--bg-base);
    padding: 24px;
  }
  .project-card {
    width: 100%;
    max-width: 520px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-xl);
    padding: 40px;
    box-shadow: var(--shadow-xl);
  }
  .logo-link {
    display: flex;
    align-items: center;
    gap: 10px;
    text-decoration: none;
    color: var(--text-primary);
    font-weight: 700;
    font-size: 20px;
    margin-bottom: 28px;
  }
  .logo-icon {
    width: 40px;
    height: 40px;
    background: var(--gradient-brand);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    color: #fff;
  }
  .success-icon {
    width: 56px;
    height: 56px;
    background: linear-gradient(135deg, #4ade80, #22d3ee);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    color: #000;
    font-weight: 900;
    margin-bottom: 16px;
  }
  h1 { font-size: 24px; margin: 0 0 8px; }
  .subtitle { font-size: 14px; color: var(--text-muted); margin: 0 0 24px; line-height: 1.5; }
  .project-form { display: flex; flex-direction: column; gap: 16px; }
  .error-banner {
    padding: 12px 14px;
    background: rgba(255, 82, 82, 0.1);
    border: 1px solid rgba(255, 82, 82, 0.3);
    border-radius: var(--radius-md);
    color: var(--status-error);
    font-size: 13px;
  }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .form-group label { font-size: 13px; font-weight: 500; color: var(--text-secondary); }
  select { appearance: none; cursor: pointer; }
  .btn-block { width: 100%; margin-top: 8px; padding: 12px; font-size: 15px; }

  /* Keys section */
  .keys-section { display: flex; flex-direction: column; gap: 12px; }
  .key-card {
    padding: 16px;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    background: var(--bg-elevated);
  }
  .key-header { display: flex; gap: 6px; margin-bottom: 6px; }
  .key-desc { font-size: 12px; color: var(--text-muted); margin: 0 0 10px; }
  .key-group { display: flex; gap: 8px; }
  .key-input {
    flex: 1;
    font-family: var(--font-mono);
    font-size: 12px;
  }
  .sdk-snippet {
    padding: 16px;
    border: 1px solid var(--border-brand);
    border-radius: var(--radius-md);
    background: linear-gradient(135deg, rgba(108, 71, 255, 0.06), rgba(0, 212, 255, 0.03));
  }
  .code-block {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.6;
    color: var(--text-primary);
    overflow-x: auto;
    white-space: pre;
  }
</style>
