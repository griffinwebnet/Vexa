import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import api from '@/lib/api'

interface AddUserModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export function AddUserModal({ open, onClose, onSuccess }: AddUserModalProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    confirmPassword: '',
    fullName: '',
    email: '',
    description: '',
    group: '',
    ou: '',
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match')
      return
    }

    setLoading(true)
    try {
      await api.post('/users', {
        username: formData.username,
        password: formData.password,
        full_name: formData.fullName,
        email: formData.email,
        description: formData.description,
        group: formData.group,
        ou_path: formData.ou,
      })
      onSuccess()
      onClose()
      setFormData({
        username: '',
        password: '',
        confirmPassword: '',
        fullName: '',
        email: '',
        description: '',
        group: '',
        ou: '',
      })
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create user')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add New User</DialogTitle>
          <DialogDescription>
            Create a new Active Directory user account
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label htmlFor="username" className="text-sm font-medium">
                Username *
              </label>
              <Input
                id="username"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                required
              />
            </div>

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
              <label htmlFor="password" className="text-sm font-medium">
                Password *
              </label>
              <Input
                id="password"
                type="password"
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                required
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-sm font-medium">
                Confirm Password *
              </label>
              <Input
                id="confirmPassword"
                type="password"
                value={formData.confirmPassword}
                onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                required
              />
            </div>

            <div className="space-y-2 col-span-2">
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
              <p className="text-xs text-muted-foreground">
                Where to place this user in the directory
              </p>
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

          {error && (
            <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Creating...' : 'Create User'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

