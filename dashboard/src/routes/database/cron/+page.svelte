<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getHeaders, getErrorMessage } from '$lib/api'

  interface CronJob {
    jobid: number
    schedule: string
    command: string
    database: string
    username: string
    active: boolean
    jobname: string
  }

  let jobs = $state<CronJob[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  let showCreateModal = $state(false)
  let isCreating = $state(false)
  let newJobName = $state('')
  let newSchedule = $state('')
  let newCommand = $state('')

  async function loadJobs() {
    try {
      loading = true
      error = null
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/cron`, { headers: getHeaders() })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to load cron jobs'))
      jobs = await resp.json()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      loading = false
    }
  }

  async function createJob() {
    if (!newJobName || !newSchedule || !newCommand) return alert('All fields are required')
    isCreating = true
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/cron`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ jobname: newJobName, schedule: newSchedule, command: newCommand })
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to create job'))
      showCreateModal = false
      newJobName = ''
      newSchedule = ''
      newCommand = ''
      await loadJobs()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    } finally {
      isCreating = false
    }
  }

  async function deleteJob(job: CronJob) {
    if (!confirm(`Are you sure you want to delete job "${job.jobname}"?`)) return
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/cron?jobid=${job.jobid}`, {
        method: 'DELETE',
        headers: getHeaders(false)
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to delete job'))
      await loadJobs()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    }
  }

  async function toggleJob(job: CronJob) {
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/cron/toggle`, {
        method: 'PATCH',
        headers: getHeaders(),
        body: JSON.stringify({ jobid: job.jobid, active: !job.active })
      })
      if (!resp.ok) throw new Error(await getErrorMessage(resp, 'Failed to toggle job'))
      await loadJobs()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Unknown error')
    }
  }

  onMount(() => loadJobs())
</script>

<svelte:head><title>Cron Jobs — OmniBase</title></svelte:head>

<div class="page-header">
  <div>
    <h1 class="page-title">Cron Jobs</h1>
    <p class="page-subtitle">Schedule recurring jobs inside PostgreSQL using pg_cron</p>
  </div>
  <button class="btn btn-primary" onclick={() => showCreateModal = true}>＋ New Job</button>
</div>

<div class="page-content">
  {#if loading}
    <div class="table-wrapper">
      <div class="skeleton" style="height: 200px; width: 100%; border-radius: var(--radius-md);"></div>
    </div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else if jobs.length === 0}
    <div class="empty-state">
      <div class="empty-icon">🕒</div>
      <h3>No Cron Jobs Found</h3>
      <p>Schedule repetitive database tasks like cleanups or aggregations directly within Postgres.</p>
      {#if error?.includes("extension is not enabled")}
        <p class="form-hint mt-2">Make sure pg_cron is enabled in the Extensions page first.</p>
      {:else}
        <button class="btn btn-primary btn-sm mt-4" onclick={() => showCreateModal = true}>Create Job</button>
      {/if}
    </div>
  {:else}
    <div class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th>Job Name</th>
            <th>Schedule</th>
            <th>Command</th>
            <th>Status</th>
            <th style="text-align: right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each jobs as j}
            <tr>
              <td class="cell-mono font-bold">{j.jobname || `job_${j.jobid}`}</td>
              <td class="cell-mono">{j.schedule}</td>
              <td class="cell-mono" style="max-width:300px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title={j.command}>{j.command}</td>
              <td>
                <span class="badge {j.active ? 'badge-success' : 'badge-neutral'}">
                  {j.active ? 'Active' : 'Paused'}
                </span>
              </td>
              <td style="display:flex;gap:6px;justify-content:flex-end">
                <button class="btn btn-secondary btn-sm" onclick={() => toggleJob(j)}>{j.active ? 'Pause' : 'Resume'}</button>
                <button class="btn btn-danger btn-sm" onclick={() => deleteJob(j)}>Delete</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

{#if showCreateModal}
  <div class="modal-backdrop" onclick={() => !isCreating && (showCreateModal = false)}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2 class="modal-title">Create Cron Job</h2>
        <button class="modal-close" onclick={() => showCreateModal = false} disabled={isCreating}>×</button>
      </div>
      <div class="modal-body" style="display:flex;flex-direction:column;gap:20px;padding-top:16px;">
        <div class="form-group">
          <label class="form-label" style="font-weight:600">Job Name</label>
          <input class="input" type="text" placeholder="e.g. daily_cleanup" bind:value={newJobName} disabled={isCreating} style="background:var(--bg-document); border-color:var(--border-default)" />
          <p class="form-hint" style="margin-top:6px">A recognizable name for this scheduled task.</p>
        </div>
        <div class="form-group">
          <label class="form-label" style="font-weight:600">Schedule (CRON Expression)</label>
          <input class="input cell-mono" type="text" placeholder="* * * * *" bind:value={newSchedule} disabled={isCreating} style="background:var(--bg-document); border-color:var(--border-default); font-size:14px; letter-spacing:1px" />
          <div style="display:flex;gap:8px;margin-top:10px;flex-wrap:wrap">
             <button class="btn btn-sm badge badge-neutral" style="cursor:pointer;font-family:var(--font-mono);font-size:11px" onclick={() => newSchedule = '* * * * *'} disabled={isCreating}>Every Min (* * * * *)</button>
             <button class="btn btn-sm badge badge-neutral" style="cursor:pointer;font-family:var(--font-mono);font-size:11px" onclick={() => newSchedule = '*/5 * * * *'} disabled={isCreating}>Every 5 Mins (*/5 * * * *)</button>
             <button class="btn btn-sm badge badge-neutral" style="cursor:pointer;font-family:var(--font-mono);font-size:11px" onclick={() => newSchedule = '0 * * * *'} disabled={isCreating}>Hourly (0 * * * *)</button>
             <button class="btn btn-sm badge badge-neutral" style="cursor:pointer;font-family:var(--font-mono);font-size:11px" onclick={() => newSchedule = '0 0 * * *'} disabled={isCreating}>Midnight (0 0 * * *)</button>
             <button class="btn btn-sm badge badge-neutral" style="cursor:pointer;font-family:var(--font-mono);font-size:11px" onclick={() => newSchedule = '0 0 * * 0'} disabled={isCreating}>Sunday (0 0 * * 0)</button>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label" style="font-weight:600">SQL Command</label>
          <div style="background:var(--bg-document);border:1px solid var(--border-default);border-radius:var(--radius-md);overflow:hidden">
             <div style="background:rgba(255,255,255,0.02);padding:8px 12px;border-bottom:1px solid var(--border-default);font-size:12px;color:var(--text-secondary);font-weight:600;display:flex;align-items:center;gap:8px">
                 <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
                 PostgreSQL Code
             </div>
             <textarea class="input cell-mono" rows="4" placeholder="e.g. DELETE FROM sessions WHERE expires_at < NOW();" bind:value={newCommand} disabled={isCreating} style="border:none;background:transparent;resize:vertical;padding:12px;font-size:13px;line-height:1.6"></textarea>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showCreateModal = false} disabled={isCreating}>Cancel</button>
        <button class="btn btn-primary" onclick={createJob} disabled={isCreating || !newJobName || !newSchedule || !newCommand}>
          {isCreating ? 'Creating...' : 'Create Job'}
        </button>
      </div>
    </div>
  </div>
{/if}
