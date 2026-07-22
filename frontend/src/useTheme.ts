import { computed, readonly, shallowRef } from 'vue'

export type ThemePreference = 'system' | 'light' | 'dark'

const stored = localStorage.getItem('freebooru.theme')
const preference = shallowRef<ThemePreference>(isTheme(stored) ? stored : 'system')
const media = typeof matchMedia === 'function'
  ? matchMedia('(prefers-color-scheme: dark)')
  : undefined
const systemDark = shallowRef(media?.matches ?? false)

function applyTheme() {
  const dark = preference.value === 'dark' || (preference.value === 'system' && systemDark.value)
  document.documentElement.dataset.theme = dark ? 'dark' : 'light'
}

media?.addEventListener('change', (event) => {
  systemDark.value = event.matches
  if (preference.value === 'system') applyTheme()
})
applyTheme()

export function useTheme() {
  const resolved = computed(() => preference.value === 'system'
    ? systemDark.value ? 'dark' : 'light'
    : preference.value)

  function select(value: ThemePreference) {
    preference.value = value
    localStorage.setItem('freebooru.theme', value)
    applyTheme()
  }

  return { preference: readonly(preference), resolved, select }
}

function isTheme(value: string | null): value is ThemePreference {
  return value === 'system' || value === 'light' || value === 'dark'
}
