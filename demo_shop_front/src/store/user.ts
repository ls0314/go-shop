import { defineStore } from 'pinia'

interface UserState {
    access_token: string
    refresh_token: string
}

export const useUserStore = defineStore('user', {
    state: (): UserState => ({
        access_token: '',
        refresh_token: ''
    }),
    getters: {
        isLogin: (state) => !!state.access_token
    },

    actions: {
        setTokens(accessToken: string, refreshToken: string) {
            this.access_token = accessToken
            this.refresh_token = refreshToken
        },

        setAccessToken(token: string) {
            this.access_token = token
        },

        clearTokens() {
            this.access_token = ''
            this.refresh_token = ''
        }
    },

    persist: {
        key: 'user',
        storage: sessionStorage,
        paths: ['access_token', 'refresh_token']
    } as any
})