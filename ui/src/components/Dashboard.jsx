import React, { useState, useEffect, lazy, Suspense } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import DashboardHeader from './dashboard/DashboardHeader';
import SidebarMobile from './dashboard/SidebarMobile';
import SidebarDesktop from './dashboard/SidebarDesktop';
import DashboardStats from './dashboard/DashboardStats';
import LoadingSpinner from './ui/LoadingSpinner';
import NotificationService from '@/services/NotificationService';

// Lazy loaded routes for better performance
const Overview = lazy(() => import('../pages/Overview'));
const RoutesList = lazy(() => import('../pages/RoutesList'));
const RouteDetail = lazy(() => import('../pages/RouteDetail'));
const RouteEdit = lazy(() => import('../pages/RouteEdit'));
const RouteCreate = lazy(() => import('../pages/RouteCreate'));
const Authentication = lazy(() => import('../pages/Authentication'));
const RateLimiting = lazy(() => import('../pages/RateLimiting'));
const CircuitBreakers = lazy(() => import('../pages/CircuitBreakers'));
const Caching = lazy(() => import('../pages/Caching'));
const Analytics = lazy(() => import('../pages/Analytics'));
const SettingsPage = lazy(() => import('../pages/Settings'));

export default function Dashboard() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [user, setUser] = useState({ name: 'Admin User', email: 'admin@example.com' });
  const [stats, setStats] = useState({
    routes: 0,
    requestsPerSecond: 0,
    responseTime: 0,
    errorRate: 0,
  });
  const [notifications, setNotifications] = useState([
    {
      id: 1,
      title: 'New error rate spike',
      description: 'Error rate increased by 5% in the last hour',
      time: '5 min ago',
      read: false,
    },
    {
      id: 2,
      title: 'Rate limit reached',
      description: 'API route /users/* hit rate limit',
      time: '20 min ago',
      read: false,
    },
    {
      id: 3,
      title: 'Circuit breaker open',
      description: 'Circuit breaker triggered for /payments/*',
      time: '1 hour ago',
      read: true,
    },
  ]);

  useEffect(() => {
    fetchStats();
    fetchNotifications();
    
    const statsInterval = setInterval(fetchStats, 5000);
    const notificationsInterval = setInterval(fetchNotifications, 30000);
    
    return () => {
      clearInterval(statsInterval);
      clearInterval(notificationsInterval);
    };
  }, []);

  const fetchStats = async () => {
    try {
      setStats({
        routes: Math.floor(Math.random() * 20) + 5,
        requestsPerSecond: Math.floor(Math.random() * 500) + 100,
        responseTime: Math.floor(Math.random() * 300) + 50,
        errorRate: (Math.random() * 5).toFixed(2),
      });
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };
  
  const fetchNotifications = async () => {
    try {
      const data = await NotificationService.getNotifications();
      setNotifications(data);
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
    }
  };

  return (
    <Router>
      <div className="flex h-screen bg-slate-50">
        {/* Mobile sidebar */}
        <SidebarMobile 
          open={sidebarOpen} 
          onOpenChange={setSidebarOpen} 
        />

        {/* Desktop sidebar */}
        <SidebarDesktop 
          user={user} 
        />

        {/* Main content */}
        <div className="flex flex-col flex-1 overflow-hidden">
          <DashboardHeader 
            setSidebarOpen={setSidebarOpen} 
            notifications={notifications}
            user={user}
          />

          <main className="relative flex-1 overflow-y-auto focus:outline-none">
            <div className="py-6">
              <div className="px-4 mx-auto max-w-7xl sm:px-6 md:px-8">
                {/* Stats cards */}
                <DashboardStats stats={stats} />

                {/* Route content */}
                <div className="py-4">
                  <Suspense fallback={<LoadingSpinner />}>
                    <Routes>
                      <Route path="/" element={<Overview />} />
                      <Route path="/routes" element={<RoutesList />} />
                      <Route path="/routes/new" element={<RouteCreate />} />
                      <Route path="/routes/:id" element={<RouteDetail />} />
                      <Route path="/routes/:id/edit" element={<RouteEdit />} />
                      <Route path="/authentication" element={<Authentication />} />
                      <Route path="/rate-limiting" element={<RateLimiting />} />
                      <Route path="/circuit-breakers" element={<CircuitBreakers />} />
                      <Route path="/caching" element={<Caching />} />
                      <Route path="/analytics" element={<Analytics />} />
                      <Route path="/settings" element={<SettingsPage />} />
                    </Routes>
                  </Suspense>
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </Router>
  );
}