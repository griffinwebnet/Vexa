import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Network, Globe, Shield, Info, Download, ExternalLink } from 'lucide-react'
import api from '../lib/api'
import type { UpdateInfo } from '../types/updates'

export default function Settings() {
  const queryClient = useQueryClient()
  const [showHeadscaleSetup, setShowHeadscaleSetup] = useState(false)
  const [fqdn, setFqdn] = useState('')
  const [setupError, setSetupError] = useState('')

  const { data: overlayStatus, isLoading } = useQuery({
    queryKey: ['overlayNetworking'],
    queryFn: async () => {
      const response = await api.get('/system/overlay-status')
      return response.data
    },
  })


  const { data: updateInfo, isLoading: isUpdateLoading } = useQuery<UpdateInfo>({
    queryKey: ['updateCheck'],
    queryFn: async () => {
      const response = await api.get('/updates/check')
      console.log('Update check response:', response.data)
      console.log('Latest version:', response.data?.latest_version)
      console.log('Status:', response.data?.status)
      console.log('Versions:', response.data?.versions)
      return response.data
    },
    refetchInterval: 1000 * 60 * 60, // Check every hour
  })

  const [showUpgradeModal, setShowUpgradeModal] = useState(false)
  const [upgradeInProgress, setUpgradeInProgress] = useState(false)
  const [upgradeStatus, setUpgradeStatus] = useState('')

  const performUpgrade = useMutation({
    mutationFn: async () => {
      return await api.post('/updates/upgrade')
    },
    onSuccess: () => {
      setUpgradeStatus('Upgrade completed successfully! Redirecting to login...')
      setTimeout(() => {
        // Redirect to login
        window.location.href = '/login'
      }, 3000)
    },
    onError: (error: any) => {
      setUpgradeStatus(`Upgrade failed: ${error.response?.data?.error || 'Unknown error'}`)
    },
  })

  const handleUpgrade = () => {
    setShowUpgradeModal(true)
  }

  const confirmUpgrade = () => {
    setShowUpgradeModal(false)
    setUpgradeInProgress(true)
    setUpgradeStatus('Starting upgrade process...')
    performUpgrade.mutate()
  }

  const setupOverlay = useMutation({
    mutationFn: async (config: { fqdn: string }) => {
      return await api.post('/system/setup-overlay', config)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overlayNetworking'] })
      setShowHeadscaleSetup(false)
    },
    onError: (error: any) => {
      setSetupError(error.response?.data?.error || 'Failed to setup overlay networking')
    },
  })

  const handleEnableOverlay = () => {
    if (overlayStatus?.enabled) {
      // Already enabled
      return
    }
    setShowHeadscaleSetup(true)
  }

  const handleSetupSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSetupError('')
    setupOverlay.mutate({ fqdn })
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground">
          Configure Vexa system settings
        </p>
      </div>

      {/* System Information */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5 text-primary" />
            System Information
          </CardTitle>
          <CardDescription>
            Version information for Vexa components
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {/* Status Badge */}
            <div className="p-3 rounded-lg bg-accent flex items-center justify-between">
              <div className="flex items-center gap-2">
                {isUpdateLoading ? (
                  <>
                    <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-primary"></div>
                    <span className="font-semibold text-muted-foreground">Checking for updates...</span>
                  </>
                ) : (
                  <>
                    <Shield className={`h-5 w-5 ${
                      updateInfo?.error ? 'text-yellow-600 dark:text-yellow-500' :
                      updateInfo?.status === 'Development Version' ? 'text-blue-600 dark:text-blue-500' :
                      updateInfo?.status === 'Update Available' ? 'text-red-600 dark:text-red-500' :
                      'text-green-600 dark:text-green-500'
                    }`} />
                    <span className={`font-semibold ${
                      updateInfo?.error ? 'text-yellow-600 dark:text-yellow-500' :
                      updateInfo?.status === 'Development Version' ? 'text-blue-600 dark:text-blue-500' :
                      updateInfo?.status === 'Update Available' ? 'text-red-600 dark:text-red-500' :
                      'text-green-600 dark:text-green-500'
                    }`}>
                      {updateInfo?.error ? 'Update Check Failed' :
                       updateInfo?.status || 'Up to Date'}
                    </span>
                  </>
                )}
              </div>
              {updateInfo?.update_available && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleUpgrade}
                  disabled={upgradeInProgress}
                  className="h-7 px-2 ml-4"
                >
                  <Download className="h-4 w-4" />
                  <span className="ml-1 text-xs">Update</span>
                </Button>
              )}
            </div>

            {/* Version Information */}
            <div className="space-y-2">
              {isUpdateLoading ? (
                <div className="text-center text-muted-foreground">Loading version information...</div>
              ) : (
                <>
                  <div className="flex justify-between items-center">
                    <span className="font-medium">Latest Release:</span>
                    <span className="text-muted-foreground font-mono">Version {updateInfo?.latest_version || 'Unknown'}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="font-medium">Current API Version:</span>
                    <span className="text-muted-foreground font-mono">Version {updateInfo?.versions?.find(v => v.component === 'api')?.version || 'Unknown'}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="font-medium">Current UI Version:</span>
                    <span className="text-muted-foreground font-mono">Version {updateInfo?.versions?.find(v => v.component === 'web')?.version || 'Unknown'}</span>
                  </div>
                </>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Overlay Networking */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Network className="h-5 w-5 text-primary" />
            Overlay Networking (Headscale)
          </CardTitle>
          <CardDescription>
            Secure mesh VPN for remote domain access
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-4 text-muted-foreground">Loading...</div>
          ) : overlayStatus?.enabled ? (
            <div className="space-y-4">
              <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
                <div className="flex items-center gap-2 mb-2">
                  <Shield className="h-5 w-5 text-green-600 dark:text-green-500" />
                  <span className="font-semibold text-green-600 dark:text-green-500">
                    Overlay Networking Enabled
                  </span>
                </div>
                <div className="text-sm space-y-2">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Public FQDN:</span>
                    <span className="font-mono">{overlayStatus.fqdn}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Mesh Domain:</span>
                    <span className="font-mono">{overlayStatus.mesh_domain}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Internal IP:</span>
                    <span className="font-mono">100.64.0.1</span>
                  </div>
                </div>
              </div>

              <div className="p-4 rounded-lg bg-accent">
                <div className="font-medium mb-2">Management URL:</div>
                <a 
                  href={`http://${overlayStatus.fqdn}:8080`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline font-mono text-sm"
                >
                  http://{overlayStatus.fqdn}:8080
                </a>
              </div>

              <Button variant="destructive" disabled>
                Disable Overlay Networking
              </Button>
            </div>
          ) : showHeadscaleSetup ? (
            <form onSubmit={handleSetupSubmit} className="space-y-4">
              <div className="p-4 rounded-lg bg-yellow-500/10 border border-yellow-500/20 text-sm">
                <div className="font-semibold mb-2">Firewall Requirements:</div>
                <ul className="space-y-1 text-muted-foreground">
                  <li>• <strong>Port 8080/tcp</strong> - Headscale control plane (HTTP)</li>
                </ul>
                <p className="text-xs text-muted-foreground mt-2">
                  Uses public Tailscale DERP relays - more resilient if your server goes down
                </p>
              </div>

              <div className="space-y-2">
                <label htmlFor="fqdn" className="text-sm font-medium">
                  Public FQDN *
                </label>
                <Input
                  id="fqdn"
                  type="text"
                  placeholder="vpn.example.com"
                  value={fqdn}
                  onChange={(e) => setFqdn(e.target.value)}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  The public domain name that points to this server
                </p>
              </div>

              <div className="p-4 rounded-lg bg-accent text-sm space-y-2">
                <div className="font-medium mb-2">What will be configured:</div>
                <ul className="space-y-1 text-muted-foreground">
                  <li>✓ Install Headscale (self-hosted Tailscale coordinator)</li>
                  <li>✓ Configure split DNS for your Vexa Domain</li>
                  <li>✓ Create mesh domain ({fqdn ? fqdn.split('.')[0] : 'example'}.mesh)</li>
                  <li>✓ Install Tailscale client on this server (100.64.0.1)</li>
                  <li>✓ Generate 15-year pre-auth key for joining devices</li>
                  <li>✓ Route all domain traffic through secure mesh network</li>
                </ul>
              </div>

              {setupError && (
                <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
                  {setupError}
                </div>
              )}

              <div className="flex gap-2">
                <Button 
                  type="button" 
                  variant="outline" 
                  onClick={() => setShowHeadscaleSetup(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={setupOverlay.isPending}>
                  {setupOverlay.isPending ? 'Setting up...' : 'Setup Overlay Network'}
                </Button>
              </div>
            </form>
          ) : (
            <div className="space-y-4">
              <div className="text-sm text-muted-foreground">
                <p className="mb-4">
                  Overlay networking creates a secure mesh VPN that allows remote users and sites 
                  to join your Vexa Domain without exposing all AD ports to the internet.
                </p>
                <div className="p-4 rounded-lg bg-accent">
                  <div className="font-medium mb-2">Benefits:</div>
                  <ul className="space-y-1">
                    <li>• Zero-trust network access</li>
                    <li>• Only 1 port needed (vs 10+ for traditional AD)</li>
                    <li>• Automatic encryption and authentication</li>
                    <li>• NAT traversal - works anywhere</li>
                    <li>• Uses public DERP relays for resilience</li>
                  </ul>
                </div>
              </div>
              <Button onClick={handleEnableOverlay}>
                <Globe className="mr-2 h-4 w-4" />
                Enable Overlay Networking
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Other Settings */}
      <Card>
        <CardHeader>
          <CardTitle>System Configuration</CardTitle>
          <CardDescription>
            Additional system settings
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            Additional settings coming soon
          </div>
        </CardContent>
      </Card>

      {/* Upgrade Confirmation Modal */}
      {showUpgradeModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background border rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold mb-4">Confirm Upgrade</h3>
            <div className="space-y-4">
              <div className="p-4 rounded-lg bg-yellow-500/10 border border-yellow-500/20">
                <div className="flex items-center gap-2 mb-2">
                  <Shield className="h-5 w-5 text-yellow-600 dark:text-yellow-500" />
                  <span className="font-semibold text-yellow-600 dark:text-yellow-500">
                    Important Notice
                  </span>
                </div>
                <div className="text-sm space-y-2">
                  <p>• The upgrade process may take up to 15 minutes (usually less than 5)</p>
                  <p>• Login and web UI services will be unavailable during the upgrade</p>
                  <p>• The system will automatically restart all services when complete</p>
                  <p>• You will be redirected to the login screen after completion</p>
                </div>
              </div>
              <div className="flex gap-2 justify-end">
                <Button
                  variant="outline"
                  onClick={() => setShowUpgradeModal(false)}
                >
                  Cancel
                </Button>
                <Button
                  onClick={confirmUpgrade}
                  className="bg-green-600 hover:bg-green-700"
                >
                  <Download className="mr-2 h-4 w-4" />
                  Start Upgrade
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Upgrade Progress Modal */}
      {upgradeInProgress && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background border rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold mb-4">Upgrading System</h3>
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                <span className="text-sm">{upgradeStatus}</span>
              </div>
              <div className="space-y-2 text-sm text-muted-foreground">
                <div>📦 Updating base system packages...</div>
                <div>🔧 Updating core system dependencies...</div>
                <div>📥 Updating main application...</div>
                <div>🔨 Rebuilding components...</div>
                <div>🔄 Restarting services...</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

