import { useState } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import api from '@/lib/api'

interface ManageUserModalProps {
  open: boolean
  onClose: () => void
  username: string
  onSuccess: () => void
}

export function ManageUserModal({ open, onClose, username, onSuccess }: ManageUserModalProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [generatedPassword, setGeneratedPassword] = useState('')
  const [mustChangePassword, setMustChangePassword] = useState(false)

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

  const handleDisableUser = async () => {
    setLoading(true)
    setError('')

    try {
      await api.post(`/users/${username}/disable`)
      onSuccess()
      onClose()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to disable user')
    } finally {
      setLoading(false)
    }
  }

  const handleToggleMustChangePassword = async () => {
    setLoading(true)
    setError('')

    try {
      await api.post(`/users/${username}/toggle-must-change-password`)
      setMustChangePassword(!mustChangePassword)
      onSuccess()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to toggle must change password flag')
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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Manage User: {username}</DialogTitle>
          <DialogDescription>
            Perform actions on this user account
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
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

          {error && (
            <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={handleResetPassword}
              disabled={loading}
            >
              Reset Password
            </Button>
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={handleToggleMustChangePassword}
              disabled={loading}
            >
              {mustChangePassword ? 'Remove' : 'Set'} "Must Change Password at Next Login"
            </Button>
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={handleDisableUser}
              disabled={loading}
            >
              Disable User
            </Button>
            <Button
              variant="destructive"
              className="w-full justify-start"
              onClick={handleDeleteUser}
              disabled={loading}
            >
              Delete User
            </Button>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

