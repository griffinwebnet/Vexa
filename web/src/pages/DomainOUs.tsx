import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { FolderTree, Plus, ChevronRight, Building2 } from 'lucide-react'
import { EditOUModal } from '../components/modals/EditOUModal'
import { AddOUModal } from '../components/modals/AddOUModal'
import api from '../lib/api'

export default function DomainOUs() {
  const queryClient = useQueryClient()
  const [expandedOUs, setExpandedOUs] = useState<Set<string>>(new Set(['root']))
  const [editingOU, setEditingOU] = useState<{ name: string; path: string } | null>(null)
  const [showAddModal, setShowAddModal] = useState(false)

  const { data: ouData, isLoading } = useQuery({
    queryKey: ['ous'],
    queryFn: async () => {
      const response = await api.get('/domain/ous')
      return response.data
    },
  })

  // Default structure if no data
  const defaultStructure = {
    name: 'Domain',
    path: 'root',
    children: [
      {
        name: 'Domain Controllers',
        path: 'OU=Domain Controllers',
        description: 'Default controllers container',
        children: [],
      },
    ],
  }

  const toggleOU = (path: string) => {
    const newExpanded = new Set(expandedOUs)
    if (newExpanded.has(path)) {
      newExpanded.delete(path)
    } else {
      newExpanded.add(path)
    }
    setExpandedOUs(newExpanded)
  }

  const renderOUTree = (ou: any, level: number = 0) => {
    const hasChildren = ou.children && ou.children.length > 0
    const isExpanded = expandedOUs.has(ou.path || 'root')

    return (
      <div key={ou.path || 'root'} className={level > 0 ? 'ml-6' : ''}>
        <div
          className={`flex items-center gap-2 p-3 rounded-lg hover:bg-accent cursor-pointer ${
            level === 0 ? 'bg-primary/10 font-semibold' : 'border mb-2'
          }`}
          onClick={() => hasChildren && toggleOU(ou.path || 'root')}
        >
          {hasChildren && (
            <ChevronRight
              className={`h-4 w-4 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
            />
          )}
          {!hasChildren && <div className="w-4" />}
          <Building2 className="h-4 w-4 text-primary" />
          <div className="flex-1">
            <div className="font-medium">{ou.name}</div>
            {ou.description && (
              <div className="text-xs text-muted-foreground">{ou.description}</div>
            )}
          </div>
          {level > 0 && (
            <Button variant="ghost" size="sm" onClick={(e) => {
              e.stopPropagation()
              setEditingOU({ name: ou.name, path: ou.path })
            }}>
              Edit
            </Button>
          )}
        </div>
        {hasChildren && isExpanded && (
          <div className="mt-2">
            {ou.children.map((child: any) => renderOUTree(child, level + 1))}
          </div>
        )}
      </div>
    )
  }

  const createOU = useMutation({
    mutationFn: async (data: { name: string; description: string; parentPath: string }) => {
      return await api.post('/domain/ous', data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ous'] })
    },
  })

  const deleteOU = useMutation({
    mutationFn: async (path: string) => {
      return await api.delete(`/domain/ous/${encodeURIComponent(path)}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ous'] })
    },
  })

  const handleSaveOU = async (newName: string) => {
    if (!editingOU) return
    
    // TODO: Add rename endpoint to API
    // For now, just log and refresh
    console.log('Rename OU:', editingOU, 'to', newName)
    
    // Refresh OU list
    await queryClient.invalidateQueries({ queryKey: ['ous'] })
  }

  const handleDeleteOU = () => {
    if (editingOU) {
      deleteOU.mutate(editingOU.path)
    }
  }

  const handleCreateOU = (name: string, description: string, parentPath: string) => {
    createOU.mutate({ name, description, parent_path: parentPath })
  }

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Organizational Units</h1>
          <p className="text-muted-foreground">
            Manage your domain's OU structure
          </p>
        </div>
        <Button onClick={() => setShowAddModal(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create OU
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FolderTree className="h-5 w-5 text-primary" />
            OU Hierarchy
          </CardTitle>
          <CardDescription>
            Your organizational structure
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">
              Loading organizational units...
            </div>
          ) : ouData ? (
            renderOUTree(ouData)
          ) : (
            renderOUTree(defaultStructure)
          )}
        </CardContent>
      </Card>

      <EditOUModal
        open={editingOU !== null}
        onClose={() => setEditingOU(null)}
        ouName={editingOU?.name || ''}
        ouPath={editingOU?.path || ''}
        onSave={handleSaveOU}
        onDelete={handleDeleteOU}
      />

      <AddOUModal
        open={showAddModal}
        onClose={() => setShowAddModal(false)}
        onSuccess={handleCreateOU}
      />
    </div>
  )
}

