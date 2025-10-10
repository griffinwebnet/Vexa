import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import api from '@/lib/api'

interface EditUserModalProps {
  open: boolean
  onClose: () => void
  username: string
  onSuccess: () => void
}

export function EditUserModal({ open, onClose, username, onSuccess }: EditUserModalProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [generatedPassword, setGeneratedPassword] = useState('')
  const [formData, setFormData] = useState({
    fullName: '',
    email: '',
    description: '',
    group: '',
    ou: '',
    enabled: true,
    mustChangePassword: false,
  })

  const { data: groups } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const response = await api.get('/groups')
      return response.data
    },
    enabled: open,
  })

  const { data: ous } = useQuery({
    queryKey: ['ous'],
    queryFn: async () => {
      const response = await api.get('/domain/ous')
      return response.data
    },
    enabled: open,
  })

  // Load user data when modal opens
  useEffect(() => {
    if (open && username) {
      loadUserData()
    }
  }, [open, username])

  const loadUserData = async () => {
    try {
      const response = await api.get(`/users/${username}`)
      const userData = response.data
      setFormData({
        fullName: userData.full_name || '',
        email: userData.email || '',
        description: userData.description || '',
        group: userData.primary_group || '',
        ou: userData.ou_path || '',
        enabled: userData.enabled !== false,
        mustChangePassword: userData.must_change_password || false,
      })
    } catch (err) {
      console.error('Failed to load user data:', err)
      setError('Failed to load user data')
    }
  }

  // Flatten OU tree for dropdown
  const flattenOUs = (ou: any, result: any[] = []): any[] => {
    if (ou.path && ou.path !== 'root') {
      result.push({ name: ou.name, path: ou.path, description: ou.description })
    }
    if (ou.children) {
      ou.children.forEach((child: any) => flattenOUs(child, result))
    }
    return result
  }

  const flatOUs = ous ? flattenOUs(ous) : []

  const handleUpdateUser = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      await api.put(`/users/${username}`, {
        full_name: formData.fullName,
        email: formData.email,
        description: formData.description,
        enabled: formData.enabled,
        group: formData.group,
        ou_path: formData.ou,
      })
      onSuccess()
      onClose()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update user')
    } finally {
      setLoading(false)
    }
  }

  const handleResetPassword = async () => {
    setLoading(true)
    setError('')
    setGeneratedPassword('')

    try {
      const response = await api.post(`/users/${username}/reset-password`)
      setGeneratedPassword(response.data.password)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to reset password')
    } finally {
      setLoading(false)
    }
  }

  const handleToggleMustChangePassword = async () => {
    setLoading(true)
    setError('')

    try {
      await api.post(`/users/${username}/toggle-must-change-password`)
      setFormData({ ...formData, mustChangePassword: !formData.mustChangePassword })
      onSuccess()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to toggle must change password flag')
    } finally {
      setLoading(false)
    }
  }

  const handleDisableUser = async () => {
    setLoading(true)
    setError('')

    try {
      if (formData.enabled) {
        await api.post(`/users/${username}/disable`)
      } else {
        await api.post(`/users/${username}/enable`)
      }
      setFormData({ ...formData, enabled: !formData.enabled })
      onSuccess()
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to ${formData.enabled ? 'disable' : 'enable'} user`)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteUser = async () => {
    if (!confirm(`Are you sure you want to delete user "${username}"? This cannot be undone.`)) {
      return
    }

    setLoading(true)
    setError('')

    try {
      await api.delete(`/users/${username}`)
      onSuccess()
      onClose()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete user')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit User: {username}</DialogTitle>
          <DialogDescription>
            Modify user account settings and properties
          </DialogDescription>
        </DialogHeader>

        <form id="edit-user-form" onSubmit={handleUpdateUser} className="space-y-6">
          {/* Generated Password Display */}
          {generatedPassword && (
            <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
              <div className="font-semibold text-green-600 dark:text-green-500 mb-2">
                New Password Generated
              </div>
              <div className="font-mono text-lg bg-background p-3 rounded border select-all">
                {generatedPassword}
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Save this password - it cannot be recovered
              </p>
            </div>
          )}

          {/* Error Display */}
          {error && (
            <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          {/* User Information */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">User Information</h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label htmlFor="fullName" className="text-sm font-medium">
                  Full Name
                </label>
                <Input
                  id="fullName"
                  value={formData.fullName}
                  onChange={(e) => setFormData({ ...formData, fullName: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="email" className="text-sm font-medium">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                />
              </div>

              <div className="space-y-2 col-span-2">
                <label htmlFor="description" className="text-sm font-medium">
                  Description
                </label>
                <Input
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>
            </div>
          </div>

          {/* Group and OU Settings */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Directory Settings</h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label htmlFor="group" className="text-sm font-medium">
                  Primary Group
                </label>
                <select
                  id="group"
                  value={formData.group}
                  onChange={(e) => setFormData({ ...formData, group: e.target.value })}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background"
                >
                  <option value="">Domain Users (default)</option>
                  {groups?.groups?.map((group: any) => (
                    <option key={group.name} value={group.name}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-2">
                <label htmlFor="ou" className="text-sm font-medium">
                  Organizational Unit
                </label>
                <select
                  id="ou"
                  value={formData.ou}
                  onChange={(e) => setFormData({ ...formData, ou: e.target.value })}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background"
                >
                  <option value="">Domain Root (default)</option>
                  {flatOUs.map((ou) => (
                    <option key={ou.path} value={ou.path}>
                      {ou.name} {ou.description ? `- ${ou.description}` : ''}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* Account Settings */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Account Settings</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <label className="text-sm font-medium">Account Status</label>
                  <p className="text-xs text-muted-foreground">
                    {formData.enabled ? 'Account is enabled' : 'Account is disabled'}
                  </p>
                </div>
                <Button
                  type="button"
                  variant={formData.enabled ? "destructive" : "outline"}
                  onClick={handleDisableUser}
                  disabled={loading}
                >
                  {formData.enabled ? 'Disable' : 'Enable'} Account
                </Button>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <label className="text-sm font-medium">Must Change Password</label>
                  <p className="text-xs text-muted-foreground">
                    {formData.mustChangePassword 
                      ? 'User must change password at next login' 
                      : 'User can use current password'
                    }
                  </p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleToggleMustChangePassword}
                  disabled={loading}
                >
                  {formData.mustChangePassword ? 'Remove' : 'Set'} Flag
                </Button>
              </div>
            </div>
          </div>

          {/* Password Management */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Password Management</h3>
            <Button
              type="button"
              variant="outline"
              onClick={handleResetPassword}
              disabled={loading}
              className="w-full"
            >
              Reset Password (Generate New)
            </Button>
          </div>

          {/* Danger Zone */}
          <div className="space-y-4 border-t pt-6">
            <h3 className="text-lg font-medium text-destructive">Danger Zone</h3>
            <Button
              type="button"
              variant="destructive"
              onClick={handleDeleteUser}
              disabled={loading}
              className="w-full"
            >
              Delete User Account
            </Button>
          </div>
        </form>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" form="edit-user-form" disabled={loading}>
            {loading ? 'Saving...' : 'Save Changes'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
