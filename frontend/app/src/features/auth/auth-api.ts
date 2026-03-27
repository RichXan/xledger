import { requestEnvelope, requestRaw } from '@/lib/api'

export interface SendCodeResponse {
  code_sent: boolean
}

export interface VerifyCodeResponse {
  access_token: string
  refresh_token: string
}

export interface CurrentUserResponse {
  email: string
  name?: string
}

export interface RefreshResponse {
  access_token: string
  refresh_token: string
}

export interface LogoutResponse {
  logged_out: boolean
}

export interface PasswordAuthResponse {
  access_token: string
  refresh_token: string
}

export interface ChangePasswordResponse {
  changed: boolean
}

export function sendCode(email: string) {
  return requestEnvelope<SendCodeResponse>('/auth/send-code', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export function verifyCode(email: string, code: string) {
  return requestEnvelope<VerifyCodeResponse>('/auth/verify-code', {
    method: 'POST',
    body: JSON.stringify({ email, code }),
  })
}

export function registerWithPassword(input: { email: string; password: string; displayName?: string }) {
  return requestEnvelope<PasswordAuthResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({
      email: input.email,
      password: input.password,
      display_name: input.displayName ?? '',
    }),
  })
}

export function loginWithPassword(email: string, password: string) {
  return requestEnvelope<PasswordAuthResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function getCurrentUser(accessToken: string) {
  return requestEnvelope<CurrentUserResponse>('/auth/me', {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

export function refreshSession(refreshToken: string) {
  return requestRaw<RefreshResponse>('/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export function logout(refreshToken: string) {
  return requestRaw<LogoutResponse>('/auth/logout', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${refreshToken}`,
    },
  })
}

export function updateProfile(accessToken: string, displayName: string) {
  return requestEnvelope<CurrentUserResponse>('/auth/profile', {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ display_name: displayName }),
  })
}

export function changePassword(accessToken: string, oldPassword: string, newPassword: string) {
  return requestEnvelope<ChangePasswordResponse>('/auth/change-password', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({
      old_password: oldPassword,
      new_password: newPassword,
    }),
  })
}
