import { useState, useEffect } from 'react'
import { useAuthStore } from '../stores/authStore'
import { Card } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../components/ui/Dialog'
import { User, Key, Shield, CheckCircle, AlertCircle } from 'lucide-react'
import api from '../lib/api'

export default function SelfService() {
  const { username } = useAuthStore()
  const [isChangePasswordOpen, setIsChangePasswordOpen] = useState(false)
  const [isUpdateProfileOpen, setIsUpdateProfileOpen] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  
  // Form states
  const [passwordForm, setPasswordForm] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
  
  const [profileForm, setProfileForm] = useState({
    fullName: '',
    email: ''
  })

  // Load current profile
  useEffect(() => {
    const loadProfile = async () => {
      try {
        const response = await api.get(`/users/${username}`)
        setProfileForm({
          fullName: response.data.full_name || '',
          email: response.data.email || ''
        })
      } catch (err) {
        console.error('Failed to load profile:', err)
      }
    }
    loadProfile()
  }, [username])

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')

    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      setError('New passwords do not match')
      return
    }

    try {
      const response = await api.post('/users/change-password', {
        current_password: passwordForm.currentPassword,
        new_password: passwordForm.newPassword
      })
      
      // If a new token is returned, update it in localStorage
      if (response.data.token) {
        localStorage.setItem('token', response.data.token)
      }
      
      setSuccess('Password changed successfully!')
      setIsChangePasswordOpen(false)
      setPasswordForm({ currentPassword: '', newPassword: '', confirmPassword: '' })
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to change password')
    }
  }

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')

    try {
      await api.post('/users/update-profile', {
        full_name: profileForm.fullName,
        email: profileForm.email
      })
      setSuccess('Profile updated successfully!')
      setIsUpdateProfileOpen(false)
      
      // Reload profile to show updated information
      try {
        const response = await api.get(`/users/${username}`)
        setProfileForm({
          fullName: response.data.user?.full_name || '',
          email: response.data.user?.email || ''
        })
      } catch (error) {
        console.error('Failed to reload profile:', error)
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update profile')
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Self Service</h1>
        <p className="text-muted-foreground">
          Manage your account settings
        </p>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-destructive/15 border border-destructive/20">
          <div className="flex items-center gap-2 text-sm text-destructive">
            <AlertCircle className="h-5 w-5" />
            <span>{error}</span>
          </div>
        </div>
      )}

      {success && (
        <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
          <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
            <CheckCircle className="h-5 w-5" />
            <span>{success}</span>
          </div>
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card className="p-4 flex flex-col items-center justify-center text-center space-y-4">
          <User className="h-12 w-12 text-primary" />
          <h3 className="text-lg font-semibold">Account Information</h3>
          <p className="text-sm text-muted-foreground">
            Update your name and contact details
          </p>
          <Button onClick={() => setIsUpdateProfileOpen(true)}>
            Update Profile
          </Button>
        </Card>

        <Card className="p-4 flex flex-col items-center justify-center text-center space-y-4">
          <Key className="h-12 w-12 text-primary" />
          <h3 className="text-lg font-semibold">Password Management</h3>
          <p className="text-sm text-muted-foreground">
            Change your account password
          </p>
          <Button onClick={() => setIsChangePasswordOpen(true)}>
            Change Password
          </Button>
        </Card>

        <Card className="p-4 flex flex-col items-center justify-center text-center space-y-4">
          <Shield className="h-12 w-12 text-primary" />
          <h3 className="text-lg font-semibold">Security Settings</h3>
          <p className="text-sm text-muted-foreground">
            View your security status
          </p>
          <div className="text-sm">
            <div className="flex items-center gap-2">
              <span className="font-medium">Username:</span>
              <span className="text-muted-foreground">{username}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="font-medium">Last Login:</span>
              <span className="text-muted-foreground">Today</span>
            </div>
          </div>
        </Card>
      </div>

      {/* Change Password Dialog */}
      <Dialog open={isChangePasswordOpen} onOpenChange={setIsChangePasswordOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Change Password</DialogTitle>
          </DialogHeader>
          <form onSubmit={handlePasswordChange} className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Current Password</label>
              <Input
                type="password"
                value={passwordForm.currentPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, currentPassword: e.target.value }))}
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">New Password</label>
              <Input
                type="password"
                value={passwordForm.newPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, newPassword: e.target.value }))}
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Confirm New Password</label>
              <Input
                type="password"
                value={passwordForm.confirmPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, confirmPassword: e.target.value }))}
                required
              />
            </div>
            <Button type="submit" className="w-full">
              Change Password
            </Button>
          </form>
        </DialogContent>
      </Dialog>

      {/* Update Profile Dialog */}
      <Dialog open={isUpdateProfileOpen} onOpenChange={setIsUpdateProfileOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update Profile</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleProfileUpdate} className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Full Name</label>
              <Input
                type="text"
                value={profileForm.fullName}
                onChange={(e) => setProfileForm(prev => ({ ...prev, fullName: e.target.value }))}
                placeholder="Enter your full name"
              />
              <p className="text-xs text-muted-foreground">Leave blank to keep current value</p>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Email Address</label>
              <Input
                type="email"
                value={profileForm.email}
                onChange={(e) => setProfileForm(prev => ({ ...prev, email: e.target.value }))}
                placeholder="Enter your email address"
              />
              <p className="text-xs text-muted-foreground">Leave blank to keep current value</p>
            </div>
            <Button type="submit" className="w-full">
              Update Profile
            </Button>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}