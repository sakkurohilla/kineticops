import React, { ReactNode, useState, useEffect } from 'react';
import Navbar from './Navbar';
import Sidebar from './Sidebar';

interface LayoutProps {
  children: ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [sidebarWidth, setSidebarWidth] = useState(80); // Default collapsed width

  useEffect(() => {
    // Listen for sidebar hover events if needed
    // For now, we'll use CSS transitions
  }, []);

  return (
    <div className="flex min-h-screen bg-gray-100">
      <Sidebar />
      {/* Main content area with dynamic margin */}
      <div className="flex-1 flex flex-col ml-20 transition-all duration-300">
        <Navbar />
        <main className="flex-1">{children}</main>
      </div>
    </div>
  );
};

export default Layout;
