import React, { createContext, useState, useEffect } from "react";

interface AuthContextType {
  token: string | null;
  role: string | null;
  setToken: (token: string | null, role?: string | null) => void;
}

export const AuthContext = createContext<AuthContextType>({
  token: null,
  role: null,
  setToken: () => {},
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setTokenState] = useState<string | null>(() => localStorage.getItem("authToken"));
  const [role, setRole] = useState<string | null>(null);

  const setToken = (token: string | null, roleArg?: string | null) => {
    setTokenState(token);
    setRole(roleArg || null);

    if (token) {
      localStorage.setItem("authToken", token);
      if (roleArg) {
        localStorage.setItem("userRole", roleArg);
      }
    } else {
      localStorage.removeItem("authToken");
      localStorage.removeItem("userRole");
    }
  };

  useEffect(() => {
    const storedRole = localStorage.getItem("userRole");
    if (storedRole) {
      setRole(storedRole);
    }
  }, []);

  return (
    <AuthContext.Provider value={{ token, role, setToken }}>
      {children}
    </AuthContext.Provider>
  );
}
