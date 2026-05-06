import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('pwa service worker cache policy', () => {
  it('only caches explicit safe API responses and never caches broad transaction URLs', () => {
    const sw = readFileSync(resolve(__dirname, '../public/sw.js'), 'utf8')

    expect(sw).toContain('function isCacheableAPIRequest')
    expect(sw).toContain('response.status === 200')
    expect(sw).toContain('!request.headers.has(\'Authorization\')')
    expect(sw).not.toContain("'/api/transactions':")
  })
})
