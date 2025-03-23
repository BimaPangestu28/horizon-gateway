import React, { useEffect, useState } from 'react';
import { FileText } from 'lucide-react';

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
import { Separator } from '@/components/ui/separator';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export default function LoggingSettings({ settings, updateSettings }) {
  // Initialize local state with sensible defaults merged with passed settings
  const [loggingSettings, setLoggingSettings] = useState({
    level: 'info',
    format: 'json',
    output: 'stdout',
    logFile: '',
    enableAccessLogs: true,
    enableErrorLogs: true,
    enableMetricsLogs: true,
    enableAuditLogs: false,
    requestSampling: 100, // percentage
    ...settings, // Override defaults with any passed settings
  });
  
  // Sync local state with parent component when settings prop changes
  useEffect(() => {
    if (Object.keys(settings).length > 0) {
      setLoggingSettings({
        level: 'info',
        format: 'json',
        output: 'stdout',
        logFile: '',
        enableAccessLogs: true,
        enableErrorLogs: true,
        enableMetricsLogs: true,
        enableAuditLogs: false,
        requestSampling: 100,
        ...settings,
      });
    }
  }, [settings]);
  
  // Update local state and propagate changes to parent
  const handleChange = (field, value) => {
    const updatedSettings = {
      ...loggingSettings,
      [field]: value,
    };
    
    setLoggingSettings(updatedSettings);
    updateSettings(updatedSettings);
  };
  
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <FileText className="h-5 w-5 mr-2" />
          Logging Settings
        </CardTitle>
        <CardDescription>
          Configure how the API Gateway logs activities and errors.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="logLevel">Log Level</Label>
              <Select
                value={loggingSettings.level}
                onValueChange={(value) => handleChange('level', value)}
              >
                <SelectTrigger id="logLevel">
                  <SelectValue placeholder="Select log level" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="debug">Debug</SelectItem>
                  <SelectItem value="info">Info</SelectItem>
                  <SelectItem value="warn">Warning</SelectItem>
                  <SelectItem value="error">Error</SelectItem>
                  <SelectItem value="fatal">Fatal</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-slate-500">
                The minimum severity level of logs to output.
              </p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="logFormat">Log Format</Label>
              <Select
                value={loggingSettings.format}
                onValueChange={(value) => handleChange('format', value)}
              >
                <SelectTrigger id="logFormat">
                  <SelectValue placeholder="Select log format" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="json">JSON</SelectItem>
                  <SelectItem value="text">Plain Text</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-slate-500">
                The format in which logs will be output.
              </p>
            </div>
          </div>
          
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="logOutput">Log Output</Label>
              <Select
                value={loggingSettings.output}
                onValueChange={(value) => handleChange('output', value)}
              >
                <SelectTrigger id="logOutput">
                  <SelectValue placeholder="Select log output" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="stdout">Standard Output</SelectItem>
                  <SelectItem value="file">File</SelectItem>
                  <SelectItem value="both">Both</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-slate-500">
                Where logs should be written.
              </p>
            </div>
            
            {(loggingSettings.output === 'file' || loggingSettings.output === 'both') && (
              <div className="space-y-2">
                <Label htmlFor="logFile">Log File Path</Label>
                <Input
                  id="logFile"
                  placeholder="/var/log/horizon/gateway.log"
                  value={loggingSettings.logFile}
                  onChange={(e) => handleChange('logFile', e.target.value)}
                />
                <p className="text-sm text-slate-500">
                  The file path where logs will be written.
                </p>
              </div>
            )}
          </div>
        </div>
        
        <Separator />
        
        <div className="space-y-4">
          <h3 className="text-sm font-medium">Log Types</h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex items-center space-x-2">
              <Switch
                id="enableAccessLogs"
                checked={loggingSettings.enableAccessLogs}
                onCheckedChange={(checked) => handleChange('enableAccessLogs', checked)}
              />
              <Label htmlFor="enableAccessLogs">Enable Access Logs</Label>
            </div>
            
            <div className="flex items-center space-x-2">
              <Switch
                id="enableErrorLogs"
                checked={loggingSettings.enableErrorLogs}
                onCheckedChange={(checked) => handleChange('enableErrorLogs', checked)}
              />
              <Label htmlFor="enableErrorLogs">Enable Error Logs</Label>
            </div>
            
            <div className="flex items-center space-x-2">
              <Switch
                id="enableMetricsLogs"
                checked={loggingSettings.enableMetricsLogs}
                onCheckedChange={(checked) => handleChange('enableMetricsLogs', checked)}
              />
              <Label htmlFor="enableMetricsLogs">Enable Metrics Logs</Label>
            </div>
            
            <div className="flex items-center space-x-2">
              <Switch
                id="enableAuditLogs"
                checked={loggingSettings.enableAuditLogs}
                onCheckedChange={(checked) => handleChange('enableAuditLogs', checked)}
              />
              <Label htmlFor="enableAuditLogs">Enable Audit Logs</Label>
            </div>
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="requestSampling">
              Request Sampling Rate ({loggingSettings.requestSampling}%)
            </Label>
            <Input
              id="requestSampling"
              type="range"
              min="0"
              max="100"
              value={loggingSettings.requestSampling}
              onChange={(e) => handleChange('requestSampling', parseInt(e.target.value))}
            />
            <p className="text-sm text-slate-500">
              The percentage of requests to log. Reducing this can improve performance in high-traffic systems.
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}