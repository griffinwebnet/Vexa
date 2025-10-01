import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Shield, Lock, Clock, History } from 'lucide-react'
import api from '../lib/api'

type PolicyPreset = 'insecure' | 'standard' | 'enhanced' | 'extreme'

const passwordComplexityPresets = {
  insecure: {
    name: 'Insecure',
    minLength: 6,
    uppercase: false,
    lowercase: false,
    numbers: false,
    symbols: false,
    description: '6 characters, no requirements',
  },
  standard: {
    name: 'Standard',
    minLength: 8,
    uppercase: true,
    lowercase: true,
    numbers: true,
    symbols: true,
    description: '8 chars, upper, lower, number, symbol',
  },
  enhanced: {
    name: 'Enhanced',
    minLength: 12,
    uppercase: true,
    lowercase: true,
    numbers: true,
    symbols: true,
    sequential: false,
    noUsername: true,
    description: '12 chars, no sequential, no username parts',
  },
  extreme: {
    name: 'Extreme',
    minLength: 16,
    uppercase: true,
    lowercase: true,
    numbers: true,
    symbols: true,
    sequential: false,
    noUsername: true,
    description: '16 chars, maximum security',
  },
}

const expirationPresets = {
  insecure: { days: 0, name: 'Never' },
  standard: { days: 365, name: '365 Days' },
  enhanced: { days: 180, name: '180 Days' },
  extreme: { days: 90, name: '90 Days' },
}

const historyPresets = {
  insecure: { count: 0, name: 'No Restriction' },
  standard: { count: 3, name: 'Last 3 Passwords' },
  enhanced: { count: 5, name: 'Last 5 Passwords' },
  extreme: { count: 24, name: 'Never Reuse (24)' },
}

export default function DomainPolicies() {
  const queryClient = useQueryClient()
  const [complexityPreset, setComplexityPreset] = useState<PolicyPreset>('standard')
  const [expirationPreset, setExpirationPreset] = useState<PolicyPreset>('standard')
  const [historyPreset, setHistoryPreset] = useState<PolicyPreset>('standard')

  const { data: policies, isLoading } = useQuery({
    queryKey: ['domainPolicies'],
    queryFn: async () => {
      const response = await api.get('/domain/policies')
      return response.data
    },
  })

  const updatePolicies = useMutation({
    mutationFn: async (data: any) => {
      return await api.put('/domain/policies', data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domainPolicies'] })
    },
  })

  const handleApplyPresets = () => {
    const complexity = passwordComplexityPresets[complexityPreset]
    const expiration = expirationPresets[expirationPreset]
    const history = historyPresets[historyPreset]

    updatePolicies.mutate({
      password_complexity: {
        min_length: complexity.minLength,
        require_uppercase: complexity.uppercase,
        require_lowercase: complexity.lowercase,
        require_numbers: complexity.numbers,
        require_symbols: complexity.symbols,
      },
      password_expiration_days: expiration.days,
      password_history_count: history.count,
    })
  }

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Default Domain Policies</h1>
          <p className="text-muted-foreground">
            Configure password and security policies
          </p>
        </div>
        <Button onClick={handleApplyPresets} disabled={updatePolicies.isPending}>
          {updatePolicies.isPending ? 'Applying...' : 'Apply Changes'}
        </Button>
      </div>

      <div className="grid gap-6">
        {/* Password Complexity */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Lock className="h-5 w-5 text-primary" />
              Password Complexity
            </CardTitle>
            <CardDescription>
              Set minimum requirements for user passwords
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="grid md:grid-cols-4 gap-4">
                {(Object.keys(passwordComplexityPresets) as PolicyPreset[]).map((preset) => {
                  const config = passwordComplexityPresets[preset]
                  return (
                    <div
                      key={preset}
                      className={`p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                        complexityPreset === preset
                          ? 'border-primary bg-primary/5'
                          : 'border-border hover:border-primary/50'
                      }`}
                      onClick={() => setComplexityPreset(preset)}
                    >
                      <div className="font-semibold mb-1">{config.name}</div>
                      <div className="text-sm text-muted-foreground">{config.description}</div>
                    </div>
                  )
                })}
              </div>
              
              <div className="p-4 rounded-lg bg-accent/50">
                <div className="font-medium mb-2">Current Selection: {passwordComplexityPresets[complexityPreset].name}</div>
                <ul className="text-sm space-y-1 text-muted-foreground">
                  <li>• Minimum length: {passwordComplexityPresets[complexityPreset].minLength} characters</li>
                  {passwordComplexityPresets[complexityPreset].uppercase && <li>• Requires uppercase letter</li>}
                  {passwordComplexityPresets[complexityPreset].lowercase && <li>• Requires lowercase letter</li>}
                  {passwordComplexityPresets[complexityPreset].numbers && <li>• Requires number</li>}
                  {passwordComplexityPresets[complexityPreset].symbols && <li>• Requires symbol</li>}
                  {passwordComplexityPresets[complexityPreset].sequential === false && <li>• No sequential characters</li>}
                  {passwordComplexityPresets[complexityPreset].noUsername && <li>• Cannot contain username parts</li>}
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Password Expiration */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-primary" />
              Password Expiration
            </CardTitle>
            <CardDescription>
              How often users must change passwords
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid md:grid-cols-4 gap-4">
              {(Object.keys(expirationPresets) as PolicyPreset[]).map((preset) => {
                const config = expirationPresets[preset]
                return (
                  <div
                    key={preset}
                    className={`p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                      expirationPreset === preset
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                    onClick={() => setExpirationPreset(preset)}
                  >
                    <div className="font-semibold mb-1">{config.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {config.days === 0 ? 'Passwords never expire' : `Expires every ${config.days} days`}
                    </div>
                  </div>
                )
              })}
            </div>
          </CardContent>
        </Card>

        {/* Password History */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <History className="h-5 w-5 text-primary" />
              Password Reuse Policy
            </CardTitle>
            <CardDescription>
              Prevent users from reusing recent passwords
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid md:grid-cols-4 gap-4">
              {(Object.keys(historyPresets) as PolicyPreset[]).map((preset) => {
                const config = historyPresets[preset]
                return (
                  <div
                    key={preset}
                    className={`p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                      historyPreset === preset
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                    onClick={() => setHistoryPreset(preset)}
                  >
                    <div className="font-semibold mb-1">{config.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {config.count === 0 ? 'No restrictions' : `Cannot reuse ${config.name.toLowerCase()}`}
                    </div>
                  </div>
                )
              })}
            </div>
          </CardContent>
        </Card>

      </div>
    </div>
  )
}

