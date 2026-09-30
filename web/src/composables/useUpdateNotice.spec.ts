// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, it, expect } from 'vitest'
import { createUpdateNotice } from './useUpdateNotice'

function setup(versions: string[]) {
  let t = 0
  const queue = [...versions]
  const notice = createUpdateNotice(async () => queue.shift() ?? versions[versions.length - 1], () => t)
  return { notice, advance: (ms: number) => { t += ms } }
}

describe('update notice', () => {
  it('flags a version change after the page loaded', async () => {
    const { notice, advance } = setup(['1.10.10', '1.10.10', '1.11.0'])
    await notice.check()
    advance(60_000)
    await notice.check()
    expect(notice.available.value).toBe(false)
    advance(60_000)
    await notice.check()
    expect(notice.available.value).toBe(true)
  })

  it('checks at most once a minute', async () => {
    const { notice, advance } = setup(['1.10.10', '1.11.0'])
    await notice.check()
    advance(1_000)
    await notice.check()
    expect(notice.available.value).toBe(false)
  })

  it('never flags a dev build', async () => {
    const { notice, advance } = setup(['dev', 'dev'])
    await notice.check()
    advance(60_000)
    await notice.check()
    expect(notice.available.value).toBe(false)
  })

  it('ignores a failed check', async () => {
    let t = 0
    let calls = 0
    const notice = createUpdateNotice(async () => {
      calls++
      if (calls === 2) throw new Error('restarting')
      return calls === 1 ? '1.10.10' : '1.10.10'
    }, () => t)
    await notice.check()
    t += 60_000
    await notice.check()
    t += 60_000
    await notice.check()
    expect(notice.available.value).toBe(false)
  })
})
