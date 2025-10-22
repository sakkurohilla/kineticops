import React, { useState, useEffect } from "react";
import AuthPages from "./AuthCard";
import Dashboard from "./Dashboard";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

export const AuthContext = React.createContext<{
  token: string | null;
  setToken: (token: string | null) => void;
}>({
  token: null,
  setToken: () => {},
});

export default function App() {
  // Initialize from localStorage
  const [token, setTokenState] = useState<string | null>(() => localStorage.getItem("authToken"));

  const setToken = (t: string | null) => {
    setTokenState(t);
    if (t) localStorage.setItem("authToken", t);
    else localStorage.removeItem("authToken");
  };

  useEffect(() => {
    // Sync with other tabs/windows:
    const handleStorage = () => setTokenState(localStorage.getItem("authToken"));
    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, []);

  return (
    <AuthContext.Provider value={{ token, setToken }}>
      <BrowserRouter>
        <Routes>
          <Route path="/dashboard" element={token ? <Dashboard /> : <Navigate to="/" replace />} />
          <Route path="/" element={token ? <Navigate to="/dashboard" replace /> : <AuthPages />} />
        </Routes>
      </BrowserRouter>
    </AuthContext.Provider>
  );
}
