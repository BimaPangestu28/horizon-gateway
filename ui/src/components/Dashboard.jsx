import React, { useState, useEffect } from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  useLocation,
} from 'react-router-dom';
import {
  Home,
  List,
  Shield,
  Settings,
  Database,
  AlertTriangle,
  BarChart2,
  Clock,
  Menu,
  X,
  LogOut,
  ChevronDown,
  Bell,
  Search,
  MoreHorizontal,
  Zap,
  Layers,
  User,
  PieChart,
  Cpu,
} from 'lucide-react';

// Shadcn UI components
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from '@/components/ui/card';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

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

export default function Dashboard() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [user, setUser] = useState({ name: 'Admin User' });
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
        errorRate: (Math.random() * 5).toFixed(2),
      });
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  return (
    <Router>
      <div className="flex h-screen bg-slate-50">
        {/* Mobile sidebar */}
        <div className="md:hidden">
          <Sheet open={sidebarOpen} onOpenChange={setSidebarOpen}>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="px-4 border-r border-slate-200 md:hidden"
              >
                <Menu className="w-6 h-6" />
                <span className="sr-only">Open sidebar</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="p-0 w-72 border-r-0">
              <div className="flex flex-col h-full bg-white">
                <div className="flex items-center h-16 px-4 border-b border-slate-200 bg-slate-900 text-white">
                  <div className="flex items-center gap-2">
                    <div className="rounded-md bg-indigo-600 p-1">
                      <Zap className="h-6 w-6 text-white" />
                    </div>
                    <span className="text-xl font-bold">Horizon Gateway</span>
                  </div>
                </div>
                <SidebarContent />
              </div>
            </SheetContent>
          </Sheet>
        </div>

        {/* Static sidebar for desktop */}
        <div className="hidden md:flex md:w-72 md:flex-col">
          <div className="flex flex-col flex-1 min-h-0 bg-slate-900 text-white">
            <div className="flex items-center h-16 px-6 border-b border-slate-700">
              <div className="flex items-center gap-2">
                <div className="rounded-md bg-indigo-600 p-1">
                  <Zap className="h-6 w-6 text-white" />
                </div>
                <span className="text-xl font-bold">Horizon Gateway</span>
              </div>
            </div>
            <SidebarContent />
            <div className="p-4 mt-auto border-t border-slate-700">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    className="w-full justify-start text-slate-300 hover:text-white hover:bg-slate-800 px-2"
                  >
                    <div className="flex items-center">
                      <Avatar className="h-8 w-8 mr-2">
                        <AvatarImage
                          src="https://github.com/shadcn.png"
                          alt="User"
                        />
                        <AvatarFallback className="bg-indigo-600">
                          AD
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex flex-col items-start">
                        <span className="text-sm font-medium">Admin User</span>
                        <span className="text-xs text-slate-400">
                          admin@example.com
                        </span>
                      </div>
                    </div>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56">
                  <DropdownMenuLabel>My Account</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem>
                    <User className="mr-2 h-4 w-4" />
                    <span>Profile</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <Settings className="mr-2 h-4 w-4" />
                    <span>Settings</span>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem>
                    <LogOut className="mr-2 h-4 w-4" />
                    <span>Sign out</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>

        {/* Main content */}
        <div className="flex flex-col flex-1 overflow-hidden">
          <header className="relative z-10 flex items-center justify-between h-16 px-4 bg-white border-b border-slate-200 shadow-sm">
            <div className="flex items-center flex-1 gap-4">
              <Button
                variant="ghost"
                size="icon"
                className="md:hidden"
                onClick={() => setSidebarOpen(true)}
              >
                <Menu className="h-6 w-6" />
              </Button>
              <div className="relative hidden md:block w-64">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-slate-400" />
                <Input placeholder="Search..." className="pl-8 bg-slate-50" />
              </div>
            </div>
            <div className="flex items-center gap-3">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button variant="outline" size="icon" className="relative">
                      <Bell className="h-5 w-5" />
                      {notifications.filter((n) => !n.read).length > 0 && (
                        <span className="absolute top-1 right-1 flex h-2 w-2">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                          <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
                        </span>
                      )}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    You have {notifications.filter((n) => !n.read).length}{' '}
                    unread notifications
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <MoreHorizontal className="h-5 w-5" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem>View Documentation</DropdownMenuItem>
                  <DropdownMenuItem>Check for Updates</DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem>Support</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    className="flex items-center gap-2 md:hidden"
                  >
                    <Avatar className="h-8 w-8">
                      <AvatarFallback className="bg-indigo-600 text-white">
                        AD
                      </AvatarFallback>
                    </Avatar>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuLabel>My Account</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem>
                    <User className="mr-2 h-4 w-4" />
                    <span>Profile</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <Settings className="mr-2 h-4 w-4" />
                    <span>Settings</span>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem>
                    <LogOut className="mr-2 h-4 w-4" />
                    <span>Sign out</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </header>

          <main className="relative flex-1 overflow-y-auto focus:outline-none">
            <div className="py-6">
              <div className="px-4 mx-auto max-w-7xl sm:px-6 md:px-8">
                {/* Stats cards */}
                <div className="grid grid-cols-1 gap-4 mb-8 sm:grid-cols-2 lg:grid-cols-4">
                  <StatCard
                    title="Total Routes"
                    value={stats.routes}
                    icon={<Layers className="w-5 h-5" />}
                    trend="up"
                    trendValue="12%"
                    description="Active API routes"
                    color="blue"
                  />
                  <StatCard
                    title="Requests/sec"
                    value={stats.requestsPerSecond}
                    icon={<BarChart2 className="w-5 h-5" />}
                    trend="up"
                    trendValue="8.2%"
                    description="Avg. over 5 min"
                    color="purple"
                  />
                  <StatCard
                    title="Avg Response"
                    value={`${stats.responseTime} ms`}
                    icon={<Clock className="w-5 h-5" />}
                    trend="down"
                    trendValue="4.3%"
                    description="Across all routes"
                    color="green"
                  />
                  <StatCard
                    title="Error Rate"
                    value={`${stats.errorRate}%`}
                    icon={<AlertTriangle className="w-5 h-5" />}
                    trend={parseFloat(stats.errorRate) < 2 ? 'down' : 'up'}
                    trendValue="0.7%"
                    description="5xx responses"
                    color={
                      parseFloat(stats.errorRate) < 2
                        ? 'green'
                        : parseFloat(stats.errorRate) < 5
                        ? 'amber'
                        : 'red'
                    }
                  />
                </div>

                {/* Route content */}
                <div className="py-4">
                  <Routes>
                    <Route path="/" element={<Overview />} />
                    <Route path="/routes" element={<RoutesList />} />
                    <Route path="/routes/new" element={<RouteCreate />} />
                    <Route path="/routes/:id" element={<RouteDetail />} />
                    <Route path="/routes/:id/edit" element={<RouteEdit />} />
                    <Route
                      path="/authentication"
                      element={<Authentication />}
                    />
                    <Route path="/rate-limiting" element={<RateLimiting />} />
                    <Route
                      path="/circuit-breakers"
                      element={<CircuitBreakers />}
                    />
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
    <div className="flex flex-col flex-1 py-4 overflow-y-auto">
      <nav className="flex-1 px-2 space-y-1">
        <div className="px-3 pb-2 text-xs font-semibold text-slate-400 uppercase tracking-wider">
          Main
        </div>
        <NavItem to="/" icon={<Home />} label="Overview" />
        <NavItem to="/routes" icon={<Layers />} label="Routes" />
        <NavItem to="/analytics" icon={<PieChart />} label="Analytics" />

        <Separator className="my-4 bg-slate-700" />

        <div className="px-3 pb-2 text-xs font-semibold text-slate-400 uppercase tracking-wider">
          Security
        </div>
        <NavItem
          to="/authentication"
          icon={<Shield />}
          label="Authentication"
        />
        <NavItem to="/rate-limiting" icon={<Clock />} label="Rate Limiting" />

        <Separator className="my-4 bg-slate-700" />

        <div className="px-3 pb-2 text-xs font-semibold text-slate-400 uppercase tracking-wider">
          Reliability
        </div>
        <NavItem
          to="/circuit-breakers"
          icon={<AlertTriangle />}
          label="Circuit Breakers"
        />
        <NavItem to="/caching" icon={<Database />} label="Caching" />
        <NavItem to="/settings" icon={<Settings />} label="Settings" />
      </nav>
    </div>
  );
}

function NavItem({ to, icon, label }) {
  const location = useLocation();
  const isActive = location.pathname === to;

  return (
    <Link to={to}>
      <Button
        variant="ghost"
        className={`w-full justify-start gap-3 mb-1 ${
          isActive
            ? 'bg-slate-800 text-white hover:bg-slate-700'
            : 'text-slate-300 hover:bg-slate-800 hover:text-white'
        }`}
      >
        <span>{icon}</span>
        <span>{label}</span>
        {isActive && (
          <span className="absolute left-0 rounded-r-md inset-y-1 w-1 bg-indigo-500"></span>
        )}
      </Button>
    </Link>
  );
}

function StatCard({
  title,
  value,
  icon,
  description,
  trend,
  trendValue,
  color,
}) {
  const colors = {
    blue: {
      bg: 'bg-blue-50',
      text: 'text-blue-600',
      icon: 'bg-blue-100',
      light: 'text-blue-500',
      trend: { up: 'text-green-600', down: 'text-red-600' },
    },
    green: {
      bg: 'bg-emerald-50',
      text: 'text-emerald-600',
      icon: 'bg-emerald-100',
      light: 'text-emerald-500',
      trend: { up: 'text-emerald-600', down: 'text-emerald-600' },
    },
    amber: {
      bg: 'bg-amber-50',
      text: 'text-amber-600',
      icon: 'bg-amber-100',
      light: 'text-amber-500',
      trend: { up: 'text-red-600', down: 'text-green-600' },
    },
    red: {
      bg: 'bg-rose-50',
      text: 'text-rose-600',
      icon: 'bg-rose-100',
      light: 'text-rose-500',
      trend: { up: 'text-red-600', down: 'text-green-600' },
    },
    purple: {
      bg: 'bg-violet-50',
      text: 'text-violet-600',
      icon: 'bg-violet-100',
      light: 'text-violet-500',
      trend: { up: 'text-green-600', down: 'text-red-600' },
    },
  }[color || 'blue'];

  return (
    <Card className="overflow-hidden">
      <CardHeader className="p-0">
        <div
          className={`h-1 w-full ${colors.text.replace('text', 'bg')}`}
        ></div>
      </CardHeader>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">{title}</p>
            <div className="flex items-baseline mt-1">
              <p className="text-2xl font-semibold">{value}</p>
              <p className={`ml-2 text-xs font-medium ${colors.light}`}>
                {description}
              </p>
            </div>
          </div>
          <div className={`p-2 rounded-full ${colors.icon}`}>{icon}</div>
        </div>
        <div className="flex items-center mt-4">
          <Badge
            variant="outline"
            className={`gap-1 ${
              trend === 'up' ? colors.trend.up : colors.trend.down
            }`}
          >
            {trend === 'up' ? (
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 20 20"
                fill="currentColor"
                className="w-3 h-3"
              >
                <path
                  fillRule="evenodd"
                  d="M12.577 4.878a.75.75 0 01.919-.53l4.78 1.281a.75.75 0 01.531.919l-1.281 4.78a.75.75 0 01-1.449-.387l.81-3.022a19.407 19.407 0 00-5.594 5.203.75.75 0 01-1.139.093L7 10.06l-4.72 4.72a.75.75 0 01-1.06-1.061l5.25-5.25a.75.75 0 011.06 0l3.074 3.073a20.923 20.923 0 015.545-4.931l-3.042-.815a.75.75 0 01-.53-.919z"
                  clipRule="evenodd"
                />
              </svg>
            ) : (
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 20 20"
                fill="currentColor"
                className="w-3 h-3"
              >
                <path
                  fillRule="evenodd"
                  d="M1.22 5.222a.75.75 0 011.06 0L7 9.94l3.172-3.172a.75.75 0 011.06 0l3.1 3.1a20.98 20.98 0 015.9-4.964.75.75 0 01.694 1.325 19.494 19.494 0 00-5.506 4.625l3.037.9a.75.75 0 01-.23 1.471l-4.8-1.425a.75.75 0 01-.535-.767l.2-4.8a.75.75 0 011.467-.246l.7 3.36a21.47 21.47 0 015.393-4.761.75.75 0 01.818 1.257 20.97 20.97 0 00-5.168 4.561l2.935.87a.75.75 0 01-.231 1.47l-4.8-1.425a.75.75 0 01-.535-.767l.2-4.783.002-.045a.75.75 0 011.466-.246l.28 1.11a20.87 20.87 0 016.229-3.581.75.75 0 01.599 1.374 19.548 19.548 0 00-4.653 2.9l.292 1.76a.75.75 0 01-.713.883l-4.8.4a.75.75 0 01-.817-.744v-.049l.64-4.759a.75.75 0 011.5.201l-.205 1.533a22.704 22.704 0 016.258-3.087.75.75 0 11.548 1.391 20.384 20.384 0 00-7.786 4.998l1.943 1.95a.75.75 0 11-1.06 1.06l-5.25-5.25a.75.75 0 010-1.06l5.25-5.25a.75.75 0 011.06 0l3.536 3.536a17.55 17.55 0 016.191-3.716.75.75 0 01.578 1.382 16.18 16.18 0 00-6.33 4.078l2.424 2.425a.75.75 0 01-1.06 1.06L7 9.06 3.28 12.78a.75.75 0 01-1.06-1.06l5.25-5.25a.75.75 0 011.06 0l3.134 3.133a21.79 21.79 0 015.728-4.586.75.75 0 01.675 1.334 20.226 20.226 0 00-6.286 5.343.75.75 0 01-1.195.144L7 8.06l-4.72 4.72a.75.75 0 01-1.06-1.06l5.25-5.25a.75.75 0 011.06 0l3.134 3.133a21.792 21.792 0 015.728-4.586.75.75 0 01.675 1.334 20.238 20.238 0 00-6.286 5.343.75.75 0 01-1.195.144L7 8.06l-4.72 4.72a.75.75 0 01-1.06-1.06l5.25-5.25a.75.75 0 011.06 0l3.134 3.133a21.792 21.792 0 015.728-4.586.75.75 0 01.675 1.334 20.238 20.238 0 00-6.286 5.343.75.75 0 01-1.195.144L7 8.06l-4.72 4.72a.75.75 0 01-1.06-1.06l5.25-5.25a.75.75 0 011.06 0l3.134 3.133a21.792 21.792 0 015.728-4.586.75.75 0 01.675 1.334 20.238 20.238 0 00-6.286 5.343.75.75 0 01-1.195.144L7 8.06l-4.72 4.72a.75.75 0 01-1.06-1.06l5.25-5.25a.75.75 0 011.06 0l3.074 3.073a20.923 20.923 0 015.545-4.93l-3.042-.815a.75.75 0 01-.53-.919z"
                  clipRule="evenodd"
                />
              </svg>
            )}
            {trendValue}
          </Badge>
          <span className="ml-2 text-xs text-slate-500">vs last period</span>
        </div>
      </CardContent>
    </Card>
  );
}
