import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AppProvider } from './context/AppContext';
import ProtectedRoute from './components/common/ProtectedRoute';
import Layout from './components/layout/Layout';

// Import pages
import Login from './pages/Login/Login';
import Register from './pages/Register/Register';
import ForgotPassword from './pages/Auth/ForgotPassword';
import MFASetup from './pages/Auth/MFASetup';

// Dashboard Page with Layout
const DashboardPage = () => (
  <Layout>
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <h1 className="text-4xl font-bold text-dark mb-8">Dashboard</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="card bg-primary-50 border-l-4 border-primary-600">
          <h3 className="text-lg font-semibold text-primary-700">Total Hosts</h3>
          <p className="text-3xl font-bold text-primary-900 mt-2">0</p>
        </div>
        
        <div className="card bg-success/10 border-l-4 border-success">
          <h3 className="text-lg font-semibold text-green-700">Online</h3>
          <p className="text-3xl font-bold text-green-900 mt-2">0</p>
        </div>
        
        <div className="card bg-warning/10 border-l-4 border-warning">
          <h3 className="text-lg font-semibold text-orange-700">Warnings</h3>
          <p className="text-3xl font-bold text-orange-900 mt-2">0</p>
        </div>
        
        <div className="card bg-error/10 border-l-4 border-error">
          <h3 className="text-lg font-semibold text-red-700">Critical</h3>
          <p className="text-3xl font-bold text-red-900 mt-2">0</p>
        </div>
      </div>
      
      <div className="card">
        <h2 className="text-2xl font-bold text-dark mb-4">Recent Activity</h2>
        <p className="text-gray-600">No recent activity</p>
      </div>
    </div>
  </Layout>
);

const App: React.FC = () => {
  return (
    <AppProvider>
      <Router>
        <div className="min-h-screen">
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/forgot-password" element={<ForgotPassword />} />
            
            {/* Semi-Protected Routes */}
            <Route path="/mfa-setup" element={<MFASetup />} />

            {/* Protected Routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardPage />
                </ProtectedRoute>
              }
            />

            {/* Default Redirects */}
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </div>
      </Router>
    </AppProvider>
  );
};

export default App;
