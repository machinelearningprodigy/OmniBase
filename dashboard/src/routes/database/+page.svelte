<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import { goto } from '$app/navigation'
  import { apiFetch, getErrorMessage, getHeaders, getOmniBaseUrl } from '$lib/api'

  // Types
  interface Table {
    name: string
    schema: string
    row_count?: number
    size?: string
    has_rls: boolean
  }

  interface ColumnMeta {
    name: string
    type: string
    is_nullable?: boolean
    is_primary?: boolean
    default?: string
  }

  interface EditingCell {
    rowKey: string
    columnName: string
    draft: string
  }

  // State
  let tables = $state<Table[]>([])
  let loading = $state(true)
  let selectedTable = $state<string | null>(null)
  let selectedTableMeta = $state<Table | null>(null)
  let tableData = $state<Record<string, unknown>[]>([])
  let columns = $state<ColumnMeta[]>([])
  let dataLoading = $state(false)
  let error = $state<string | null>(null)
  let searchQuery = $state('')
  let currentSchema = $state('public')
  let listError = $state<string | null>(null)
  let tableNotice = $state<string | null>(null)
  
  // Row edits & selection
  let editingCell = $state<EditingCell | null>(null)
  let rowMutationError = $state<string | null>(null)
  let rowMutationLoading = $state<string | null>(null)
  let selectedRows = $state<string[]>([])
  
  // Modals & Menus
  let showCreateModal = $state(false)
  let showInsertModal = $state(false)
  let showDeleteTableModal = $state(false)
  let tableActionLoading = $state(false)
  let showTableSettingsMenu = $state(false)
  let showRowMenu = $state<string | null>(null)

  let insertRowValues = $state<Record<string, string>>({})
  let insertLoading = $state(false)
  let insertError = $state<string | null>(null)
  
  // Custom confirmation modal state
  let confirmDeleteAction = $state<{ type: 'single' | 'bulk', row?: Record<string, unknown> } | null>(null)
  
  // Create table state
  let newTableName = $state('')
  let isCreating = $state(false)
  let newTableColumns = $state([
    { name: 'id', type: 'uuid', is_primary: true, is_nullable: false, default: 'gen_random_uuid()' },
    { name: 'created_at', type: 'timestamp with time zone', is_primary: false, is_nullable: false, default: 'now()' }
  ])

  // Context Menu outside click helper
  function handleWindowClick(e: MouseEvent) {
    const target = e.target as HTMLElement
    if (!target.closest('.dropdown-menu-container') && !target.closest('.dropdown-trigger')) {
      showTableSettingsMenu = false
      showRowMenu = null
    }
  }

  onMount(() => {
    window.addEventListener('click', handleWindowClick)
    return () => window.removeEventListener('click', handleWindowClick)
  })

  // API Headers
  function buildRestHeaders(schema: string, includeContentType = false) {
    const headers: Record<string, string> = { ...getHeaders(includeContentType) }
    if (schema !== 'public') {
      headers['Accept-Profile'] = schema
      if (includeContentType) headers['Content-Profile'] = schema
    }
    return headers
  }

  function normalizeTables(payload: Table[] | Table | null | undefined) {
    if (Array.isArray(payload)) return payload
    if (payload && typeof payload === 'object') return [payload]
    return []
  }

  function getPendingCreatedTable() {
    try {
      const fromQuery = $page.url.searchParams.get('table')
      const schemaFromQuery = $page.url.searchParams.get('schema')
      if (fromQuery) return { table: fromQuery, schema: schemaFromQuery || currentSchema }
      const raw = localStorage.getItem('omnibase.last_created_table')
      if (!raw) return null
      const parsed = JSON.parse(raw)
      if (!parsed?.table || !parsed?.createdAt || Date.now() - parsed.createdAt > 5 * 60 * 1000) {
        localStorage.removeItem('omnibase.last_created_table')
        return null
      }
      return { table: parsed.table as string, schema: (parsed.schema as string) || currentSchema }
    } catch {
      localStorage.removeItem('omnibase.last_created_table')
      return null
    }
  }

  function clearPendingCreatedTable() {
    localStorage.removeItem('omnibase.last_created_table')
  }

  function mergePendingTable(items: Table[]) {
    const pending = getPendingCreatedTable()
    if (!pending || pending.schema !== currentSchema) return items
    if (items.some((table) => table.name === pending.table && table.schema === pending.schema)) return items
    return [{ name: pending.table, schema: pending.schema, row_count: 0, size: '0 bytes', has_rls: false }, ...items]
  }

  function getPrimaryColumns() {
    return columns.filter((column) => column.is_primary)
  }

  function getRowKey(row: Record<string, unknown>) {
    const primaryColumns = getPrimaryColumns()
    if (primaryColumns.length === 0) return JSON.stringify(row) // Fallback for no PK
    return primaryColumns.map((column) => `${column.name}:${String(row[column.name])}`).join('|')
  }

  function encodeFilterValue(value: unknown) {
    if (value === null || value === undefined) return 'is.null'
    return `eq.${encodeURIComponent(String(value))}`
  }

  function getRowFilterQuery(row: Record<string, unknown>) {
    const primaryColumns = getPrimaryColumns()
    if (primaryColumns.length === 0) return null
    return primaryColumns.map((column) => `${encodeURIComponent(column.name)}=${encodeFilterValue(row[column.name])}`).join('&')
  }

  function getDisplayValue(value: unknown) {
    if (value === null || value === undefined) return ''
    if (typeof value === 'object') return JSON.stringify(value)
    return String(value)
  }

  function truncate(value: string, max = 80) {
    return value.length > max ? `${value.slice(0, max)}...` : value
  }

  function parseColumnValue(column: ColumnMeta, draft: string) {
    const trimmed = draft.trim()
    if (trimmed === '') return null
    const type = column.type.toLowerCase()
    if (type === 'bool' || type === 'boolean') return trimmed === 'true' || trimmed === '1' || trimmed.toLowerCase() === 'yes'
    if (['int2', 'int4', 'int8', 'float4', 'float8', 'numeric'].includes(type)) {
      const numericValue = Number(trimmed)
      return Number.isNaN(numericValue) ? draft : numericValue
    }
    if (type === 'json' || type === 'jsonb') return JSON.parse(draft)
    return draft
  }

  function getPolicyStarterSQL() {
    if (!selectedTableMeta) return ''
    const qualified = `${selectedTableMeta.schema}."${selectedTableMeta.name}"`
    const hasUserID = columns.some((column) => column.name === 'user_id')
    if (hasUserID) {
      return [
        `ALTER TABLE ${qualified} ENABLE ROW LEVEL SECURITY;`,
        '',
        `CREATE POLICY "${selectedTableMeta.name}_select_own" ON ${qualified} FOR SELECT TO authenticated USING (auth.uid() = user_id);`,
        `CREATE POLICY "${selectedTableMeta.name}_insert_own" ON ${qualified} FOR INSERT TO authenticated WITH CHECK (auth.uid() = user_id);`,
        `CREATE POLICY "${selectedTableMeta.name}_update_own" ON ${qualified} FOR UPDATE TO authenticated USING (auth.uid() = user_id) WITH CHECK (auth.uid() = user_id);`
      ].join('\n')
    }
    return [
      `ALTER TABLE ${qualified} ENABLE ROW LEVEL SECURITY;`,
      '',
      `CREATE POLICY "${selectedTableMeta.name}_authenticated_read" ON ${qualified} FOR SELECT TO authenticated USING (true);`,
      `CREATE POLICY "${selectedTableMeta.name}_authenticated_write" ON ${qualified} FOR ALL TO authenticated USING (true) WITH CHECK (true);`,
      '',
      `-- OmniBase note: tighten the predicates above for production use.`
    ].join('\n')
  }

  function openPolicyStarter() {
    const sql = getPolicyStarterSQL()
    if (!sql) return
    window.location.href = `/database/editor?query=${encodeURIComponent(sql)}`
  }

  async function loadTables(focusTable?: string) {
    try {
      loading = true
      error = null
      listError = null
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/tables?schema=${currentSchema}`, { headers: getHeaders() })
      if (!resp.ok) {
        listError = await getErrorMessage(resp, `Failed to load tables: ${resp.statusText}`)
        error = listError
        tables = []
        return
      }
      tables = mergePendingTable(normalizeTables(await resp.json()))
      const nextTable = focusTable || $page.url.searchParams.get('table')
      if (nextTable) {
        const match = tables.find((table) => table.name === nextTable)
        if (match) {
          tableNotice = `Table ${match.schema}.${match.name} is ready.`
          clearPendingCreatedTable()
          await selectTable(match)
        }
      }
    } catch {
      listError = 'Could not connect to the API gateway. Is it running?'
      error = listError
      tables = []
    } finally {
      loading = false
    }
  }

  async function loadColumns(table: Table) {
    const resp = await apiFetch(`${getOmniBaseUrl()}/pg/columns?schema=${table.schema}&table=${table.name}`, {
      headers: getHeaders(false)
    })
    if (!resp.ok) throw new Error(await getErrorMessage(resp, `Failed to load columns: ${resp.statusText}`))
    columns = (await resp.json()) as ColumnMeta[]
  }

  async function selectTable(table: Table) {
    selectedTable = table.name
    selectedTableMeta = table
    currentSchema = table.schema || currentSchema
    dataLoading = true
    error = null
    rowMutationError = null
    editingCell = null
    selectedRows = []
    showTableSettingsMenu = false

    // Update URL to match current view
    if ($page.url.searchParams.get('table') !== table.name || $page.url.searchParams.get('schema') !== currentSchema) {
      goto(`/database?schema=${currentSchema}&table=${table.name}`, { replaceState: true, keepFocus: true })
    }

    try {
      await loadColumns(table)
      const resp = await apiFetch(`${getOmniBaseUrl()}/rest/v1/${table.name}?limit=50&select=*`, {
        headers: buildRestHeaders(table.schema)
      })
      if (!resp.ok) {
        error = await getErrorMessage(resp, `Failed to load data: ${resp.statusText}`)
        tableData = []
        return
      }
      tableData = await resp.json()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not connect to the API gateway. Is it running?'
      tableData = []
      columns = []
    } finally {
      dataLoading = false
    }
  }

  async function focusTableFromRoute() {
    const nextTable = $page.url.searchParams.get('table')
    const nextSchema = $page.url.searchParams.get('schema')
    if (nextSchema && nextSchema !== currentSchema) {
      currentSchema = nextSchema
      localStorage.setItem('omnibase.current_schema', currentSchema)
      await loadTables(nextTable || undefined)
      return
    }
    if (!nextTable || tables.length === 0) return
    const match = tables.find((table) => table.name === nextTable && table.schema === currentSchema)
    if (!match || selectedTable === match.name) return
    tableNotice = `Table ${match.schema}.${match.name} is ready.`
    clearPendingCreatedTable()
    await selectTable(match)
  }

  function addColumn() {
    newTableColumns = [...newTableColumns, { name: '', type: 'text', is_primary: false, is_nullable: true, default: '' }]
  }

  async function createTable() {
    if (!newTableName) return alert('Table name is required')
    try {
      isCreating = true
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/tables`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ name: newTableName, schema: currentSchema, columns: newTableColumns.filter((column) => column.name) })
      })
      if (!resp.ok) {
        alert(`Error: ${await getErrorMessage(resp, 'Failed to create table')}`)
        return
      }
      showCreateModal = false
      tableNotice = `Table ${currentSchema}.${newTableName} created.`
      localStorage.setItem('omnibase.last_created_table', JSON.stringify({ schema: currentSchema, table: newTableName, createdAt: Date.now() }))
      window.dispatchEvent(new CustomEvent('omnibase:table-created', { detail: { schema: currentSchema, table: newTableName } }))
      newTableName = ''
      newTableColumns = [
        { name: 'id', type: 'uuid', is_primary: true, is_nullable: false, default: 'gen_random_uuid()' },
        { name: 'created_at', type: 'timestamp with time zone', is_primary: false, is_nullable: false, default: 'now()' }
      ]
      await loadTables(JSON.parse(localStorage.getItem('omnibase.last_created_table') || '{}').table)
    } catch {
      alert('Failed to connect to gateway')
    } finally {
      isCreating = false
    }
  }

  function openInsertModal() {
    insertError = null
    const initial: Record<string, string> = {}
    for (const column of columns) initial[column.name] = ''
    insertRowValues = initial
    showInsertModal = true
  }

  function closeInsertModal() {
    showInsertModal = false
    insertRowValues = {}
    insertError = null
  }

  async function submitInsert() {
    if (!selectedTable || !selectedTableMeta) return
    insertLoading = true
    insertError = null
    const row: Record<string, unknown> = {}
    for (const column of columns) {
      const raw = insertRowValues[column.name] ?? ''
      if (raw === '') continue
      try {
        row[column.name] = parseColumnValue(column, raw)
      } catch {
        insertError = `Column ${column.name} expects valid ${column.type} data.`
        insertLoading = false
        return
      }
    }
    try {
      const headers = buildRestHeaders(currentSchema, true)
      headers.Prefer = 'return=representation'
      const resp = await apiFetch(`${getOmniBaseUrl()}/rest/v1/${selectedTable}`, { method: 'POST', headers, body: JSON.stringify(row) })
      if (!resp.ok) {
        insertError = await getErrorMessage(resp, resp.statusText)
        return
      }
      await selectTable(selectedTableMeta)
      closeInsertModal()
    } catch {
      insertError = 'Failed to connect to API'
    } finally {
      insertLoading = false
    }
  }

  function beginEdit(row: Record<string, unknown>, column: ColumnMeta) {
    if (column.is_primary || rowMutationLoading) return
    if (getPrimaryColumns().length === 0) {
      rowMutationError = 'Inline editing requires a primary key on the table.'
      return
    }
    editingCell = { rowKey: getRowKey(row), columnName: column.name, draft: getDisplayValue(row[column.name]) }
    rowMutationError = null
  }

  function cancelEdit() {
    editingCell = null
  }

  async function saveEdit(row: Record<string, unknown>, column: ColumnMeta) {
    if (!selectedTable || !selectedTableMeta || !editingCell) return
    const rowFilter = getRowFilterQuery(row)
    if (!rowFilter) {
      rowMutationError = 'Inline editing requires a primary key on the table.'
      return
    }
    let nextValue: unknown
    try {
      nextValue = parseColumnValue(column, editingCell.draft)
    } catch {
      rowMutationError = `Column ${column.name} expects valid ${column.type} data.`
      return
    }
    rowMutationLoading = editingCell.rowKey
    rowMutationError = null
    try {
      const headers = buildRestHeaders(currentSchema, true)
      headers.Prefer = 'return=representation'
      const resp = await apiFetch(`${getOmniBaseUrl()}/rest/v1/${selectedTable}?${rowFilter}`, {
        method: 'PATCH',
        headers,
        body: JSON.stringify({ [column.name]: nextValue })
      })
      if (!resp.ok) {
        rowMutationError = await getErrorMessage(resp, `Failed to update ${column.name}`)
        return
      }
      const data = await resp.json()
      if (Array.isArray(data) && data.length === 0) {
        rowMutationError = 'Row update matched 0 database records. Please refresh.'
        return
      }
      await selectTable(selectedTableMeta)
      editingCell = null
      tableNotice = `${column.name} updated on ${selectedTableMeta.schema}.${selectedTableMeta.name}.`
    } catch {
      rowMutationError = 'Failed to save row changes.'
    } finally {
      rowMutationLoading = null
    }
  }

  function toggleRowSelection(row: Record<string, unknown>) {
    const key = getRowKey(row)
    if (selectedRows.includes(key)) {
      selectedRows = selectedRows.filter((k) => k !== key)
    } else {
      selectedRows = [...selectedRows, key]
    }
  }

  function toggleAllRows() {
    if (selectedRows.length === tableData.length) {
      selectedRows = []
    } else {
      selectedRows = tableData.map((row) => getRowKey(row))
    }
  }

  async function deleteSelectedRows() {
    console.log('deleteSelectedRows: started', { selectedRows })
    if (!selectedTableMeta || selectedRows.length === 0) {
      console.warn('deleteSelectedRows: No table selected or rows selected')
      return
    }
    confirmDeleteAction = { type: 'bulk' }
  }

  async function executeBulkDelete() {
    confirmDeleteAction = null
    rowMutationLoading = 'bulk-delete'
    rowMutationError = null
    let deletedCount = 0
    let failedDueToPolicy = false
    try {
      const headers = buildRestHeaders(currentSchema, true)
      headers.Prefer = 'return=representation'
      for (const rowKey of selectedRows) {
        const row = tableData.find((r) => getRowKey(r) === rowKey)
        if (!row) continue
        const rowFilter = getRowFilterQuery(row)
        if (rowFilter) {
          const url = `${getOmniBaseUrl()}/rest/v1/${selectedTableMeta!.name}?${rowFilter}`
          const resp = await apiFetch(url, {
            method: 'DELETE',
            headers
          })
          if (resp.ok) {
            const data = await resp.json()
            if (Array.isArray(data) && data.length > 0) {
              deletedCount++
            } else {
              failedDueToPolicy = true
            }
          }
        }
      }
      
      if (deletedCount === 0 && selectedRows.length > 0 && failedDueToPolicy) {
        rowMutationError = 'No rows were deleted. Check if Row Level Security (RLS) is blocking your DELETE request.'
      } else if (deletedCount < selectedRows.length) {
        rowMutationError = `Deleted ${deletedCount} of ${selectedRows.length} rows. Some deletions may have been blocked by RLS.`
      } else {
        tableNotice = `Successfully deleted ${deletedCount} rows.`
        selectedRows = []
      }
      await selectTable(selectedTableMeta!)
    } catch (err) {
      rowMutationError = 'Failed to delete selected rows.'
    } finally {
      rowMutationLoading = null
    }
  }

  async function deleteRow(row: Record<string, unknown>) {
    if (!selectedTableMeta) return
    const rowFilter = getRowFilterQuery(row)
    if (!rowFilter) {
      rowMutationError = 'Row deletion requires a primary key on the table.'
      return
    }
    confirmDeleteAction = { type: 'single', row }
  }

  async function executeSingleDelete(row: Record<string, unknown>) {
    confirmDeleteAction = null
    const rowFilter = getRowFilterQuery(row)
    if (!rowFilter || !selectedTableMeta) return
    
    rowMutationLoading = getRowKey(row)
    rowMutationError = null
    try {
      const headers = buildRestHeaders(currentSchema, true)
      headers.Prefer = 'return=representation'
      const resp = await apiFetch(`${getOmniBaseUrl()}/rest/v1/${selectedTableMeta.name}?${rowFilter}`, {
        method: 'DELETE',
        headers
      })
      if (!resp.ok) {
        rowMutationError = await getErrorMessage(resp, 'Failed to delete row')
        return
      }
      const data = await resp.json()
      if (Array.isArray(data) && data.length > 0) {
        await selectTable(selectedTableMeta)
        tableNotice = `Row removed from ${selectedTableMeta.schema}.${selectedTableMeta.name}.`
      } else {
        rowMutationError = 'Deletion failed. This row may be protected by Row Level Security (RLS) policies.'
      }
    } catch {
      rowMutationError = 'Failed to delete row.'
    } finally {
      rowMutationLoading = null
      showRowMenu = null
    }
  }

  async function duplicateRow(row: Record<string, unknown>) {
    if (!selectedTableMeta) return
    rowMutationLoading = getRowKey(row)
    
    // Duplicate everything except primary keys
    const newRow = { ...row }
    for (const column of columns) {
      if (column.is_primary) {
        delete newRow[column.name]
      }
    }
    
    try {
      const headers = buildRestHeaders(currentSchema, true)
      headers.Prefer = 'return=representation'
      const resp = await apiFetch(`${getOmniBaseUrl()}/rest/v1/${selectedTableMeta.name}`, { 
        method: 'POST', 
        headers, 
        body: JSON.stringify(newRow) 
      })
      if (!resp.ok) {
        rowMutationError = await getErrorMessage(resp, 'Failed to duplicate row')
        return
      }
      await selectTable(selectedTableMeta)
      tableNotice = `Row duplicated successfully.`
    } catch {
      rowMutationError = 'Failed to connect to API'
    } finally {
      rowMutationLoading = null
      showRowMenu = null
    }
  }

  function copyJson(row: Record<string, unknown>) {
    navigator.clipboard.writeText(JSON.stringify(row, null, 2))
    tableNotice = "Row JSON copied to clipboard!"
    showRowMenu = null
  }

  async function toggleRLS(enabled: boolean) {
    if (!selectedTableMeta) return
    const currentTable = selectedTableMeta
    tableActionLoading = true
    rowMutationError = null
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/rls`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ schema: currentTable.schema, table: currentTable.name, enabled })
      })
      if (!resp.ok) {
        rowMutationError = await getErrorMessage(resp, 'Failed to update RLS')
        return
      }
      tables = tables.map((table) => table.name === currentTable.name && table.schema === currentTable.schema ? { ...table, has_rls: enabled } : table)
      selectedTableMeta = { ...currentTable, has_rls: enabled }
      tableNotice = enabled ? `RLS enabled on ${currentTable.schema}.${currentTable.name}.` : `RLS disabled on ${currentTable.schema}.${currentTable.name}.`
    } catch {
      rowMutationError = 'Failed to update RLS.'
    } finally {
      tableActionLoading = false
    }
  }

  async function confirmDeleteTable() {
    if (!selectedTableMeta) return
    const currentTable = selectedTableMeta
    tableActionLoading = true
    rowMutationError = null
    try {
      const resp = await apiFetch(`${getOmniBaseUrl()}/pg/tables?schema=${encodeURIComponent(currentTable.schema)}&table=${encodeURIComponent(currentTable.name)}`, {
        method: 'DELETE',
        headers: getHeaders(false)
      })
      if (!resp.ok) {
        rowMutationError = await getErrorMessage(resp, 'Failed to delete table')
        return
      }
      const deletedName = currentTable.name
      tables = tables.filter((table) => !(table.name === currentTable.name && table.schema === currentTable.schema))
      selectedTable = null
      selectedTableMeta = null
      tableData = []
      columns = []
      showDeleteTableModal = false
      tableNotice = `Table ${currentSchema}.${deletedName} deleted.`
      
      // Clear the table from the URL so loadTables doesn't try to re-focus it
      await goto(`/database?schema=${currentSchema}`, { replaceState: true })
      await loadTables()
    } catch {
      rowMutationError = 'Failed to delete table.'
    } finally {
      tableActionLoading = false
    }
  }
  
  async function truncateTable() {
     if (!selectedTableMeta) return
     if (!confirm(`Are you sure you want to EMPTY ALL ROWS from ${selectedTableMeta.name}?`)) return
     // OmniBase doesn't have a truncate endpoint natively yet, but we drop rows if PK exists
     if (getPrimaryColumns().length === 0) {
        rowMutationError = "Truncating without PK via API requires executing a SQL query."
        return
     }
     tableActionLoading = true
     try {
       // Delete all rows hack without filter (if rest API permits it)
       // Usually we need to use a query wrapper
       const sql = `TRUNCATE TABLE ${selectedTableMeta.schema}."${selectedTableMeta.name}";`
       window.location.href = `/database/editor?query=${encodeURIComponent(sql)}`
     } finally {
        tableActionLoading = false
        showTableSettingsMenu = false
     }
  }

  async function generateAIAssist() {
     alert("OmniBase AI Assist: Coming soon! We will automatically generate dummy data and test queries for this table.")
  }

  function exportCSV() {
    if (tableData.length === 0) return
    const keys = Object.keys(tableData[0])
    let csv = keys.join(',') + '\n'
    for (const record of tableData) {
      csv += keys.map(k => {
         let val = record[k] === null ? '' : String(record[k])
         if (val.includes(',') || val.includes('"') || val.includes('\n')) {
             val = `"${val.replace(/"/g, '""')}"`
         }
         return val
      }).join(',') + '\n'
    }
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedTableMeta?.name || 'export'}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  let filteredTables = $derived(tables.filter((table) => table.name.toLowerCase().includes(searchQuery.toLowerCase())))
  let posture = $derived(
    selectedTableMeta 
      ? (!selectedTableMeta.has_rls ? { tone: 'warning', label: 'API exposure warning', copy: `${selectedTableMeta.schema}.${selectedTableMeta.name} is writable without policies.` } : { tone: 'success', label: 'RLS protection active', copy: `${selectedTableMeta.schema}.${selectedTableMeta.name} enforces row-level checks.` })
      : { tone: 'neutral', label: 'No table selected', copy: 'Pick a table to inspect.' }
  )

  onMount(() => {
    const schemaFromQuery = $page.url.searchParams.get('schema')
    currentSchema = schemaFromQuery || localStorage.getItem('omnibase.current_schema') || 'public'
    localStorage.setItem('omnibase.current_schema', currentSchema)
    const pending = getPendingCreatedTable()
    loadTables(pending?.schema === currentSchema ? pending.table : undefined)

    const handleSchemaChange = (e: CustomEvent<string>) => {
      currentSchema = e.detail
      selectedTable = null
      selectedTableMeta = null
      tableNotice = null
      loadTables()
    }

    const handleTableCreated = (e: Event) => {
      const detail = (e as CustomEvent<{ schema: string; table: string }>).detail
      if (!detail?.table) return
      currentSchema = detail.schema || 'public'
      localStorage.setItem('omnibase.current_schema', currentSchema)
      loadTables(detail.table)
    }

    const handleFocus = () => {
      const pendingTable = getPendingCreatedTable()
      if (!pendingTable) return
      currentSchema = pendingTable.schema || currentSchema
      localStorage.setItem('omnibase.current_schema', currentSchema)
      loadTables(pendingTable.table)
    }

    window.addEventListener('omnibase:schema-change', handleSchemaChange as EventListener)
    window.addEventListener('omnibase:table-created', handleTableCreated as EventListener)
    window.addEventListener('focus', handleFocus)
    return () => {
      window.removeEventListener('omnibase:schema-change', handleSchemaChange as EventListener)
      window.removeEventListener('omnibase:table-created', handleTableCreated as EventListener)
      window.removeEventListener('focus', handleFocus)
    }
  })

  $effect(() => {
    void focusTableFromRoute()
  })
</script>

<svelte:head>
  <title>Table Editor - OmniBase</title>
  <meta name="description" content="Browse, edit, and secure your database tables" />
</svelte:head>

<div class="table-editor-shell">
  <aside class="table-sidebar">
    <div class="table-sidebar-header">
      <div class="flex items-center justify-between" style="margin-bottom: 12px;">
        <span class="sidebar-title">Tables</span>
        <button class="btn btn-primary btn-sm" onclick={() => showCreateModal = true}>+ New</button>
      </div>
      <input class="input" type="text" placeholder="Filter tables..." bind:value={searchQuery} />
    </div>

    <div class="table-sidebar-body">
      {#if listError}
        <div class="sidebar-message sidebar-error"><strong>Could not load tables</strong><span>{listError}</span></div>
      {:else if tableNotice}
        <div class="sidebar-message sidebar-success"><strong>{tableNotice}</strong><span>The display refreshed.</span></div>
      {/if}

      {#if loading}
        {#each Array(5) as _}
          <div class="skeleton" style="height: 36px; margin-bottom: 4px; border-radius: var(--radius-sm);"></div>
        {/each}
      {:else if filteredTables.length === 0}
        <div class="empty-state" style="padding: 30px 10px;"><p>No tables found</p></div>
      {:else}
        {#each filteredTables as table}
          <button class="nav-item {selectedTable === table.name ? 'active' : ''}" onclick={() => selectTable(table)}>
            <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">
              <div style="display: flex; align-items: center; gap: 8px;">
                <span class="table-icon">[]</span>
                <span class="table-name">{table.name}</span>
              </div>
              <div class="flex items-center gap-2">
                {#if !table.has_rls}
                  <span class="badge sb-badge-unrestricted">UNRESTRICTED</span>
                {/if}
              </div>
            </div>
          </button>
        {/each}
      {/if}
    </div>
  </aside>

  <div class="table-content">
    {#if !selectedTableMeta}
      <div class="empty-state" style="height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center;">
        <svg fill="currentColor" viewBox="0 0 24 24" style="width: 48px; color: var(--text-muted); opacity: 0.5;"><path d="M4 6H20V18H4V6M4 4C2.9 4 2 4.9 2 6V18C2 19.1 2.9 20 4 20H20C21.1 20 22 19.1 22 18V6C22 4.9 21.1 4 20 4H4Z" /></svg>
        <h3 style="margin-top: 16px;">Select a table to explore</h3>
        <p>Choose a table from the sidebar to view, edit, and secure your data.</p>
        <button class="btn btn-secondary" style="margin-top: 16px;" onclick={() => loadTables()}>{loading ? 'Loading...' : 'Refresh Schema'}</button>
      </div>
    {:else}
      <div class="page-header table-header">
        <div style="display: flex; justify-content: space-between; width: 100%; align-items: center; align-content: center;">
           <div class="table-heading-row" style="display: flex; gap: 12px; align-items: center;">
             <h1 class="page-title table-title" style="margin: 0; display: flex; align-items: center; gap: 8px;">
               {selectedTableMeta.schema}.{selectedTableMeta.name}
               <div class="dropdown-container">
                  <button class="icon-button dropdown-trigger" onclick={() => showTableSettingsMenu = !showTableSettingsMenu}>▾</button>
                  {#if showTableSettingsMenu}
                    <div class="dropdown-menu-container animate-fade-in" style="left: 0;">
                      <a href="/database/editor?table={selectedTableMeta.name}" class="dropdown-item">Open in SQL Editor</a>
                      <button class="dropdown-item" onclick={exportCSV}>Export to CSV</button>
                      <button class="dropdown-item" onclick={() => showTableSettingsMenu = false}>Duplicate Table...</button>
                      <div class="dropdown-divider"></div>
                      <button class="dropdown-item danger" onclick={truncateTable}>Truncate Table</button>
                      <button class="dropdown-item danger" onclick={() => { showTableSettingsMenu = false; showDeleteTableModal = true; }}>Delete Table</button>
                    </div>
                  {/if}
               </div>
             </h1>
           </div>
           
           <div class="table-actions" style="display: flex; gap: 10px;">
             <!-- RLS Policies Tab Mimic -->
             <button class="btn btn-secondary btn-sm" onclick={() => toggleRLS(!(selectedTableMeta?.has_rls ?? false))} disabled={tableActionLoading}>
                <span style="display: flex; align-items: center; gap: 6px;">
                  <span style="color: {selectedTableMeta.has_rls ? '#00e676' : '#ff5252'}; font-size: 14px;">🛡</span>
                  {selectedTableMeta?.has_rls ? 'RLS Policies' : 'Enable RLS'}
                </span>
             </button>
             
             <button class="btn btn-secondary btn-sm" onclick={generateAIAssist}>
               <span style="background: var(--gradient-brand); -webkit-background-clip: text; color: transparent;">✨ AI Enhance</span>
             </button>
             <button class="btn btn-danger btn-sm" style="background: transparent; border-color: #ff5252; color: #ff5252;" onclick={() => showDeleteTableModal = true}>Delete Table</button>
             <button class="btn btn-primary btn-sm" onclick={openInsertModal} style="background: #24b47e; border-color: #24b47e; color: #fff;">Insert Row</button>
           </div>
        </div>
      </div>

      <div class="page-content table-page-content">
        {#if !selectedTableMeta.has_rls}
           <div style="background: rgba(255, 82, 82, 0.08); border: 1px solid rgba(255, 82, 82, 0.2); border-radius: 8px; padding: 12px 16px; margin-bottom: 16px; display: flex; align-items: center; justify-content: space-between;">
              <div style="display: flex; align-items: center; gap: 12px;">
                 <span style="color: #ff5252; font-size: 20px;">!</span>
                 <div>
                   <h4 style="margin: 0; font-size: 14px; color: #fff;">Table is Unrestricted</h4>
                   <p style="margin: 2px 0 0 0; font-size: 13px; color: var(--text-muted);">Users can access and modify any data in this table. Add RLS policies to restrict operations.</p>
                 </div>
              </div>
              <button class="btn btn-primary btn-sm" style="background: #ff5252; color: #fff; border: none;" onclick={() => toggleRLS(true)}>Enable RLS</button>
           </div>
        {/if}

        {#if rowMutationError}
          <div class="card" style="margin-bottom: 16px; border-color: rgba(255,82,82,0.3); background: rgba(255,82,82,0.05);">
            <div style="color: var(--status-error); font-size: 13px;"><strong>Action failed</strong><br>{rowMutationError}</div>
          </div>
        {/if}

        {#if dataLoading}
          <div class="card" style="margin-top: 16px;">{#each Array(8) as _}<div class="skeleton" style="height: 44px; margin-bottom: 4px;"></div>{/each}</div>
        {:else if error}
          <div class="card" style="margin-top: 16px; border-color: rgba(255,82,82,0.3); background: rgba(255,82,82,0.05);">
            <div style="color: var(--status-error); font-size: 13px;"><strong>Error loading data</strong><br>{error}</div>
          </div>
        {:else if columns.length === 0}
          <div class="empty-state" style="margin-top: 20px;"><h3>No columns available</h3><p>Use SQL Editor or recreate the table with valid columns.</p></div>
        {:else if tableData.length === 0}
          <div class="empty-state" style="margin-top: 20px; text-align: center;">
            <div style="font-size: 40px; margin-bottom: 10px; opacity: 0.5;">0</div>
            <h3>Table is empty</h3>
            <p>Add the first row to start tracking your data.</p>
            <button class="btn btn-primary btn-sm" style="margin-top: 12px; background: #24b47e; border: none;" onclick={openInsertModal}>Insert Row</button>
          </div>
        {:else}
          <div class="table-wrapper animate-fade-in sb-table">
             <div class="sb-table-topbar" style="display: flex; justify-content: space-between; align-items: center; min-height: 52px; padding: 0 16px; background: #1a1a22; border-bottom: 1px solid #2a2a35;">
               <span style="font-size: 11px; font-weight: 500; color: var(--text-muted); letter-spacing: 0.02em;">{tableData.length} records • Double click to edit</span>
               {#if selectedRows.length > 0}
                 <div class="animate-fade-in" style="display: flex; gap: 12px; align-items: center;">
                    <span style="font-size: 12px; color: #24b47e; font-weight: 600; background: rgba(36, 180, 126, 0.1); padding: 2px 8px; border-radius: 4px;">{selectedRows.length} selected</span>
                    <button class="btn btn-danger btn-sm" onclick={deleteSelectedRows} disabled={rowMutationLoading === 'bulk-delete'} style="padding: 6px 16px; font-size: 12px; height: 32px; border: none; background: #ff5252; color: #fff; font-weight: 600; border-radius: 6px; box-shadow: 0 2px 8px rgba(255, 82, 82, 0.2); transition: all 0.2s;">
                      {rowMutationLoading === 'bulk-delete' ? 'Deleting...' : `Delete Rows`}
                    </button>
                    <button class="icon-button" onclick={() => selectedRows = []} title="Clear selection" style="color: var(--text-muted);">✕</button>
                 </div>
               {/if}
             </div>
             <div style="overflow-x: auto; width: 100%; padding-bottom: 20px;">
              <table>
                <thead>
                <tr>
                  <th style="width: 40px; text-align: center;">
                    <input type="checkbox" style="cursor: pointer;" checked={tableData.length > 0 && selectedRows.length === tableData.length} onchange={toggleAllRows} />
                  </th>
                  <th style="width: 40px; text-align: center;">#</th>
                  {#each columns as column}
                    <th>
                       <div style="display: flex; align-items: center; gap: 6px;">
                          {#if column.is_primary}
                             <span style="color: #edbb36;" title="Primary Key">🔑</span>
                          {/if}
                          <span>{column.name}</span>
                          <span class="column-meta" title="Type: {column.type}">{column.type.split(' ')[0]}</span>
                       </div>
                    </th>
                  {/each}
                  <th style="width: 50px;"></th>
                </tr>
              </thead>
              <tbody>
                {#each tableData as row, rowIndex}
                  <tr style="background: {selectedRows.includes(getRowKey(row)) ? 'rgba(255, 255, 255, 0.05)' : 'transparent'};">
                    <td style="text-align: center;">
                      <input type="checkbox" style="cursor: pointer;" checked={selectedRows.includes(getRowKey(row))} onchange={() => toggleRowSelection(row)} />
                    </td>
                    <td class="row-index">{rowIndex + 1}</td>
                    {#each columns as column}
                      <!-- The entire TD acts as interactive cell -->
                      <td class:cell-editing={editingCell?.rowKey === getRowKey(row) && editingCell?.columnName === column.name}
                          class:cell-interactive={true}
                          onclick={() => {
                              if(editingCell?.rowKey !== getRowKey(row) || editingCell?.columnName !== column.name) {
                                  beginEdit(row, column)
                              }
                          }}
                          style="padding: 0; min-width: 150px; cursor: cell;"
                      >
                        {#if editingCell?.rowKey === getRowKey(row) && editingCell?.columnName === column.name}
                          <div class="cell-editor">
                            <input
                              class="spreadsheet-input"
                              value={editingCell.draft}
                              autoFocus
                              onclick={(e) => e.stopPropagation()}
                              oninput={(event) => editingCell = { ...editingCell!, draft: (event.target as HTMLInputElement).value }}
                              onkeydown={(event) => {
                                if (event.key === 'Enter') void saveEdit(row, column)
                                if (event.key === 'Escape') cancelEdit()
                              }}
                              onblur={() => saveEdit(row, column)}
                            />
                          </div>
                        {:else}
                          <div class="cell-display" title={column.is_primary ? 'Primary keys are read-only' : 'Click to edit'}>
                            {#if row[column.name] === null}
                              <span class="null-text">NULL</span>
                            {:else}
                              <span class="truncate">{truncate(getDisplayValue(row[column.name]), 100)}</span>
                            {/if}
                          </div>
                        {/if}
                      </td>
                    {/each}
                    <td style="text-align: right; overflow: visible;">
                       <div class="dropdown-container">
                          <button class="icon-button dropdown-trigger row-menu-btn" onclick={(e) => { e.stopPropagation(); showRowMenu = getRowKey(row) }}>⋮</button>
                          {#if showRowMenu === getRowKey(row)}
                            <div class="dropdown-menu-container animate-fade-in" style="right: 0;">
                              <button class="dropdown-item" onclick={() => copyJson(row)}>Copy Row JSON</button>
                              <button class="dropdown-item" onclick={() => duplicateRow(row)} disabled={rowMutationLoading !== null}>Clone Row</button>
                              <div class="dropdown-divider"></div>
                              <button class="dropdown-item danger" onclick={() => deleteRow(row)} disabled={rowMutationLoading !== null}>Delete Row</button>
                            </div>
                          {/if}
                       </div>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
            </div>
          </div>
        {/if}

        {#if getPrimaryColumns().length === 0}
          <div class="metadata-note">OmniBase note: inline editing and row deletion require a primary key. Add one in SQL Editor to enable spreadsheet-style mutations.</div>
        {/if}
      </div>
    {/if}
  </div>
</div>

{#if showInsertModal && selectedTable}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 520px; border-radius: 12px;">
      <div class="modal-header"><h3>Insert row into {selectedTable}</h3><button class="btn-close" onclick={closeInsertModal}>x</button></div>
      <div class="modal-body">
        {#if insertError}<div class="error-banner" style="margin-bottom: 16px;">{insertError}</div>{/if}
        <div style="display: flex; flex-direction: column; gap: 12px;">
          {#each columns as column}
            <div class="form-group" style="display: flex; align-items:center; justify-content: space-between;">
              <label for="ins-{column.name}" style="width: 140px; font-weight: 500;">{column.name} <br/><span style="color: var(--text-muted); font-size: 11px; font-weight: 400;">{column.type}</span></label>
              <input id="ins-{column.name}" type="text" class="input" style="flex: 1; background: #1a1a24; border: 1px solid #333;" placeholder={column.type === 'uuid' ? 'leave empty for default UUID' : 'NULL'} value={insertRowValues[column.name] ?? ''} oninput={(event) => insertRowValues = { ...insertRowValues, [column.name]: (event.target as HTMLInputElement).value }} />
            </div>
          {/each}
        </div>
      </div>
      <div class="modal-footer"><button class="btn btn-secondary" onclick={closeInsertModal}>Cancel</button><button class="btn btn-primary" onclick={submitInsert} disabled={insertLoading} style="background: #24b47e;">{insertLoading ? 'Inserting...' : 'Save'}</button></div>
    </div>
  </div>
{/if}

{#if confirmDeleteAction}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 440px; border-top: 4px solid #ff5252;">
      <div class="modal-header"><h3 style="color: #ff5252;">Confirm Deletion</h3><button class="btn-close" onclick={() => confirmDeleteAction = null}>x</button></div>
      <div class="modal-body">
        <p style="margin: 0; color: var(--text-secondary); line-height: 1.6;">
          Are you sure you want to delete 
          <strong>{confirmDeleteAction.type === 'bulk' ? `${selectedRows.length} selected rows` : 'this row'}</strong>? 
          This action is permanent and cannot be reversed.
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => confirmDeleteAction = null}>Cancel</button>
        <button class="btn btn-primary" style="background: #ff5252; border-color: #ff5252;" 
                onclick={() => {
                  if (confirmDeleteAction.type === 'bulk') executeBulkDelete()
                  else if (confirmDeleteAction.row) executeSingleDelete(confirmDeleteAction.row)
                }}
        >
          Confirm Delete
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showCreateModal}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 600px;">
      <div class="modal-header"><h3>Create a new table</h3><button class="btn-close" onclick={() => showCreateModal = false}>x</button></div>
      <div class="modal-body">
        <div class="form-group mb-4"><label for="tableName">Name</label><input type="text" id="tableName" bind:value={newTableName} class="input" placeholder="e.g. profiles" /></div>
        <div class="mb-2" style="font-size: 13px; font-weight: 600;">Columns</div>
        <div style="display: flex; flex-direction: column; gap: 8px;">
          {#each newTableColumns as column, i}
            <div style="display: grid; grid-template-columns: 1fr 1fr 80px 40px; gap: 8px; align-items: center;">
              <input type="text" bind:value={column.name} class="input input-sm" placeholder="Column name" />
              <select bind:value={column.type} class="input input-sm">
                <option value="uuid">uuid</option>
                <option value="text">text</option>
                <option value="int8">bigint</option>
                <option value="bool">boolean</option>
                <option value="timestamp with time zone">timestamptz</option>
                <option value="jsonb">jsonb</option>
              </select>
              <label style="font-size: 11px; display: flex; align-items: center; gap: 4px;"><input type="checkbox" bind:checked={column.is_primary} /> PK</label>
              <button class="btn btn-secondary btn-sm" style="color: var(--status-error);" onclick={() => newTableColumns = newTableColumns.filter((_, idx) => idx !== i)}>x</button>
            </div>
          {/each}
          <button class="btn btn-secondary btn-sm" style="align-self: flex-start; margin-top: 4px;" onclick={addColumn}>+ Add Column</button>
        </div>
      </div>
      <div class="modal-footer"><button class="btn btn-secondary" onclick={() => showCreateModal = false}>Cancel</button><button class="btn btn-primary" onclick={createTable} disabled={isCreating}>{isCreating ? 'Creating...' : 'Create Table'}</button></div>
    </div>
  </div>
{/if}

{#if showDeleteTableModal && selectedTableMeta}
  <div class="modal-overlay">
    <div class="modal-card" style="width: 480px; border-top: 4px solid #ff5252;">
      <div class="modal-header"><h3 style="color: #ff5252;">Delete Table {selectedTableMeta.schema}.{selectedTableMeta.name}?</h3><button class="btn-close" onclick={() => showDeleteTableModal = false}>x</button></div>
      <div class="modal-body"><p style="margin: 0; color: var(--text-secondary); line-height: 1.5;">This drops the table entirely, destroying all data within it. This action cannot be undone.</p></div>
      <div class="modal-footer"><button class="btn btn-secondary" onclick={() => showDeleteTableModal = false}>Cancel</button><button class="btn btn-primary" style="background: #ff5252; border-color: #ff5252;" onclick={confirmDeleteTable} disabled={tableActionLoading}>{tableActionLoading ? 'Deleting...' : 'Delete Table'}</button></div>
    </div>
  </div>
{/if}

<style>
  .table-editor-shell { display: grid; grid-template-columns: 280px 1fr; height: calc(100vh - var(--topbar-height)); overflow: hidden; }
  .table-sidebar { background: #16161a; border-right: 1px solid var(--border-subtle); display: flex; flex-direction: column; overflow: hidden; }
  .table-sidebar-header { padding: 16px; border-bottom: 1px solid var(--border-subtle); }
  .table-sidebar-body { overflow-y: auto; flex: 1; padding: 10px; }
  .sidebar-title { font-size: 12px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 8px; }
  
  .sb-badge-unrestricted { font-size: 9px; padding: 2px 5px; color: #ff5252; background: rgba(255,82,82,0.1); border: 1px solid rgba(255,82,82,0.3); border-radius: 4px; font-weight: 700; letter-spacing: 0.03em;}
  
  .nav-item { padding: 8px 10px; border-radius: 6px; }
  .nav-item.active { background: #26262a; }
  .table-icon { font-family: monospace; font-size: 12px; opacity: 0.5; }
  .table-name { font-size: 13px; font-family: var(--font-sans); color: #dedede; }

  .table-content { display: flex; flex-direction: column; overflow: hidden; background: #0b0b0e; }
  .table-header { padding: 16px 24px; border-bottom: 1px solid var(--border-subtle); background: #121217; }
  .table-page-content { padding: 24px; overflow: auto; flex: 1; }

  .sb-table {
    border: 1px solid #2a2a30;
    border-radius: 8px;
    background: #15151a;
    box-shadow: 0 4px 20px rgba(0,0,0,0.2);
    display: flex;
    flex-direction: column;
  }
  .sb-table-topbar { background: #1a1a20; padding: 8px 12px; border-bottom: 1px solid #2a2a30; }

  .sb-table table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
    font-family: var(--font-mono);
  }

  .sb-table th {
    background: #1a1a20;
    color: var(--text-muted);
    font-size: 12px;
    font-weight: 600;
    text-transform: none;
    padding: 10px 14px;
    border-bottom: 1px solid #2a2a30;
    border-right: 1px solid #2a2a30;
    text-align: left;
    white-space: nowrap;
  }

  .sb-table td {
    padding: 0;
    border-bottom: 1px solid #2a2a30;
    border-right: 1px solid #2a2a30;
    color: #dedede;
    height: 38px;
    position: relative;
  }
  
  .sb-table tr:hover { background: rgba(255,255,255,0.02); }

  .row-index {
     background: #1a1a20;
     color: var(--text-muted);
     text-align: center;
     font-size: 11px;
     width: 40px;
  }

  /* Spreadsheet Cells */
  .cell-display {
    padding: 8px 14px;
    height: 100%;
    display: flex;
    align-items: center;
    width: 100%;
  }

  .cell-interactive:hover .cell-display {
    background: rgba(255,255,255,0.03);
  }

  .cell-editing {
    background: #1e1e24;
    box-shadow: inset 0 0 0 1.5px var(--brand-primary);
  }

  .spreadsheet-input {
    width: 100%;
    height: 38px;
    padding: 8px 14px;
    background: transparent;
    border: none;
    color: #fff;
    font-family: var(--font-mono);
    font-size: 13px;
    outline: none;
  }

  .column-meta { color: var(--text-muted); font-size: 10px; margin-left: 6px; }
  .null-text { color: var(--text-muted); font-style: italic; opacity: 0.6; }
  
  /* Dropdowns */
  .dropdown-container { position: relative; display: inline-block; }
  .dropdown-menu-container {
     position: absolute;
     top: calc(100% + 4px);
     min-width: 180px;
     background: #1e1e24;
     border: 1px solid #333;
     border-radius: 8px;
     padding: 6px;
     box-shadow: 0 8px 30px rgba(0,0,0,0.5);
     z-index: 100;
  }
  .dropdown-item {
     display: block;
     width: 100%;
     text-align: left;
     padding: 8px 12px;
     background: transparent;
     border: none;
     color: #dedede;
     font-size: 13px;
     border-radius: 4px;
     cursor: pointer;
     font-family: var(--font-sans);
  }
  .dropdown-item:hover { background: #2a2a32; }
  .dropdown-item.danger { color: #ff5252; }
  .dropdown-item.danger:hover { background: rgba(255,82,82,0.1); }
  .dropdown-divider { height: 1px; background: #333; margin: 4px 0; }
  
  .icon-button {
     background: transparent;
     border: none;
     color: var(--text-muted);
     cursor: pointer;
     padding: 4px 8px;
     border-radius: 4px;
  }
  .icon-button:hover { background: rgba(255,255,255,0.1); color: #fff; }

  /* Modals */
  .modal-overlay { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.7); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; }
  .modal-card { background: var(--bg-surface); border: 1px solid var(--border-subtle); border-radius: var(--radius-lg); box-shadow: var(--shadow-xl); display: flex; flex-direction: column; max-height: 90vh; }
  .modal-header { padding: 16px 20px; border-bottom: 1px solid var(--border-subtle); display: flex; justify-content: space-between; align-items: center; }
  .modal-header h3 { font-size: 16px; margin: 0; }
  .modal-body { padding: 20px; overflow-y: auto; }
  .modal-footer { padding: 16px 20px; border-top: 1px solid var(--border-subtle); display: flex; justify-content: flex-end; gap: 12px; }
  .btn-close { background: none; border: none; color: var(--text-muted); font-size: 24px; cursor: pointer; }
</style>
