export interface WebAppSettings {
  deviceAutoRefreshSec: number
}

const SETTINGS_KEY = 'rvcs_web_settings'
const DEFAULT_SETTINGS: WebAppSettings = {
  deviceAutoRefreshSec: 15
}

function normalizeRefreshSec(value: unknown): number {
  const n = Number(value)
  if (!Number.isFinite(n)) return DEFAULT_SETTINGS.deviceAutoRefreshSec
  return Math.min(120, Math.max(5, Math.round(n)))
}

export function getWebSettings(): WebAppSettings {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (!raw) return { ...DEFAULT_SETTINGS }
    const parsed = JSON.parse(raw)
    return {
      deviceAutoRefreshSec: normalizeRefreshSec(parsed?.deviceAutoRefreshSec)
    }
  } catch {
    return { ...DEFAULT_SETTINGS }
  }
}

export function saveWebSettings(settings: Partial<WebAppSettings>): WebAppSettings {
  const merged: WebAppSettings = {
    ...getWebSettings(),
    ...settings
  }
  merged.deviceAutoRefreshSec = normalizeRefreshSec(merged.deviceAutoRefreshSec)
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(merged))
  return merged
}

export function resetWebSettings(): WebAppSettings {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(DEFAULT_SETTINGS))
  return { ...DEFAULT_SETTINGS }
}

export function getDeviceAutoRefreshIntervalMs(): number {
  return getWebSettings().deviceAutoRefreshSec * 1000
}


