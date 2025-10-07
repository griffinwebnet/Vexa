import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { 
  Shield, 
  Eye, 
  Settings, 
  Search, 
  Filter, 
  Download, 
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  User,
  Activity,
  BarChart3
} from 'lucide-react'
import { Button } from '../components/ui/Button'
import { Card } from '../components/ui/Card'
import { Input } from '../components/ui/Input'
import api from '../lib/api'

interface AuditLogEntry {
  timestamp: string
  user: string
  action: string
  category: string
  resource: string
  details: Record<string, any>
  ip_address?: string
  user_agent?: string
  session_id?: string
  success: boolean
  error?: string
}

interface AuditLogResponse {
  entries: AuditLogEntry[]
  total: number
  page: number
  limit: number
  has_more: boolean
}

interface LogStats {
  total_entries: number
  entries_today: number
  failed_logins: number
  successful_logins: number
  user_actions: number
  system_actions: number
  categories: Record<string, number>
}

export default function Security() {
  const [activeTab, setActiveTab] = useState<'audit' | 'logs' | 'settings'>('audit')
  const [filters, setFilters] = useState({
    user: '',
    category: '',
    action: '',
    success: '',
    start_date: '',
    end_date: ''
  })
  const [page, setPage] = useState(1)
  const [logType, setLogType] = useState<'debug' | 'info' | 'warn' | 'error'>('info')
  const [logLevel, setLogLevel] = useState('INFO')

  // Fetch audit logs
  const { data: auditData, isLoading: auditLoading, refetch: refetchAudit } = useQuery({
    queryKey: ['audit-logs', page, filters],
    queryFn: async () => {
      const params = new URLSearchParams({
        page: page.toString(),
        limit: '50',
        ...Object.fromEntries(Object.entries(filters).filter(([_, v]) => v !== ''))
      })
      const response = await api.get(`/audit/logs?${params}`)
      return response.data as AuditLogResponse
    },
  })

  // Fetch log statistics
  const { data: logStats, isLoading: statsLoading } = useQuery({
    queryKey: ['log-stats'],
    queryFn: async () => {
      const response = await api.get('/audit/stats')
      return response.data as LogStats
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  // Fetch system logs
  const { data: systemLogs, isLoading: logsLoading, refetch: refetchLogs } = useQuery({
    queryKey: ['system-logs', logType, page],
    queryFn: async () => {
      const params = new URLSearchParams({
        page: page.toString(),
        limit: '100'
      })
      const response = await api.get(`/audit/logs/${logType}?${params}`)
      return response.data
    },
    enabled: activeTab === 'logs'
  })

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
    setPage(1) // Reset to first page when filtering
  }

  const handleLogLevelChange = async (newLevel: string) => {
    try {
      await api.post('/audit/log-level', { level: newLevel })
      setLogLevel(newLevel)
    } catch (error) {
      console.error('Failed to change log level:', error)
    }
  }

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString()
  }

  const getCategoryColor = (category: string) => {
    const colors = {
      authentication: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
      user_management: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
      group_management: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300',
      computer_management: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300',
      system_management: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
      data_access: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300',
      security: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
    }
    return colors[category as keyof typeof colors] || 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300'
  }

  const getSuccessIcon = (success: boolean) => {
    return success ? (
      <CheckCircle className="h-4 w-4 text-green-500" />
    ) : (
      <XCircle className="h-4 w-4 text-red-500" />
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <Shield className="h-8 w-8 text-blue-600" />
            Security & Audit
          </h1>
          <p className="text-muted-foreground mt-2">
            Monitor system activities, view audit trails, and manage security settings
          </p>
        </div>
        <Button onClick={() => refetchAudit()} disabled={auditLoading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${auditLoading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Statistics Cards */}
      {logStats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Total Entries</p>
                <p className="text-2xl font-bold">{logStats.total_entries.toLocaleString()}</p>
              </div>
              <BarChart3 className="h-8 w-8 text-blue-500" />
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Today's Activity</p>
                <p className="text-2xl font-bold">{logStats.entries_today}</p>
              </div>
              <Clock className="h-8 w-8 text-green-500" />
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Failed Logins</p>
                <p className="text-2xl font-bold text-red-600">{logStats.failed_logins}</p>
              </div>
              <AlertTriangle className="h-8 w-8 text-red-500" />
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Successful Logins</p>
                <p className="text-2xl font-bold text-green-600">{logStats.successful_logins}</p>
              </div>
              <CheckCircle className="h-8 w-8 text-green-500" />
            </div>
          </Card>
        </div>
      )}

      {/* Tab Navigation */}
      <div className="border-b border-border">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('audit')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'audit'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300'
            }`}
          >
            <Eye className="h-4 w-4 inline mr-2" />
            Audit Trail
          </button>
          <button
            onClick={() => setActiveTab('logs')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'logs'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300'
            }`}
          >
            <Activity className="h-4 w-4 inline mr-2" />
            System Logs
          </button>
          <button
            onClick={() => setActiveTab('settings')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'settings'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300'
            }`}
          >
            <Settings className="h-4 w-4 inline mr-2" />
            Settings
          </button>
        </nav>
      </div>

      {/* Audit Trail Tab */}
      {activeTab === 'audit' && (
        <div className="space-y-4">
          {/* Filters */}
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-4">
              <Filter className="h-4 w-4" />
              <h3 className="font-medium">Filters</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">User</label>
                <Input
                  placeholder="Filter by user..."
                  value={filters.user}
                  onChange={(e) => handleFilterChange('user', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Category</label>
                <select
                  className="w-full px-3 py-2 border border-input rounded-md bg-background"
                  value={filters.category}
                  onChange={(e) => handleFilterChange('category', e.target.value)}
                >
                  <option value="">All Categories</option>
                  <option value="authentication">Authentication</option>
                  <option value="user_management">User Management</option>
                  <option value="group_management">Group Management</option>
                  <option value="computer_management">Computer Management</option>
                  <option value="system_management">System Management</option>
                  <option value="data_access">Data Access</option>
                  <option value="security">Security</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Action</label>
                <Input
                  placeholder="Filter by action..."
                  value={filters.action}
                  onChange={(e) => handleFilterChange('action', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Status</label>
                <select
                  className="w-full px-3 py-2 border border-input rounded-md bg-background"
                  value={filters.success}
                  onChange={(e) => handleFilterChange('success', e.target.value)}
                >
                  <option value="">All Status</option>
                  <option value="true">Success</option>
                  <option value="false">Failed</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Start Date</label>
                <Input
                  type="date"
                  value={filters.start_date}
                  onChange={(e) => handleFilterChange('start_date', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">End Date</label>
                <Input
                  type="date"
                  value={filters.end_date}
                  onChange={(e) => handleFilterChange('end_date', e.target.value)}
                />
              </div>
            </div>
          </Card>

          {/* Audit Log Entries */}
          <Card>
            <div className="p-4 border-b border-border">
              <div className="flex items-center justify-between">
                <h3 className="font-medium">Audit Trail</h3>
                <div className="text-sm text-muted-foreground">
                  {auditData && `${auditData.total} total entries`}
                </div>
              </div>
            </div>
            <div className="divide-y divide-border">
              {auditLoading ? (
                <div className="p-8 text-center">
                  <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-2" />
                  <p>Loading audit logs...</p>
                </div>
              ) : auditData?.entries.length === 0 ? (
                <div className="p-8 text-center text-muted-foreground">
                  <Eye className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>No audit entries found</p>
                </div>
              ) : (
                auditData?.entries.map((entry, index) => (
                  <div key={index} className="p-4 hover:bg-muted/50">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          {getSuccessIcon(entry.success)}
                          <span className="font-medium">{entry.action}</span>
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${getCategoryColor(entry.category)}`}>
                            {entry.category.replace('_', ' ')}
                          </span>
                          <span className="text-sm text-muted-foreground">
                            {formatTimestamp(entry.timestamp)}
                          </span>
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <User className="h-3 w-3" />
                            {entry.user}
                          </span>
                          <span>Resource: {entry.resource}</span>
                          {entry.ip_address && <span>IP: {entry.ip_address}</span>}
                        </div>
                        {Object.keys(entry.details).length > 0 && (
                          <details className="mt-2">
                            <summary className="text-sm text-muted-foreground cursor-pointer hover:text-foreground">
                              View Details
                            </summary>
                            <pre className="mt-2 p-2 bg-muted rounded text-xs overflow-x-auto">
                              {JSON.stringify(entry.details, null, 2)}
                            </pre>
                          </details>
                        )}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
            
            {/* Pagination */}
            {auditData && auditData.total > 0 && (
              <div className="p-4 border-t border-border flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                  Page {auditData.page} of {Math.ceil(auditData.total / auditData.limit)}
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={auditData.page <= 1}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(p => p + 1)}
                    disabled={!auditData.has_more}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </Card>
        </div>
      )}

      {/* System Logs Tab */}
      {activeTab === 'logs' && (
        <div className="space-y-4">
          <Card className="p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">System Logs</h3>
              <div className="flex items-center gap-2">
                <select
                  className="px-3 py-2 border border-input rounded-md bg-background"
                  value={logType}
                  onChange={(e) => setLogType(e.target.value as any)}
                >
                  <option value="debug">Debug</option>
                  <option value="info">Info</option>
                  <option value="warn">Warning</option>
                  <option value="error">Error</option>
                </select>
                <Button onClick={() => refetchLogs()} disabled={logsLoading}>
                  <RefreshCw className={`h-4 w-4 ${logsLoading ? 'animate-spin' : ''}`} />
                </Button>
              </div>
            </div>
            
            <div className="bg-black text-green-400 p-4 rounded-lg font-mono text-sm max-h-96 overflow-y-auto">
              {logsLoading ? (
                <div className="flex items-center gap-2">
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  Loading logs...
                </div>
              ) : systemLogs?.entries?.length === 0 ? (
                <div className="text-gray-500">No log entries found</div>
              ) : (
                systemLogs?.entries?.map((log: string, index: number) => (
                  <div key={index} className="mb-1">
                    {log}
                  </div>
                ))
              )}
            </div>
          </Card>
        </div>
      )}

      {/* Settings Tab */}
      {activeTab === 'settings' && (
        <div className="space-y-4">
          <Card className="p-6">
            <h3 className="text-lg font-medium mb-4">Logging Configuration</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">Log Level</label>
                <select
                  className="w-full px-3 py-2 border border-input rounded-md bg-background"
                  value={logLevel}
                  onChange={(e) => handleLogLevelChange(e.target.value)}
                >
                  <option value="DEBUG">DEBUG</option>
                  <option value="INFO">INFO</option>
                  <option value="WARN">WARN</option>
                  <option value="ERROR">ERROR</option>
                </select>
                <p className="text-sm text-muted-foreground mt-1">
                  Controls the minimum level of messages to be logged
                </p>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}
