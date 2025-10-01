import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Users, UsersRound, Monitor, Activity } from 'lucide-react'
import api from '../lib/api'

export default function Dashboard() {
  const { data: domainStatus } = useQuery({
    queryKey: ['domainStatus'],
    queryFn: async () => {
      const response = await api.get('/domain/status')
      return response.data
    },
  })

  const stats = [
    {
      name: 'Total Users',
      value: '0',
      icon: Users,
      description: 'Active directory users',
      color: 'text-blue-500',
    },
    {
      name: 'Groups',
      value: '0',
      icon: UsersRound,
      description: 'Security groups',
      color: 'text-green-500',
    },
    {
      name: 'Computers',
      value: '0',
      icon: Monitor,
      description: 'Domain-joined devices',
      color: 'text-purple-500',
    },
    {
      name: 'Domain Status',
      value: domainStatus?.dc_ready ? 'Online' : 'Offline',
      icon: Activity,
      description: 'Domain controller',
      color: domainStatus?.dc_ready ? 'text-green-500' : 'text-red-500',
    },
  ]

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome to Vexa Active Directory Management
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <Card key={stat.name}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {stat.name}
              </CardTitle>
              <stat.icon className={`h-5 w-5 ${stat.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground">
                {stat.description}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Domain Information */}
      <Card>
        <CardHeader>
          <CardTitle>Domain Information</CardTitle>
          <CardDescription>
            Current Active Directory domain configuration
          </CardDescription>
        </CardHeader>
        <CardContent>
          {domainStatus?.provisioned ? (
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="font-medium">Domain:</span>
                <span className="text-muted-foreground">
                  {domainStatus.domain || 'N/A'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">Realm:</span>
                <span className="text-muted-foreground">
                  {domainStatus.realm || 'N/A'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">DC Status:</span>
                <span
                  className={
                    domainStatus.dc_ready
                      ? 'text-green-500'
                      : 'text-destructive'
                  }
                >
                  {domainStatus.dc_ready ? 'Running' : 'Stopped'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">DNS Status:</span>
                <span
                  className={
                    domainStatus.dns_ready
                      ? 'text-green-500'
                      : 'text-destructive'
                  }
                >
                  {domainStatus.dns_ready ? 'Running' : 'Stopped'}
                </span>
              </div>
            </div>
          ) : (
            <div className="text-center py-8">
              <p className="text-muted-foreground mb-4">
                No domain has been provisioned yet
              </p>
              <a
                href="/setup"
                className="text-primary hover:underline font-medium"
              >
                Set up your domain →
              </a>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

