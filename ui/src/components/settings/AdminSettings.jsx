import React, { useEffect, useState } from 'react';
import { User, X } from 'lucide-react';

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
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

export default function AdminSettings({ settings, updateSettings }) {
  // Initialize local state with sensible defaults merged with passed settings
  const [adminSettings, setAdminSettings] = useState({
    enableAdminAPI: true,
    enableAdminUI: true,
    adminAPIAuth: true,
    corsAllowOrigins: ['*'],
    maxRequestSize: 5242880, // 5MB
    rateLimitRequests: 100,
    rateLimitPeriod: '1m',
    sessionTimeout: '24h',
    ...settings, // Override defaults with any passed settings
  });
  
  // For adding new CORS origins
  const [newCorsOrigin, setNewCorsOrigin] = useState('');
  
  // Sync local state with parent component when settings prop changes
  useEffect(() => {
    if (Object.keys(settings).length > 0) {
      setAdminSettings({
        enableAdminAPI: true,
        enableAdminUI: true,
        adminAPIAuth: true,
        corsAllowOrigins: ['*'],
        maxRequestSize: 5242880,
        rateLimitRequests: 100,
        rateLimitPeriod: '1m',
        sessionTimeout: '24h',
        ...settings,
      });
    }
  }, [settings]);
  
  // Update local state and propagate changes to parent
  const handleChange = (field, value) => {
    const updatedSettings = {
      ...adminSettings,
      [field]: value,
    };
    
    setAdminSettings(updatedSettings);
    updateSettings(updatedSettings);
  };
  
  const handleAddCorsOrigin = () => {
    if (!newCorsOrigin.trim()) return;
    
    const updatedOrigins = [...adminSettings.corsAllowOrigins, newCorsOrigin.trim()];
    handleChange('corsAllowOrigins', updatedOrigins);
    setNewCorsOrigin('');
  };
  
  const handleRemoveCorsOrigin = (origin) => {
    const updatedOrigins = adminSettings.corsAllowOrigins.filter(item => item !== origin);
    handleChange('corsAllowOrigins', updatedOrigins);
  };
  
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <User className="h-5 w-5 mr-2" />
          Admin Interface Settings
        </CardTitle>
        <CardDescription>
          Configure the admin API and user interface behavior.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch
                id="enableAdminAPI"
                checked={adminSettings.enableAdminAPI}
                onCheckedChange={(checked) => handleChange('enableAdminAPI', checked)}
              />
              <Label htmlFor="enableAdminAPI">Enable Admin API</Label>
            </div>
            
            <div className="flex items-center space-x-2">
              <Switch
                id="enableAdminUI"
                checked={adminSettings.enableAdminUI}
                onCheckedChange={(checked) => handleChange('enableAdminUI', checked)}
              />
              <Label htmlFor="enableAdminUI">Enable Admin UI</Label>
            </div>
            
            <div className="flex items-center space-x-2">
              <Switch
                id="adminAPIAuth"
                checked={adminSettings.adminAPIAuth}
                onCheckedChange={(checked) => handleChange('adminAPIAuth', checked)}
              />
              <Label htmlFor="adminAPIAuth">Require Authentication for Admin API</Label>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="sessionTimeout">Session Timeout</Label>
              <Input
                id="sessionTimeout"
                placeholder="24h"
                value={adminSettings.sessionTimeout}
                onChange={(e) => handleChange('sessionTimeout', e.target.value)}
              />
              <p className="text-sm text-slate-500">
                How long admin sessions remain valid before requiring re-authentication.
              </p>
            </div>
          </div>
          
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="maxRequestSize">Maximum Request Size (bytes)</Label>
              <Input
                id="maxRequestSize"
                type="number"
                value={adminSettings.maxRequestSize}
                onChange={(e) => handleChange('maxRequestSize', parseInt(e.target.value))}
              />
              <p className="text-sm text-slate-500">
                Maximum size of requests to the admin API (5242880 = 5MB).
              </p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="rateLimitRequests">Rate Limit (requests)</Label>
              <Input
                id="rateLimitRequests"
                type="number"
                value={adminSettings.rateLimitRequests}
                onChange={(e) => handleChange('rateLimitRequests', parseInt(e.target.value))}
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="rateLimitPeriod">Rate Limit Period</Label>
              <Input
                id="rateLimitPeriod"
                placeholder="1m"
                value={adminSettings.rateLimitPeriod}
                onChange={(e) => handleChange('rateLimitPeriod', e.target.value)}
              />
              <p className="text-sm text-slate-500">
                Period for rate limiting (e.g., 1m = 100 requests per minute).
              </p>
            </div>
          </div>
        </div>
        
        <Separator />
        
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>CORS Allow Origins</Label>
            <div className="flex flex-wrap gap-2">
              {adminSettings.corsAllowOrigins.map((origin) => (
                <Badge key={origin} variant="secondary" className="gap-1">
                  {origin}
                  <X 
                    className="h-3 w-3 cursor-pointer" 
                    onClick={() => handleRemoveCorsOrigin(origin)}
                  />
                </Badge>
              ))}
            </div>
            <div className="flex gap-2 mt-1">
              <Input
                placeholder="Add origin (e.g., https://admin.example.com)"
                value={newCorsOrigin}
                onChange={(e) => setNewCorsOrigin(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleAddCorsOrigin();
                  }
                }}
              />
              <Button 
                type="button" 
                size="sm"
                onClick={handleAddCorsOrigin}
              >
                Add
              </Button>
            </div>
            <p className="text-sm text-slate-500">
              Domains allowed to access the Admin API (use * for all origins).
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}