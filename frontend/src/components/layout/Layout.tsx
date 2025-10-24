import React, { ReactNode } from 'react';
import Navbar from './Navbar';
import Sidebar from './Sidebar';

interface LayoutProps {
  children: ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="flex min-h-screen bg-light">
      <Sidebar />
      <div className="flex-1 flex flex-col">
        <Navbar />
        <main className="flex-1">{children}</main>
      </div>
    </div>
  );
};

export default Layout;
