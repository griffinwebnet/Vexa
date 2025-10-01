import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { FolderTree, Plus } from 'lucide-react'

export default function OrganizationalUnits() {
  return (
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Organizational Units</h1>
          <p className="text-muted-foreground">
            Manage domain OUs and structure
          </p>
        </div>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create OU
        </Button>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FolderTree className="h-5 w-5 text-primary" />
              OU Structure
            </CardTitle>
            <CardDescription>
              Organizational unit hierarchy
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="p-3 rounded-lg border hover:bg-accent">
                <div className="font-medium">Domain Controllers</div>
                <div className="text-sm text-muted-foreground">Default OU for DCs</div>
              </div>
              <div className="p-3 rounded-lg border hover:bg-accent">
                <div className="font-medium">Users</div>
                <div className="text-sm text-muted-foreground">Default users container</div>
              </div>
              <div className="p-3 rounded-lg border hover:bg-accent">
                <div className="font-medium">Computers</div>
                <div className="text-sm text-muted-foreground">Default computers container</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>OU Management</CardTitle>
            <CardDescription>
              Create custom organizational units
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              Organizational Units allow you to organize users, groups, and computers into a hierarchical structure that mirrors your organization.
            </p>
            <div className="rounded-md bg-primary/10 p-4 text-sm">
              <p className="font-medium text-primary mb-2">Coming Soon</p>
              <p className="text-muted-foreground">
                Full OU management including creation, modification, delegation, and GPO linking.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

