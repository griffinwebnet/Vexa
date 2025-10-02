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
  const setupComplete = localStorage.getItem('vexa-setup-complete')
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }
  
  if (!setupComplete && window.location.pathname !== '/wizard') {
    return <Navigate to="/wizard" />
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
              <PrivateRoute>
                <SetupWizard />
              </PrivateRoute>
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
            <Route path="self-service" element={<SelfService />} />
            
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

