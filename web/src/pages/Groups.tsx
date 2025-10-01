import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { FolderPlus, Users } from 'lucide-react'
import { AddGroupModal } from '../components/modals/AddGroupModal'
import { ManageGroupModal } from '../components/modals/ManageGroupModal'
import api from '../lib/api'

export default function Groups() {
  const queryClient = useQueryClient()
  const [showAddModal, setShowAddModal] = useState(false)
  const [managingGroup, setManagingGroup] = useState<string | null>(null)

  const { data: groupsData, isLoading, refetch } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const response = await api.get('/groups')
      return response.data
    },
  })

  const deleteGroup = useMutation({
    mutationFn: async (groupName: string) => {
      return await api.delete(`/groups/${groupName}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] })
    },
  })

  const handleSaveGroup = async (newName: string) => {
    // TODO: Implement rename via API
    console.log('Rename group:', managingGroup, 'to', newName)
    await queryClient.invalidateQueries({ queryKey: ['groups'] })
  }

  const handleDeleteGroup = () => {
    if (managingGroup) {
      deleteGroup.mutate(managingGroup)
    }
  }

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Groups</h1>
          <p className="text-muted-foreground">
            Manage Active Directory groups
          </p>
        </div>
        <Button onClick={() => setShowAddModal(true)}>
          <FolderPlus className="mr-2 h-4 w-4" />
          Add Group
        </Button>
      </div>

      <AddGroupModal
        open={showAddModal}
        onClose={() => setShowAddModal(false)}
        onSuccess={() => refetch()}
      />

      <ManageGroupModal
        open={managingGroup !== null}
        onClose={() => setManagingGroup(null)}
        groupName={managingGroup || ''}
        onSave={handleSaveGroup}
        onDelete={handleDeleteGroup}
      />

      <Card>
        <CardHeader>
          <CardTitle>Security Groups</CardTitle>
          <CardDescription>
            {groupsData?.count || 0} groups in the directory
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">
              Loading groups...
            </div>
          ) : groupsData?.groups && groupsData.groups.length > 0 ? (
            <div className="space-y-2">
              {groupsData.groups.map((group: any) => (
                <div
                  key={group.name}
                  className="flex items-center justify-between p-4 rounded-lg border hover:bg-accent"
                >
                  <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
                      <Users className="h-5 w-5 text-primary" />
                    </div>
                    <div>
                      <p className="font-medium">{group.name}</p>
                      {group.description && (
                        <p className="text-sm text-muted-foreground">
                          {group.description}
                        </p>
                      )}
                    </div>
                  </div>
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => setManagingGroup(group.name)}
                  >
                    Manage
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              No groups found
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

