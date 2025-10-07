import { useQuery } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { 
  Monitor, 
  Settings, 
  Copy, 
  CheckCircle, 
  Info,
  User,
  Calendar,
  Key,
  Network,
  Activity
} from 'lucide-react'
import api from '../lib/api'
import { useState } from 'react'

interface MachineDetails {
  id: string
  name: string
  hostname: string
  creator: string
  created: string
  lastSeen: string
  keyExpiry: string
  nodeKey: string
  tailscaleIPv4: string
  tailscaleIPv6: string
  shortDomain: string
  status: 'online' | 'offline' | 'expired'
  managedBy: string
}

export default function MachineDetails() {
  const { machineId } = useParams<{ machineId: string }>()
  const navigate = useNavigate()
  const [copied, setCopied] = useState<string | null>(null)

  const { data: machine, isLoading } = useQuery({
    queryKey: ['machine', machineId],
    queryFn: async () => {
      const response = await api.get(`/machines/${machineId}`)
      return response.data as MachineDetails
    },
    enabled: !!machineId,
  })

  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(label)
      setTimeout(() => setCopied(null), 2000)
    } catch (error) {
      console.error('Failed to copy to clipboard:', error)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'bg-green-500'
      case 'offline': return 'bg-gray-500'
      case 'expired': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }


  if (isLoading) {
    return (
      <div className="p-6">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/4 mb-4"></div>
          <div className="h-32 bg-muted rounded"></div>
        </div>
      </div>
    )
  }

  if (!machine) {
    return (
      <div className="p-6">
        <div className="text-center py-8">
          <h1 className="text-2xl font-semibold mb-2">Machine Not Found</h1>
          <p className="text-muted-foreground mb-4">The requested machine could not be found.</p>
          <Button onClick={() => navigate('/computers')}>
            Back to Computers
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className={`w-3 h-3 rounded-full ${getStatusColor(machine.status)}`}></div>
          <h1 className="text-2xl font-semibold">{machine.name}</h1>
        </div>
        <Button variant="outline">
          <Settings className="h-4 w-4 mr-2" />
          Machine Settings
        </Button>
      </div>

      {/* Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Info className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-medium">Managed by</span>
              </div>
              <div className="flex items-center gap-2">
                <User className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm">{machine.managedBy}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">Status</span>
              <span className="text-sm bg-muted px-2 py-1 rounded">
                {machine.keyExpiry === 'Never' ? 'No expiry' : machine.keyExpiry}
              </span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Machine Details */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Monitor className="h-5 w-5" />
            Machine Details
          </CardTitle>
          <CardDescription>
            Information about this machine's network. Used to debug connection issues.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Left Column - Machine Attributes */}
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Creator</span>
                <span className="text-sm">{machine.creator}</span>
              </div>
              
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Machine name</span>
                <span className="text-sm">{machine.name}</span>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <span className="text-sm text-muted-foreground">OS hostname</span>
                  <Info className="h-3 w-3 text-muted-foreground" />
                </div>
                <span className="text-sm">{machine.hostname}</span>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <span className="text-sm text-muted-foreground">ID</span>
                  <Info className="h-3 w-3 text-muted-foreground" />
                </div>
                <span className="text-sm">{machine.id}</span>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <span className="text-sm text-muted-foreground">Node key</span>
                  <Info className="h-3 w-3 text-muted-foreground" />
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-mono">{machine.nodeKey}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(machine.nodeKey, 'nodekey')}
                  >
                    {copied === 'nodekey' ? (
                      <CheckCircle className="h-3 w-3 text-green-500" />
                    ) : (
                      <Copy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <Calendar className="h-3 w-3 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Created</span>
                </div>
                <span className="text-sm">{machine.created}</span>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <Activity className="h-3 w-3 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Last Seen</span>
                </div>
                <span className="text-sm">{machine.lastSeen}</span>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <Key className="h-3 w-3 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Key expiry</span>
                </div>
                <span className="text-sm">{machine.keyExpiry}</span>
              </div>
            </div>

            {/* Right Column - Addresses */}
            <div className="space-y-4">
              <h3 className="text-sm font-medium">ADDRESSES</h3>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <span className="text-sm text-muted-foreground">Tailscale IPv4</span>
                  <Info className="h-3 w-3 text-muted-foreground" />
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-mono">{machine.tailscaleIPv4}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(machine.tailscaleIPv4, 'ipv4')}
                  >
                    {copied === 'ipv4' ? (
                      <CheckCircle className="h-3 w-3 text-green-500" />
                    ) : (
                      <Copy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <span className="text-sm text-muted-foreground">Tailscale IPv6</span>
                  <Info className="h-3 w-3 text-muted-foreground" />
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-mono">{machine.tailscaleIPv6}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(machine.tailscaleIPv6, 'ipv6')}
                  >
                    {copied === 'ipv6' ? (
                      <CheckCircle className="h-3 w-3 text-green-500" />
                    ) : (
                      <Copy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1">
                  <span className="text-sm text-muted-foreground">Short domain</span>
                  <Info className="h-3 w-3 text-muted-foreground" />
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-mono">{machine.shortDomain}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(machine.shortDomain, 'domain')}
                  >
                    {copied === 'domain' ? (
                      <CheckCircle className="h-3 w-3 text-green-500" />
                    ) : (
                      <Copy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Actions */}
      <div className="flex gap-4">
        <Button onClick={() => navigate('/computers')}>
          Back to Computers
        </Button>
        <Button variant="outline">
          <Network className="h-4 w-4 mr-2" />
          Test Connection
        </Button>
      </div>
    </div>
  )
}
