import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { 
  Home, 
  List, 
  Shield, 
  Users, 
  Settings, 
  Key, 
  Database, 
  AlertTriangle, 
  BarChart2, 
  Clock,
  Menu, 
  X,
  LogOut,
  ChevronDown
} from 'react-feather';

// Pages
import Overview from '../pages/Overview';
import RoutesList from '../pages/RoutesList';
import RouteDetail from '../pages/RouteDetail';
import RouteEdit from '../pages/RouteEdit';
import RouteCreate from '../pages/RouteCreate';
import Authentication from '../pages/Authentication';
import RateLimiting from '../pages/RateLimiting';
import CircuitBreakers from '../pages/CircuitBreakers';
import Caching from '../pages/Caching';
import Analytics from '../pages/Analytics';
import SettingsPage from '../pages/Settings';

// API Service
import ApiService from '../services/ApiService';

export default function Dashboard() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [user, setUser] = useState({ name: "Admin User" });
  const [stats, setStats] = useState({
    routes: 0,
    requestsPerSecond: 0,
    responseTime: 0,
    errorRate: 0
  });

  useEffect(() => {
    // Fetch initial stats
    fetchStats();
    
    // Set up polling for real-time stats
    const interval = setInterval(fetchStats, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchStats = async () => {
    try {
      // In a real implementation, this would call the API
      // For now, simulate with random values
      setStats({
        routes: Math.floor(Math.random() * 20) + 5,
        requestsPerSecond: Math.floor(Math.random() * 500) + 100,
        responseTime: Math.floor(Math.random() * 300) + 50,
        errorRate: (Math.random() * 5).toFixed(2)
      });
    } catch (error) {
      console.error("Failed to fetch stats:", error);
    }
  };

  return (
    <Router>
      <div className="flex h-screen bg-gray-100">
        {/* Mobile sidebar */}
        <div className="md:hidden">
          {sidebarOpen && (
            <div 
              className="fixed inset-0 z-40 flex"
              onClick={() => setSidebarOpen(false)}
            >
              <div className="fixed inset-0 bg-gray-600 bg-opacity-75" />
              <div className="relative flex flex-col flex-1 w-full max-w-xs bg-white">
                <div className="absolute top-0 right-0 pt-2 -mr-12">
                  <button
                    className="flex items-center justify-center w-10 h-10 ml-1 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
                    onClick={() => setSidebarOpen(false)}
                  >
                    <span className="sr-only">Close sidebar</span>
                    <X className="w-6 h-6 text-white" />
                  </button>
                </div>
                <SidebarContent />
              </div>
            </div>
          )}
        </div>

        {/* Static sidebar for desktop */}
        <div className="hidden md:flex md:flex-shrink-0">
          <div className="flex flex-col w-64">
            <div className="flex flex-col flex-1 min-h-0 bg-white border-r border-gray-200">
              <div className="flex flex-col flex-1 pt-5 pb-4 overflow-y-auto">
                <div className="flex items-center flex-shrink-0 px-4">
                  <span className="text-xl font-bold text-indigo-600">Horizon Gateway</span>
                </div>
                <SidebarContent />
              </div>
            </div>
          </div>
        </div>

        {/* Main content */}
        <div className="flex flex-col flex-1 overflow-hidden">
          <div className="relative z-10 flex flex-shrink-0 h-16 bg-white shadow">
            <button
              className="px-4 text-gray-500 border-r border-gray-200 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500 md:hidden"
              onClick={() => setSidebarOpen(true)}
            >
              <span className="sr-only">Open sidebar</span>
              <Menu className="w-6 h-6" />
            </button>
            <div className="flex justify-between flex-1 px-4">
              <div className="flex flex-1">
                <div className="flex items-center flex-1 md:ml-0">
                  <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
                </div>
              </div>
              <div className="flex items-center ml-4 md:ml-6">
                <div className="relative ml-3">
                  <div>
                    <button className="flex items-center max-w-xs text-sm bg-white rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                      <span className="sr-only">Open user menu</span>
                      <div className="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-semibold">
                        {user.name.charAt(0)}
                      </div>
                      <span className="ml-2 text-gray-700">{user.name}</span>
                      <ChevronDown className="w-4 h-4 ml-1 text-gray-400" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <main className="relative flex-1 overflow-y-auto focus:outline-none">
            <div className="py-6">
              <div className="px-4 mx-auto max-w-7xl sm:px-6 md:px-8">
                {/* Stats cards */}
                <div className="grid grid-cols-1 gap-4 mb-8 sm:grid-cols-2 lg:grid-cols-4">
                  <StatCard title="Total Routes" value={stats.routes} icon={<List className="w-6 h-6" />} />
                  <StatCard title="Requests/sec" value={stats.requestsPerSecond} icon={<BarChart2 className="w-6 h-6" />} />
                  <StatCard title="Avg Response" value={`${stats.responseTime} ms`} icon={<Clock className="w-6 h-6" />} />
                  <StatCard title="Error Rate" value={`${stats.errorRate}%`} icon={<AlertTriangle className="w-6 h-6" />} />
                </div>
                
                {/* Route content */}
                <div className="py-4">
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
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </Router>
  );
}

function SidebarContent() {
  return (
    <nav className="px-2 mt-5 space-y-1">
      <NavItem to="/" icon={<Home />} label="Overview" />
      <NavItem to="/routes" icon={<List />} label="Routes" />
      <NavItem to="/authentication" icon={<Shield />} label="Authentication" />
      <NavItem to="/rate-limiting" icon={<Clock />} label="Rate Limiting" />
      <NavItem to="/circuit-breakers" icon={<AlertTriangle />} label="Circuit Breakers" />
      <NavItem to="/caching" icon={<Database />} label="Caching" />
      <NavItem to="/analytics" icon={<BarChart2 />} label="Analytics" />
      <NavItem to="/settings" icon={<Settings />} label="Settings" />
      
      <div className="pt-8">
        <div className="px-2 space-y-1">
          <button className="flex items-center w-full px-2 py-2 text-sm font-medium text-gray-600 rounded-md group hover:bg-gray-50 hover:text-gray-900">
            <LogOut className="w-5 h-5 mr-3 text-gray-400 group-hover:text-gray-500" />
            Sign out
          </button>
        </div>
      </div>
    </nav>
  );
}

function NavItem({ to, icon, label }) {
  return (
    <Link
      to={to}
      className="flex items-center px-2 py-2 text-sm font-medium text-gray-600 rounded-md group hover:bg-gray-50 hover:text-gray-900"
    >
      <div className="w-5 h-5 mr-3 text-gray-400 group-hover:text-gray-500">
        {icon}
      </div>
      {label}
    </Link>
  );
}

function StatCard({ title, value, icon }) {
  return (
    <div className="px-4 py-5 overflow-hidden bg-white rounded-lg shadow">
      <div className="flex items-center">
        <div className="p-3 rounded-md bg-indigo-50 text-indigo-600">
          {icon}
        </div>
        <div className="ml-5 w-0 flex-1">
          <dl>
            <dt className="text-sm font-medium text-gray-500 truncate">{title}</dt>
            <dd>
              <div className="text-lg font-medium text-gray-900">{value}</div>
            </dd>
          </dl>
        </div>
      </div>
    </div>
  );
}