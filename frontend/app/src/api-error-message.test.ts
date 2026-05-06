import { ApiError, getFriendlyApiErrorMessage } from './lib/api'

describe('friendly API error messages', () => {
  const t = (key: string, options?: { defaultValue?: string }) => {
    const messages: Record<string, string> = {
      'errors.authExpired': 'Your session expired. Please sign in again.',
      'errors.businessRule': 'Check the highlighted fields and try again.',
      'errors.network': 'Network connection failed. Please retry.',
      'errors.unknown': 'Something went wrong.',
    }
    return messages[key] ?? options?.defaultValue ?? key
  }

  it('maps API codes to consistent user-facing messages', () => {
    expect(getFriendlyApiErrorMessage(new ApiError('业务规则不满足', 409, 'BUSINESS_RULE_VIOLATION'), t)).toBe(
      'Check the highlighted fields and try again.',
    )
    expect(getFriendlyApiErrorMessage(new ApiError('未认证或凭证无效', 401, 'AUTH_UNAUTHORIZED'), t)).toBe(
      'Your session expired. Please sign in again.',
    )
    expect(getFriendlyApiErrorMessage(new TypeError('Failed to fetch'), t)).toBe(
      'Network connection failed. Please retry.',
    )
  })
})
