import { useState } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface AddOUModalProps {
  open: boolean
  onClose: () => void
  onSuccess: (name: string, description: string, parentPath: string) => void
}

export function AddOUModal({ open, onClose, onSuccess }: AddOUModalProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [parentPath, setParentPath] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSuccess(name, description, parentPath)
    onClose()
    setName('')
    setDescription('')
    setParentPath('')
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Organizational Unit</DialogTitle>
          <DialogDescription>
            Add a new OU to your domain structure
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="ouName" className="text-sm font-medium">
              OU Name *
            </label>
            <Input
              id="ouName"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Engineering"
              required
            />
          </div>

          <div className="space-y-2">
            <label htmlFor="ouDescription" className="text-sm font-medium">
              Description
            </label>
            <Input
              id="ouDescription"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Engineering department"
            />
          </div>

          <div className="space-y-2">
            <label htmlFor="parentPath" className="text-sm font-medium">
              Parent OU Path (optional)
            </label>
            <Input
              id="parentPath"
              value={parentPath}
              onChange={(e) => setParentPath(e.target.value)}
              placeholder="OU=Corporate,DC=example,DC=com"
            />
            <p className="text-xs text-muted-foreground">
              Leave empty to create at domain root
            </p>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit">
              Create OU
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

