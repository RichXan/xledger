// frontend/app/src/features/offline/offline-queue.ts
import Dexie from 'dexie'

export interface QueuedRequest {
  id?: number
  method: string
  url: string
  body?: string
  timestamp: number
}

class OfflineQueueDB extends Dexie {
  queue!: Dexie.Table<QueuedRequest, number>

  constructor() {
    super('xledger-offline-queue')
    this.version(1).stores({
      queue: '++id, method, url, timestamp',
    })
  }
}

const queueDb = new OfflineQueueDB()

export async function enqueueRequest(request: Omit<QueuedRequest, 'id'>): Promise<void> {
  await queueDb.queue.add(request as QueuedRequest)
}

export async function getQueuedRequests(): Promise<QueuedRequest[]> {
  return queueDb.queue.toArray()
}

export async function dequeueRequest(id: number): Promise<void> {
  await queueDb.queue.delete(id)
}

export async function clearQueue(): Promise<void> {
  await queueDb.queue.clear()
}

export async function processQueue(): Promise<void> {
  const requests = await getQueuedRequests()

  for (const request of requests) {
    try {
      const options: RequestInit = {
        method: request.method,
        headers: {
          'Content-Type': 'application/json',
        },
      }
      if (request.body) {
        options.body = request.body
      }

      const response = await fetch(request.url, options)
      if (response.ok && request.id) {
        await dequeueRequest(request.id)
      }
    } catch {
      // Network error, leave in queue for retry
      break
    }
  }
}