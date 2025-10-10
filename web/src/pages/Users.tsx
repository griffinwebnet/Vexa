import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { UserPlus } from 'lucide-react'
import { AddUserModal } from '../components/modals/AddUserModal'
import { EditUserModal } from '../components/modals/EditUserModal'
import api from '../lib/api'

export default function Users() {
  const [showAddModal, setShowAddModal] = useState(false)
  const [editingUser, setEditingUser] = useState<string | null>(null)
  
  const { data: usersData, isLoading, refetch } = useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await api.get('/users')
      return response.data
    },
  })

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
        <h1 className="text-3xl font-bold">Manage Directory Users</h1>
        <p className="text-muted-foreground">
          AD-compatible users
        </p>
        </div>
        <Button onClick={() => setShowAddModal(true)}>
          <UserPlus className="mr-2 h-4 w-4" />
          Add User
        </Button>
      </div>

      <AddUserModal
        open={showAddModal}
        onClose={() => setShowAddModal(false)}
        onSuccess={() => refetch()}
      />

      <EditUserModal
        open={editingUser !== null}
        onClose={() => setEditingUser(null)}
        username={editingUser || ''}
        onSuccess={() => refetch()}
      />

      <Card>
        <CardHeader>
          <CardTitle>User List</CardTitle>
          <CardDescription>
            {usersData?.count || 0} users in the directory
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">
              Loading users...
            </div>
          ) : usersData?.users && usersData.users.length > 0 ? (
            <div className="space-y-2">
              {usersData.users.map((user: any) => (
                <div
                  key={user.username}
                  className="flex items-center justify-between p-4 rounded-lg border hover:bg-accent"
                >
                  <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-medium">
                      {user.username.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <p className="font-medium">{user.username}</p>
                      {user.full_name && (
                        <p className="text-sm text-muted-foreground">
                          {user.full_name}
                        </p>
                      )}
                    </div>
                  </div>
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => setEditingUser(user.username)}
                  >
                    Edit
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              No users found
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

