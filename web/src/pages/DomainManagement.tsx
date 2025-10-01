import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { useQuery } from '@tanstack/react-query'
import { Server, FolderTree, Shield, Settings as SettingsIcon } from 'lucide-react'
import { Link } from 'react-router-dom'
import api from '../lib/api'

export default function DomainManagement() {
  const { data: domainStatus } = useQuery({
    queryKey: ['domainStatus'],
    queryFn: async () => {
      const response = await api.get('/domain/status')
      return response.data
    },
  })

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Domain Management</h1>
        <p className="text-muted-foreground">
          Configure and manage your Active Directory domain
        </p>
      </div>

      {/* Domain Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5 text-primary" />
            Domain Information
          </CardTitle>
          <CardDescription>
            Current domain configuration and status
          </CardDescription>
        </CardHeader>
        <CardContent>
          {domainStatus?.provisioned ? (
            <div className="grid md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex justify-between p-3 rounded-lg bg-accent/50">
                  <span className="font-medium">Domain Name:</span>
                  <span className="text-muted-foreground">{domainStatus.domain || 'N/A'}</span>
                </div>
                <div className="flex justify-between p-3 rounded-lg bg-accent/50">
                  <span className="font-medium">Realm:</span>
                  <span className="text-muted-foreground">{domainStatus.realm || 'N/A'}</span>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between p-3 rounded-lg bg-accent/50">
                  <span className="font-medium">DC Status:</span>
                  <span className={domainStatus.dc_ready ? 'text-green-500' : 'text-destructive'}>
                    {domainStatus.dc_ready ? 'Running' : 'Stopped'}
                  </span>
                </div>
                <div className="flex justify-between p-3 rounded-lg bg-accent/50">
                  <span className="font-medium">DNS Status:</span>
                  <span className={domainStatus.dns_ready ? 'text-green-500' : 'text-destructive'}>
                    {domainStatus.dns_ready ? 'Running' : 'Stopped'}
                  </span>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8">
              <p className="text-muted-foreground mb-4">
                No domain has been provisioned yet
              </p>
              <Link to="/setup">
                <Button>Provision Domain</Button>
              </Link>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Management Sections Grid */}
      <div className="grid md:grid-cols-3 gap-6">
        <Link to="/domain/ous">
          <Card className="hover:border-primary transition-colors cursor-pointer h-full">
            <CardHeader>
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
                <FolderTree className="h-6 w-6 text-primary" />
              </div>
              <CardTitle>Organizational Units</CardTitle>
              <CardDescription>
                Manage OU structure and hierarchy
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Create and organize OUs for departments, locations, or business units.
              </p>
            </CardContent>
          </Card>
        </Link>

        <Link to="/domain/policies">
          <Card className="hover:border-primary transition-colors cursor-pointer h-full">
            <CardHeader>
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
                <Shield className="h-6 w-6 text-primary" />
              </div>
              <CardTitle>Default Policies</CardTitle>
              <CardDescription>
                Password and security policies
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Configure password complexity, expiration, and account lockout policies.
              </p>
            </CardContent>
          </Card>
        </Link>

        <Card className="opacity-50 h-full">
          <CardHeader>
            <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
              <SettingsIcon className="h-6 w-6 text-primary" />
            </div>
            <CardTitle>Advanced Settings</CardTitle>
            <CardDescription>
              Domain-wide configuration
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Functional levels, replication, and advanced domain settings.
            </p>
            <p className="text-xs text-yellow-600 dark:text-yellow-500 mt-2">Coming soon</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

