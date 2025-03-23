import React, { useState, useEffect } from 'react';
import {
  Database,
  RefreshCw,
  Plus,
  Settings,
  Trash2,
  Clock,
  Filter,
  Search,
  ChevronDown,
  Check,
  X,
  MinusCircle,
  PlusCircle,
  Save,
  ExternalLink,
  Edit,
  RotateCw,
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

import ApiService from '@/services/ApiService';

export default function Caching() {
  const [cacheConfigs, setCacheConfigs] = useState([]);
  const [filteredConfigs, setFilteredConfigs] = useState([]);
  const [statistics, setStatistics] = useState({
    total: 0,
    enabled: 0,
    hitRate: 0,
    savedBandwidth: 0,
  });

  const [loading, setLoading] = useState(true);
  const [clearingCache, setClearingCache] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState('all');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  const [newCache, setNewCache] = useState({
    route: '',
    type: 'memory',
    ttl: '5m',
    maxSize: 1048576, // 1MB
    methods: ['GET'],
    cacheKeyTemplate: '{method}:{path}:{query}',
    ignoreQueryParams: [],
    varyHeaders: [],
    neverCache: [],
    alwaysCache: [],
    respectCacheControl: true,
    includeHost: false,
    enabled: true,
  });

  const [selectedConfig, setSelectedConfig] = useState(null);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editCache, setEditCache] = useState({});

  const [tempIgnoreQueryParam, setTempIgnoreQueryParam] = useState('');
  const [tempVaryHeader, setTempVaryHeader] = useState('');
  const [tempNeverCache, setTempNeverCache] = useState('');
  const [tempAlwaysCache, setTempAlwaysCache] = useState('');

  useEffect(() => {
    fetchCacheConfigs();
  }, []);

  useEffect(() => {
    // Apply filters when cache configs or filters change
    applyFilters();
  }, [cacheConfigs, searchQuery, typeFilter]);

  const fetchCacheConfigs = async () => {
    setLoading(true);
    try {
      // In a real implementation, this would call the API
      // For now, use sample data
      const data = generateSampleData();
      setCacheConfigs(data);

      // Calculate statistics
      const stats = {
        total: data.length,
        enabled: data.filter((config) => config.enabled).length,
        hitRate: calculateAverageHitRate(data),
        savedBandwidth: calculateSavedBandwidth(data),
      };
      setStatistics(stats);

      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch cache configs:', error);
      setLoading(false);
    }
  };

  const generateSampleData = () => {
    return [
      {
        id: 1,
        route: '/api/products/*',
        type: 'memory',
        ttl: '5m',
        maxSize: 1048576, // 1MB
        methods: ['GET'],
        cacheKeyTemplate: '{method}:{path}:{query}',
        ignoreQueryParams: ['sort', 'page'],
        varyHeaders: ['Accept-Language'],
        neverCache: ['/api/products/special/*'],
        alwaysCache: ['/api/products/popular/*'],
        respectCacheControl: true,
        includeHost: false,
        enabled: true,
        stats: {
          hits: 12500,
          misses: 2500,
          hitRate: 83.33,
          size: 768432,
          itemCount: 243,
          avgResponseSize: 3163,
          oldestItem: '3 hours ago',
          newestItem: '2 minutes ago',
          savedBandwidth: 39.84, // MB
        },
      },
      {
        id: 2,
        route: '/api/categories/*',
        type: 'memory',
        ttl: '30m',
        maxSize: 524288, // 512KB
        methods: ['GET'],
        cacheKeyTemplate: '{path}:{query}',
        ignoreQueryParams: [],
        varyHeaders: [],
        neverCache: [],
        alwaysCache: [],
        respectCacheControl: true,
        includeHost: false,
        enabled: true,
        stats: {
          hits: 8900,
          misses: 1100,
          hitRate: 89.0,
          size: 215040,
          itemCount: 48,
          avgResponseSize: 4480,
          oldestItem: '25 minutes ago',
          newestItem: '30 seconds ago',
          savedBandwidth: 38.49, // MB
        },
      },
      {
        id: 3,
        route: '/api/users/*/profile',
        type: 'redis',
        ttl: '60m',
        maxSize: 2097152, // 2MB
        methods: ['GET'],
        cacheKeyTemplate: '{method}:{path}:{query}',
        ignoreQueryParams: ['t'],
        varyHeaders: ['Accept-Language', 'Accept'],
        neverCache: ['/api/users/admin/*'],
        alwaysCache: [],
        respectCacheControl: true,
        includeHost: false,
        enabled: false,
        stats: {
          hits: 0,
          misses: 0,
          hitRate: 0,
          size: 0,
          itemCount: 0,
          avgResponseSize: 0,
          oldestItem: 'N/A',
          newestItem: 'N/A',
          savedBandwidth: 0,
        },
      },
      {
        id: 4,
        route: '/api/content/*',
        type: 'redis',
        ttl: '24h',
        maxSize: 10485760, // 10MB
        methods: ['GET'],
        cacheKeyTemplate: '{path}',
        ignoreQueryParams: ['source', 'ref', 'utm_*'],
        varyHeaders: [],
        neverCache: [],
        alwaysCache: ['/api/content/static/*'],
        respectCacheControl: false,
        includeHost: false,
        enabled: true,
        stats: {
          hits: 25400,
          misses: 4600,
          hitRate: 84.67,
          size: 8125440,
          itemCount: 156,
          avgResponseSize: 52086,
          oldestItem: '22 hours ago',
          newestItem: '5 minutes ago',
          savedBandwidth: 1278.52, // MB
        },
      },
      {
        id: 5,
        route: '/api/search/*',
        type: 'memory',
        ttl: '2m',
        maxSize: 5242880, // 5MB
        methods: ['GET'],
        cacheKeyTemplate: '{method}:{path}:{query}',
        ignoreQueryParams: ['page', 'count'],
        varyHeaders: [],
        neverCache: [],
        alwaysCache: [],
        respectCacheControl: true,
        includeHost: false,
        enabled: true,
        stats: {
          hits: 18500,
          misses: 7500,
          hitRate: 71.15,
          size: 3670016,
          itemCount: 350,
          avgResponseSize: 10486,
          oldestItem: '1 minute ago',
          newestItem: '10 seconds ago',
          savedBandwidth: 188.72, // MB
        },
      },
    ];
  };

  const calculateAverageHitRate = (data) => {
    const enabledConfigs = data.filter((config) => config.enabled);
    if (enabledConfigs.length === 0) return 0;

    const totalHits = enabledConfigs.reduce(
      (sum, config) => sum + config.stats.hits,
      0,
    );
    const totalRequests = enabledConfigs.reduce(
      (sum, config) => sum + config.stats.hits + config.stats.misses,
      0,
    );

    return totalRequests > 0
      ? ((totalHits / totalRequests) * 100).toFixed(2)
      : 0;
  };

  const calculateSavedBandwidth = (data) => {
    return data
      .filter((config) => config.enabled)
      .reduce((sum, config) => sum + config.stats.savedBandwidth, 0)
      .toFixed(2);
  };

  const applyFilters = () => {
    let filtered = [...cacheConfigs];

    // Apply search filter
    if (searchQuery) {
      filtered = filtered.filter((config) =>
        config.route.toLowerCase().includes(searchQuery.toLowerCase()),
      );
    }

    // Apply type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter((config) => config.type === typeFilter);
    }

    setFilteredConfigs(filtered);
  };

  const handleToggleEnabled = (id, enabled) => {
    const updatedConfigs = cacheConfigs.map((config) =>
      config.id === id ? { ...config, enabled: !enabled } : config,
    );
    setCacheConfigs(updatedConfigs);
  };

  const handleClearCache = async (id) => {
    setClearingCache(true);
    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Reset cache stats
      const updatedConfigs = cacheConfigs.map((config) => {
        if (config.id === id) {
          return {
            ...config,
            stats: {
              ...config.stats,
              hits: 0,
              misses: 0,
              hitRate: 0,
              itemCount: 0,
              size: 0,
              oldestItem: 'N/A',
              newestItem: 'N/A',
              savedBandwidth: 0,
            },
          };
        }
        return config;
      });

      setCacheConfigs(updatedConfigs);
    } catch (error) {
      console.error('Failed to clear cache:', error);
    } finally {
      setClearingCache(false);
    }
  };

  const handleCreateCache = () => {
    const newId = Math.max(0, ...cacheConfigs.map((config) => config.id)) + 1;
    const createdConfig = {
      id: newId,
      ...newCache,
      stats: {
        hits: 0,
        misses: 0,
        hitRate: 0,
        size: 0,
        itemCount: 0,
        avgResponseSize: 0,
        oldestItem: 'N/A',
        newestItem: 'N/A',
        savedBandwidth: 0,
      },
    };

    setCacheConfigs([...cacheConfigs, createdConfig]);
    setIsCreateDialogOpen(false);

    // Reset form
    setNewCache({
      route: '',
      type: 'memory',
      ttl: '5m',
      maxSize: 1048576, // 1MB
      methods: ['GET'],
      cacheKeyTemplate: '{method}:{path}:{query}',
      ignoreQueryParams: [],
      varyHeaders: [],
      neverCache: [],
      alwaysCache: [],
      respectCacheControl: true,
      includeHost: false,
      enabled: true,
    });
    setTempIgnoreQueryParam('');
    setTempVaryHeader('');
    setTempNeverCache('');
    setTempAlwaysCache('');
  };

  const handleEditCache = () => {
    const updatedConfigs = cacheConfigs.map((config) =>
      config.id === selectedConfig ? { ...config, ...editCache } : config,
    );

    setCacheConfigs(updatedConfigs);
    setIsEditDialogOpen(false);
    setSelectedConfig(null);
    setEditCache({});
  };

  const openEditDialog = (id) => {
    const config = cacheConfigs.find((config) => config.id === id);
    if (config) {
      setSelectedConfig(id);
      setEditCache({ ...config });
      setIsEditDialogOpen(true);
    }
  };

  const addMethod = (method, isEdit = false) => {
    if (isEdit) {
      if (!editCache.methods.includes(method)) {
        setEditCache({
          ...editCache,
          methods: [...editCache.methods, method],
        });
      }
    } else {
      if (!newCache.methods.includes(method)) {
        setNewCache({
          ...newCache,
          methods: [...newCache.methods, method],
        });
      }
    }
  };

  const removeMethod = (method, isEdit = false) => {
    if (isEdit) {
      setEditCache({
        ...editCache,
        methods: editCache.methods.filter((m) => m !== method),
      });
    } else {
      setNewCache({
        ...newCache,
        methods: newCache.methods.filter((m) => m !== method),
      });
    }
  };

  const handleAddIgnoreQueryParam = (isEdit = false) => {
    if (!tempIgnoreQueryParam.trim()) return;

    if (isEdit) {
      setEditCache({
        ...editCache,
        ignoreQueryParams: [
          ...editCache.ignoreQueryParams,
          tempIgnoreQueryParam.trim(),
        ],
      });
    } else {
      setNewCache({
        ...newCache,
        ignoreQueryParams: [
          ...newCache.ignoreQueryParams,
          tempIgnoreQueryParam.trim(),
        ],
      });
    }

    setTempIgnoreQueryParam('');
  };

  const handleRemoveIgnoreQueryParam = (param, isEdit = false) => {
    if (isEdit) {
      setEditCache({
        ...editCache,
        ignoreQueryParams: editCache.ignoreQueryParams.filter(
          (p) => p !== param,
        ),
      });
    } else {
      setNewCache({
        ...newCache,
        ignoreQueryParams: newCache.ignoreQueryParams.filter(
          (p) => p !== param,
        ),
      });
    }
  };

  const handleAddVaryHeader = (isEdit = false) => {
    if (!tempVaryHeader.trim()) return;

    if (isEdit) {
      setEditCache({
        ...editCache,
        varyHeaders: [...editCache.varyHeaders, tempVaryHeader.trim()],
      });
    } else {
      setNewCache({
        ...newCache,
        varyHeaders: [...newCache.varyHeaders, tempVaryHeader.trim()],
      });
    }

    setTempVaryHeader('');
  };

  const handleRemoveVaryHeader = (header, isEdit = false) => {
    if (isEdit) {
      setEditCache({
        ...editCache,
        varyHeaders: editCache.varyHeaders.filter((h) => h !== header),
      });
    } else {
      setNewCache({
        ...newCache,
        varyHeaders: newCache.varyHeaders.filter((h) => h !== header),
      });
    }
  };

  const handleAddNeverCache = (isEdit = false) => {
    if (!tempNeverCache.trim()) return;

    if (isEdit) {
      setEditCache({
        ...editCache,
        neverCache: [...editCache.neverCache, tempNeverCache.trim()],
      });
    } else {
      setNewCache({
        ...newCache,
        neverCache: [...newCache.neverCache, tempNeverCache.trim()],
      });
    }

    setTempNeverCache('');
  };

  const handleRemoveNeverCache = (path, isEdit = false) => {
    if (isEdit) {
      setEditCache({
        ...editCache,
        neverCache: editCache.neverCache.filter((p) => p !== path),
      });
    } else {
      setNewCache({
        ...newCache,
        neverCache: newCache.neverCache.filter((p) => p !== path),
      });
    }
  };

  const handleAddAlwaysCache = (isEdit = false) => {
    if (!tempAlwaysCache.trim()) return;

    if (isEdit) {
      setEditCache({
        ...editCache,
        alwaysCache: [...editCache.alwaysCache, tempAlwaysCache.trim()],
      });
    } else {
      setNewCache({
        ...newCache,
        alwaysCache: [...newCache.alwaysCache, tempAlwaysCache.trim()],
      });
    }

    setTempAlwaysCache('');
  };

  const handleRemoveAlwaysCache = (path, isEdit = false) => {
    if (isEdit) {
      setEditCache({
        ...editCache,
        alwaysCache: editCache.alwaysCache.filter((p) => p !== path),
      });
    } else {
      setNewCache({
        ...newCache,
        alwaysCache: newCache.alwaysCache.filter((p) => p !== path),
      });
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            Cache Management
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Improve performance and reduce backend load by caching API
            responses.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchCacheConfigs}>
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
                Add Cache Config
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[650px]">
              <DialogHeader>
                <DialogTitle>Create New Cache Configuration</DialogTitle>
                <DialogDescription>
                  Configure caching for your API routes to improve performance
                  and reduce backend load.
                </DialogDescription>
              </DialogHeader>
              <div className="max-h-[70vh] overflow-y-auto py-4 pr-2">
                <div className="space-y-6">
                  {/* Basic Settings */}
                  <div className="space-y-4">
                    <h3 className="text-md font-semibold">Basic Settings</h3>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="grid w-full items-center gap-1.5">
                        <Label htmlFor="route">Route Pattern</Label>
                        <Input
                          id="route"
                          placeholder="/api/products/*"
                          value={newCache.route}
                          onChange={(e) =>
                            setNewCache({ ...newCache, route: e.target.value })
                          }
                        />
                        <p className="text-xs text-slate-500">
                          The path pattern to apply this cache to.
                        </p>
                      </div>

                      <div className="grid w-full items-center gap-1.5">
                        <Label htmlFor="type">Cache Type</Label>
                        <Select
                          value={newCache.type}
                          onValueChange={(value) =>
                            setNewCache({ ...newCache, type: value })
                          }
                        >
                          <SelectTrigger id="type">
                            <SelectValue placeholder="Select type" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="memory">Memory</SelectItem>
                            <SelectItem value="redis">Redis</SelectItem>
                          </SelectContent>
                        </Select>
                        <p className="text-xs text-slate-500">
                          Type of storage for cached data.
                        </p>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="grid w-full items-center gap-1.5">
                        <Label htmlFor="ttl">Time to Live (TTL)</Label>
                        <Input
                          id="ttl"
                          placeholder="5m"
                          value={newCache.ttl}
                          onChange={(e) =>
                            setNewCache({ ...newCache, ttl: e.target.value })
                          }
                        />
                        <p className="text-xs text-slate-500">
                          How long to keep items in cache (e.g., 10s, 5m, 1h,
                          7d)
                        </p>
                      </div>

                      <div className="grid w-full items-center gap-1.5">
                        <Label htmlFor="maxSize">Maximum Size (bytes)</Label>
                        <Input
                          id="maxSize"
                          type="number"
                          placeholder="1048576"
                          value={newCache.maxSize}
                          onChange={(e) =>
                            setNewCache({
                              ...newCache,
                              maxSize: parseInt(e.target.value),
                            })
                          }
                        />
                        <p className="text-xs text-slate-500">
                          Maximum size of responses to cache (1048576 = 1MB)
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* HTTP Methods */}
                  <Separator />
                  <div className="space-y-4">
                    <h3 className="text-md font-semibold">HTTP Methods</h3>
                    <div className="space-y-2">
                      <Label>Methods to Cache</Label>
                      <div className="flex flex-wrap gap-2">
                        {[
                          'GET',
                          'POST',
                          'PUT',
                          'DELETE',
                          'PATCH',
                          'HEAD',
                          'OPTIONS',
                        ].map((method) => (
                          <Badge
                            key={method}
                            variant={
                              newCache.methods.includes(method)
                                ? 'default'
                                : 'outline'
                            }
                            className="cursor-pointer"
                            onClick={() => {
                              newCache.methods.includes(method)
                                ? removeMethod(method)
                                : addMethod(method);
                            }}
                          >
                            {method}
                            {newCache.methods.includes(method) ? (
                              <X className="ml-1 h-3 w-3" />
                            ) : (
                              <Plus className="ml-1 h-3 w-3" />
                            )}
                          </Badge>
                        ))}
                      </div>
                      <p className="text-xs text-slate-500">
                        Path patterns that should always be cached, regardless
                        of cache-control headers.
                      </p>
                    </div>
                  </div>

                  {/* Cache Behavior */}
                  <Separator />
                  <div className="space-y-4">
                    <h3 className="text-md font-semibold">Cache Behavior</h3>

                    <div className="flex items-center space-x-2">
                      <Switch
                        id="respectCacheControl"
                        checked={newCache.respectCacheControl}
                        onCheckedChange={(checked) =>
                          setNewCache({
                            ...newCache,
                            respectCacheControl: checked,
                          })
                        }
                      />
                      <Label htmlFor="respectCacheControl">
                        Respect Cache-Control Headers
                      </Label>
                      <p className="text-xs text-slate-500 ml-2">
                        Honor cache-control headers from the origin server.
                      </p>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Switch
                        id="includeHost"
                        checked={newCache.includeHost}
                        onCheckedChange={(checked) =>
                          setNewCache({ ...newCache, includeHost: checked })
                        }
                      />
                      <Label htmlFor="includeHost">
                        Include Host in Cache Key
                      </Label>
                      <p className="text-xs text-slate-500 ml-2">
                        Add hostname to cache keys (useful for multi-tenant
                        APIs).
                      </p>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Switch
                        id="enabled"
                        checked={newCache.enabled}
                        onCheckedChange={(checked) =>
                          setNewCache({ ...newCache, enabled: checked })
                        }
                      />
                      <Label htmlFor="enabled">Enable Caching</Label>
                    </div>
                  </div>
                </div>
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setIsCreateDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button onClick={handleCreateCache}>Create</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Cache Configs"
          value={statistics.total}
          description={`${statistics.enabled} active`}
          icon={<Database className="h-5 w-5" />}
          color="blue"
        />
        <StatCard
          title="Hit Rate"
          value={`${statistics.hitRate}%`}
          description="Average cache effectiveness"
          icon={<PlusCircle className="h-5 w-5" />}
          color={
            parseFloat(statistics.hitRate) > 80
              ? 'green'
              : parseFloat(statistics.hitRate) > 50
              ? 'blue'
              : 'amber'
          }
        />
        <StatCard
          title="Bandwidth Saved"
          value={`${statistics.savedBandwidth} MB`}
          description="Total bandwidth reduction"
          icon={<RotateCw className="h-5 w-5" />}
          color="purple"
        />
        <StatCard
          title="Status"
          value={statistics.enabled > 0 ? 'Active' : 'Inactive'}
          description={
            statistics.enabled > 0
              ? `${statistics.enabled} configs enabled`
              : 'No active cache configs'
          }
          icon={<Clock className="h-5 w-5" />}
          color={statistics.enabled > 0 ? 'green' : 'amber'}
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
                Cache Type:{' '}
                {typeFilter === 'all'
                  ? 'All'
                  : typeFilter === 'memory'
                  ? 'Memory'
                  : 'Redis'}
              </span>
              <ChevronDown className="h-4 w-4 ml-1" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-48">
            <DropdownMenuLabel>Filter by Type</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => setTypeFilter('all')}>
              All Types
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setTypeFilter('memory')}>
              <Database className="h-4 w-4 mr-2 text-blue-500" />
              Memory
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setTypeFilter('redis')}>
              <Database className="h-4 w-4 mr-2 text-red-500" />
              Redis
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Cache Configs Table */}
      <Card>
        <CardHeader className="px-6 py-5">
          <CardTitle>Cache Configurations</CardTitle>
          <CardDescription>
            {loading
              ? 'Loading cache configurations...'
              : filteredConfigs.length > 0
              ? `Showing ${filteredConfigs.length} of ${cacheConfigs.length} cache configurations`
              : 'No cache configurations found'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center items-center py-8">
              <RefreshCw className="h-8 w-8 animate-spin text-slate-400" />
            </div>
          ) : filteredConfigs.length === 0 ? (
            <div className="text-center py-8 text-slate-500">
              <Database className="h-12 w-12 mx-auto mb-4 text-slate-300" />
              <p className="text-lg font-medium mb-1">
                No cache configurations found
              </p>
              <p className="text-sm">
                {searchQuery || typeFilter !== 'all'
                  ? 'Try adjusting your search or filters'
                  : 'Create a cache configuration to improve API performance'}
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[250px]">Route</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>TTL</TableHead>
                  <TableHead>Hit Rate</TableHead>
                  <TableHead>Items</TableHead>
                  <TableHead>Size</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredConfigs.map((config) => (
                  <TableRow key={config.id}>
                    <TableCell className="font-medium">
                      {config.route}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          config.type === 'memory' ? 'secondary' : 'outline'
                        }
                        className={
                          config.type === 'redis'
                            ? 'text-red-600 border-red-200 bg-red-50'
                            : ''
                        }
                      >
                        {config.type.charAt(0).toUpperCase() +
                          config.type.slice(1)}
                      </Badge>
                    </TableCell>
                    <TableCell>{config.ttl}</TableCell>
                    <TableCell>
                      <div className="flex items-center">
                        <span
                          className={
                            config.stats.hitRate > 80
                              ? 'text-green-600'
                              : config.stats.hitRate > 50
                              ? 'text-blue-600'
                              : config.stats.hitRate > 0
                              ? 'text-amber-600'
                              : 'text-slate-500'
                          }
                        >
                          {config.stats.hitRate}%
                        </span>
                        <span className="text-xs text-slate-500 ml-2">
                          ({config.stats.hits}/
                          {config.stats.hits + config.stats.misses})
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>{config.stats.itemCount}</TableCell>
                    <TableCell>
                      {config.stats.size > 0
                        ? `${(config.stats.size / (1024 * 1024)).toFixed(2)} MB`
                        : '0 MB'}
                    </TableCell>
                    <TableCell>
                      <Switch
                        checked={config.enabled}
                        onCheckedChange={() =>
                          handleToggleEnabled(config.id, config.enabled)
                        }
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          disabled={
                            clearingCache ||
                            !config.enabled ||
                            config.stats.itemCount === 0
                          }
                          onClick={() => handleClearCache(config.id)}
                          title="Clear Cache"
                        >
                          <MinusCircle className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEditDialog(config.id)}
                          title="Edit Configuration"
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

      {/* Edit Dialog */}
      {selectedConfig !== null && (
        <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
          <DialogContent className="sm:max-w-[650px]">
            <DialogHeader>
              <DialogTitle>Edit Cache Configuration</DialogTitle>
              <DialogDescription>
                Update caching settings for this API route.
              </DialogDescription>
            </DialogHeader>
            <div className="max-h-[70vh] overflow-y-auto py-4 pr-2">
              <div className="space-y-6">
                {/* Basic Settings */}
                <div className="space-y-4">
                  <h3 className="text-md font-semibold">Basic Settings</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="grid w-full items-center gap-1.5">
                      <Label htmlFor="edit-route">Route Pattern</Label>
                      <Input
                        id="edit-route"
                        placeholder="/api/products/*"
                        value={editCache.route || ''}
                        onChange={(e) =>
                          setEditCache({ ...editCache, route: e.target.value })
                        }
                      />
                    </div>

                    <div className="grid w-full items-center gap-1.5">
                      <Label htmlFor="edit-type">Cache Type</Label>
                      <Select
                        value={editCache.type || 'memory'}
                        onValueChange={(value) =>
                          setEditCache({ ...editCache, type: value })
                        }
                      >
                        <SelectTrigger id="edit-type">
                          <SelectValue placeholder="Select type" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="memory">Memory</SelectItem>
                          <SelectItem value="redis">Redis</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="grid w-full items-center gap-1.5">
                      <Label htmlFor="edit-ttl">Time to Live (TTL)</Label>
                      <Input
                        id="edit-ttl"
                        placeholder="5m"
                        value={editCache.ttl || ''}
                        onChange={(e) =>
                          setEditCache({ ...editCache, ttl: e.target.value })
                        }
                      />
                    </div>

                    <div className="grid w-full items-center gap-1.5">
                      <Label htmlFor="edit-maxSize">Maximum Size (bytes)</Label>
                      <Input
                        id="edit-maxSize"
                        type="number"
                        placeholder="1048576"
                        value={editCache.maxSize || 0}
                        onChange={(e) =>
                          setEditCache({
                            ...editCache,
                            maxSize: parseInt(e.target.value),
                          })
                        }
                      />
                    </div>
                  </div>
                </div>

                {/* HTTP Methods */}
                <Separator />
                <div className="space-y-4">
                  <h3 className="text-md font-semibold">HTTP Methods</h3>
                  <div className="space-y-2">
                    <Label>Methods to Cache</Label>
                    <div className="flex flex-wrap gap-2">
                      {[
                        'GET',
                        'POST',
                        'PUT',
                        'DELETE',
                        'PATCH',
                        'HEAD',
                        'OPTIONS',
                      ].map((method) => (
                        <Badge
                          key={method}
                          variant={
                            editCache.methods?.includes(method)
                              ? 'default'
                              : 'outline'
                          }
                          className="cursor-pointer"
                          onClick={() => {
                            editCache.methods?.includes(method)
                              ? removeMethod(method, true)
                              : addMethod(method, true);
                          }}
                        >
                          {method}
                          {editCache.methods?.includes(method) ? (
                            <X className="ml-1 h-3 w-3" />
                          ) : (
                            <Plus className="ml-1 h-3 w-3" />
                          )}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>

                {/* Cache Behavior */}
                <Separator />
                <div className="space-y-4">
                  <h3 className="text-md font-semibold">Cache Behavior</h3>

                  <div className="flex items-center space-x-2">
                    <Switch
                      id="edit-respectCacheControl"
                      checked={editCache.respectCacheControl}
                      onCheckedChange={(checked) =>
                        setEditCache({
                          ...editCache,
                          respectCacheControl: checked,
                        })
                      }
                    />
                    <Label htmlFor="edit-respectCacheControl">
                      Respect Cache-Control Headers
                    </Label>
                  </div>

                  <div className="flex items-center space-x-2">
                    <Switch
                      id="edit-includeHost"
                      checked={editCache.includeHost}
                      onCheckedChange={(checked) =>
                        setEditCache({ ...editCache, includeHost: checked })
                      }
                    />
                    <Label htmlFor="edit-includeHost">
                      Include Host in Cache Key
                    </Label>
                  </div>

                  <div className="flex items-center space-x-2">
                    <Switch
                      id="edit-enabled"
                      checked={editCache.enabled}
                      onCheckedChange={(checked) =>
                        setEditCache({ ...editCache, enabled: checked })
                      }
                    />
                    <Label htmlFor="edit-enabled">Enable Caching</Label>
                  </div>
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsEditDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button onClick={handleEditCache}>
                <Save className="h-4 w-4 mr-2" />
                Save Changes
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}

function StatCard({ title, value, description, icon, color }) {
  const colors = {
    blue: {
      bg: 'bg-blue-50',
      text: 'text-blue-600',
      icon: 'bg-blue-100',
      light: 'text-blue-500',
    },
    green: {
      bg: 'bg-emerald-50',
      text: 'text-emerald-600',
      icon: 'bg-emerald-100',
      light: 'text-emerald-500',
    },
    amber: {
      bg: 'bg-amber-50',
      text: 'text-amber-600',
      icon: 'bg-amber-100',
      light: 'text-amber-500',
    },
    red: {
      bg: 'bg-rose-50',
      text: 'text-rose-600',
      icon: 'bg-rose-100',
      light: 'text-rose-500',
    },
    purple: {
      bg: 'bg-violet-50',
      text: 'text-violet-600',
      icon: 'bg-violet-100',
      light: 'text-violet-500',
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
      </CardContent>
    </Card>
  );
}