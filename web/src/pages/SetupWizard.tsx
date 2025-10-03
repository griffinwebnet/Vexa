import { useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Server, GitMerge, Plus, Copy, Terminal } from 'lucide-react'

type SetupMode = 'new' | 'join' | 'migrate' | null

export default function SetupWizard() {
  const navigate = useNavigate()
  const [mode, setMode] = useState<SetupMode>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [cliOutput, setCliOutput] = useState<string[]>([])
  const [showOutput, setShowOutput] = useState(false)
  const outputRef = useRef<HTMLDivElement>(null)
  
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
    setCliOutput([])
    setShowOutput(true)

    if (!formData.realm.trim()) {
      setError('Realm is required')
      return
    }

    setLoading(true)
    try {
      // Get DNS servers based on provider
      let dnsServers = formData.customDnsServers
      if (formData.dnsProvider !== 'custom') {
        const provider = dnsProviders.find(p => p.value === formData.dnsProvider)
        dnsServers = provider?.servers || '1.1.1.1,1.0.0.1'
      }

      const domainName = getDomainFromRealm(formData.realm)
      
      // Use the new streaming endpoint
      const response = await fetch('/api/v1/domain/provision-with-output', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          domain: domainName,
          realm: formData.realm,
          dns_forwarder: dnsServers,
        })
      })

      if (!response.ok) {
        throw new Error('Failed to start provisioning')
      }

      const reader = response.body?.getReader()
      const decoder = new TextDecoder()

      if (reader) {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          const chunk = decoder.decode(value)
          const lines = chunk.split('\n')
          
          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6))
                if (data.type === 'output') {
                  setCliOutput(prev => [...prev, data.content])
                  // Auto-scroll to bottom
                  setTimeout(() => {
                    if (outputRef.current) {
                      outputRef.current.scrollTop = outputRef.current.scrollHeight
                    }
                  }, 100)
                } else if (data.type === 'complete') {
                  localStorage.setItem('vexa-setup-complete', 'true')
                  navigate('/')
                }
              } catch (e) {
                // Ignore parsing errors for incomplete chunks
              }
            }
          }
        }
      }
    } catch (err: any) {
      setError(err.message || 'Failed to provision domain')
      setCliOutput(prev => [...prev, `ERROR: ${err.message || 'Unknown error'}`])
    } finally {
      setLoading(false)
    }
  }

  const copyOutputToClipboard = () => {
    const outputText = cliOutput.join('\n')
    navigator.clipboard.writeText(outputText).then(() => {
      // Could add a toast notification here
    }).catch(err => {
      console.error('Failed to copy to clipboard:', err)
    })
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

              {showOutput && (
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <label className="text-sm font-medium flex items-center gap-2">
                      <Terminal className="h-4 w-4" />
                      CLI Output
                    </label>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={copyOutputToClipboard}
                      className="flex items-center gap-2"
                    >
                      <Copy className="h-4 w-4" />
                      Copy
                    </Button>
                  </div>
                  <div
                    ref={outputRef}
                    className="bg-black text-green-400 p-4 rounded-md font-mono text-sm h-64 overflow-y-auto border"
                  >
                    {cliOutput.length === 0 ? (
                      <div className="text-gray-500">Waiting for output...</div>
                    ) : (
                      cliOutput.map((line, index) => (
                        <div key={index} className="mb-1">
                          {line.startsWith('ERROR:') ? (
                            <span className="text-red-400">{line}</span>
                          ) : line.startsWith('STDOUT:') ? (
                            <span className="text-green-400">{line}</span>
                          ) : line.startsWith('STDERR:') ? (
                            <span className="text-yellow-400">{line}</span>
                          ) : (
                            <span>{line}</span>
                          )}
                        </div>
                      ))
                    )}
                  </div>
                </div>
              )}

              <div className="flex gap-2 pt-4">
                <Button type="button" variant="outline" onClick={() => setMode(null)}>
                  Back
                </Button>
                <Button type="submit" className="flex-1" disabled={loading}>
                  {loading ? 'Provisioning...' : 'Provision Domain'}
                </Button>
                {!loading && showOutput && cliOutput.length > 0 && (
                  <Button 
                    type="button" 
                    onClick={() => {
                      localStorage.setItem('vexa-setup-complete', 'true')
                      navigate('/')
                    }}
                    className="flex-1"
                  >
                    Continue to Dashboard
                  </Button>
                )}
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    )
  }

  return null
}

