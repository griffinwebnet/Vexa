import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  username: string | null
  isAdmin: boolean
  isAuthenticated: boolean
  login: (token: string, username: string, isAdmin: boolean) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      username: null,
      isAdmin: false,
      isAuthenticated: false,
      login: (token, username, isAdmin) =>
        set({ token, username, isAdmin, isAuthenticated: true }),
      logout: () =>
        set({ token: null, username: null, isAdmin: false, isAuthenticated: false }),
    }),
    {
      name: 'vexa-auth',
    }
  )
)

