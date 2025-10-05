import { useState, useRef, useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Server, GitMerge, Plus, Loader2, CheckCircle, AlertCircle } from 'lucide-react'

type SetupMode = 'new' | 'join' | 'migrate' | null

export default function SetupWizard() {
  const queryClient = useQueryClient()
  const [mode, setMode] = useState<SetupMode>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [currentStatus, setCurrentStatus] = useState('')
  const [domainName, setDomainName] = useState('')
  const [provisioningState, setProvisioningState] = useState<'idle' | 'provisioning' | 'success' | 'error'>('idle')
  
  // Ref to track if component is mounted
  const isMounted = useRef(true)
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      isMounted.current = false
    }
  }, [])
  
  // Debug provisioning state changes
  console.log('Current provisioning state:', provisioningState)
  
  const [formData, setFormData] = useState({
    realm: '',
    dnsProvider: 'cloudflare',
    customDnsServers: '',
  })

  const dnsProviders = [
    { value: 'cloudflare', label: 'Cloudflare (1.1.1.1)', servers: '1.1.1.1,1.0.0.1' },
    { value: 'google', label: 'Google (8.8.8.8)', servers: '8.8.8.8,8.8.4.4' },
    { value: 'quad9', label: 'Quad9 (9.9.9.9)', servers: '9.9.9.9,149.112.112.112' },
    { value: 'custom', label: 'Custom DNS Servers', servers: '' },
  ]

  // Auto-generate domain name from realm
  const getDomainFromRealm = (realm: string) => {
    if (!realm) return ''
    const domainPart = realm.split('.')[0]
    return domainPart.toUpperCase()
  }

  const handleSetupNew = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setProvisioningState('provisioning')
    setCurrentStatus('Starting domain provisioning...')

    if (!formData.realm.trim()) {
      setError('Realm is required')
      return
    }
    
    console.log('=== DOMAIN PROVISIONING START ===')

    setLoading(true)
    try {
      // Get DNS servers based on provider
      let dnsServers = formData.customDnsServers
      if (formData.dnsProvider !== 'custom') {
        const provider = dnsProviders.find(p => p.value === formData.dnsProvider)
        dnsServers = provider?.servers || '1.1.1.1,1.0.0.1'
      }

      const domain = getDomainFromRealm(formData.realm)
      setDomainName(domain)
      
      // Use the new streaming endpoint (no authentication required during setup)
      const response = await fetch('/api/v1/domain/provision-with-output', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          domain: domain,
          realm: formData.realm,
          dns_forwarder: dnsServers,
        })
      })
      
      console.log('Response status:', response.status, response.statusText)
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const reader = response.body?.getReader()
      const decoder = new TextDecoder()

      // Add a fallback timeout to check completion status
      const fallbackTimeout = setTimeout(async () => {
        console.log('Fallback timeout reached, checking domain status...')
        try {
          const statusResponse = await fetch('/api/v1/domain/status')
          const statusData = await statusResponse.json()
          if (statusData.provisioned) {
            console.log('Domain is provisioned, setting success state')
            setCurrentStatus('Domain provisioning completed successfully!')
            setProvisioningState('success')
            setTimeout(() => {
              localStorage.setItem('vexa-setup-complete', 'true')
              localStorage.setItem('vexa-domain-info', JSON.stringify({
                domain: domainName,
                realm: formData.realm
              }))
              queryClient.invalidateQueries({ queryKey: ['domainStatus'] })
              window.location.href = '/'
            }, 2000)
          }
        } catch (err) {
          console.log('Fallback status check failed:', err)
        }
      }, 30000) // 30 second fallback

      if (reader) {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          const chunk = decoder.decode(value)
          console.log('Received chunk:', chunk) // Console logging for debugging
          const lines = chunk.split('\n')
          
          for (const line of lines) {
            const trimmedLine = line.trim()
            if (trimmedLine.startsWith('data: ')) {
              try {
                const jsonStr = trimmedLine.slice(6).trim()
                console.log('JSON string to parse:', jsonStr)
                const data = JSON.parse(jsonStr)
                console.log('Parsed SSE data:', data) // Console logging for debugging
                
                if (data.type === 'output') {
                  const content = data.content
                  console.log('SSE Output:', content)
                  
                  // Handle errors
                  if (content.startsWith('ERROR:')) {
                    setError(content.replace('ERROR: ', ''))
                    setProvisioningState('error')
                    return
                  }
                  
                  // Update status for any meaningful content
                  console.log('Updating status to:', content)
                  setCurrentStatus(content)
                  
                  // Check for completion messages
                  if (content.includes('Domain provisioning completed') || 
                      content.includes('Samba AD DC service started successfully') ||
                      content.includes('provisioning completed successfully')) {
                    console.log('FORCING SUCCESS STATE due to completion message:', content)
                    clearTimeout(fallbackTimeout)
                    setProvisioningState('success')
                    
                    // Auto-redirect after 2 seconds
                    setTimeout(() => {
                      console.log('Auto-redirecting to dashboard from completion message')
                      localStorage.setItem('vexa-setup-complete', 'true')
                      localStorage.setItem('vexa-domain-info', JSON.stringify({
                        domain: domainName,
                        realm: formData.realm
                      }))
                      queryClient.invalidateQueries({ queryKey: ['domainStatus'] })
                      window.location.href = '/'
                    }, 2000)
                  }
                } else if (data.type === 'complete') {
                  console.log('=== COMPLETE EVENT RECEIVED ===')
                  console.log('Complete event data:', data)
                  clearTimeout(fallbackTimeout)
                  setCurrentStatus('Domain provisioning completed successfully!')
                  setProvisioningState('success')
                  
                  // Auto-redirect after 2 seconds
                  setTimeout(() => {
                    console.log('Auto-redirecting to dashboard')
                    localStorage.setItem('vexa-setup-complete', 'true')
                    localStorage.setItem('vexa-domain-info', JSON.stringify({
                      domain: domainName,
                      realm: formData.realm
                    }))
                    queryClient.invalidateQueries({ queryKey: ['domainStatus'] })
                    window.location.href = '/'
                  }, 2000)
                } else {
                  console.log('Unknown event type:', data.type, 'Data:', data)
                }
              } catch (e) {
                console.error('Failed to parse SSE data:', e)
              }
            }
          }
        }
      }
    } catch (err: any) {
      console.error('Domain provisioning error:', err)
      setError(err.message || 'Failed to provision domain')
      setProvisioningState('error')
    } finally {
      setLoading(false)
      console.log('=== DOMAIN PROVISIONING END ===')
    }
  }


  if (!mode) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/10 via-background to-background p-8">
        <div className="w-full max-w-4xl">
          <div className="text-center mb-12">
            <h1 className="text-4xl font-bold bg-gradient-to-r from-yellow-400 to-amber-600 bg-clip-text text-transparent mb-4">
              Welcome to Vexa
            </h1>
            <p className="text-muted-foreground text-lg">
              Let's set up your directory environment
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            <Card className="hover:border-primary transition-colors cursor-pointer" onClick={() => setMode('new')}>
              <CardHeader>
                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
                  <Plus className="h-6 w-6 text-primary" />
                </div>
                <CardTitle>New Domain</CardTitle>
                <CardDescription>
                  Set up a brand new Vexa Domain
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Create a fresh AD environment from scratch with Vexa as the primary domain controller.
                </p>
              </CardContent>
            </Card>

            <Card className="hover:border-primary transition-colors cursor-pointer opacity-50">
              <CardHeader>
                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
                  <Server className="h-6 w-6 text-primary" />
                </div>
                <CardTitle>Join as DC</CardTitle>
                <CardDescription>
                  Join existing domain as secondary controller
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Add Vexa as an additional domain controller to your existing AD infrastructure.
                </p>
                <p className="text-xs text-yellow-600 dark:text-yellow-500 mt-2">Coming soon</p>
              </CardContent>
            </Card>

            <Card className="hover:border-primary transition-colors cursor-pointer opacity-50">
              <CardHeader>
                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
                  <GitMerge className="h-6 w-6 text-primary" />
                </div>
                <CardTitle>Migrate Domain</CardTitle>
                <CardDescription>
                  Migrate existing AD domain to Vexa
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Transfer your existing Vexa Domain to Vexa infrastructure.
                </p>
                <p className="text-xs text-yellow-600 dark:text-yellow-500 mt-2">Coming soon</p>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    )
  }

  if (mode === 'new') {
    // Show provisioning progress interface
    if (provisioningState !== 'idle') {
      return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/10 via-background to-background p-8">
          <Card className="w-full max-w-2xl">
            <CardHeader className="text-center">
              <CardTitle>Domain Setup</CardTitle>
              <CardDescription>
                {domainName && `Domain Name: ${domainName}`}
              </CardDescription>
            </CardHeader>
            <CardContent className="text-center space-y-6">
              {/* Progress Indicator */}
              <div className="flex justify-center">
                {provisioningState === 'provisioning' && (
                  <Loader2 className="h-12 w-12 animate-spin text-primary" />
                )}
                {provisioningState === 'success' && (
                  <CheckCircle className="h-12 w-12 text-green-500" />
                )}
                {provisioningState === 'error' && (
                  <AlertCircle className="h-12 w-12 text-red-500" />
                )}
              </div>

              {/* Status Message */}
              <div className="space-y-4">
                {provisioningState === 'provisioning' && (
                  <div>
                    <p className="text-lg font-medium">Provisioning Domain...</p>
                    <p className="text-sm text-muted-foreground">{currentStatus}</p>
                  </div>
                )}

                {provisioningState === 'success' && (
                  <div>
                    <p className="text-lg font-medium text-green-600">Welcome to {domainName}!</p>
                    <p className="text-sm text-muted-foreground">Redirecting to dashboard...</p>
                  </div>
                )}

                {provisioningState === 'error' && (
                  <div>
                    <p className="text-lg font-medium text-red-600">Setup Failed</p>
                    <p className="text-sm text-muted-foreground">{error}</p>
                  </div>
                )}
              </div>

              {/* Action Buttons */}
              <div className="flex justify-center">
                {provisioningState === 'error' && (
                  <Button 
                    variant="outline" 
                    onClick={() => {
                      setProvisioningState('idle')
                      setError('')
                      setCurrentStatus('')
                    }}
                  >
                    Back to Setup
                  </Button>
                )}
                
                {provisioningState === 'success' && (
                  <Button 
                    onClick={() => {
                      localStorage.setItem('vexa-setup-complete', 'true')
                      // Store domain info for immediate display
                      localStorage.setItem('vexa-domain-info', JSON.stringify({
                        domain: domainName,
                        realm: formData.realm
                      }))
                      queryClient.invalidateQueries({ queryKey: ['domainStatus'] })
                      window.location.href = '/'
                    }}
                  >
                    Go to Dashboard
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      )
    }

    // Show setup form
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/10 via-background to-background p-8">
        <Card className="w-full max-w-2xl">
          <CardHeader>
            <CardTitle>New Domain Setup</CardTitle>
            <CardDescription>
              Configure your new Vexa Domain
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSetupNew} className="space-y-4">

              <div className="space-y-2">
                <label htmlFor="realm" className="text-sm font-medium">
                  Realm
                </label>
                <Input
                  id="realm"
                  type="text"
                  placeholder="mydomain.local"
                  value={formData.realm}
                  onChange={(e) => setFormData({ ...formData, realm: e.target.value })}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Kerberos realm (e.g., mydomain.local)
                </p>
                {formData.realm && (
                  <div className="rounded-md bg-accent p-2 text-sm">
                    Domain Name: <span className="font-mono">{getDomainFromRealm(formData.realm)}</span>
                    <p className="text-xs text-muted-foreground mt-1">
                      Domain name auto-generated from realm
                    </p>
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <label htmlFor="dns_provider" className="text-sm font-medium">
                  DNS Provider
                </label>
                <select
                  id="dns_provider"
                  value={formData.dnsProvider}
                  onChange={(e) => setFormData({ ...formData, dnsProvider: e.target.value })}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {dnsProviders.map((provider) => (
                    <option key={provider.value} value={provider.value}>
                      {provider.label}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-muted-foreground">
                  DNS provider for upstream queries
                </p>
              </div>

              {formData.dnsProvider === 'custom' && (
                <div className="space-y-2">
                  <label htmlFor="customDnsServers" className="text-sm font-medium">
                    Custom DNS Servers
                  </label>
                  <Input
                    id="customDnsServers"
                    type="text"
                    placeholder="8.8.8.8,8.8.4.4"
                    value={formData.customDnsServers}
                    onChange={(e) => setFormData({ ...formData, customDnsServers: e.target.value })}
                    required
                  />
                  <p className="text-xs text-muted-foreground">
                    Comma-separated list of DNS servers
                  </p>
                </div>
              )}

              {error && (
                <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
                  {error}
                </div>
              )}

              <div className="flex gap-2 pt-4">
                <Button type="button" variant="outline" onClick={() => setMode(null)}>
                  Back
                </Button>
                <Button type="submit" className="flex-1" disabled={loading}>
                  {loading ? 'Provisioning...' : 'Provision Domain'}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    )
  }

  return null
}

