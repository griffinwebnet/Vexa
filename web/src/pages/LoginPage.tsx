import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import api from '../lib/api'

// Get version from package.json
const VERSION = '0.2.102'

export default function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  // Check domain status on mount to show appropriate message
  useEffect(() => {
    const checkDomainStatus = async () => {
      try {
        const response = await api.get('/domain/status')
        const status = response.data
        
        // If no domain is provisioned, show setup message
        if (!status.provisioned) {
          console.log('No domain provisioned - user needs to authenticate for setup')
        }
      } catch (err) {
        console.error('Failed to check domain status:', err)
      }
    }
    
    checkDomainStatus()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    console.log('=== LOGIN ATTEMPT ===')
    console.log('Username:', username)

    try {
      const response = await api.post('/auth/login', {
        username,
        password,
      })

      console.log('Login response:', response.data)

      const { token, username: user, is_admin, is_domain_user } = response.data
      console.log('Extracted values:', { user, is_admin, is_domain_user })
      login(token, user, is_admin, is_domain_user)
      
      // Debug: Check what was stored in auth store
      console.log('Auth store after login:', {
        isAuthenticated: true,
        username: user,
        isAdmin: is_admin,
        isDomainUser: is_domain_user
      })

      // Redirect logic based on user type and domain status
      if (!is_domain_user && is_admin) {
        // Local admin - check if domain exists to determine where to go
        try {
          const statusResponse = await api.get('/domain/status')
          const status = statusResponse.data
          
          if (!status.provisioned) {
            console.log('Local admin authenticated, no domain exists - redirecting to wizard')
            navigate('/wizard')
          } else {
            console.log('Local admin authenticated, domain exists - redirecting to dashboard')
            navigate('/')
          }
        } catch (err) {
          console.error('Failed to check domain status after login:', err)
          navigate('/')
        }
      } else if (is_domain_user) {
        // Domain user - go to dashboard (will redirect to self-service if not admin)
        console.log('Domain user authenticated, redirecting to dashboard')
        navigate('/')
      } else {
        // This shouldn't happen based on our auth logic, but just in case
        console.log('Authenticated user, redirecting to dashboard')
        navigate('/')
      }
    } catch (err: any) {
      console.error('=== LOGIN FAILED ===')
      console.error('Error:', err.response?.data?.error)
      
      setError(err.response?.data?.error || 'Login failed. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-gradient-to-br from-primary/10 via-background to-background">
      {/* Version number in bottom right corner */}
      <div className="absolute bottom-4 right-4 text-xs text-muted-foreground/60 font-mono">
        v{VERSION}
      </div>
      
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-3xl font-bold text-center bg-gradient-to-r from-yellow-400 to-amber-600 bg-clip-text text-transparent">
            Vexa
          </CardTitle>
          <CardDescription className="text-center">
            Sign in to your directory management console
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="username" className="text-sm font-medium">
                Username
              </label>
              <Input
                id="username"
                type="text"
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                autoComplete="username"
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium">
                Password
              </label>
              <Input
                id="password"
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
              />
            </div>
            {error && (
              <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
                {error}
              </div>
            )}
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? 'Signing in...' : 'Sign In'}
            </Button>
          </form>
          <div className="mt-4 text-center text-sm text-muted-foreground">
            Authenticate with your Linux PAM or directory credentials
          </div>
        </CardContent>
      </Card>
    </div>
  )
}