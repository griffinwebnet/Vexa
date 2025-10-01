import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Network, Globe, Shield } from 'lucide-react'
import api from '../lib/api'

export default function Settings() {
  const queryClient = useQueryClient()
  const [overlayEnabled, setOverlayEnabled] = useState(false)
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
                  <li>✓ Configure split DNS for your AD domain</li>
                  <li>✓ Create mesh domain ({fqdn ? fqdn.split('.')[0] : 'example'}.mesh)</li>
                  <li>✓ Install Tailscale client on this server (100.64.0.1)</li>
                  <li>✓ Generate 15-year pre-auth key for joining</li>
                  <li>✓ Route all AD traffic through secure mesh network</li>
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
                  to join your Active Directory domain without exposing all AD ports to the internet.
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
    </div>
  )
}

