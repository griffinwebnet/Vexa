import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { 
  Globe, 
  Server, 
  CheckCircle, 
  XCircle, 
  ArrowRightLeft, 
  Settings,
  RefreshCw,
  ExternalLink
} from 'lucide-react'
import api from '../lib/api'

export default function DNS() {
  const [externalDns, setExternalDns] = useState({
    primary: '1.1.1.1',
    secondary: '1.0.0.1'
  })
  const [isUpdating, setIsUpdating] = useState(false)

  // Fetch DNS status from API
  const { data: sambaDnsStatus, refetch: refetchDnsStatus } = useQuery({
    queryKey: ['dnsStatus'],
    queryFn: async () => {
      const response = await api.get('/dns/status')
      return response.data
    },
  })

  // Update external DNS state when API data loads
  useEffect(() => {
    if (sambaDnsStatus?.forwarders && sambaDnsStatus.forwarders.length >= 2) {
      setExternalDns({
        primary: sambaDnsStatus.forwarders[0],
        secondary: sambaDnsStatus.forwarders[1]
      })
    }
  }, [sambaDnsStatus])

  const handleUpdateForwarders = async () => {
    setIsUpdating(true)
    try {
      await api.put('/dns/forwarders', externalDns)
      // Refresh DNS status to get updated forwarders
      refetchDnsStatus()
    } catch (error) {
      console.error('Failed to update DNS forwarders:', error)
    } finally {
      setIsUpdating(false)
    }
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">DNS Management</h1>
        <p className="text-muted-foreground">
          Manage Samba DNS settings and external DNS forwarders
        </p>
      </div>

      {/* Samba DNS Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-6 w-6" />
            Samba DNS Server Status
          </CardTitle>
          <CardDescription>
            Built-in DNS server for your domain
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 md:grid-cols-2">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="font-medium">DNS Server Status</span>
                <div className="flex items-center gap-2">
                  {sambaDnsStatus?.enabled ? (
                    <>
                      <CheckCircle className="h-5 w-5 text-green-500" />
                      <span className="text-green-600">Active</span>
                    </>
                  ) : (
                    <>
                      <XCircle className="h-5 w-5 text-red-500" />
                      <span className="text-red-600">Inactive</span>
                    </>
                  )}
                </div>
              </div>
              
              <div className="space-y-2">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Domain</label>
                  <p className="text-lg font-mono">{sambaDnsStatus?.domain || 'Loading...'}</p>
                </div>
                <div>
                  <label className="text-sm font-medium text-muted-foreground">DNS Server</label>
                  <p className="text-lg font-mono">{sambaDnsStatus?.dns_server || '127.0.0.1'}:{sambaDnsStatus?.port || '53'}</p>
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-muted-foreground">Active Zones</label>
                <div className="space-y-2 mt-2">
                  {sambaDnsStatus?.zones?.map((zone: any, index: number) => (
                    <div key={index} className="flex items-center justify-between p-3 bg-muted rounded-lg">
                      <div>
                        <p className="font-medium font-mono">{zone.name}</p>
                        <p className="text-sm text-muted-foreground">{zone.type} Zone</p>
                      </div>
                      <div className="text-right">
                        <p className="text-sm font-medium">{zone.records} records</p>
                      </div>
                    </div>
                  )) || (
                    <div className="text-center py-4 text-muted-foreground">
                      Loading zones...
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>

          <div className="mt-6 p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
            <h4 className="font-medium text-blue-900 dark:text-blue-100 mb-2 flex items-center gap-2">
              <Globe className="h-4 w-4" />
              How Samba DNS Works
            </h4>
            <p className="text-sm text-blue-700 dark:text-blue-300">
              Samba automatically manages DNS for your domain including SRV records for Active Directory services, 
              host records for domain-joined computers, and zone management. No additional DNS server setup required.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* External DNS Forwarders */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ArrowRightLeft className="h-6 w-6" />
            External DNS Forwarders
          </CardTitle>
          <CardDescription>
            Configure external DNS servers for internet lookups
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <label className="text-sm font-medium mb-2 block">Primary DNS Server</label>
                <Input
                  value={externalDns.primary}
                  onChange={(e) => setExternalDns({ ...externalDns, primary: e.target.value })}
                  placeholder="1.1.1.1"
                  className="font-mono"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Primary DNS server for external queries
                </p>
              </div>
              <div>
                <label className="text-sm font-medium mb-2 block">Secondary DNS Server</label>
                <Input
                  value={externalDns.secondary}
                  onChange={(e) => setExternalDns({ ...externalDns, secondary: e.target.value })}
                  placeholder="1.0.0.1"
                  className="font-mono"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Backup DNS server for external queries
                </p>
              </div>
            </div>

            {/* Popular DNS Services */}
            <div>
              <label className="text-sm font-medium mb-3 block">Quick Setup - Popular DNS Services</label>
              <div className="grid gap-2 md:grid-cols-3">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setExternalDns({ primary: '1.1.1.1', secondary: '1.0.0.1' })}
                  className="justify-start"
                >
                  <ExternalLink className="h-4 w-4 mr-2" />
                  Cloudflare (1.1.1.1)
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setExternalDns({ primary: '8.8.8.8', secondary: '8.8.4.4' })}
                  className="justify-start"
                >
                  <ExternalLink className="h-4 w-4 mr-2" />
                  Google DNS (8.8.8.8)
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setExternalDns({ primary: '9.9.9.9', secondary: '149.112.112.112' })}
                  className="justify-start"
                >
                  <ExternalLink className="h-4 w-4 mr-2" />
                  Quad9 (9.9.9.9)
                </Button>
              </div>
            </div>

            <div className="flex gap-2">
              <Button 
                onClick={handleUpdateForwarders}
                disabled={isUpdating}
              >
                {isUpdating ? (
                  <>
                    <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                    Updating...
                  </>
                ) : (
                  <>
                    <Settings className="h-4 w-4 mr-2" />
                    Update DNS Forwarders
                  </>
                )}
              </Button>
            </div>

            <div className="mt-6 p-4 bg-amber-50 dark:bg-amber-950/20 rounded-lg">
              <h4 className="font-medium text-amber-900 dark:text-amber-100 mb-2">
                How DNS Forwarding Works
              </h4>
              <p className="text-sm text-amber-700 dark:text-amber-300">
                When clients query for external domains (like google.com), Samba DNS will forward these requests 
                to the configured external DNS servers. Internal domain queries are handled directly by Samba.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* DNS Query Testing */}
      <Card>
        <CardHeader>
          <CardTitle>DNS Query Testing</CardTitle>
          <CardDescription>
            Test DNS resolution for your domain and external sites
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            DNS query testing interface coming soon
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

