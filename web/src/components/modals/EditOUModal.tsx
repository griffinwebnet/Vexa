import { useState } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface EditOUModalProps {
  open: boolean
  onClose: () => void
  ouName: string
  ouPath: string
  onSave: (newName: string) => void
  onDelete: () => void
}

export function EditOUModal({ open, onClose, ouName, ouPath, onSave, onDelete }: EditOUModalProps) {
  const [newName, setNewName] = useState(ouName)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    if (confirmDelete) {
      onDelete()
    } else {
      onSave(newName)
    }
    
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Organizational Unit</DialogTitle>
          <DialogDescription>
            Rename or delete this OU
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="ouName" className="text-sm font-medium">
              OU Name
            </label>
            <Input
              id="ouName"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              disabled={confirmDelete}
              required
            />
            <p className="text-xs text-muted-foreground font-mono">
              {ouPath}
            </p>
          </div>

          <div className="flex items-center space-x-2 p-3 rounded-lg border border-destructive/50 bg-destructive/10">
            <input
              type="checkbox"
              id="confirmDelete"
              checked={confirmDelete}
              onChange={(e) => setConfirmDelete(e.target.checked)}
              className="h-4 w-4"
            />
            <label htmlFor="confirmDelete" className="text-sm font-medium cursor-pointer">
              Delete this OU (cannot be undone)
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

