// frontend/app/src/features/notifications/preference-store.ts
import { requestEnvelope } from '@/lib/api'

export interface NotificationPreferences {
  realtime_alert: boolean
  daily_digest: boolean
  weekly_digest: boolean
}

export interface PreferenceStore {
  getPrefs(): Promise<NotificationPreferences>
  updatePrefs(prefs: Partial<NotificationPreferences>): Promise<void>
}

export function createPreferenceStore(accessToken: string): PreferenceStore {
  return {
    async getPrefs() {
      const data = await requestEnvelope<{ prefs: NotificationPreferences }>('/push/preferences', {
        headers: { Authorization: `Bearer ${accessToken}` },
      })
      return data.prefs
    },
    async updatePrefs(prefs: Partial<NotificationPreferences>) {
      await requestEnvelope('/push/preferences', {
        method: 'PATCH',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(prefs),
      })
    },
  }
}