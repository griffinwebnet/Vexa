import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/authStore'
import { ThemeProvider } from './components/ThemeProvider'
import LoginPage from './pages/LoginPage'
import DashboardLayout from './layouts/DashboardLayout'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Groups from './pages/Groups'
import Computers from './pages/Computers'
import DNS from './pages/DNS'
import Settings from './pages/Settings'
import SetupWizard from './pages/SetupWizard'
import DomainManagement from './pages/DomainManagement'
import DomainOUs from './pages/DomainOUs'
import DomainPolicies from './pages/DomainPolicies'
import SelfService from './pages/SelfService'

// Simple route protection - just check if authenticated
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }
  
  return <>{children}</>
}

// Admin-only routes
function AdminRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const isAdmin = useAuthStore((state) => state.isAdmin)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }
  
  if (!isAdmin) {
    return <Navigate to="/self-service" />
  }
  
  return <>{children}</>
}

// Domain user routes (self-service)
function DomainUserRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const isDomainUser = useAuthStore((state) => state.isDomainUser)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }
  
  if (!isDomainUser) {
    return <Navigate to="/" />
  }
  
  return <>{children}</>
}

// Setup wizard route - requires local admin authentication
function SetupRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const isAdmin = useAuthStore((state) => state.isAdmin)
  const isDomainUser = useAuthStore((state) => state.isDomainUser)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }
  
  // Only local admins (not domain users) can access setup
  if (!isAdmin || isDomainUser) {
    return <Navigate to="/" />
  }
  
  return <>{children}</>
}

function App() {
  return (
    <ThemeProvider>
      <Router>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/wizard"
            element={
              <SetupRoute>
                <SetupWizard />
              </SetupRoute>
            }
          />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Dashboard />} />
            <Route path="self-service" element={
              <DomainUserRoute>
                <SelfService />
              </DomainUserRoute>
            } />
            
            <Route path="domain" element={
              <AdminRoute>
                <DomainManagement />
              </AdminRoute>
            } />
            <Route path="domain/ous" element={
              <AdminRoute>
                <DomainOUs />
              </AdminRoute>
            } />
            <Route path="domain/policies" element={
              <AdminRoute>
                <DomainPolicies />
              </AdminRoute>
            } />
            <Route path="users" element={
              <AdminRoute>
                <Users />
              </AdminRoute>
            } />
            <Route path="groups" element={
              <AdminRoute>
                <Groups />
              </AdminRoute>
            } />
            <Route path="computers" element={
              <AdminRoute>
                <Computers />
              </AdminRoute>
            } />
            <Route path="dns" element={
              <AdminRoute>
                <DNS />
              </AdminRoute>
            } />
            <Route path="settings" element={
              <AdminRoute>
                <Settings />
              </AdminRoute>
            } />
          </Route>
        </Routes>
      </Router>
    </ThemeProvider>
  )
}

export default App