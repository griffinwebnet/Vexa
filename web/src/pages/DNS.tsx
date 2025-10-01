import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/Card'

export default function DNS() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">DNS Management</h1>
        <p className="text-muted-foreground">
          Manage DNS zones and records
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>DNS Configuration</CardTitle>
          <CardDescription>
            Configure DNS zones and records for your domain
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            DNS management interface coming soon
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

