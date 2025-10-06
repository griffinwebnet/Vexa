import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Network, Globe, Shield, Info, AlertTriangle, CheckCircle, Loader2 } from 'lucide-react'
import api from '../lib/api'

export default function OverlayNetworking() {
  const queryClient = useQueryClient()
  const [showHeadscaleSetup, setShowHeadscaleSetup] = useState(false)
  const [fqdn, setFqdn] = useState('')
  const [setupError, setSetupError] = useState('')
  const [fqdnTestResult, setFqdnTestResult] = useState<any>(null)
  const [testingFqdn, setTestingFqdn] = useState(false)

  const { data: overlayStatus, isLoading } = useQuery({
    queryKey: ['overlayNetworking'],
    queryFn: async () => {
      const response = await api.get('/system/overlay-status')
      return response.data
    },
  })

  const setupOverlay = useMutation({
    mutationFn: async (data: { fqdn: string }) => {
      const response = await api.post('/system/setup-overlay', data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overlayNetworking'] })
      setShowHeadscaleSetup(false)
      setSetupError('')
    },
    onError: (error: any) => {
      setSetupError(error.response?.data?.error || 'Failed to setup overlay networking')
    },
  })

  const testFqdn = useMutation({
    mutationFn: async (fqdn: string) => {
      const response = await api.post('/system/test-fqdn', { fqdn })
      return response.data
    },
    onSuccess: (data) => {
      setFqdnTestResult(data)
    },
    onError: (error: any) => {
      setFqdnTestResult({
        accessible: false,
        reason: 'Test failed',
        message: error.response?.data?.error || 'Failed to test FQDN',
        can_proceed: false
      })
    },
  })

  const handleSetup = (e: React.FormEvent) => {
    e.preventDefault()
    if (!fqdn.trim()) {
      setSetupError('FQDN is required')
      return
    }
    setupOverlay.mutate({ fqdn })
  }

  const handleTestFqdn = () => {
    if (!fqdn.trim()) {
      setSetupError('Please enter an FQDN to test')
      return
    }
    setTestingFqdn(true)
    setFqdnTestResult(null)
    testFqdn.mutate(fqdn)
  }

  const handleFqdnChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFqdn(e.target.value)
    // Clear test result when FQDN changes
    if (fqdnTestResult) {
      setFqdnTestResult(null)
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

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center gap-3">
        <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
          <Network className="h-5 w-5 text-primary" />
        </div>
        <div>
          <h1 className="text-2xl font-semibold">Overlay Networking</h1>
          <p className="text-muted-foreground">
            Secure mesh VPN for remote access to your domain
          </p>
        </div>
      </div>

      {overlayStatus?.status === 'configured' ? (
        <div className="space-y-6">
          {/* Status Card */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <CheckCircle className="h-5 w-5 text-green-500" />
                <CardTitle>Overlay Network Active</CardTitle>
              </div>
              <CardDescription>
                Your secure mesh VPN is running and ready for remote connections
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="text-sm font-medium">Server FQDN</div>
                  <div className="text-sm text-muted-foreground">{overlayStatus.fqdn}</div>
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-medium">Mesh Domain</div>
                  <div className="text-sm text-muted-foreground">{overlayStatus.mesh_domain}</div>
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-medium">Server IP</div>
                  <div className="text-sm text-muted-foreground">{overlayStatus.server_ip}</div>
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-medium">Headscale Status</div>
                  <div className="text-sm text-muted-foreground">
                    <span className="inline-flex items-center gap-1">
                      <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                      Running
                    </span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Port Forwarding Info */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Globe className="h-5 w-5 text-blue-500" />
                <CardTitle>Remote Access Setup</CardTitle>
              </div>
              <CardDescription>
                Configure port forwarding to allow remote users to join your network
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="p-4 rounded-lg bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800">
                <div className="flex items-start gap-3">
                  <Info className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5" />
                  <div className="space-y-2">
                    <div className="font-medium text-blue-900 dark:text-blue-100">
                      Port Forwarding Required
                    </div>
                    <p className="text-sm text-blue-700 dark:text-blue-300">
                      To allow remote users to connect, forward <strong>port 50443</strong> on your router to this server.
                    </p>
                    <div className="text-sm text-blue-700 dark:text-blue-300">
                      <strong>Forward:</strong> External Port 50443 → Internal Port 50443 → {overlayStatus.server_ip}
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <div className="text-sm font-medium">Benefits of Remote Access:</div>
                <ul className="space-y-1 text-sm text-muted-foreground ml-4">
                  <li>• Remote workers can securely access your domain</li>
                  <li>• Branch offices can join your network</li>
                  <li>• Home users can connect individual devices</li>
                  <li>• All traffic is encrypted through Tailscale</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* Security Notice */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Shield className="h-5 w-5 text-green-500" />
                <CardTitle>Security Features</CardTitle>
              </div>
              <CardDescription>
                Your domain controller remains secure and internal
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="text-sm font-medium">Domain Controller</div>
                  <div className="text-sm text-green-600 dark:text-green-400">
                    ✓ Completely internal - no internet exposure
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-medium">Web UI</div>
                  <div className="text-sm text-green-600 dark:text-green-400">
                    ✓ Local network only
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-medium">Remote Access</div>
                  <div className="text-sm text-green-600 dark:text-green-400">
                    ✓ Encrypted through Tailscale mesh
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-medium">Authentication</div>
                  <div className="text-sm text-green-600 dark:text-green-400">
                    ✓ All remote connections authenticated
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Client Instructions */}
          <Card>
            <CardHeader>
              <CardTitle>Join Remote Devices</CardTitle>
              <CardDescription>
                Instructions for connecting remote devices to your mesh network
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="p-4 rounded-lg bg-muted">
                <div className="font-medium mb-2">For Remote Users:</div>
                <ol className="space-y-2 text-sm text-muted-foreground">
                  <li>1. Install Tailscale on their device</li>
                  <li>2. Use the pre-auth key provided by your admin</li>
                  <li>3. Connect to: <code className="bg-background px-2 py-1 rounded">{overlayStatus.fqdn}:50443</code></li>
                  <li>4. Access your domain resources securely</li>
                </ol>
              </div>
            </CardContent>
          </Card>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Setup Card */}
          <Card>
            <CardHeader>
              <CardTitle>Setup Overlay Network</CardTitle>
              <CardDescription>
                Create a secure mesh VPN for remote access to your domain
              </CardDescription>
            </CardHeader>
            <CardContent>
              {!showHeadscaleSetup ? (
                <div className="space-y-4">
                  <div className="text-sm text-muted-foreground">
                    <p>Create a secure mesh VPN for remote access to your domain.</p>
                  </div>

                  <Button onClick={() => setShowHeadscaleSetup(true)} className="w-full">
                    <Network className="mr-2 h-4 w-4" />
                    Setup Overlay Network
                  </Button>
                </div>
              ) : (
                <form onSubmit={handleSetup} className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Tailscale Public FQDN</label>
                    <div className="flex gap-2">
                      <Input
                        placeholder="vpn.yourcompany.com"
                        value={fqdn}
                        onChange={handleFqdnChange}
                        required
                      />
                      <Button
                        type="button"
                        variant="outline"
                        onClick={handleTestFqdn}
                        disabled={testingFqdn || !fqdn.trim()}
                      >
                        {testingFqdn ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          'Test'
                        )}
                      </Button>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Public domain name for remote Tailscale connections (different from your internal domain)
                    </p>
                  </div>

                  {/* FQDN Test Results */}
                  {fqdnTestResult && (
                    <div className={`p-4 rounded-lg border ${
                      fqdnTestResult.accessible 
                        ? 'bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800' 
                        : fqdnTestResult.can_proceed
                        ? 'bg-yellow-50 dark:bg-yellow-950/20 border-yellow-200 dark:border-yellow-800'
                        : 'bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-800'
                    }`}>
                      <div className="flex items-start gap-3">
                        {fqdnTestResult.accessible ? (
                          <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400 mt-0.5" />
                        ) : fqdnTestResult.can_proceed ? (
                          <AlertTriangle className="h-5 w-5 text-yellow-600 dark:text-yellow-400 mt-0.5" />
                        ) : (
                          <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" />
                        )}
                        <div className="space-y-2">
                          <div className={`font-medium ${
                            fqdnTestResult.accessible 
                              ? 'text-green-900 dark:text-green-100' 
                              : fqdnTestResult.can_proceed
                              ? 'text-yellow-900 dark:text-yellow-100'
                              : 'text-red-900 dark:text-red-100'
                          }`}>
                            {fqdnTestResult.accessible ? 'FQDN is Accessible' : 'FQDN Test Results'}
                          </div>
                          <p className={`text-sm ${
                            fqdnTestResult.accessible 
                              ? 'text-green-700 dark:text-green-300' 
                              : fqdnTestResult.can_proceed
                              ? 'text-yellow-700 dark:text-yellow-300'
                              : 'text-red-700 dark:text-red-300'
                          }`}>
                            {fqdnTestResult.message}
                          </p>
                          <div className={`text-xs ${
                            fqdnTestResult.accessible 
                              ? 'text-green-600 dark:text-green-400' 
                              : fqdnTestResult.can_proceed
                              ? 'text-yellow-600 dark:text-yellow-400'
                              : 'text-red-600 dark:text-red-400'
                          }`}>
                            Reason: {fqdnTestResult.reason}
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* FQDN Explanation - Simplified */}
                  <div className="p-4 rounded-lg bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800">
                    <div className="flex items-start gap-3">
                      <Info className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5" />
                      <div className="text-sm text-blue-700 dark:text-blue-300">
                        <p><strong>Domain Controller:</strong> <code>company.local</code> (internal)</p>
                        <p><strong>Tailscale FQDN:</strong> <code>vpn.company.com</code> (public endpoint)</p>
                        <p className="text-xs mt-1">These should be different domains.</p>
                      </div>
                    </div>
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
                    <Button 
                      type="submit" 
                      disabled={setupOverlay.isPending || testingFqdn}
                    >
                      {setupOverlay.isPending ? 'Setting up...' : 'Setup Overlay Network'}
                    </Button>
                    {fqdnTestResult && !fqdnTestResult.can_proceed && (
                      <Button 
                        type="button" 
                        variant="destructive"
                        onClick={() => setupOverlay.mutate({ fqdn })}
                        disabled={setupOverlay.isPending}
                      >
                        Proceed Anyway
                      </Button>
                    )}
                  </div>
                </form>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
