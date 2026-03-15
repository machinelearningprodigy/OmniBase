<script lang="ts">
  import { onMount } from 'svelte'
  import { apiFetch, getOmniBaseUrl, getErrorMessage } from '$lib/api'
  import { fade, slide, scale } from 'svelte/transition'

  let activeTab = $state('templates') // 'templates' | 'smtp' | 'security'
  
  // Data State
  let smtpSettings = $state({
    enable_custom_smtp: false,
    sender_email: '',
    sender_name: '',
    host: '',
    port: 465,
    username: '',
    password: '',
    secure: true,
    min_interval: 1
  })

  let templates = $state([]) // Raw from backend
  let saving = $state(false)
  let loading = $state(true)

  // UI State
  let editingTemplate = $state(null) 
  let showModal = $state(false)
  let modalTab = $state('html') // 'html' | 'text' | 'preview'

  // SVG Icons for professional look
  const icons = {
    signup: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><polyline points="16 11 18 13 22 9"/></svg>`,
    invite: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><line x1="19" y1="8" x2="19" y2="14"/><line x1="16" y1="11" x2="22" y2="11"/></svg>`,
    magic_link: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 4V2"/><path d="M15 16v-2"/><path d="M8 9h2"/><path d="M20 9h2"/><path d="M17.8 5.2 18.8 4.2"/><path d="M11.2 12.8 12.2 11.8"/><path d="M17.8 12.8 18.8 13.8"/><path d="M11.2 5.2 12.2 6.2"/><path d="M10 2H2v8h8V2z"/><path d="M22 14h-8v8h8v-8z"/></svg>`,
    change_email: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/><path d="M16 19h6"/><path d="M19 16l3 3-3 3"/></svg>`,
    reset_password: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>`,
    reauth: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><circle cx="12" cy="11" r="3"/><line x1="12" y1="14" x2="12" y2="14.01"/></svg>`,
    alert: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12" y2="17.01"/></svg>`,
    mail: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>`
  }

  // Built-in map for templates
  const authTemplates = $state([
    { id: 'signup', name: 'Confirm Sign Up', desc: 'Ask users to confirm their email address after signing up', icon: icons.signup },
    { id: 'invite', name: 'Invite User', desc: 'Invite users who don\'t yet have an account to sign up', icon: icons.invite },
    { id: 'magic_link', name: 'Magic Link', desc: 'Allow users to sign in via a one-time link sent to their email', icon: icons.magic_link },
    { id: 'change_email', name: 'Change Email Address', desc: 'Ask users to verify their new email address after changing it', icon: icons.change_email },
    { id: 'reset_password', name: 'Reset Password', desc: 'Allow users to reset their password if they forget it', icon: icons.reset_password },
    { id: 'reauth', name: 'Reauthentication', desc: 'Ask users to re-authenticate before performing a sensitive action', icon: icons.reauth }
  ])

  const securityTemplates = $state([
    { id: 'sec_password_changed', name: 'Password Changed', desc: 'Notify users when their password has changed', active: false },
    { id: 'sec_email_changed', name: 'Email Address Changed', desc: 'Notify users when their email address has changed', active: false },
    { id: 'sec_phone_changed', name: 'Phone Number Changed', desc: 'Notify users when their phone number has changed', active: false },
    { id: 'sec_identity_linked', name: 'Identity Linked', desc: 'Notify users when a new identity has been linked to their account', active: false },
    { id: 'sec_mfa_added', name: 'MFA Method Added', desc: 'Notify users when a new multi-factor authentication method has been added', active: false },
    { id: 'sec_mfa_removed', name: 'MFA Method Removed', desc: 'Notify users when a multi-factor authentication method has been removed', active: false }
  ])

  onMount(async () => {
    await fetchData()
    await fetchSecuritySettings()
  })

  async function fetchSecuritySettings() {
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/config/settings/security_alerts`)
      if (resp.ok) {
        const data = await resp.json()
        securityTemplates.forEach(t => {
          if (data[t.id] !== undefined) t.active = data[t.id]
        })
      }
    } catch (err) {
      console.error('Error fetching security settings', err)
    }
  }

  async function saveSecuritySettings() {
    saving = true
    const settings = {}
    securityTemplates.forEach(t => settings[t.id] = t.active)
    
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/config/settings/security_alerts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(settings)
      })
      if (!resp.ok) throw new Error('Failed to save')
      alert('Security settings saved!')
    } catch (err) {
      alert('Error saving security settings: ' + err.message)
    } finally {
      saving = false
    }
  }

  async function testEmail() {
    const email = prompt('Enter email address to send test message to:')
    if (!email) return
    
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/config/test-email`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
      })
      if (resp.ok) {
        alert('Verification email sent! Check your inbox.')
      } else {
        const error = await resp.json()
        throw new Error(error.error || 'Failed to send')
      }
    } catch (err) {
      alert('Error: ' + err.message)
    }
  }

  async function fetchData() {
    loading = true
    try {
      console.log('Fetching email configuration...')
      const [templResp, smtpResp] = await Promise.all([
        apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/config/templates`),
        apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/config/smtp`)
      ])
      
      if (templResp.ok) {
        const tData = await templResp.json()
        console.log('Templates received:', tData)
        if (Array.isArray(tData)) templates = tData
      } else {
        const err = await getErrorMessage(templResp, 'Failed templates')
        console.error('Template fetch error:', err)
      }
      
      if (smtpResp.ok) {
        const sData = await smtpResp.json()
        console.log('SMTP settings received:', sData)
        if (sData && typeof sData === 'object') {
          smtpSettings = { ...smtpSettings, ...sData }
        }
      } else {
        const err = await getErrorMessage(smtpResp, 'Failed SMTP')
        console.error('SMTP fetch error:', err)
      }
    } catch (err) {
      console.error('Error fetching email data', err)
    } finally {
      loading = false
    }
  }

  function getTemplateData(id) {
    const t = templates.find(x => x.type === id)
    if (t) return { ...t, _is_mock: false }
    // Default mock data
    return {
      type: id,
      subject: `Your ${id.replace(/_/g, ' ')} link`,
      body_html: `<p>Hello,</p><p>Please use this link: <a href="{{ .ConfirmationURL }}">Click here</a></p><p>Or use this token: <b>{{ .Token }}</b></p>`,
      body_text: `Hello, please use this link: {{ .ConfirmationURL }} or use this token: {{ .Token }}`,
      _is_mock: true
    }
  }

  function openEditor(t) {
    const data = getTemplateData(t.id)
    console.log(`Opening editor for ${t.id}`, data)
    editingTemplate = { ...data, _meta: t }
    modalTab = 'html'
    showModal = true
  }

  async function saveTemplate() {
    saving = true
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/config/templates`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: editingTemplate.type,
          subject: editingTemplate.subject,
          body_html: editingTemplate.body_html,
          body_text: editingTemplate.body_text
        })
      })
      if (!resp.ok) {
        throw new Error(await getErrorMessage(resp, 'Failed to save template'))
      }
      showModal = false
      await fetchData()
    } catch (err) {
      alert('Error saving template: ' + err.message)
    } finally {
      saving = false
    }
  }

  async function saveSMTP() {
    saving = true
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/auth/v1/admin/config/smtp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(smtpSettings)
      })
      if (!resp.ok) {
        throw new Error(await getErrorMessage(resp, 'Failed to save SMTP settings'))
      }
      alert('SMTP Settings Saved!')
    } catch (err) {
      alert('Error saving SMTP: ' + err.message)
    } finally {
      saving = false
    }
  }

  function renderPreview(htmlContent) {
    // Basic placeholder replacement for preview
    let preview = htmlContent
      .replace(/\{\{\s*\.ConfirmationURL\s*\}\}/g, 'https://omnibase.app/verify?token=example_token')
      .replace(/\{\{\s*\.Token\s*\}\}/g, '123456')
      .replace(/\{\{\s*\.Email\s*\}\}/g, 'user@example.com')
      .replace(/\{\{\s*\.SiteURL\s*\}\}/g, 'https://yourapp.com')
    
    // Wrap in some basic styling for the preview container
    return `
      <div style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; color: #333; background: #fff; border-radius: 8px;">
        <div style="padding-bottom: 20px; border-bottom: 1px solid #eee; margin-bottom: 20px;">
          <h2 style="margin: 0; color: #111;">OmniBase</h2>
        </div>
        ${preview}
        <div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666;">
          Sent by OmniBase Authentication
        </div>
      </div>
    `
  }
</script>

<svelte:head>
  <title>Email Flow & Config — OmniBase</title>
</svelte:head>

<div class="page-container">
  <!-- Header -->
  <div class="header-section">
    <div class="header-content">
      <div class="icon-pulse">
        {@html icons.mail}
      </div>
      <div>
        <h1 class="page-title">Email Configuration</h1>
        <p class="page-subtitle">Configure what emails your users receive, how they are sent, and your SMTP server.</p>
      </div>
    </div>
    <div class="tabs-container">
      <button class="nav-tab {activeTab === 'templates' ? 'active' : ''}" onclick={() => activeTab = 'templates'}>Templates</button>
      <button class="nav-tab {activeTab === 'security' ? 'active' : ''}" onclick={() => activeTab = 'security'}>Security Alerts</button>
      <button class="nav-tab {activeTab === 'smtp' ? 'active' : ''}" onclick={() => activeTab = 'smtp'}>SMTP Settings</button>
    </div>
  </div>

  <div class="main-content">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Loading configuration...</p>
      </div>
    {:else}
      {#if activeTab === 'templates'}
        <div class="grid-layout" in:fade>
          {#each authTemplates as t}
            <!-- svelte-ignore a11y-click-events-have-key-events -->
            <!-- svelte-ignore a11y-no-static-element-interactions -->
            <div class="template-card" onclick={() => openEditor(t)}>
              <div class="card-glow"></div>
              <div class="card-icon">{@html t.icon}</div>
              <div class="card-body">
                <div class="card-title-row">
                  <h3>{t.name}</h3>
                  {#if !getTemplateData(t.id)._is_mock}
                    <span class="status-badge saved">Saved in DB</span>
                  {/if}
                </div>
                <p>{t.desc}</p>
                {#if !getTemplateData(t.id)._is_mock}
                  <span class="update-text">Modified {new Date(getTemplateData(t.id).updated_at).toLocaleDateString()}</span>
                {/if}
              </div>
              <div class="card-action">
                <button class="edit-btn">Edit <span>⟶</span></button>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      {#if activeTab === 'security'}
        <div class="list-layout" in:fade>
          <div class="info-banner">
            <span class="badge new-badge">NEW</span>
            <div class="banner-text">
              <strong>Notify users about security-sensitive actions</strong>
              <span>We've expanded our templates to handle critical security alerts. Toggle them on or off below.</span>
            </div>
          </div>

          <div class="settings-list">
            {#each securityTemplates as st}
              <div class="list-item">
                <div class="item-info">
                  <h4>{st.name}</h4>
                  <p>{st.desc}</p>
                </div>
                <div class="item-controls">
                  <label class="toggle-switch">
                    <input type="checkbox" bind:checked={st.active} />
                    <span class="slider"></span>
                  </label>
                  <button class="icon-btn" onclick={() => openEditor(st)}>
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
                  </button>
                </div>
              </div>
            {/each}
          </div>

          <div class="action-bar">
            <button class="save-btn" onclick={saveSecuritySettings} disabled={saving}>
              {saving ? 'Saving...' : 'Save Security Settings'}
            </button>
          </div>
        </div>
      {/if}

      {#if activeTab === 'smtp'}
        <div class="smtp-layout" in:fade>
          <!-- Master Toggle -->
          <div class="config-panel master-panel">
            <div class="panel-header" style="border:none; padding-bottom:0;">
              <div>
                <h3>Enable Custom SMTP</h3>
                <p>Emails will be sent using your custom SMTP provider instead of the default local mailer.</p>
              </div>
              <label class="toggle-switch large">
                <input type="checkbox" bind:checked={smtpSettings.enable_custom_smtp} />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <div class="form-grid" style="opacity: {smtpSettings.enable_custom_smtp ? 1 : 0.5}; pointer-events: {smtpSettings.enable_custom_smtp ? 'auto' : 'none'}">
            <!-- Sender Details -->
            <div class="config-panel">
              <div class="panel-header">
                <h3>Sender Details</h3>
                <p>Configure the sender information for your outgoing emails.</p>
              </div>
              <div class="panel-body">
                <label class="input-group">
                  <span>Sender Email Address</span>
                  <input type="email" bind:value={smtpSettings.sender_email} placeholder="no-reply@yourdomain.com" />
                  <span class="help-text">The address emails are sent from.</span>
                </label>
                <label class="input-group">
                  <span>Sender Name</span>
                  <input type="text" bind:value={smtpSettings.sender_name} placeholder="Acme Corp" />
                  <span class="help-text">Name displayed in the recipient's inbox.</span>
                </label>
              </div>
            </div>

            <!-- Provider Settings -->
            <div class="config-panel">
              <div class="panel-header">
                <h3>SMTP Provider Settings</h3>
                <p>Enter your provider's credentials securely.</p>
              </div>
              <div class="panel-body">
                <div class="row">
                  <label class="input-group" style="flex:2">
                    <span>Host</span>
                    <input type="text" bind:value={smtpSettings.host} placeholder="smtp.resend.com" />
                  </label>
                  <label class="input-group" style="flex:1">
                    <span>Port</span>
                    <input type="number" bind:value={smtpSettings.port} placeholder="465" />
                  </label>
                </div>
                <div class="row">
                  <label class="input-group" style="flex:1">
                    <span>Username</span>
                    <input type="text" bind:value={smtpSettings.username} placeholder="Username" />
                  </label>
                  <label class="input-group" style="flex:1">
                    <span>Password</span>
                    <input type="password" bind:value={smtpSettings.password} placeholder="••••••••" />
                  </label>
                </div>
                <div class="row-settings">
                   <label class="checkbox-label">
                     <input type="checkbox" bind:checked={smtpSettings.secure} />
                     <span>Use Secure Connection (TLS/SSL)</span>
                   </label>
                </div>
              </div>
            </div>

            <!-- Rate Limits -->
            <div class="config-panel full-width">
               <div class="panel-header">
                  <h3>Rate Limits</h3>
               </div>
               <div class="panel-body flex-row">
                 <label class="input-group" style="max-width:300px">
                    <span>Minimum interval per user (seconds)</span>
                    <div class="suffix-input">
                       <input type="number" bind:value={smtpSettings.min_interval} min="1" />
                       <span class="suffix">sec</span>
                    </div>
                    <span class="help-text">Time before another email can be sent.</span>
                 </label>
               </div>
            </div>
          </div>

          <div class="action-bar">
            <button class="test-btn" onclick={testEmail}>Send Test Email</button>
            <button class="save-btn" onclick={saveSMTP} disabled={saving}>
              {saving ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</div>

<!-- Template Editor Modal -->
{#if showModal && editingTemplate}
  <div class="modal-backdrop" transition:fade={{duration: 150}}>
    <div class="modal-content" transition:scale={{start: 0.95, duration: 200}}>
      <div class="modal-header">
         <div class="modal-title">
           <span class="modal-icon">{@html editingTemplate._meta.icon || icons.mail}</span>
           <h3>Edit {editingTemplate._meta.name}</h3>
         </div>
         <button class="close-btn" onclick={() => showModal = false}>✕</button>
      </div>

      <div class="modal-body">
         <div class="info-alert">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right:8px; vertical-align:middle;"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            You can use variables like <code>{`{{ .ConfirmationURL }}`}</code>, <code>{`{{ .Email }}`}</code>, and <code>{`{{ .Token }}`}</code>.
         </div>

         <label class="input-group">
           <span>Email Subject</span>
           <input type="text" bind:value={editingTemplate.subject} />
         </label>

         <div class="editor-tabs">
            <button class="ed-tab {modalTab === 'html' ? 'active' : ''}" onclick={() => modalTab = 'html'}>HTML Body</button>
            <button class="ed-tab {modalTab === 'text' ? 'active' : ''}" onclick={() => modalTab = 'text'}>Text Body</button>
            <button class="ed-tab {modalTab === 'preview' ? 'active' : ''}" onclick={() => modalTab = 'preview'}>Preview</button>
         </div>
         
         {#if modalTab === 'html'}
           <label class="input-group" style="margin-top:0;">
             <textarea rows="12" bind:value={editingTemplate.body_html} style="font-family: 'JetBrains Mono', 'Fira Code', monospace; font-size: 13px;" placeholder="<h1>Hello</h1>..."></textarea>
           </label>
         {:else if modalTab === 'text'}
           <label class="input-group" style="margin-top:0;">
             <textarea rows="12" bind:value={editingTemplate.body_text} style="font-family: inherit; font-size: 14px;" placeholder="Hello, use this link..."></textarea>
           </label>
         {:else if modalTab === 'preview'}
           <div class="preview-container" in:fade>
              <div class="preview-browser">
                 <div class="browser-dot"></div>
                 <div class="browser-dot"></div>
                 <div class="browser-dot"></div>
              </div>
              <div class="preview-content">
                {@html renderPreview(editingTemplate.body_html)}
              </div>
           </div>
         {/if}
      </div>

      <div class="modal-footer">
         <button class="test-btn-alt" onclick={testEmail}>Send Test Email</button>
         <div style="flex:1"></div>
         <button class="btn btn-secondary" onclick={() => showModal = false}>Cancel</button>
         <button class="btn btn-primary" onclick={saveTemplate} disabled={saving}>
            {saving ? 'Saving...' : 'Save Changes'}
         </button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Premium Dark Theme Variables */
  :root {
    --bg-base: #0a0a0b;
    --bg-surface: #121214;
    --bg-surface-hover: #1c1c1f;
    --border-dim: #27272a;
    --border-focus: #4f46e5;
    --primary: #6366f1;
    --primary-glow: rgba(99, 102, 241, 0.2);
    --text-primary: #f4f4f5;
    --text-secondary: #a1a1aa;
    --text-muted: #71717a;
  }

  .page-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 32px 40px;
    font-family: 'Inter', sans-serif;
    color: var(--text-primary);
    position: relative;
    z-index: 1;
  }

  .header-section {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    margin-bottom: 40px;
    border-bottom: 1px solid var(--border-dim);
    padding-bottom: 0;
  }

  .header-content {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 24px;
  }

  .icon-pulse {
    width: 64px;
    height: 64px;
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(168, 85, 247, 0.1));
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 0 20px rgba(99, 102, 241, 0.1);
    color: var(--primary);
  }
  .icon-pulse :global(svg) {
    width: 32px;
    height: 32px;
  }

  .page-title {
    font-size: 28px;
    font-weight: 700;
    margin: 0 0 4px 0;
    letter-spacing: -0.5px;
  }

  .page-subtitle {
    color: var(--text-secondary);
    margin: 0;
    font-size: 14px;
  }

  .tabs-container {
    display: flex;
    gap: 24px;
  }

  .nav-tab {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 14px;
    font-weight: 500;
    padding: 0 0 12px 0;
    cursor: pointer;
    position: relative;
    transition: color 0.2s;
  }

  .nav-tab:hover {
    color: var(--text-primary);
  }

  .nav-tab.active {
    color: var(--primary);
  }

  .nav-tab.active::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 2px;
    background: var(--primary);
    border-radius: 2px 2px 0 0;
  }

  /* Grid Layout for Templates */
  .grid-layout {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 24px;
  }

  .template-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-dim);
    border-radius: 16px;
    padding: 24px;
    display: flex;
    align-items: flex-start;
    gap: 16px;
    position: relative;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .card-glow {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 100%;
    background: radial-gradient(circle at top right, rgba(99, 102, 241, 0.08), transparent 70%);
    opacity: 0;
    transition: opacity 0.3s;
  }

  .template-card:hover {
    transform: translateY(-2px);
    border-color: rgba(99, 102, 241, 0.3);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2), 0 0 15px rgba(99, 102, 241, 0.1);
  }

  .template-card:hover .card-glow {
    opacity: 1;
  }

  .card-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.03);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    border: 1px solid rgba(255, 255, 255, 0.05);
    color: var(--text-secondary);
  }
  .card-icon :global(svg) {
    width: 24px;
    height: 24px;
  }

  .card-body h3 {
    margin: 0 0 6px 0;
    font-size: 15px;
    font-weight: 600;
  }

  .card-body p {
    margin: 0;
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .card-action {
    position: absolute;
    right: 24px;
    top: 24px;
  }

  .edit-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 13px;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 4px;
    opacity: 0;
    transform: translateX(-10px);
    transition: all 0.3s;
  }

  .template-card:hover .edit-btn {
    opacity: 1;
    transform: translateX(0);
    color: var(--primary);
  }

  /* List Layout for Security */
  .list-layout {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .info-banner {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    padding: 16px 24px;
    background: rgba(16, 185, 129, 0.05);
    border: 1px solid rgba(16, 185, 129, 0.2);
    border-radius: 12px;
  }

  .new-badge {
    background: #10b981;
    color: #fff;
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 4px;
    font-weight: 700;
    margin-top: 2px;
  }

  .banner-text {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .banner-text strong {
    color: #10b981;
    font-size: 14px;
  }
  .banner-text span {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .settings-list {
    background: var(--bg-surface);
    border: 1px solid var(--border-dim);
    border-radius: 12px;
    overflow: hidden;
  }

  .list-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-dim);
  }
  .list-item:last-child {
    border-bottom: none;
  }

  .item-info h4 {
    margin: 0 0 4px 0;
    font-size: 14px;
    font-weight: 500;
  }
  .item-info p {
    margin: 0;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .item-controls {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .icon-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 18px;
    transition: transform 0.2s;
    padding: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
  }
  .icon-btn:hover {
    background: rgba(255,255,255,0.05);
    color: #fff;
  }

  /* SMTP Layout */
  .smtp-layout {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    transition: opacity 0.3s;
  }

  .config-panel {
    background: var(--bg-surface);
    border: 1px solid var(--border-dim);
    border-radius: 12px;
    overflow: hidden;
  }
  
  .config-panel.master-panel {
    background: linear-gradient(to right, rgba(99, 102, 241, 0.05), transparent);
    border-color: rgba(99, 102, 241, 0.2);
  }

  .config-panel.full-width {
    grid-column: 1 / -1;
  }

  .panel-header {
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-dim);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .panel-header h3 {
    margin: 0 0 4px 0;
    font-size: 15px;
    font-weight: 600;
  }

  .panel-header p {
    margin: 0;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .panel-body {
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .row {
    display: flex;
    gap: 16px;
  }
  
  .flex-row {
    flex-direction: row;
    align-items: flex-end;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .input-group span:first-child {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .help-text {
    font-size: 12px !important;
    color: var(--text-muted) !important;
  }

  input[type="text"], input[type="email"], input[type="password"], input[type="number"], textarea {
    background: #000;
    border: 1px solid var(--border-dim);
    padding: 10px 14px;
    border-radius: 8px;
    color: #fff;
    font-size: 14px;
    outline: none;
    transition: all 0.2s;
  }

  input:focus, textarea:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 2px var(--primary-glow);
  }

  .suffix-input {
    display: flex;
    align-items: center;
    background: #000;
    border: 1px solid var(--border-dim);
    border-radius: 8px;
    overflow: hidden;
  }

  .suffix-input input {
    border: none;
    border-radius: 0;
    flex: 1;
  }
  
  .suffix-input input:focus {
    box-shadow: none;
  }

  .suffix-input:focus-within {
    border-color: var(--primary);
    box-shadow: 0 0 0 2px var(--primary-glow);
  }

  .suffix {
    padding: 0 14px;
    background: #111;
    color: var(--text-secondary);
    font-size: 13px;
    border-left: 1px solid var(--border-dim);
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    cursor: pointer;
    margin-top: 8px;
  }

  .action-bar {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  .save-btn {
    background: var(--text-primary);
    color: #000;
    padding: 10px 24px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: transform 0.1s, opacity 0.2s;
  }
  .save-btn:active {
    transform: scale(0.97);
  }
  .save-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Toggle Switch */
  .toggle-switch {
    position: relative;
    display: inline-block;
    width: 36px;
    height: 20px;
  }
  .toggle-switch.large {
    width: 44px;
    height: 24px;
  }
  .toggle-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--border-dim);
    border: 1px solid rgba(255,255,255,0.1);
    transition: .4s;
    border-radius: 34px;
  }
  .slider:before {
    position: absolute;
    content: "";
    height: 14px;
    width: 14px;
    left: 2px;
    bottom: 2px;
    background-color: white;
    transition: .3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
    border-radius: 50%;
  }
  .toggle-switch.large .slider:before {
    height: 18px;
    width: 18px;
  }
  input:checked + .slider {
    background-color: var(--primary);
  }
  input:checked + .slider:before {
    transform: translateX(16px);
  }
  .toggle-switch.large input:checked + .slider:before {
    transform: translateX(20px);
  }

  /* Modal Styles */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(12px);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .modal-content {
    background: var(--bg-surface);
    border: 1px solid var(--border-dim);
    border-radius: 20px;
    width: 100%;
    max-width: 800px;
    box-shadow: 0 40px 100px -20px rgba(0, 0, 0, 1);
    overflow: hidden;
  }

  .modal-header {
    padding: 24px 32px;
    border-bottom: 1px solid var(--border-dim);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .modal-title {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .modal-icon {
    width: 36px;
    height: 36px;
    background: rgba(255,255,255,0.03);
    border: 1px solid var(--border-dim);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--primary);
  }
  .modal-icon :global(svg) {
    width: 20px;
    height: 20px;
  }

  .modal-title h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 700;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 20px;
    cursor: pointer;
    padding: 8px;
    border-radius: 8px;
  }
  .close-btn:hover { background: rgba(255,255,255,0.05); color: #fff; }

  .modal-body {
    padding: 32px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .info-alert {
    background: rgba(99, 102, 241, 0.05);
    border: 1px solid rgba(99, 102, 241, 0.2);
    padding: 14px 20px;
    border-radius: 12px;
    font-size: 13px;
    color: #c7d2fe;
    display: flex;
    align-items: center;
  }
  .info-alert code {
    background: rgba(0,0,0,0.4);
    padding: 2px 6px;
    border-radius: 4px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    color: #818cf8;
    margin: 0 4px;
  }

  .editor-tabs {
    display: flex;
    border-bottom: 1px solid var(--border-dim);
    gap: 32px;
    margin-bottom: -10px;
  }
  .ed-tab {
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-secondary);
    padding: 10px 0;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }
  .ed-tab:hover { color: #fff; }
  .ed-tab.active {
    color: var(--primary);
    border-bottom-color: var(--primary);
  }

  .preview-container {
    background: #1a1a1c;
    border: 1px solid var(--border-dim);
    border-radius: 12px;
    overflow: hidden;
  }
  .preview-browser {
    height: 32px;
    background: #232326;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 16px;
    border-bottom: 1px solid var(--border-dim);
  }
  .browser-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #3a3a3d;
  }
  .preview-content {
    background: #f8fafc;
    min-height: 200px;
    max-height: 400px;
    overflow-y: auto;
    padding: 40px;
  }

  .modal-footer {
    padding: 24px 32px;
    border-top: 1px solid var(--border-dim);
    display: flex;
    justify-content: flex-end;
    gap: 16px;
    background: rgba(0,0,0,0.1);
  }

  .btn {
    padding: 12px 28px;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    border: none;
    transition: all 0.2s;
  }
  .btn:active { transform: scale(0.98); }
  .btn:disabled { opacity: 0.5; }
  .btn-primary {
    background: var(--primary);
    color: #fff;
    box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
  }
  .btn-primary:hover:not(:disabled) {
    background: #4f46e5;
    transform: translateY(-1px);
  }
  .btn-secondary {
    background: rgba(255,255,255,0.03);
    border: 1px solid var(--border-dim);
    color: var(--text-primary);
  }
  .btn-secondary:hover {
    background: rgba(255,255,255,0.08);
  }

  /* Loaders */
  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 400px;
    gap: 16px;
  }
  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid rgba(255,255,255,0.05);
    border-top-color: var(--primary);
    border-radius: 50%;
    animation: spin 0.8s infinite cubic-bezier(0.5, 0.1, 0.4, 0.9);
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  textarea {
    resize: none;
    line-height: 1.6;
  }
  .test-btn {
    padding: 0.6rem 1.2rem;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s ease;
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: var(--text-muted);
    margin-right: 0.5rem;
  }

  .test-btn:hover {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.2);
    color: var(--text-color);
  }

  .test-btn-alt {
    padding: 0.5rem 1rem;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.2s;
  }

  .test-btn-alt:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-color);
    border-color: rgba(255, 255, 255, 0.2);
  }
  .card-title-row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin-bottom: 0.3rem;
  }
  .card-title-row h3 { margin-bottom: 0 !important; }
  .status-badge.saved {
    font-size: 10px;
    background: rgba(16, 185, 129, 0.1);
    color: #10b981;
    border: 1px solid rgba(16, 185, 129, 0.2);
    padding: 2px 6px;
    border-radius: 4px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 700;
  }
  .update-text {
    font-size: 11px;
    color: var(--text-muted);
    display: block;
    margin-top: 0.5rem;
    font-style: italic;
  }
</style>
