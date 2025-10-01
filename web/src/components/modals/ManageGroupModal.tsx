import { useState } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface ManageGroupModalProps {
  open: boolean
  onClose: () => void
  groupName: string
  onSave: (newName: string) => void
  onDelete: () => void
}

export function ManageGroupModal({ open, onClose, groupName, onSave, onDelete }: ManageGroupModalProps) {
  const [newName, setNewName] = useState(groupName)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    if (confirmDelete) {
      onDelete()
    } else {
      onSave(newName)
    }
    
    onClose()
    setConfirmDelete(false)
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Manage Group</DialogTitle>
          <DialogDescription>
            Rename or delete this security group
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="groupName" className="text-sm font-medium">
              Group Name
            </label>
            <Input
              id="groupName"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              disabled={confirmDelete}
              required
            />
          </div>

          <div className="flex items-center space-x-2 p-3 rounded-lg border border-destructive/50 bg-destructive/10">
            <input
              type="checkbox"
              id="confirmDeleteGroup"
              checked={confirmDelete}
              onChange={(e) => setConfirmDelete(e.target.checked)}
              className="h-4 w-4"
            />
            <label htmlFor="confirmDeleteGroup" className="text-sm font-medium cursor-pointer">
              Delete this group (cannot be undone)
            </label>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button 
              type="submit" 
              variant={confirmDelete ? "destructive" : "default"}
            >
              {confirmDelete ? "Confirm Deletion" : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

