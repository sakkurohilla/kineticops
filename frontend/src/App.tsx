import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AppProvider } from './context/AppContext';
import { ThemeProvider } from './context/ThemeContext';
import ProtectedRoute from './components/common/ProtectedRoute';
import ErrorBoundary from './components/common/ErrorBoundary';
import Logs from './pages/Logs/Logs';

// Import pages
import Login from './pages/Login/Login';
import Register from './pages/Register/Register';
import ForgotPassword from './pages/Auth/ForgotPassword';
import ResetPassword from './pages/Auth/ResetPassword';
import MFASetup from './pages/Auth/MFASetup';
import Dashboard from './pages/Dashboard/Dashboard';
import AlertRules from './pages/AlertRules/AlertRules';
import Metrics from './pages/Metrics/Metrics';
import Hosts from './pages/Hosts/Hosts';
import HostDetails from './components/hosts/HostDetails';
import Workflow from './pages/Workflow/Workflow';
import Process from './pages/Process/Process';
import Services from './pages/Services/Services';
import Synthetics from './pages/Synthetics/Synthetics';
import Settings from './pages/Settings/Settings';
import Profile from './pages/Profile/Profile';

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <ThemeProvider>
        <AppProvider>
          <Router>
          <Routes>
          {/* Public Routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/forgot-password" element={<ForgotPassword />} />
          <Route path="/reset-password" element={<ResetPassword />} />
          
          {/* Semi-Protected Routes */}
          <Route path="/mfa-setup" element={<MFASetup />} />

          {/* Protected Routes */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          />

          <Route
            path="/hosts"
            element={
              <ProtectedRoute>
                <Hosts />
              </ProtectedRoute>
            }
          />

          <Route
            path="/hosts/:id"
            element={
              <ProtectedRoute>
                <HostDetails />
              </ProtectedRoute>
            }
          />

          <Route
            path="/metrics"
            element={
              <ProtectedRoute>
                <Metrics />
              </ProtectedRoute>
            }
          />

          <Route
            path="/alerts"
            element={
              <ProtectedRoute>
                <AlertRules />
              </ProtectedRoute>
            }
          />

          <Route
            path="/logs"
            element={
              <ProtectedRoute>
                <Logs />
              </ProtectedRoute>
            }
          />

          <Route
            path="/workflow"
            element={
              <ProtectedRoute>
                <Workflow />
              </ProtectedRoute>
            }
          />

          <Route
            path="/process"
            element={
              <ProtectedRoute>
                <Process />
              </ProtectedRoute>
            }
          />

          <Route
            path="/services"
            element={
              <ProtectedRoute>
                <Services />
              </ProtectedRoute>
            }
          />

          <Route
            path="/synthetics"
            element={
              <ProtectedRoute>
                <Synthetics />
              </ProtectedRoute>
            }
          />

          <Route
            path="/settings"
            element={
              <ProtectedRoute>
                <Settings />
              </ProtectedRoute>
            }
          />

          <Route
            path="/profile"
            element={
              <ProtectedRoute>
                <Profile />
              </ProtectedRoute>
            }
          />

          {/* Default Redirects */}
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
        </Router>
      </AppProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
};

export default App;