import { readonly, shallowRef } from 'vue'

import { getStatus, type ApplicationStatus } from './api'

export type LoadState = 'loading' | 'ready' | 'error'

export function useApplicationStatus() {
  const state = shallowRef<LoadState>('loading')
  const status = shallowRef<ApplicationStatus>()
  const errorMessage = shallowRef('')

  async function reload() {
    state.value = 'loading'
    errorMessage.value = ''

    try {
      status.value = await getStatus()
      state.value = 'ready'
    } catch (error) {
      status.value = undefined
      errorMessage.value = error instanceof Error ? error.message : 'Unable to reach FreeBooru'
      state.value = 'error'
    }
  }

  return {
    state: readonly(state),
    status: readonly(status),
    errorMessage: readonly(errorMessage),
    reload,
  }
}
