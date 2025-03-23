import React, { useState, useEffect } from 'react';
import {
  AlertTriangle,
  Plus,
  RefreshCw,
  Power,
  Settings,
  CheckCircle,
  XCircle,
  AlertOctagon,
  Zap,
  Clock,
  Activity,
  Filter,
  ChevronDown,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Switch } from '@/components/ui/switch';
import { Progress } from '@/components/ui/progress';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';

export default function CircuitBreakers() {
  const [circuitBreakers, setCircuitBreakers] = useState([]);
  const [filteredBreakers, setFilteredBreakers] = useState([]);
  const [statistics, setStatistics] = useState({
    total: 0,
    open: 0,
    halfOpen: 0,
    closed: 0,
  });
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [newBreaker, setNewBreaker] = useState({
    route: '',
    type: 'error',
    errorThreshold: 50,
    minRequests: 20,
    samplingWindow: '1m',
    openStateDuration: '30s',
    halfOpenMaxRequests: 5,
    enabled: true,
  });

  useEffect(() => {
    fetchCircuitBreakers();
  }, []);

  useEffect(() => {
    // Apply filters when circuit breakers or filters change
    applyFilters();
  }, [circuitBreakers, searchQuery, statusFilter]);

  const fetchCircuitBreakers = async () => {
    setLoading(true);
    try {
      // Simulate API call with sample data
      const data = generateSampleData();
      setCircuitBreakers(data);

      // Calculate statistics
      const stats = {
        total: data.length,
        open: data.filter((cb) => cb.state === 'open').length,
        halfOpen: data.filter((cb) => cb.state === 'half_open').length,
        closed: data.filter((cb) => cb.state === 'closed').length,
      };
      setStatistics(stats);

      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch circuit breakers:', error);
      setLoading(false);
    }
  };

  const generateSampleData = () => {
    const sampleData = [
      {
        id: 1,
        route: '/api/users/*',
        type: 'error',
        state: 'closed',
        errorRate: 3,
        requestsInWindow: 245,
        failureCount: 7,
        lastTripped: '3 days ago',
        errorThreshold: 50,
        minRequests: 20,
        samplingWindow: '1m',
        openStateDuration: '30s',
        halfOpenMaxRequests: 5,
        enabled: true,
      },
      {
        id: 2,
        route: '/api/payments/*',
        type: 'timeout',
        state: 'open',
        errorRate: 62,
        requestsInWindow: 87,
        failureCount: 54,
        lastTripped: '5 minutes ago',
        errorThreshold: 60,
        minRequests: 10,
        samplingWindow: '1m',
        openStateDuration: '60s',
        halfOpenMaxRequests: 3,
        enabled: true,
      },
      {
        id: 3,
        route: '/api/products/*',
        type: 'concurrency',
        state: 'half_open',
        errorRate: 18,
        requestsInWindow: 125,
        failureCount: 22,
        lastTripped: '15 minutes ago',
        errorThreshold: 40,
        minRequests: 15,
        samplingWindow: '2m',
        openStateDuration: '45s',
        halfOpenMaxRequests: 5,
        enabled: true,
      },
      {
        id: 4,
        route: '/api/auth/*',
        type: 'error',
        state: 'closed',
        errorRate: 1,
        requestsInWindow: 534,
        failureCount: 5,
        lastTripped: '2 weeks ago',
        errorThreshold: 25,
        minRequests: 50,
        samplingWindow: '5m',
        openStateDuration: '1m',
        halfOpenMaxRequests: 10,
        enabled: true,
      },
      {
        id: 5,
        route: '/api/orders/*',
        type: 'hybrid',
        state: 'closed',
        errorRate: 12,
        requestsInWindow: 230,
        failureCount: 28,
        lastTripped: '2 days ago',
        errorThreshold: 30,
        minRequests: 30,
        samplingWindow: '3m',
        openStateDuration: '1m',
        halfOpenMaxRequests: 5,
        enabled: true,
      },
      {
        id: 6,
        route: '/api/external/weather/*',
        type: 'timeout',
        state: 'open',
        errorRate: 78,
        requestsInWindow: 42,
        failureCount: 33,
        lastTripped: '2 minutes ago',
        errorThreshold: 40,
        minRequests: 10,
        samplingWindow: '1m',
        openStateDuration: '2m',
        halfOpenMaxRequests: 3,
        enabled: true,
      },
      {
        id: 7,
        route: '/api/search/*',
        type: 'error',
        state: 'closed',
        errorRate: 8,
        requestsInWindow: 320,
        failureCount: 26,
        lastTripped: '1 day ago',
        errorThreshold: 30,
        minRequests: 20,
        samplingWindow: '2m',
        openStateDuration: '30s',
        halfOpenMaxRequests: 5,
        enabled: false,
      },
    ];

    return sampleData;
  };

  const applyFilters = () => {
    let filtered = [...circuitBreakers];

    // Apply search filter
    if (searchQuery) {
      filtered = filtered.filter((cb) =>
        cb.route.toLowerCase().includes(searchQuery.toLowerCase()),
      );
    }

    // Apply status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter((cb) => cb.state === statusFilter);
    }

    setFilteredBreakers(filtered);
  };

  const handleToggleEnabled = (id, enabled) => {
    const updatedBreakers = circuitBreakers.map((cb) =>
      cb.id === id ? { ...cb, enabled: !enabled } : cb,
    );
    setCircuitBreakers(updatedBreakers);
  };

  const handleResetCircuitBreaker = (id) => {
    const updatedBreakers = circuitBreakers.map((cb) =>
      cb.id === id ? { ...cb, state: 'closed', lastTripped: 'just now' } : cb,
    );
    setCircuitBreakers(updatedBreakers);
  };

  const handleCreateCircuitBreaker = () => {
    const newId = Math.max(...circuitBreakers.map((cb) => cb.id)) + 1;
    const createdBreaker = {
      id: newId,
      ...newBreaker,
      state: 'closed',
      errorRate: 0,
      requestsInWindow: 0,
      failureCount: 0,
      lastTripped: 'never',
    };

    setCircuitBreakers([...circuitBreakers, createdBreaker]);
    setIsCreateDialogOpen(false);

    // Reset form
    setNewBreaker({
      route: '',
      type: 'error',
      errorThreshold: 50,
      minRequests: 20,
      samplingWindow: '1m',
      openStateDuration: '30s',
      halfOpenMaxRequests: 5,
      enabled: true,
    });
  };

  const circuitBreakerStateInfo = {
    open: {
      color: 'destructive',
      icon: <XCircle className="h-4 w-4 mr-1" />,
      description: 'Requests are failing, traffic is blocked',
    },
    half_open: {
      color: 'warning',
      icon: <AlertOctagon className="h-4 w-4 mr-1" />,
      description: 'Testing limited traffic after failure',
    },
    closed: {
      color: 'success',
      icon: <CheckCircle className="h-4 w-4 mr-1" />,
      description: 'Working normally, all traffic allowed',
    },
  };

  const circuitBreakerTypeInfo = {
    error: {
      icon: <AlertTriangle className="h-4 w-4 mr-1" />,
      description: 'Trips on error percentage',
    },
    timeout: {
      icon: <Clock className="h-4 w-4 mr-1" />,
      description: 'Trips on timeout percentage',
    },
    concurrency: {
      icon: <Activity className="h-4 w-4 mr-1" />,
      description: 'Trips on concurrent connections',
    },
    hybrid: {
      icon: <Zap className="h-4 w-4 mr-1" />,
      description: 'Combines multiple conditions',
    },
  };

  const getErrorRateClass = (rate) => {
    if (rate < 10) return 'text-green-600';
    if (rate < 30) return 'text-amber-600';
    return 'text-red-600';
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            Circuit Breakers
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Prevent cascading failures by automatically detecting and isolating
            failing services.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchCircuitBreakers}>
            <RefreshCw className="h-4 w-4 mr-1" />
            Refresh
          </Button>
          <Dialog
            open={isCreateDialogOpen}
            onOpenChange={setIsCreateDialogOpen}
          >
            <DialogTrigger asChild>
              <Button size="sm">
                <Plus className="h-4 w-4 mr-1" />
                Add Circuit Breaker
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[560px]">
              <DialogHeader>
                <DialogTitle>Create New Circuit Breaker</DialogTitle>
                <DialogDescription>
                  Configure a new circuit breaker to protect your API routes
                  from cascading failures.
                </DialogDescription>
              </DialogHeader>
              <div className="py-4 space-y-4">
                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="route">Route Pattern</Label>
                  <Input
                    id="route"
                    placeholder="/api/users/*"
                    value={newBreaker.route}
                    onChange={(e) =>
                      setNewBreaker({ ...newBreaker, route: e.target.value })
                    }
                  />
                  <p className="text-xs text-slate-500">
                    The path pattern to apply this circuit breaker to.
                  </p>
                </div>

                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="type">Circuit Breaker Type</Label>
                  <Select
                    value={newBreaker.type}
                    onValueChange={(value) =>
                      setNewBreaker({ ...newBreaker, type: value })
                    }
                  >
                    <SelectTrigger id="type">
                      <SelectValue placeholder="Select type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="error">Error Rate</SelectItem>
                      <SelectItem value="timeout">Timeout</SelectItem>
                      <SelectItem value="concurrency">Concurrency</SelectItem>
                      <SelectItem value="hybrid">Hybrid</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-xs text-slate-500">
                    The condition that will trigger the circuit breaker.
                  </p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="grid w-full items-center gap-1.5">
                    <Label htmlFor="errorThreshold">Error Threshold (%)</Label>
                    <Input
                      id="errorThreshold"
                      type="number"
                      min="1"
                      max="100"
                      value={newBreaker.errorThreshold}
                      onChange={(e) =>
                        setNewBreaker({
                          ...newBreaker,
                          errorThreshold: parseInt(e.target.value),
                        })
                      }
                    />
                  </div>
                  <div className="grid w-full items-center gap-1.5">
                    <Label htmlFor="minRequests">Minimum Requests</Label>
                    <Input
                      id="minRequests"
                      type="number"
                      min="1"
                      value={newBreaker.minRequests}
                      onChange={(e) =>
                        setNewBreaker({
                          ...newBreaker,
                          minRequests: parseInt(e.target.value),
                        })
                      }
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="grid w-full items-center gap-1.5">
                    <Label htmlFor="samplingWindow">Sampling Window</Label>
                    <Input
                      id="samplingWindow"
                      placeholder="1m"
                      value={newBreaker.samplingWindow}
                      onChange={(e) =>
                        setNewBreaker({
                          ...newBreaker,
                          samplingWindow: e.target.value,
                        })
                      }
                    />
                    <p className="text-xs text-slate-500">e.g., 30s, 1m, 5m</p>
                  </div>
                  <div className="grid w-full items-center gap-1.5">
                    <Label htmlFor="openStateDuration">
                      Open State Duration
                    </Label>
                    <Input
                      id="openStateDuration"
                      placeholder="30s"
                      value={newBreaker.openStateDuration}
                      onChange={(e) =>
                        setNewBreaker({
                          ...newBreaker,
                          openStateDuration: e.target.value,
                        })
                      }
                    />
                    <p className="text-xs text-slate-500">e.g., 30s, 1m, 5m</p>
                  </div>
                </div>

                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="halfOpenMaxRequests">
                    Half-Open Max Requests
                  </Label>
                  <Input
                    id="halfOpenMaxRequests"
                    type="number"
                    min="1"
                    value={newBreaker.halfOpenMaxRequests}
                    onChange={(e) =>
                      setNewBreaker({
                        ...newBreaker,
                        halfOpenMaxRequests: parseInt(e.target.value),
                      })
                    }
                  />
                  <p className="text-xs text-slate-500">
                    Maximum requests to allow in half-open state.
                  </p>
                </div>

                <div className="flex items-center space-x-2">
                  <Switch
                    id="enabled"
                    checked={newBreaker.enabled}
                    onCheckedChange={(checked) =>
                      setNewBreaker({ ...newBreaker, enabled: checked })
                    }
                  />
                  <Label htmlFor="enabled">Enable Circuit Breaker</Label>
                </div>
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setIsCreateDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button onClick={handleCreateCircuitBreaker}>Create</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Breakers"
          value={statistics.total}
          description="Configured circuit breakers"
          icon={<AlertTriangle className="h-5 w-5" />}
          className="bg-slate-100"
        />
        <StatCard
          title="Open Circuits"
          value={statistics.open}
          description="Blocking traffic currently"
          icon={<XCircle className="h-5 w-5" />}
          className="bg-red-50 text-red-700"
        />
        <StatCard
          title="Half-Open Circuits"
          value={statistics.halfOpen}
          description="Testing traffic flow"
          icon={<AlertOctagon className="h-5 w-5" />}
          className="bg-amber-50 text-amber-700"
        />
        <StatCard
          title="Closed Circuits"
          value={statistics.closed}
          description="Functioning normally"
          icon={<CheckCircle className="h-5 w-5" />}
          className="bg-green-50 text-green-700"
        />
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-slate-500" />
          <Input
            type="text"
            placeholder="Search routes..."
            className="pl-9"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="outline"
              className="w-full sm:w-auto flex justify-between items-center gap-1"
            >
              <Filter className="h-4 w-4 mr-1" />
              <span>
                Status:{' '}
                {statusFilter === 'all'
                  ? 'All'
                  : statusFilter === 'open'
                  ? 'Open'
                  : statusFilter === 'half_open'
                  ? 'Half-Open'
                  : 'Closed'}
              </span>
              <ChevronDown className="h-4 w-4 ml-1" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-48">
            <DropdownMenuLabel>Filter by Status</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => setStatusFilter('all')}>
              All States
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setStatusFilter('open')}>
              <XCircle className="h-4 w-4 mr-2 text-red-500" />
              Open
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setStatusFilter('half_open')}>
              <AlertOctagon className="h-4 w-4 mr-2 text-amber-500" />
              Half-Open
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setStatusFilter('closed')}>
              <CheckCircle className="h-4 w-4 mr-2 text-green-500" />
              Closed
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Circuit Breakers Table */}
      <Card>
        <CardHeader className="px-6 py-5">
          <CardTitle>Circuit Breakers</CardTitle>
          <CardDescription>
            {loading
              ? 'Loading circuit breakers...'
              : filteredBreakers.length > 0
              ? `Showing ${filteredBreakers.length} of ${circuitBreakers.length} circuit breakers`
              : 'No circuit breakers found'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center items-center py-8">
              <RefreshCw className="h-8 w-8 animate-spin text-slate-400" />
            </div>
          ) : filteredBreakers.length === 0 ? (
            <div className="text-center py-8 text-slate-500">
              <AlertTriangle className="h-12 w-12 mx-auto mb-4 text-slate-300" />
              <p className="text-lg font-medium mb-1">
                No circuit breakers found
              </p>
              <p className="text-sm">
                {searchQuery || statusFilter !== 'all'
                  ? 'Try adjusting your search or filters'
                  : 'Create a circuit breaker to protect your API routes'}
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[260px]">Route</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>State</TableHead>
                  <TableHead>Error Rate</TableHead>
                  <TableHead>Requests</TableHead>
                  <TableHead>Last Tripped</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredBreakers.map((circuitBreaker) => (
                  <TableRow key={circuitBreaker.id}>
                    <TableCell className="font-medium">
                      {circuitBreaker.route}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center">
                        <Badge variant="outline" className="flex items-center">
                          {circuitBreakerTypeInfo[circuitBreaker.type].icon}
                          {circuitBreaker.type.charAt(0).toUpperCase() +
                            circuitBreaker.type.slice(1)}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          circuitBreakerStateInfo[circuitBreaker.state].color
                        }
                        className="flex items-center"
                      >
                        {circuitBreakerStateInfo[circuitBreaker.state].icon}
                        {circuitBreaker.state === 'half_open'
                          ? 'Half-Open'
                          : circuitBreaker.state.charAt(0).toUpperCase() +
                            circuitBreaker.state.slice(1)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span
                          className={getErrorRateClass(
                            circuitBreaker.errorRate,
                          )}
                        >
                          {circuitBreaker.errorRate}%
                        </span>
                        <Progress
                          value={circuitBreaker.errorRate}
                          max={100}
                          className={`h-2 w-16 ${
                            circuitBreaker.errorRate < 10
                              ? 'bg-slate-100'
                              : circuitBreaker.errorRate < 30
                              ? 'bg-amber-100'
                              : 'bg-red-100'
                          }`}
                          indicatorClassName={
                            circuitBreaker.errorRate < 10
                              ? 'bg-green-500'
                              : circuitBreaker.errorRate < 30
                              ? 'bg-amber-500'
                              : 'bg-red-500'
                          }
                        />
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col">
                        <span>{circuitBreaker.requestsInWindow} total</span>
                        <span className="text-xs text-slate-500">
                          {circuitBreaker.failureCount} failures
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>{circuitBreaker.lastTripped}</TableCell>
                    <TableCell>
                      <Switch
                        checked={circuitBreaker.enabled}
                        onCheckedChange={() =>
                          handleToggleEnabled(
                            circuitBreaker.id,
                            circuitBreaker.enabled,
                          )
                        }
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          disabled={circuitBreaker.state === 'closed'}
                          onClick={() =>
                            handleResetCircuitBreaker(circuitBreaker.id)
                          }
                          title="Reset Circuit Breaker"
                        >
                          <Power className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          title="Edit Circuit Breaker"
                        >
                          <Settings className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({ title, value, description, icon, className = '' }) {
  return (
    <Card className={className}>
      <CardContent className="pt-6">
        <div className="flex justify-between">
          <div>
            <p className="text-sm font-medium">{title}</p>
            <p className="text-3xl font-bold mt-1">{value}</p>
            <p className="text-xs text-slate-500 mt-1">{description}</p>
          </div>
          <div className="h-12 w-12 rounded-lg bg-white flex items-center justify-center shadow-sm">
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// Missing component definition for Search icon
function Search(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="11" cy="11" r="8" />
      <path d="m21 21-4.3-4.3" />
    </svg>
  );
}
