import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AppProvider } from './context/AppContext';
import ProtectedRoute from './components/common/ProtectedRoute';

// Login Page
const LoginPage = () => (
  <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50 flex items-center justify-center p-8">
    <div className="card max-w-md w-full">
      <h1 className="text-3xl font-bold text-primary-700 mb-2">KineticOps</h1>
      <p className="text-gray-600 mb-6">Infrastructure Monitoring Platform</p>
      
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input 
            type="email" 
            placeholder="you@example.com" 
            className="input" 
          />
        </div>
        
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input 
            type="password" 
            placeholder="••••••••" 
            className="input" 
          />
        </div>
        
        <button className="btn btn-primary w-full">
          Sign In
        </button>
        
        <button className="btn btn-secondary w-full">
          Sign In with OAuth
        </button>
      </div>
      
      <div className="mt-6 pt-6 border-t border-gray-200">
        <p className="text-sm text-center text-gray-600">
          Don't have an account? <a href="/register" className="text-primary-600 hover:text-primary-700 font-medium">Sign up</a>
        </p>
      </div>
    </div>
  </div>
);

// Register Page
const RegisterPage = () => (
  <div className="min-h-screen bg-gradient-to-br from-secondary-50 to-primary-50 flex items-center justify-center p-8">
    <div className="card max-w-md w-full">
      <h1 className="text-3xl font-bold text-secondary-700 mb-2">Create Account</h1>
      <p className="text-gray-600 mb-6">Join KineticOps today</p>
      
      <div className="space-y-4">
        <input type="text" placeholder="Full Name" className="input" />
        <input type="email" placeholder="Email" className="input" />
        <input type="password" placeholder="Password" className="input" />
        <button className="btn btn-secondary w-full">Create Account</button>
      </div>
      
      <p className="text-sm text-center text-gray-600 mt-4">
        Already have an account? <a href="/login" className="text-secondary-600 hover:text-secondary-700 font-medium">Sign in</a>
      </p>
    </div>
  </div>
);

// Dashboard Page
const DashboardPage = () => (
  <div className="min-h-screen bg-light p-8">
    <div className="max-w-7xl mx-auto">
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
  </div>
);

const App: React.FC = () => {
  return (
    <AppProvider>
      <Router>
        <div className="min-h-screen">
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />

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
