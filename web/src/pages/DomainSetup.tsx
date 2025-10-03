import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'
import api from '../lib/api'

export default function DomainSetup() {
  const [formData, setFormData] = useState({
    domain: '',
    realm: '',
    dns_provider: 'cloudflare',
    dns_forwarder: '',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  const dnsProviders = [
    { value: 'cloudflare', label: 'Cloudflare (1.1.1.1)', servers: '1.1.1.1,1.0.0.1' },
    { value: 'google', label: 'Google (8.8.8.8)', servers: '8.8.8.8,8.8.4.4' },
    { value: 'quad9', label: 'Quad9 (9.9.9.9)', servers: '9.9.9.9,149.112.112.112' },
    { value: 'custom', label: 'Custom DNS Servers', servers: '' },
  ]

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess(false)

    setLoading(true)

    try {
      // Get DNS servers based on provider
      let dnsServers = formData.dns_forwarder
      if (formData.dns_provider !== 'custom') {
        const provider = dnsProviders.find(p => p.value === formData.dns_provider)
        dnsServers = provider?.servers || '1.1.1.1,1.0.0.1'
      }

      await api.post('/domain/provision', {
        domain: formData.domain,
        realm: formData.realm,
        dns_forwarder: dnsServers,
      })

      setSuccess(true)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to provision domain')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Domain Setup</h1>
        <p className="text-muted-foreground">
          Provision your Vexa Domain controller
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Domain Configuration</CardTitle>
          <CardDescription>
            Configure your new Vexa Domain
          </CardDescription>
        </CardHeader>
        <CardContent>
          {success ? (
            <div className="text-center py-8">
              <div className="rounded-full h-16 w-16 bg-green-100 dark:bg-green-900 mx-auto flex items-center justify-center mb-4">
                <svg
                  className="h-8 w-8 text-green-600 dark:text-green-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
              <h3 className="text-lg font-semibold mb-2">
                Domain Provisioned Successfully!
              </h3>
              <p className="text-muted-foreground mb-4">
                Your Vexa Domain controller is now running. You can now create users and manage your domain.
              </p>
              <Button onClick={() => window.location.href = '/dashboard'}>
                Go to Dashboard
              </Button>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="domain" className="text-sm font-medium">
                  Domain Name
                </label>
                <Input
                  id="domain"
                  type="text"
                  placeholder="MYDOMAIN"
                  value={formData.domain}
                  onChange={(e) =>
                    setFormData({ ...formData, domain: e.target.value })
                  }
                  required
                />
                <p className="text-xs text-muted-foreground">
                  NetBIOS domain name (e.g., MYDOMAIN)
                </p>
              </div>

              <div className="space-y-2">
                <label htmlFor="realm" className="text-sm font-medium">
                  Realm
                </label>
                <Input
                  id="realm"
                  type="text"
                  placeholder="mydomain.local"
                  value={formData.realm}
                  onChange={(e) =>
                    setFormData({ ...formData, realm: e.target.value })
                  }
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Kerberos realm (e.g., mydomain.local)
                </p>
              </div>

              <div className="space-y-2">
                <label htmlFor="dns_provider" className="text-sm font-medium">
                  DNS Provider
                </label>
                <select
                  id="dns_provider"
                  value={formData.dns_provider}
                  onChange={(e) => setFormData({ ...formData, dns_provider: e.target.value })}
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

              {formData.dns_provider === 'custom' && (
                <div className="space-y-2">
                  <label htmlFor="dns_forwarder" className="text-sm font-medium">
                    Custom DNS Servers
                  </label>
                  <Input
                    id="dns_forwarder"
                    type="text"
                    placeholder="8.8.8.8,8.8.4.4"
                    value={formData.dns_forwarder}
                    onChange={(e) =>
                      setFormData({ ...formData, dns_forwarder: e.target.value })
                    }
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

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? 'Provisioning...' : 'Provision Domain'}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

