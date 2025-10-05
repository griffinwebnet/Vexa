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

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const isDomainUser = useAuthStore((state) => state.isDomainUser)

  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }

  // If it's a local admin user (not domain user), they should be in wizard
  if (!isDomainUser && window.location.pathname !== '/wizard') {
    return <Navigate to="/wizard" />
  }

  // If it's a domain user and they're trying to access wizard, redirect to dashboard
  if (isDomainUser && window.location.pathname === '/wizard') {
    return <Navigate to="/" />
  }

  return <>{children}</>
}

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

function DomainUserRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const isDomainUser = useAuthStore((state) => state.isDomainUser)
  const isAdmin = useAuthStore((state) => state.isAdmin)
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }

  // Domain users get self-service access, admins get full access
  if (isDomainUser && !isAdmin) {
    return <>{children}</>
  }

  // Non-domain users (local admins) don't get self-service access
  if (!isDomainUser) {
    return <Navigate to="/" />
  }

  return <>{children}</>
}

function SetupRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const isDomainUser = useAuthStore((state) => state.isDomainUser)

  // If not authenticated, redirect to login
  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }

  // If it's a domain user, they shouldn't access wizard
  if (isDomainUser) {
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
              <PrivateRoute>
                <DashboardLayout />
              </PrivateRoute>
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

