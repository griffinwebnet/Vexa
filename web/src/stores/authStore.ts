import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  username: string | null
  isAdmin: boolean
  isDomainUser: boolean
  isAuthenticated: boolean
  login: (token: string, username: string, isAdmin: boolean, isDomainUser: boolean) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      username: null,
      isAdmin: false,
      isDomainUser: false,
      isAuthenticated: false,
      login: (token, username, isAdmin, isDomainUser) =>
        set({ token, username, isAdmin, isDomainUser, isAuthenticated: true }),
      logout: () => {
        set({ token: null, username: null, isAdmin: false, isDomainUser: false, isAuthenticated: false })
        // Force redirect to login page
        window.location.href = '/login'
      },
    }),
    {
      name: 'vexa-auth',
    }
  )
)

