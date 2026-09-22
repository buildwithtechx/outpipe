import { Outlet } from '@tanstack/react-router';
import { useState } from 'react';
import { AdminHeader } from './admin-header';
import { AdminSidebar } from './admin-sidebar';

export function AdminDashboardLayout({ children }: { children?: React.ReactNode }) {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <div className="flex h-screen flex-col bg-black text-white antialiased overflow-hidden">
      {/* Top Admin Header */}
      <AdminHeader mobileOpen={mobileOpen} setMobileOpen={setMobileOpen} />

      {/* Main Container with Admin Sidebar + Dynamic Admin View */}
      <div className="flex flex-1 min-h-0 overflow-hidden relative">
        <AdminSidebar mobileOpen={mobileOpen} setMobileOpen={setMobileOpen} />

        <main className="flex-1 min-w-0 overflow-y-auto bg-black p-4 sm:p-6 lg:p-8">
          <div className="mx-auto max-w-6xl">
            {children || <Outlet />}
          </div>
        </main>
      </div>
    </div>
  );
}
