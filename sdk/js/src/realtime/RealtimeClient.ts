import type { AuthClient } from '../auth/AuthClient'
import type { RealtimeEventType, RealtimePayload, PresenceState } from '../types'

interface RealtimeClientOptions {
  url: string
  headers: Record<string, string>
  auth: AuthClient
}

type ChannelEventCallback<T> = (payload: RealtimePayload<T>) => void
type PresenceCallback = (state: PresenceState) => void
type BroadcastCallback = (payload: Record<string, unknown>) => void

export type ChannelState = 'SUBSCRIBED' | 'UNSUBSCRIBED' | 'CLOSED' | 'CHANNEL_ERROR'

/**
 * RealtimeClient manages WebSocket connections to OmniBase's realtime service.
 *
 * @example
 * ```typescript
 * // Subscribe to all changes on the 'posts' table
 * const channel = omni.channel('posts-changes')
 *   .on('INSERT', (payload) => console.log('New post:', payload.new))
 *   .on('UPDATE', (payload) => console.log('Updated:', payload.new))
 *   .on('DELETE', (payload) => console.log('Deleted:', payload.old))
 *   .subscribe()
 *
 * // Filtered subscription
 * const channel = omni.channel('my-posts')
 *   .on('*', { filter: 'user_id=eq.123' }, (payload) => console.log(payload))
 *   .subscribe()
 *
 * // Presence (see who's online)
 * const channel = omni.channel('room-1')
 *   .on('presence', { event: 'sync' }, (state) => console.log('Online users:', state))
 *   .subscribe(async (status) => {
 *     if (status === 'SUBSCRIBED') {
 *       await channel.track({ userId: 'me', cursor: { x: 100, y: 200 } })
 *     }
 *   })
 * ```
 */
export class RealtimeClient {
  private wsUrl: string
  private headers: Record<string, string>
  private auth: AuthClient
  private ws: WebSocket | null = null
  private channels: Map<string, Channel> = new Map()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null

  constructor(options: RealtimeClientOptions) {
    this.wsUrl = options.url
    this.headers = options.headers
    this.auth = options.auth
  }

  /**
   * Get or create a channel for subscriptions.
   */
  channel(name: string): Channel {
    if (!this.channels.has(name)) {
      this.channels.set(name, new Channel(name, this))
    }
    return this.channels.get(name)!
  }

  /**
   * Remove and unsubscribe a specific channel.
   */
  removeChannel(channel: Channel): void {
    channel._unsubscribe()
    this.channels.delete(channel.name)
  }

  /**
   * Remove all channels and close the WebSocket connection.
   */
  removeAllChannels(): void {
    this.channels.forEach((ch) => ch._unsubscribe())
    this.channels.clear()
    this._disconnect()
  }

  _connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return

    const token = this.auth.getAccessToken()
    const url = token ? `${this.wsUrl}?apikey=${token}` : this.wsUrl

    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      this.reconnectAttempts = 0
      // Resubscribe all channels after reconnect
      this.channels.forEach((ch) => ch._sendSubscription())
    }

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        this._routeMessage(msg)
      } catch {}
    }

    this.ws.onclose = () => {
      this._scheduleReconnect()
    }

    this.ws.onerror = () => {
      // Error handling — reconnect will be triggered by onclose
    }
  }

  _disconnect(): void {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.ws?.close()
    this.ws = null
  }

  _send(msg: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
    }
  }

  private _routeMessage(msg: Record<string, unknown>): void {
    const channelName = msg.channel as string
    if (channelName && this.channels.has(channelName)) {
      this.channels.get(channelName)!._handleMessage(msg)
    }

    // Broadcast to all channels for table-level events
    if (msg.type === 'broadcast') {
      this.channels.forEach((ch) => ch._handleMessage(msg))
    }
  }

  private _scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return

    const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 30000)
    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++
      this._connect()
    }, delay)
  }
}

/**
 * Channel represents a realtime subscription channel.
 */
export class Channel {
  readonly name: string
  private _client: RealtimeClient
  private _state: ChannelState = 'UNSUBSCRIBED'
  private _listeners: Array<{
    type: 'db' | 'presence' | 'broadcast'
    event: string
    table?: string
    schema?: string
    filter?: string
    callback: ChannelEventCallback<unknown> | PresenceCallback | BroadcastCallback
  }> = []

  constructor(name: string, client: RealtimeClient) {
    this.name = name
    this._client = client
  }

  /**
   * Subscribe to database change events on a table.
   */
  on<T = Record<string, unknown>>(
    event: RealtimeEventType,
    options: { schema?: string; table?: string; filter?: string },
    callback: ChannelEventCallback<T>
  ): this
  on<T = Record<string, unknown>>(
    event: RealtimeEventType,
    callback: ChannelEventCallback<T>
  ): this
  on<T = Record<string, unknown>>(
    event: RealtimeEventType,
    optionsOrCallback:
      | { schema?: string; table?: string; filter?: string }
      | ChannelEventCallback<T>,
    callback?: ChannelEventCallback<T>
  ): this {
    if (typeof optionsOrCallback === 'function') {
      this._listeners.push({
        type: 'db',
        event: event.toUpperCase(),
        callback: optionsOrCallback as ChannelEventCallback<unknown>,
      })
    } else {
      this._listeners.push({
        type: 'db',
        event: event.toUpperCase(),
        table: optionsOrCallback.table,
        schema: optionsOrCallback.schema ?? 'public',
        filter: optionsOrCallback.filter,
        callback: callback! as ChannelEventCallback<unknown>,
      })
    }
    return this
  }

  /**
   * Subscribe to presence events (who's online in this channel).
   */
  presence(event: 'sync' | 'join' | 'leave', callback: PresenceCallback): this {
    this._listeners.push({ type: 'presence', event, callback })
    return this
  }

  /**
   * Subscribe to broadcast messages in this channel.
   */
  broadcast(event: string, callback: BroadcastCallback): this {
    this._listeners.push({ type: 'broadcast', event, callback })
    return this
  }

  /**
   * Start the subscription. Call after chaining .on() handlers.
   */
  subscribe(callback?: (status: ChannelState) => void): this {
    this._client._connect()
    this._sendSubscription()

    if (callback) {
      setTimeout(() => callback('SUBSCRIBED'), 500)
    }
    this._state = 'SUBSCRIBED'
    return this
  }

  /**
   * Track presence for the current user in this channel.
   */
  async track(state: Record<string, unknown>): Promise<void> {
    this._client._send({
      type: 'presence',
      channel: this.name,
      event: 'track',
      payload: state,
    })
  }

  /**
   * Broadcast an ephemeral message to all clients in this channel.
   */
  async send(type: string, event: string, payload: Record<string, unknown>): Promise<void> {
    this._client._send({
      type: 'broadcast',
      channel: this.name,
      event,
      payload,
    })
  }

  /**
   * Unsubscribe from this channel.
   */
  unsubscribe(): this {
    this._unsubscribe()
    return this
  }

  _sendSubscription(): void {
    this._client._send({
      type: 'subscribe',
      channel: this.name,
      config: {
        broadcast: this._listeners.some((l) => l.type === 'broadcast'),
        presence: this._listeners.some((l) => l.type === 'presence'),
        postgres_changes: this._listeners
          .filter((l) => l.type === 'db')
          .map((l) => ({
            event: l.event,
            schema: l.schema ?? 'public',
            table: l.table ?? '*',
            filter: l.filter,
          })),
      },
    })
  }

  _unsubscribe(): void {
    this._state = 'UNSUBSCRIBED'
    this._client._send({ type: 'unsubscribe', channel: this.name })
  }

  _handleMessage(msg: Record<string, unknown>): void {
    const eventType = String(msg.event ?? msg.eventType ?? '').toUpperCase()
    const schema = String(msg.schema ?? 'public')
    const table = String(msg.table ?? '')

    for (const listener of this._listeners) {
      if (listener.type === 'db') {
        const eventMatch = listener.event === '*' || listener.event === eventType
        const tableMatch = !listener.table || listener.table === '*' || listener.table === table
        const schemaMatch = !listener.schema || listener.schema === schema
        if (eventMatch && tableMatch && schemaMatch) {
          ;(listener.callback as ChannelEventCallback<unknown>)(msg.payload as RealtimePayload)
        }
      } else if (listener.type === 'broadcast' && msg.type === 'broadcast') {
        if (listener.event === '*' || listener.event === String(msg.event)) {
          ;(listener.callback as BroadcastCallback)(msg.payload as Record<string, unknown>)
        }
      } else if (listener.type === 'presence' && msg.type === 'presence') {
        ;(listener.callback as PresenceCallback)(msg.state as PresenceState)
      }
    }
  }
}
