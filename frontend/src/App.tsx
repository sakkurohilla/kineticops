import React, { useContext } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthContext, AuthProvider } from "./context/AuthContext";
import AuthPages from "./AuthCard";
import Dashboard from "./Dashboard";
import Workspaces from "./Workspaces";

function AppRoutes() {
  const { token } = useContext(AuthContext);

  if (!token) {
    return (
      <Routes>
        <Route path="/*" element={<AuthPages />} />
      </Routes>
    );
  }

  return (
    <Routes>
      <Route path="/dashboard" element={<Dashboard />} />
      <Route path="/workspaces" element={<Workspaces />} />
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </AuthProvider>
  );
}
