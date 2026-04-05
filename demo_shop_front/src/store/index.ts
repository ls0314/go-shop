import { createPinia } from 'pinia'
import persistedState from 'pinia-plugin-persistedstate'
import { useUserStore } from './user'

const pinia = createPinia()

pinia.use(persistedState)

export default pinia

export let getUserStore: typeof useUserStore

export function initStores() {
    getUserStore = () => useUserStore(pinia)
}

