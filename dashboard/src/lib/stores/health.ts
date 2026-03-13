import { writable } from 'svelte/store'

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

const SERVICES: Pick<ServiceStatus, 'name' | 'url'>[] = [
  { name: 'API Gateway', url: 'http://localhost:8000' },
  { name: 'Auth Service', url: 'http://localhost:9001' },
  { name: 'Storage Service', url: 'http://localhost:9003' },
  { name: 'Realtime Service', url: 'http://localhost:9004' },
  { name: 'PostgREST', url: 'http://localhost:3000' },
]

function createHealthStore() {
  const { subscribe, set, update } = writable<HealthState>({
    services: SERVICES.map(s => ({ ...s, status: 'unknown' })),
    lastChecked: null,
    loading: false,
  })

  async function checkService(service: Pick<ServiceStatus, 'name' | 'url'>): Promise<ServiceStatus> {
    const start = Date.now()
    try {
      const resp = await fetch(`${service.url}/health`, {
        signal: AbortSignal.timeout(5000),
      })
      const latency = Date.now() - start
      return {
        ...service,
        status: resp.ok ? 'healthy' : 'degraded',
        latency,
      }
    } catch {
      return { ...service, status: 'down' }
    }
  }

  async function refresh() {
    update(s => ({ ...s, loading: true }))
    const results = await Promise.all(SERVICES.map(checkService))
    set({ services: results, lastChecked: new Date(), loading: false })
  }

  return { subscribe, refresh }
}

export const serviceHealth = createHealthStore()
