import React, { useEffect, useState } from 'react';
import { Server } from 'lucide-react';

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

export default function ServerSettings({ settings, updateSettings }) {
  // Initialize local state with sensible defaults merged with passed settings
  const [serverSettings, setServerSettings] = useState({
    httpPort: 8080,
    adminPort: 8081,
    readTimeout: '30s',
    writeTimeout: '30s',
    idleTimeout: '120s',
    maxHeaderSize: 1048576, // 1MB
    enableHTTPS: false,
    tlsCertFile: '',
    tlsKeyFile: '',
    gracefulShutdownTimeout: '10s',
    ...settings, // Override defaults with any passed settings
  });
  
  // Sync local state with parent component when settings prop changes
  useEffect(() => {
    if (Object.keys(settings).length > 0) {
      setServerSettings({
        httpPort: 8080,
        adminPort: 8081,
        readTimeout: '30s',
        writeTimeout: '30s',
        idleTimeout: '120s',
        maxHeaderSize: 1048576,
        enableHTTPS: false,
        tlsCertFile: '',
        tlsKeyFile: '',
        gracefulShutdownTimeout: '10s',
        ...settings,
      });
    }
  }, [settings]);
  
  // Update local state and propagate changes to parent
  const handleChange = (field, value) => {
    const updatedSettings = {
      ...serverSettings,
      [field]: value,
    };
    
    setServerSettings(updatedSettings);
    updateSettings(updatedSettings);
  };
  
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Server className="h-5 w-5 mr-2" />
          Server Settings
        </CardTitle>
        <CardDescription>
          Configure the core server settings for your API Gateway.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="httpPort">HTTP Port</Label>
              <Input
                id="httpPort"
                type="number"
                min="1"
                max="65535"
                value={serverSettings.httpPort}
                onChange={(e) => handleChange('httpPort', parseInt(e.target.value))}
              />
              <p className="text-sm text-slate-500">
                The port on which the API Gateway will listen for HTTP traffic.
              </p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="adminPort">Admin Port</Label>
              <Input
                id="adminPort"
                type="number"
                min="1"
                max="65535"
                value={serverSettings.adminPort}
                onChange={(e) => handleChange('adminPort', parseInt(e.target.value))}
              />
              <p className="text-sm text-slate-500">
                The port on which the Admin API/UI will be available.
              </p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="maxHeaderSize">Max Header Size (bytes)</Label>
              <Input
                id="maxHeaderSize"
                type="number"
                value={serverSettings.maxHeaderSize}
                onChange={(e) => handleChange('maxHeaderSize', parseInt(e.target.value))}
              />
              <p className="text-sm text-slate-500">
                Maximum size of request headers.
              </p>
            </div>
          </div>
          
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="readTimeout">Read Timeout</Label>
              <Input
                id="readTimeout"
                placeholder="30s"
                value={serverSettings.readTimeout}
                onChange={(e) => handleChange('readTimeout', e.target.value)}
              />
              <p className="text-sm text-slate-500">
                Maximum time to read the entire request (e.g., 30s, 1m).
              </p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="writeTimeout">Write Timeout</Label>
              <Input
                id="writeTimeout"
                placeholder="30s"
                value={serverSettings.writeTimeout}
                onChange={(e) => handleChange('writeTimeout', e.target.value)}
              />
              <p className="text-sm text-slate-500">
                Maximum time to write the response (e.g., 30s, 1m).
              </p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="idleTimeout">Idle Timeout</Label>
              <Input
                id="idleTimeout"
                placeholder="120s"
                value={serverSettings.idleTimeout}
                onChange={(e) => handleChange('idleTimeout', e.target.value)}
              />
              <p className="text-sm text-slate-500">
                Maximum time for keep-alive connections (e.g., 120s, 5m).
              </p>
            </div>
          </div>
        </div>
        
        <Separator />
        
        <div className="space-y-4">
          <div className="flex items-center space-x-2">
            <Switch
              id="enableHTTPS"
              checked={serverSettings.enableHTTPS}
              onCheckedChange={(checked) => handleChange('enableHTTPS', checked)}
            />
            <Label htmlFor="enableHTTPS">Enable HTTPS</Label>
          </div>
          
          {serverSettings.enableHTTPS && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pl-6 mt-2">
              <div className="space-y-2">
                <Label htmlFor="tlsCertFile">TLS Certificate File Path</Label>
                <Input
                  id="tlsCertFile"
                  placeholder="/path/to/cert.pem"
                  value={serverSettings.tlsCertFile}
                  onChange={(e) => handleChange('tlsCertFile', e.target.value)}
                />
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="tlsKeyFile">TLS Key File Path</Label>
                <Input
                  id="tlsKeyFile"
                  placeholder="/path/to/key.pem"
                  value={serverSettings.tlsKeyFile}
                  onChange={(e) => handleChange('tlsKeyFile', e.target.value)}
                />
              </div>
            </div>
          )}
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="gracefulShutdownTimeout">Graceful Shutdown Timeout</Label>
          <Input
            id="gracefulShutdownTimeout"
            placeholder="10s"
            value={serverSettings.gracefulShutdownTimeout}
            onChange={(e) => handleChange('gracefulShutdownTimeout', e.target.value)}
          />
          <p className="text-sm text-slate-500">
            Time to wait for active connections to complete before shutdown.
          </p>
        </div>
      </CardContent>
    </Card>
  );
}