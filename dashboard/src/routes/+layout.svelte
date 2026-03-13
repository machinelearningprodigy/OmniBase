<script lang="ts">
  import '../app.css'
  import Sidebar from '$lib/components/Sidebar.svelte'
  import Topbar from '$lib/components/Topbar.svelte'
  import { page } from '$app/stores'
  import { onMount } from 'svelte'
  import { authStore } from '$lib/stores/auth'

  // Public routes that don't need layout
  const publicRoutes = ['/auth/login', '/auth/setup']
  let isPublic = $derived(publicRoutes.some(r => $page.url.pathname.startsWith(r)))

  onMount(() => {
    authStore.init()
  })

  let { children } = $props()
</script>

{#if isPublic}
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
