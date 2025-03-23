import React, { useEffect, useState } from 'react';
import { Activity } from 'lucide-react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export default function MonitoringSettings({ settings, updateSettings }) {
  const [monitoringSettings, setMonitoringSettings] = useState({
    enablePrometheus: true,
    metricsEndpoint: '/metrics',
    enableHealthCheck: true,
    healthCheckEndpoint: '/health',
    enableProfiling: false,
    profilingEndpoint: '/debug/pprof',
    statsInterval: '10s',
    enableTracing: false,
    tracingExporter: 'jaeger',
    tracingEndpoint: 'http://localhost:14268/api/traces',
    tracingSamplingRate: 0.1,
    ...settings,
  });
  
  useEffect(() => {
    if (Object.keys(settings).length > 0) {
      setMonitoringSettings({
        enablePrometheus: true,
        metricsEndpoint: '/metrics',
        enableHealthCheck: true,
        healthCheckEndpoint: '/health',
        enableProfiling: false,
        profilingEndpoint: '/debug/pprof',
        statsInterval: '10s',
        enableTracing: false,
        tracingExporter: 'jaeger',
        tracingEndpoint: 'http://localhost:14268/api/traces',
        tracingSamplingRate: 0.1,
        ...settings,
      });
    }
  }, [settings]);
  
  const handleChange = (field, value) => {
    const updatedSettings = {
      ...monitoringSettings,
      [field]: value,
    };
    
    setMonitoringSettings(updatedSettings);
    updateSettings(updatedSettings);
  };
  
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Activity className="h-5 w-5 mr-2" />
          Monitoring Settings
        </CardTitle>
        <CardDescription>
          Configure metrics, health checks, and observability features.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch
                id="enablePrometheus"
                checked={monitoringSettings.enablePrometheus}
                onCheckedChange={(checked) => handleChange('enablePrometheus', checked)}
              />
              <Label htmlFor="enablePrometheus">Enable Prometheus Metrics</Label>
            </div>
            
            {monitoringSettings.enablePrometheus && (
              <div className="space-y-2 pl-6">
                <Label htmlFor="metricsEndpoint">Metrics Endpoint</Label>
                <Input
                  id="metricsEndpoint"
                  placeholder="/metrics"
                  value={monitoringSettings.metricsEndpoint}
                  onChange={(e) => handleChange('metricsEndpoint', e.target.value)}
                />
                <p className="text-sm text-slate-500">
                  The endpoint where Prometheus metrics will be exposed.
                </p>
              </div>
            )}
            
            <div className="flex items-center space-x-2">
              <Switch
                id="enableHealthCheck"
                checked={monitoringSettings.enableHealthCheck}
                onCheckedChange={(checked) => handleChange('enableHealthCheck', checked)}
              />
              <Label htmlFor="enableHealthCheck">Enable Health Check Endpoint</Label>
            </div>
            
            {monitoringSettings.enableHealthCheck && (
              <div className="space-y-2 pl-6">
                <Label htmlFor="healthCheckEndpoint">Health Check Endpoint</Label>
                <Input
                  id="healthCheckEndpoint"
                  placeholder="/health"
                  value={monitoringSettings.healthCheckEndpoint}
                  onChange={(e) => handleChange('healthCheckEndpoint', e.target.value)}
                />
                <p className="text-sm text-slate-500">
                  The endpoint where system health will be reported.
                </p>
              </div>
            )}
            
            <div className="space-y-2">
              <Label htmlFor="statsInterval">Statistics Collection Interval</Label>
              <Input
                id="statsInterval"
                placeholder="10s"
                value={monitoringSettings.statsInterval}
                onChange={(e) => handleChange('statsInterval', e.target.value)}
              />
              <p className="text-sm text-slate-500">
                How frequently to collect performance statistics.
              </p>
            </div>
          </div>
          
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch
                id="enableProfiling"
                checked={monitoringSettings.enableProfiling}
                onCheckedChange={(checked) => handleChange('enableProfiling', checked)}
              />
              <Label htmlFor="enableProfiling">Enable Performance Profiling</Label>
            </div>
            
            {monitoringSettings.enableProfiling && (
              <div className="space-y-2 pl-6">
                <Label htmlFor="profilingEndpoint">Profiling Endpoint</Label>
                <Input
                  id="profilingEndpoint"
                  placeholder="/debug/pprof"
                  value={monitoringSettings.profilingEndpoint}
                  onChange={(e) => handleChange('profilingEndpoint', e.target.value)}
                />
                <p className="text-sm text-slate-500 text-red-400">
                  Warning: Enable only in development or debugging environments.
                </p>
              </div>
            )}
            
            <div className="flex items-center space-x-2">
              <Switch
                id="enableTracing"
                checked={monitoringSettings.enableTracing}
                onCheckedChange={(checked) => handleChange('enableTracing', checked)}
              />
              <Label htmlFor="enableTracing">Enable Distributed Tracing</Label>
            </div>
            
            {monitoringSettings.enableTracing && (
              <div className="space-y-4 pl-6">
                <div className="space-y-2">
                  <Label htmlFor="tracingExporter">Tracing Exporter</Label>
                  <Select
                    value={monitoringSettings.tracingExporter}
                    onValueChange={(value) => handleChange('tracingExporter', value)}
                  >
                    <SelectTrigger id="tracingExporter">
                      <SelectValue placeholder="Select exporter" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="jaeger">Jaeger</SelectItem>
                      <SelectItem value="zipkin">Zipkin</SelectItem>
                      <SelectItem value="otlp">OTLP</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="tracingEndpoint">Tracing Endpoint</Label>
                  <Input
                    id="tracingEndpoint"
                    placeholder="http://localhost:14268/api/traces"
                    value={monitoringSettings.tracingEndpoint}
                    onChange={(e) => handleChange('tracingEndpoint', e.target.value)}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="tracingSamplingRate">
                    Sampling Rate: {monitoringSettings.tracingSamplingRate * 100}%
                  </Label>
                  <Input
                    id="tracingSamplingRate"
                    type="range"
                    min="0"
                    max="1"
                    step="0.01"
                    value={monitoringSettings.tracingSamplingRate}
                    onChange={(e) => handleChange('tracingSamplingRate', parseFloat(e.target.value))}
                  />
                  <p className="text-sm text-slate-500">
                    Percentage of requests to trace. Lower values reduce overhead.
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}