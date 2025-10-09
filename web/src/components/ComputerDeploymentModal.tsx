import { useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/Dialog'
import { Button } from './ui/Button'
import { Card } from './ui/Card'
import { 
  Monitor, 
  Download,
  Copy, 
  CheckCircle, 
  AlertCircle,
  ExternalLink,
  Building,
  Globe,
  Zap,
  XCircle
} from 'lucide-react'
import api from '../lib/api'

interface DeploymentScript {
  id: string
  name: string
  description: string
  icon: string
  enabled: boolean
  requirements: string[]
}

interface ComputerDeploymentModalProps {
  isOpen: boolean
  onClose: () => void
  domainName?: string
  domainController?: string
}

export default function ComputerDeploymentModal({ 
  isOpen, 
  onClose, 
  domainName = 'example.local',
  domainController = 'dc.example.local'
}: ComputerDeploymentModalProps) {
  const [selectedScript, setSelectedScript] = useState<string>('')
  const [generatedCommand, setGeneratedCommand] = useState<string>('')
  const [isGenerating, setIsGenerating] = useState(false)
  const [copied, setCopied] = useState(false)
  const [deploymentScripts, setDeploymentScripts] = useState<DeploymentScript[]>([])
  const [headscaleEnabled, setHeadscaleEnabled] = useState(true)
  const [isLoadingScripts, setIsLoadingScripts] = useState(false)
  const [error, setError] = useState<string>('')

  // Fetch deployment scripts on modal open and reset state on close
  useEffect(() => {
    if (isOpen) {
      fetchDeploymentScripts()
    } else {
      // Reset all state when modal closes
      setSelectedScript('')
      setGeneratedCommand('')
      setError('')
      setCopied(false)
      setIsGenerating(false)
    }
  }, [isOpen])

  // Auto-generate command when script is selected
  useEffect(() => {
    if (selectedScript) {
      // Reset state when changing scripts
      setGeneratedCommand('')
      setError('')
      setCopied(false)
      
      // Generate new command
      generateCommand()
    }
  }, [selectedScript])

  const fetchDeploymentScripts = async () => {
    setIsLoadingScripts(true)
    try {
      const response = await api.get('/deployment/scripts')
      setDeploymentScripts(response.data.scripts)
      setHeadscaleEnabled(response.data.headscale_enabled || false)
    } catch (error) {
      console.error('Failed to fetch deployment scripts:', error)
      // Fallback to default scripts if API fails
      setDeploymentScripts([
        {
          id: 'tailscale-domain',
          name: 'Domain Join with Tailscale',
          description: 'Download Tailscale, join domain, and connect to Tailnet',
          icon: '🔗',
          enabled: true,
          requirements: ['Administrator privileges', 'Network access to domain controller', 'Headscale server configured']
        },
        {
          id: 'domain-only',
          name: 'Domain Join Only',
          description: 'Join computer to domain without Tailscale',
          icon: '🏢',
          enabled: true,
          requirements: ['Administrator privileges', 'Network access to domain controller']
        },
        {
          id: 'tailnet-add',
          name: 'Add to Tailnet',
          description: 'Add existing domain-joined computer to Tailnet',
          icon: '🌐',
          enabled: true,
          requirements: ['Administrator privileges', 'Computer already domain-joined', 'Headscale server configured']
        }
      ])
    } finally {
      setIsLoadingScripts(false)
    }
  }

  const generateCommand = async () => {
    if (!selectedScript) return

    setIsGenerating(true)
    setError('')
    try {
      const response = await api.post('/deployment/generate', {
        script_type: selectedScript,
        domain_name: domainName,
        domain_controller: domainController,
        computer_name: undefined // Let the script auto-detect
      })
      
      setGeneratedCommand(response.data.command)
    } catch (error: any) {
      console.error('Failed to generate command:', error)
      const errorMessage = error?.response?.data?.error || 'Failed to generate deployment command. Please try again.'
      setError(errorMessage)
    } finally {
      setIsGenerating(false)
    }
  }

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(generatedCommand)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      console.error('Failed to copy to clipboard:', error)
    }
  }

  const downloadScript = async () => {
    try {
      // Get the script content from the API
      const response = await api.get(`/deployment/scripts/${getScriptFileName()}`)
      const scriptContent = response.data
      
      // Create blob and download
      const blob = new Blob([scriptContent], { type: 'text/plain' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = getScriptFileName()
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download script:', error)
    }
  }

  const getScriptFileName = () => {
    switch (selectedScript) {
      case 'tailscale-domain': return 'domain-join-with-tailscale.bat'
      case 'domain-only': return 'domain-join-only.bat'
      case 'tailnet-add': return 'tailnet-add.bat'
      default: return 'deployment-script.bat'
    }
  }

  const getScriptIcon = (icon: string) => {
    switch (icon) {
      case '🔗': return <Zap className="h-6 w-6 text-blue-500" />
      case '🏢': return <Building className="h-6 w-6 text-green-500" />
      case '🌐': return <Globe className="h-6 w-6 text-purple-500" />
      default: return <Monitor className="h-6 w-6 text-gray-500" />
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Monitor className="h-6 w-6" />
            Computer Deployment
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          {/* Script Selection */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Choose Deployment Option</h3>
            {!headscaleEnabled && (
              <div className="mb-4 p-3 bg-amber-50 dark:bg-amber-950/20 rounded-lg">
                <div className="flex items-center gap-2 text-amber-700 dark:text-amber-300">
                  <XCircle className="h-4 w-4" />
                  <span className="text-sm font-medium">Headscale not enabled</span>
                </div>
                <p className="text-xs text-amber-600 dark:text-amber-400 mt-1">
                  Tailscale deployment options are disabled. Only domain-only deployment is available.
                </p>
              </div>
            )}
            {isLoadingScripts ? (
              <div className="text-center py-8 text-muted-foreground">
                Loading deployment options...
              </div>
            ) : (
              <div className="grid gap-4 md:grid-cols-3">
                {deploymentScripts.map((script) => (
                  <Card 
                    key={script.id}
                    className={`transition-all ${
                      !script.enabled 
                        ? 'opacity-50 cursor-not-allowed bg-muted/30' 
                        : selectedScript === script.id 
                          ? 'ring-2 ring-primary bg-primary/5 cursor-pointer' 
                          : 'hover:bg-muted/50 cursor-pointer'
                    }`}
                    onClick={() => script.enabled && setSelectedScript(script.id)}
                  >
                    <div className="p-4">
                      <div className="flex items-center gap-3 mb-2">
                        {getScriptIcon(script.icon)}
                        <h4 className="font-medium">{script.name}</h4>
                        {!script.enabled && (
                          <XCircle className="h-4 w-4 text-muted-foreground ml-auto" />
                        )}
                      </div>
                      <p className="text-sm text-muted-foreground mb-3">
                        {script.description}
                      </p>
                      <div className="space-y-1">
                        <p className="text-xs font-medium text-muted-foreground">Requirements:</p>
                        {script.requirements.map((req, index) => (
                          <div key={index} className="flex items-center gap-2 text-xs">
                            <AlertCircle className="h-3 w-3 text-amber-500" />
                            <span>{req}</span>
                          </div>
                        ))}
                      </div>
                      {!script.enabled && (
                        <div className="mt-3 p-2 bg-red-50 dark:bg-red-950/20 rounded">
                          <p className="text-xs text-red-600 dark:text-red-400 font-medium">
                            Not available - Headscale not configured
                          </p>
                        </div>
                      )}
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </div>

          {/* Show info about automatic setup */}
          {selectedScript && (selectedScript === 'tailscale-domain' || selectedScript === 'tailnet-add') && !error && (
            <div className="p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
              <div className="flex items-center gap-2 text-blue-700 dark:text-blue-300">
                <CheckCircle className="h-4 w-4" />
                <span className="text-sm font-medium">Automatic Setup</span>
              </div>
              <p className="text-xs text-blue-600 dark:text-blue-400 mt-1">
                A pre-auth key will be automatically generated and included in the deployment command. No manual configuration needed.
              </p>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="p-4 bg-red-50 dark:bg-red-950/20 rounded-lg">
              <div className="flex items-center gap-2 text-red-700 dark:text-red-300">
                <AlertCircle className="h-5 w-5" />
                <span className="text-sm font-semibold">Error Generating Deployment Command</span>
              </div>
              <p className="text-sm text-red-600 dark:text-red-400 mt-2">
                {error}
              </p>
              {error.includes('infrastructure key') && (
                <div className="mt-3 p-3 bg-red-100 dark:bg-red-900/30 rounded border border-red-200 dark:border-red-800">
                  <p className="text-xs text-red-700 dark:text-red-300 font-medium mb-2">
                    This indicates a problem with the Headscale setup:
                  </p>
                  <ol className="text-xs text-red-600 dark:text-red-400 space-y-1 list-decimal list-inside">
                    <li>Go to the Overlay Networking page</li>
                    <li>Check if Headscale is running and properly configured</li>
                    <li>Re-run the Headscale setup if needed</li>
                    <li>The infrastructure user and pre-auth key should be created automatically during setup</li>
                  </ol>
                </div>
              )}
              <div className="mt-3">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setError('')
                    generateCommand()
                  }}
                  className="text-red-700 dark:text-red-300 border-red-300 dark:border-red-700"
                >
                  Try Again
                </Button>
              </div>
            </div>
          )}

          {/* Loading State */}
          {isGenerating && !error && (
            <div className="p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg text-center">
              <div className="flex items-center justify-center gap-2 text-blue-700 dark:text-blue-300">
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-700 dark:border-blue-300"></div>
                <span className="text-sm font-medium">Generating deployment command...</span>
              </div>
            </div>
          )}

          {/* Generated Script */}
          {generatedCommand && (
            <div>
              <h3 className="text-lg font-semibold mb-4">Deployment Script</h3>
              <Card>
                <div className="p-4">
                  <div className="flex items-center justify-between mb-3">
                    <p className="text-sm font-medium">Download the deployment script:</p>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => downloadScript()}
                      >
                        <Download className="h-4 w-4 mr-2" />
                        Download Script
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={copyToClipboard}
                        disabled={copied}
                      >
                        {copied ? (
                          <>
                            <CheckCircle className="h-4 w-4 mr-2 text-green-500" />
                            Copied!
                          </>
                        ) : (
                          <>
                            <Copy className="h-4 w-4 mr-2" />
                            Copy Command
                          </>
                        )}
                      </Button>
                    </div>
                  </div>
                  <pre className="bg-muted p-3 rounded-lg text-sm overflow-x-auto">
                    <code>{generatedCommand}</code>
                  </pre>
                </div>
              </Card>

              <div className="mt-4 p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
                <h4 className="font-medium text-blue-900 dark:text-blue-100 mb-2 flex items-center gap-2">
                  <ExternalLink className="h-4 w-4" />
                  Instructions
                </h4>
                <ol className="text-sm text-blue-700 dark:text-blue-300 space-y-1">
                  <li>1. Download the batch script to a USB drive</li>
                  <li>2. Copy the script to the target computer</li>
                  <li>3. Double-click the .bat file (UAC will prompt for elevation)</li>
                  <li>4. Follow the on-screen prompts</li>
                </ol>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
