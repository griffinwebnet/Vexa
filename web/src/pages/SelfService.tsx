import { useState } from 'react'
import { useAuthStore } from '../stores/authStore'
import { Card } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../components/ui/Dialog'
import { User, Key, Shield, Mail } from 'lucide-react'

export default function SelfService() {
  const { username } = useAuthStore()
  const [isChangePasswordOpen, setIsChangePasswordOpen] = useState(false)
  const [isUpdateProfileOpen, setIsUpdateProfileOpen] = useState(false)
  
  // Form states
  const [passwordForm, setPasswordForm] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
  
  const [profileForm, setProfileForm] = useState({
    fullName: 'John Smith',
    email: 'john.smith@example.com'
  })

  const handlePasswordChange = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Implement password change API call
    console.log('Password change requested:', passwordForm)
    setIsChangePasswordOpen(false)
    setPasswordForm({ currentPassword: '', newPassword: '', confirmPassword: '' })
  }

  const handleProfileUpdate = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Implement profile update API call
    console.log('Profile update requested:', profileForm)
    setIsUpdateProfileOpen(false)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Self Service Portal</h1>
        <p className="text-muted-foreground">
          Manage your account settings and preferences
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Account Information */}
        <Card>
          <div className="p-6">
            <div className="flex items-center gap-3 mb-4">
              <User className="h-6 w-6 text-primary" />
              <h2 className="text-xl font-semibold">Account Information</h2>
            </div>
            
            <div className="space-y-3">
              <div>
                <label className="text-sm font-medium text-muted-foreground">Username</label>
                <p className="text-lg">{username}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">Full Name</label>
                <p className="text-lg">{profileForm.fullName}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">Email</label>
                <p className="text-lg">{profileForm.email}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">Role</label>
                <p className="text-lg">Domain User</p>
              </div>
            </div>

            <Button 
              className="mt-4" 
              variant="outline"
              onClick={() => setIsUpdateProfileOpen(true)}
            >
              <Mail className="h-4 w-4 mr-2" />
              Update Profile
            </Button>

            <Dialog open={isUpdateProfileOpen} onOpenChange={setIsUpdateProfileOpen}>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Update Profile Information</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleProfileUpdate} className="space-y-4">
                  <div>
                    <label className="text-sm font-medium">Full Name</label>
                    <Input
                      value={profileForm.fullName}
                      onChange={(e) => setProfileForm({ ...profileForm, fullName: e.target.value })}
                      placeholder="Enter your full name"
                    />
                  </div>
                  <div>
                    <label className="text-sm font-medium">Email Address</label>
                    <Input
                      type="email"
                      value={profileForm.email}
                      onChange={(e) => setProfileForm({ ...profileForm, email: e.target.value })}
                      placeholder="Enter your email address"
                    />
                  </div>
                  <div className="flex gap-2 justify-end">
                    <Button type="button" variant="outline" onClick={() => setIsUpdateProfileOpen(false)}>
                      Cancel
                    </Button>
                    <Button type="submit">Update Profile</Button>
                  </div>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        </Card>

        {/* Security */}
        <Card>
          <div className="p-6">
            <div className="flex items-center gap-3 mb-4">
              <Shield className="h-6 w-6 text-primary" />
              <h2 className="text-xl font-semibold">Security</h2>
            </div>
            
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium">Password</h3>
                  <p className="text-sm text-muted-foreground">
                    Last changed 30 days ago
                  </p>
                </div>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={() => setIsChangePasswordOpen(true)}
                >
                  <Key className="h-4 w-4 mr-2" />
                  Change Password
                </Button>

                <Dialog open={isChangePasswordOpen} onOpenChange={setIsChangePasswordOpen}>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Change Password</DialogTitle>
                    </DialogHeader>
                    <form onSubmit={handlePasswordChange} className="space-y-4">
                      <div>
                        <label className="text-sm font-medium">Current Password</label>
                        <Input
                          type="password"
                          value={passwordForm.currentPassword}
                          onChange={(e) => setPasswordForm({ ...passwordForm, currentPassword: e.target.value })}
                          placeholder="Enter current password"
                          required
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium">New Password</label>
                        <Input
                          type="password"
                          value={passwordForm.newPassword}
                          onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
                          placeholder="Enter new password"
                          required
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium">Confirm New Password</label>
                        <Input
                          type="password"
                          value={passwordForm.confirmPassword}
                          onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
                          placeholder="Confirm new password"
                          required
                        />
                      </div>
                      <div className="flex gap-2 justify-end">
                        <Button type="button" variant="outline" onClick={() => setIsChangePasswordOpen(false)}>
                          Cancel
                        </Button>
                        <Button type="submit">Change Password</Button>
                      </div>
                    </form>
                  </DialogContent>
                </Dialog>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium">Two-Factor Authentication</h3>
                  <p className="text-sm text-muted-foreground">
                    Add an extra layer of security
                  </p>
                </div>
                <Button variant="outline" size="sm" disabled>
                  Enable 2FA
                </Button>
              </div>
            </div>
          </div>
        </Card>

        {/* Group Memberships */}
        <Card>
          <div className="p-6">
            <div className="flex items-center gap-3 mb-4">
              <User className="h-6 w-6 text-primary" />
              <h2 className="text-xl font-semibold">Group Memberships</h2>
            </div>
            
            <div className="space-y-2">
              <div className="flex items-center justify-between py-2 px-3 bg-muted rounded-lg">
                <span className="font-medium">Domain Users</span>
                <span className="text-sm text-muted-foreground">Primary Group</span>
              </div>
              <div className="flex items-center justify-between py-2 px-3 bg-muted rounded-lg">
                <span className="font-medium">IT Staff</span>
                <span className="text-sm text-muted-foreground">Member</span>
              </div>
              <div className="flex items-center justify-between py-2 px-3 bg-muted rounded-lg">
                <span className="font-medium">Sales</span>
                <span className="text-sm text-muted-foreground">Member</span>
              </div>
            </div>
            
            <p className="text-sm text-muted-foreground mt-4">
              Contact your administrator to request group membership changes.
            </p>
          </div>
        </Card>

        {/* Quick Actions */}
        <Card>
          <div className="p-6">
            <div className="flex items-center gap-3 mb-4">
              <Shield className="h-6 w-6 text-primary" />
              <h2 className="text-xl font-semibold">Quick Actions</h2>
            </div>
            
            <div className="space-y-3">
              <Button className="w-full justify-start" variant="outline">
                <Mail className="h-4 w-4 mr-2" />
                Update Email Address
              </Button>
              <Button className="w-full justify-start" variant="outline">
                <Key className="h-4 w-4 mr-2" />
                Change Password
              </Button>
              <Button className="w-full justify-start" variant="outline" disabled>
                <Shield className="h-4 w-4 mr-2" />
                Enable Two-Factor Auth
              </Button>
            </div>
            
            <div className="mt-6 p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
              <h4 className="font-medium text-blue-900 dark:text-blue-100 mb-2">
                Need Help?
              </h4>
              <p className="text-sm text-blue-700 dark:text-blue-300">
                Contact your IT administrator for assistance with account management or security settings.
              </p>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}
