import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'
import api from '../lib/api'

export default function DomainSetup() {
  const [formData, setFormData] = useState({
    domain: '',
    realm: '',
    admin_password: '',
    confirm_password: '',
    dns_forwarder: '8.8.8.8',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess(false)

    if (formData.admin_password !== formData.confirm_password) {
      setError('Passwords do not match')
      return
    }

    setLoading(true)

    try {
      await api.post('/domain/provision', {
        domain: formData.domain,
        realm: formData.realm,
        admin_password: formData.admin_password,
        dns_forwarder: formData.dns_forwarder,
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
          Provision your Samba Active Directory Domain Controller
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Domain Configuration</CardTitle>
          <CardDescription>
            Configure your new Active Directory domain
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
              <p className="text-muted-foreground">
                Your Active Directory domain controller is now running
              </p>
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
                <label htmlFor="admin_password" className="text-sm font-medium">
                  Administrator Password
                </label>
                <Input
                  id="admin_password"
                  type="password"
                  placeholder="Enter strong password"
                  value={formData.admin_password}
                  onChange={(e) =>
                    setFormData({ ...formData, admin_password: e.target.value })
                  }
                  required
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="confirm_password" className="text-sm font-medium">
                  Confirm Password
                </label>
                <Input
                  id="confirm_password"
                  type="password"
                  placeholder="Confirm password"
                  value={formData.confirm_password}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      confirm_password: e.target.value,
                    })
                  }
                  required
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="dns_forwarder" className="text-sm font-medium">
                  DNS Forwarder
                </label>
                <Input
                  id="dns_forwarder"
                  type="text"
                  placeholder="8.8.8.8"
                  value={formData.dns_forwarder}
                  onChange={(e) =>
                    setFormData({ ...formData, dns_forwarder: e.target.value })
                  }
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

