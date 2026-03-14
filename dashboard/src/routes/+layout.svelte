<script lang="ts">
  import '../app.css'
  import Sidebar from '$lib/components/Sidebar.svelte'
  import Topbar from '$lib/components/Topbar.svelte'
  import { browser } from '$app/environment'
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import { onMount } from 'svelte'
  import { authStore } from '$lib/stores/auth'

  const publicRoutes = ['/auth/login', '/auth/signup', '/auth/setup', '/auth/forgot-password', '/auth/reset-password', '/projects/new']
  let isPublic = $derived(publicRoutes.some((route) => $page.url.pathname.startsWith(route)))
  let authReady = $state(false)

  onMount(() => {
    const initAuth = async () => {
      if (typeof window !== 'undefined' && window.location.hash) {
        const hash = window.location.hash.slice(1)
        const params = new URLSearchParams(hash)
        const accessToken = params.get('access_token')
        const refreshToken = params.get('refresh_token')

        if (accessToken) {
          await authStore.setSessionFromHash(accessToken, refreshToken || '')
          window.history.replaceState(null, '', window.location.pathname + window.location.search)
          authReady = true
          return
        }
      }

      await authStore.init()
      authReady = true
    }

    initAuth()
  })

  $effect(() => {
    if (!browser || !authReady || $authStore.loading) {
      return
    }

    const pathname = $page.url.pathname
    const redirectTarget = encodeURIComponent(pathname + $page.url.search)

    if (!$authStore.user && !isPublic) {
      goto(`/auth/setup?redirect=${redirectTarget}`, { replaceState: true })
      return
    }

    if ($authStore.user && (pathname === '/auth/login' || pathname === '/auth/signup' || pathname === '/auth/setup')) {
      const redirect = $page.url.searchParams.get('redirect')
      // After login, check if user has any projects — if not, redirect to project creation
      if ($authStore.activeProject) {
        goto(redirect || '/', { replaceState: true })
      } else {
        // No project yet — send to project creation (but only after brief delay to allow project fetch to complete)
        setTimeout(async () => {
          if (!$authStore.activeProject) {
            goto(redirect ? `/projects/new?redirect=${encodeURIComponent(redirect)}` : '/projects/new', { replaceState: true })
          } else {
            goto(redirect || '/', { replaceState: true })
          }
        }, 800)
      }
      return
    }

    // Redirect to authenticated dashboard pages
    if ($authStore.user && pathname === '/projects/new') {
      // Already has a project? Only redirect if intentionally navigated here
    }
  })

  let { children } = $props()
</script>

{#if !authReady && !isPublic}
  <div class="auth-splash">
    <div class="auth-splash-card">
      <div class="auth-splash-logo">O</div>
      <h1>Preparing OmniBase</h1>
      <p>Checking your local project and session.</p>
    </div>
  </div>
{:else if isPublic}
  {@render children()}
{:else}
  <div class="app-layout">
    <Topbar />
    <Sidebar />
    <main class="main-content">
      {@render children()}
    </main>
  </div>
{/if}

<style>
  .auth-splash {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background:
      radial-gradient(circle at top, rgba(108, 71, 255, 0.22), transparent 32%),
      linear-gradient(180deg, #09090f 0%, #111122 100%);
  }

  .auth-splash-card {
    width: 100%;
    max-width: 420px;
    padding: 32px;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-xl);
    background: rgba(17, 17, 24, 0.9);
    box-shadow: var(--shadow-lg);
    text-align: center;
  }

  .auth-splash-logo {
    width: 56px;
    height: 56px;
    margin: 0 auto 18px;
    border-radius: 16px;
    background: var(--gradient-brand);
    display: grid;
    place-items: center;
    font-size: 22px;
    font-weight: 800;
    color: #fff;
  }

  h1 {
    margin: 0 0 8px;
    font-size: 24px;
  }

  p {
    margin: 0;
    color: var(--text-secondary);
  }
</style>
