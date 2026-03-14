import { writable } from 'svelte/store'
import { getOmniBaseUrl } from '$lib/api'

export interface ServiceStatus {
  name: string
  url: string
  status: 'healthy' | 'degraded' | 'down' | 'unknown'
  latency?: number
}

interface HealthState {
  services: ServiceStatus[]
  lastChecked: Date | null
  loading: boolean
}

// Default initial state — these will be populated from gateway /health/services
const DEFAULT_SERVICES: ServiceStatus[] = [
  { name: 'API Gateway', url: 'http://localhost:8000', status: 'unknown' },
  { name: 'Auth Service', url: 'internal:9001', status: 'unknown' },
  { name: 'Storage Service', url: 'internal:9003', status: 'unknown' },
  { name: 'Realtime Service', url: 'internal:9004', status: 'unknown' },
  { name: 'Functions Service', url: 'internal:9006', status: 'unknown' },
  { name: 'Database Service', url: 'internal:9005', status: 'unknown' },
  { name: 'PostgREST', url: 'internal:3000', status: 'unknown' },
]

function createHealthStore() {
  const { subscribe, set, update } = writable<HealthState>({
    services: DEFAULT_SERVICES,
    lastChecked: null,
    loading: false,
  })

  // Check gateway itself (can be reached directly by browser)
  async function checkGateway(omnibaseUrl: string): Promise<ServiceStatus> {
    const start = Date.now()
    try {
      const resp = await fetch(`${omnibaseUrl}/health`, {
        signal: AbortSignal.timeout(5000),
      })
      const latency = Date.now() - start
      return {
        name: 'API Gateway',
        url: omnibaseUrl,
        status: resp.ok ? 'healthy' : 'degraded',
        latency,
      }
    } catch {
      return { name: 'API Gateway', url: omnibaseUrl, status: 'down' }
    }
  }

  async function refresh() {
    update(s => ({ ...s, loading: true }))
    const omnibaseUrl = getOmniBaseUrl()

    // Check gateway directly first
    const gatewayStatus = await checkGateway(omnibaseUrl)

    // If gateway is up, use its /health/services to get internal service status
    if (gatewayStatus.status === 'healthy') {
      try {
        const resp = await fetch(`${omnibaseUrl}/health/services`, {
          signal: AbortSignal.timeout(8000),
        })
        if (resp.ok) {
          const data = await resp.json()
          const internalServices: ServiceStatus[] = (data.services || []).map((s: any) => ({
            name: s.name,
            url: s.url,
            status: s.status as ServiceStatus['status'],
            latency: s.latency_ms,
          }))
          set({
            services: [gatewayStatus, ...internalServices],
            lastChecked: new Date(),
            loading: false,
          })
          return
        }
      } catch {
        // fall through to show gateway status only
      }
    }

    // Gateway is down or /health/services failed — show what we know
    set({
      services: [
        gatewayStatus,
        ...DEFAULT_SERVICES.slice(1).map(s => ({
          ...s,
          status: (gatewayStatus.status === 'down' ? 'down' : 'unknown') as ServiceStatus['status'],
        })),
      ],
      lastChecked: new Date(),
      loading: false,
    })
  }

  return { subscribe, refresh }
}

export const serviceHealth = createHealthStore()
