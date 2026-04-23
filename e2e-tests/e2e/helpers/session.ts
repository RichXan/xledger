import type { Page } from '@playwright/test'
import { XledgerApiClient } from './api-client'

export interface AuthSessionPayload {
  accessToken: string
  refreshToken: string
  email: string
  name: string | null
}

export interface PreparedSession {
  email: string
  password: string
  authSession: AuthSessionPayload
}

export async function prepareSession(apiClient: XledgerApiClient, runTag: string): Promise<PreparedSession> {
  const sanitizedTag = runTag.replace(/[^a-zA-Z0-9]/g, '').toLowerCase().slice(0, 24)
  const nonce = `${Date.now()}-${Math.floor(Math.random() * 10_000)}`
  const email = `e2e.${sanitizedTag}.${nonce}@example.com`
  const password = 'E2E-pass-1234'
  const displayName = `E2E ${sanitizedTag || 'runner'}`

  const tokens = await apiClient.registerOrLogin({
    email,
    password,
    displayName,
  })
  const me = await apiClient.getMe(tokens.access_token)

  return {
    email,
    password,
    authSession: {
      accessToken: tokens.access_token,
      refreshToken: tokens.refresh_token,
      email: me.email,
      name: me.name ?? null,
    },
  }
}

export async function injectSession(page: Page, session: AuthSessionPayload) {
  await page.addInitScript((payload) => {
    window.localStorage.setItem('xledger.auth', JSON.stringify(payload))
    window.localStorage.setItem('i18nextLng', 'en')
    window.localStorage.setItem('pwa-onboarding-dismissed', '1')
  }, session)
}
