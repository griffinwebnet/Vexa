import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Monitor, Wifi, Network, Circle } from 'lucide-react'
import api from '../lib/api'

type ConnectionType = 'local' | 'overlay' | 'offline'

interface Computer {
  name: string
  dns_name: string
  operating_system: string
  last_logon: string
  connection_type: ConnectionType
  online: boolean
  ip_address?: string
  overlay_ip?: string
}

export default function Computers() {
  const { data: computersData, isLoading } = useQuery({
    queryKey: ['computers'],
    queryFn: async () => {
      const response = await api.get('/computers')
      return response.data
    },
    refetchInterval: 10000, // Refresh every 10 seconds
  })

  const { data: overlayStatus } = useQuery({
    queryKey: ['overlayNetworking'],
    queryFn: async () => {
      const response = await api.get('/system/overlay-status')
      return response.data
    },
  })

  const getStatusDot = (computer: Computer) => {
    if (!computer.online) {
      return <Circle className="h-3 w-3 fill-gray-400 text-gray-400" />
    }
    if (computer.connection_type === 'overlay') {
      return <Circle className="h-3 w-3 fill-green-500 text-green-500" />
    }
    return <Circle className="h-3 w-3 fill-blue-500 text-blue-500" />
  }

  const getConnectionBadge = (computer: Computer) => {
    if (!computer.online) {
      return (
        <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400">
          Offline
        </span>
      )
    }

    if (computer.connection_type === 'overlay') {
      return (
        <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300">
          <Network className="h-3 w-3" />
          Overlay Connected
        </span>
      )
    }

    return (
      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300">
        <Wifi className="h-3 w-3" />
        Locally Connected
      </span>
    )
  }

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Computers</h1>
          <p className="text-muted-foreground">
            Domain-joined computers and devices
          </p>
        </div>
        {overlayStatus?.enabled && (
          <div className="text-sm text-muted-foreground">
            <span className="flex items-center gap-2">
              <Network className="h-4 w-4 text-primary" />
              Overlay networking enabled
            </span>
          </div>
        )}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Domain Computers</CardTitle>
          <CardDescription>
            {computersData?.count || 0} computers joined to the domain
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">
              Loading computers...
            </div>
          ) : computersData?.computers && computersData.computers.length > 0 ? (
            <div className="space-y-2">
              {computersData.computers.map((computer: Computer) => (
                <div
                  key={computer.name}
                  className="flex items-center justify-between p-4 rounded-lg border hover:bg-accent"
                >
                  <div className="flex items-center gap-4">
                    {getStatusDot(computer)}
                    <Monitor className="h-8 w-8 text-muted-foreground" />
                    <div>
                      <p className="font-medium">{computer.name}</p>
                      <p className="text-sm text-muted-foreground">
                        {computer.dns_name}
                      </p>
                      {computer.operating_system && (
                        <p className="text-xs text-muted-foreground">
                          {computer.operating_system}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    {computer.online && computer.ip_address && (
                      <div className="text-right text-xs text-muted-foreground">
                        <div className="font-mono">{computer.ip_address}</div>
                        {computer.overlay_ip && (
                          <div className="font-mono text-green-600 dark:text-green-400">
                            {computer.overlay_ip}
                          </div>
                        )}
                      </div>
                    )}
                    {getConnectionBadge(computer)}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              No computers have joined the domain yet
            </div>
          )}
        </CardContent>
      </Card>

      {overlayStatus?.enabled && (
        <Card className="border-primary/50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Network className="h-5 w-5 text-primary" />
              Connection Status Legend
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-2">
                <Circle className="h-3 w-3 fill-green-500 text-green-500" />
                <span className="font-medium">Green:</span>
                <span className="text-muted-foreground">Online via overlay network (Headscale)</span>
              </div>
              <div className="flex items-center gap-2">
                <Circle className="h-3 w-3 fill-blue-500 text-blue-500" />
                <span className="font-medium">Blue:</span>
                <span className="text-muted-foreground">Online via local network</span>
              </div>
              <div className="flex items-center gap-2">
                <Circle className="h-3 w-3 fill-gray-400 text-gray-400" />
                <span className="font-medium">Gray:</span>
                <span className="text-muted-foreground">Offline</span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

