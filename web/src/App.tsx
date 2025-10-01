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
import DomainSetup from './pages/DomainSetup'
import Settings from './pages/Settings'
import SetupWizard from './pages/SetupWizard'
import DomainManagement from './pages/DomainManagement'
import DomainOUs from './pages/DomainOUs'
import DomainPolicies from './pages/DomainPolicies'

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
            <Route path="setup" element={<DomainSetup />} />
            <Route path="domain" element={<DomainManagement />} />
            <Route path="domain/ous" element={<DomainOUs />} />
            <Route path="domain/policies" element={<DomainPolicies />} />
            <Route path="users" element={<Users />} />
            <Route path="groups" element={<Groups />} />
            <Route path="computers" element={<Computers />} />
            <Route path="dns" element={<DNS />} />
            <Route path="settings" element={<Settings />} />
          </Route>
        </Routes>
      </Router>
    </ThemeProvider>
  )
}

export default App

