import { createPinia } from 'pinia'
import persistedState from 'pinia-plugin-persistedstate'
import { useUserStore } from './modules/user'

const pinia = createPinia()

pinia.use(persistedState)

export default pinia

export const getUserStore = () => {
    return useUserStore(pinia)
}