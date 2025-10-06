import { useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/Dialog'
import { Button } from './ui/Button'
import { Card } from './ui/Card'
import { 
  Monitor, 
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
  const [, setIsGenerating] = useState(false)
  const [copied, setCopied] = useState(false)
  const [deploymentScripts, setDeploymentScripts] = useState<DeploymentScript[]>([])
  const [headscaleEnabled, setHeadscaleEnabled] = useState(true)
  const [isLoadingScripts, setIsLoadingScripts] = useState(false)

  // Fetch deployment scripts on modal open
  useEffect(() => {
    if (isOpen) {
      fetchDeploymentScripts()
    }
  }, [isOpen])

  // Auto-generate command when script is selected
  useEffect(() => {
    if (selectedScript && !generatedCommand) {
      generateCommand()
    }
  }, [selectedScript, generatedCommand])

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
    try {
      const response = await api.post('/deployment/generate', {
        script_type: selectedScript,
        domain_name: domainName,
        domain_controller: domainController,
        computer_name: undefined // Let the script auto-detect
      })
      
      setGeneratedCommand(response.data.command)
    } catch (error) {
      console.error('Failed to generate command:', error)
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
          {selectedScript && (selectedScript === 'tailscale-domain' || selectedScript === 'tailnet-add') && (
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

          {/* Generated Command */}
          {generatedCommand && (
            <div>
              <h3 className="text-lg font-semibold mb-4">Deployment Command</h3>
              <Card>
                <div className="p-4">
                  <div className="flex items-center justify-between mb-3">
                    <p className="text-sm font-medium">Copy and run this command in PowerShell as Administrator:</p>
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
                          Copy
                        </>
                      )}
                    </Button>
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
                  <li>1. Copy the command above</li>
                  <li>2. Open PowerShell as Administrator</li>
                  <li>3. Paste and run the command</li>
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
