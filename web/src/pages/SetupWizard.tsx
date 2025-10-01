import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Server, GitMerge, Plus } from 'lucide-react'
import api from '../lib/api'

type SetupMode = 'new' | 'join' | 'migrate' | null

export default function SetupWizard() {
  const navigate = useNavigate()
  const [mode, setMode] = useState<SetupMode>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  
  const [formData, setFormData] = useState({
    domain: '',
    realm: '',
    adminPassword: '',
    confirmPassword: '',
    dnsForwarder: '8.8.8.8',
  })

  const handleSetupNew = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (formData.adminPassword !== formData.confirmPassword) {
      setError('Passwords do not match')
      return
    }

    setLoading(true)
    try {
      await api.post('/domain/provision', {
        domain: formData.domain,
        realm: formData.realm,
        admin_password: formData.adminPassword,
        dns_forwarder: formData.dnsForwarder,
      })
      
      localStorage.setItem('vexa-setup-complete', 'true')
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to provision domain')
    } finally {
      setLoading(false)
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
              Let's set up your Active Directory environment
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
                  Set up a brand new Active Directory domain
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
                  Transfer your existing Active Directory domain to Vexa infrastructure.
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
              Configure your new Active Directory domain
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSetupNew} className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="domain" className="text-sm font-medium">
                  Domain Name
                </label>
                <Input
                  id="domain"
                  type="text"
                  placeholder="MYDOMAIN"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
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
                  onChange={(e) => setFormData({ ...formData, realm: e.target.value })}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Kerberos realm (e.g., mydomain.local)
                </p>
              </div>

              <div className="space-y-2">
                <label htmlFor="adminPassword" className="text-sm font-medium">
                  Administrator Password
                </label>
                <Input
                  id="adminPassword"
                  type="password"
                  placeholder="Enter strong password"
                  value={formData.adminPassword}
                  onChange={(e) => setFormData({ ...formData, adminPassword: e.target.value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="confirmPassword" className="text-sm font-medium">
                  Confirm Password
                </label>
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="Confirm password"
                  value={formData.confirmPassword}
                  onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="dnsForwarder" className="text-sm font-medium">
                  DNS Forwarder
                </label>
                <Input
                  id="dnsForwarder"
                  type="text"
                  placeholder="8.8.8.8"
                  value={formData.dnsForwarder}
                  onChange={(e) => setFormData({ ...formData, dnsForwarder: e.target.value })}
                />
                <p className="text-xs text-muted-foreground">
                  External DNS server for forwarding (optional)
                </p>
              </div>

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

