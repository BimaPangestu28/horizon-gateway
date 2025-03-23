import React, { useEffect, useState } from 'react';
import { Shield, X } from 'lucide-react';

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
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export default function SecuritySettings({ settings, updateSettings }) {
  // Initialize local state with sensible defaults merged with passed settings
  const [securitySettings, setSecuritySettings] = useState({
    enableTLS: false,
    minTLSVersion: 'TLS1.2',
    enableHSTS: false,
    hstsMaxAge: 31536000, // 1 year
    enableCSRF: true,
    csrfTokenExpiry: '24h',
    enableRateLimiting: true,
    globalRateLimitRequests: 1000,
    globalRateLimitPeriod: '1m',
    ipWhitelist: [],
    ipBlacklist: [],
    ...settings, // Override defaults with any passed settings
  });

  // For adding new IPs
  const [newIP, setNewIP] = useState('');

  // Sync local state with parent component when settings prop changes
  useEffect(() => {
    if (Object.keys(settings).length > 0) {
      setSecuritySettings({
        enableTLS: false,
        minTLSVersion: 'TLS1.2',
        enableHSTS: false,
        hstsMaxAge: 31536000, // 1 year
        enableCSRF: true,
        csrfTokenExpiry: '24h',
        enableRateLimiting: true,
        globalRateLimitRequests: 1000,
        globalRateLimitPeriod: '1m',
        ipWhitelist: [],
        ipBlacklist: [],
        ...settings,
      });
    }
  }, [settings]);

  // Update local state and propagate changes to parent
  const handleChange = (field, value) => {
    const updatedSettings = {
      ...securitySettings,
      [field]: value,
    };

    setSecuritySettings(updatedSettings);
    updateSettings(updatedSettings);
  };

  const handleAddIP = (list) => {
    if (!newIP.trim()) return;

    if (list === 'whitelist') {
      handleChange('ipWhitelist', [
        ...securitySettings.ipWhitelist,
        newIP.trim(),
      ]);
    } else {
      handleChange('ipBlacklist', [
        ...securitySettings.ipBlacklist,
        newIP.trim(),
      ]);
    }

    setNewIP('');
  };

  const handleRemoveIP = (ip, list) => {
    if (list === 'whitelist') {
      handleChange(
        'ipWhitelist',
        securitySettings.ipWhitelist.filter((item) => item !== ip),
      );
    } else {
      handleChange(
        'ipBlacklist',
        securitySettings.ipBlacklist.filter((item) => item !== ip),
      );
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Shield className="h-5 w-5 mr-2" />
          Security Settings
        </CardTitle>
        <CardDescription>
          Configure security features and protection mechanisms.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch
                id="enableTLS"
                checked={securitySettings.enableTLS}
                onCheckedChange={(checked) =>
                  handleChange('enableTLS', checked)
                }
              />
              <Label htmlFor="enableTLS">Enforce TLS for All Connections</Label>
            </div>

            {securitySettings.enableTLS && (
              <div className="space-y-2 pl-6">
                <Label htmlFor="minTLSVersion">Minimum TLS Version</Label>
                <Select
                  value={securitySettings.minTLSVersion}
                  onValueChange={(value) =>
                    handleChange('minTLSVersion', value)
                  }
                >
                  <SelectTrigger id="minTLSVersion">
                    <SelectValue placeholder="Select TLS version" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="TLS1.0">
                      TLS 1.0 (Not Recommended)
                    </SelectItem>
                    <SelectItem value="TLS1.1">
                      TLS 1.1 (Not Recommended)
                    </SelectItem>
                    <SelectItem value="TLS1.2">TLS 1.2</SelectItem>
                    <SelectItem value="TLS1.3">TLS 1.3</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-sm text-slate-500">
                  Minimum TLS version required for connections.
                </p>
              </div>
            )}

            <div className="flex items-center space-x-2">
              <Switch
                id="enableHSTS"
                checked={securitySettings.enableHSTS}
                onCheckedChange={(checked) =>
                  handleChange('enableHSTS', checked)
                }
              />
              <Label htmlFor="enableHSTS">
                Enable HTTP Strict Transport Security (HSTS)
              </Label>
            </div>

            {securitySettings.enableHSTS && (
              <div className="space-y-2 pl-6">
                <Label htmlFor="hstsMaxAge">HSTS Max Age (seconds)</Label>
                <Input
                  id="hstsMaxAge"
                  type="number"
                  value={securitySettings.hstsMaxAge}
                  onChange={(e) =>
                    handleChange('hstsMaxAge', parseInt(e.target.value))
                  }
                />
                <p className="text-sm text-slate-500">
                  Time browsers should remember to use HTTPS (31536000 = 1
                  year).
                </p>
              </div>
            )}

            <div className="flex items-center space-x-2">
              <Switch
                id="enableCSRF"
                checked={securitySettings.enableCSRF}
                onCheckedChange={(checked) =>
                  handleChange('enableCSRF', checked)
                }
              />
              <Label htmlFor="enableCSRF">Enable CSRF Protection</Label>
            </div>

            {securitySettings.enableCSRF && (
              <div className="space-y-2 pl-6">
                <Label htmlFor="csrfTokenExpiry">CSRF Token Expiry</Label>
                <Input
                  id="csrfTokenExpiry"
                  placeholder="24h"
                  value={securitySettings.csrfTokenExpiry}
                  onChange={(e) =>
                    handleChange('csrfTokenExpiry', e.target.value)
                  }
                />
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch
                id="enableRateLimiting"
                checked={securitySettings.enableRateLimiting}
                onCheckedChange={(checked) =>
                  handleChange('enableRateLimiting', checked)
                }
              />
              <Label htmlFor="enableRateLimiting">
                Enable Global Rate Limiting
              </Label>
            </div>

            {securitySettings.enableRateLimiting && (
              <div className="space-y-4 pl-6">
                <div className="space-y-2">
                  <Label htmlFor="globalRateLimitRequests">
                    Global Rate Limit (requests)
                  </Label>
                  <Input
                    id="globalRateLimitRequests"
                    type="number"
                    value={securitySettings.globalRateLimitRequests}
                    onChange={(e) =>
                      handleChange(
                        'globalRateLimitRequests',
                        parseInt(e.target.value),
                      )
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="globalRateLimitPeriod">
                    Rate Limit Period
                  </Label>
                  <Input
                    id="globalRateLimitPeriod"
                    placeholder="1m"
                    value={securitySettings.globalRateLimitPeriod}
                    onChange={(e) =>
                      handleChange('globalRateLimitPeriod', e.target.value)
                    }
                  />
                  <p className="text-sm text-slate-500">
                    Period for global rate limiting (e.g., 1m = 1000 requests
                    per minute).
                  </p>
                </div>
              </div>
            )}

            <div className="space-y-2">
              <Label>IP Whitelist</Label>
              <div className="flex flex-wrap gap-2">
                {securitySettings.ipWhitelist.map((ip) => (
                  <Badge key={ip} variant="secondary" className="gap-1">
                    {ip}
                    <X
                      className="h-3 w-3 cursor-pointer"
                      onClick={() => handleRemoveIP(ip, 'whitelist')}
                    />
                  </Badge>
                ))}
              </div>
              <div className="flex gap-2 mt-1">
                <Input
                  placeholder="Add IP address or CIDR"
                  value={newIP}
                  onChange={(e) => setNewIP(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddIP('whitelist');
                    }
                  }}
                />
                <Button
                  type="button"
                  size="sm"
                  onClick={() => handleAddIP('whitelist')}
                >
                  Add
                </Button>
              </div>
              <p className="text-sm text-slate-500">
                IP addresses always allowed to access the gateway.
              </p>
            </div>

            <div className="space-y-2">
              <Label>IP Blacklist</Label>
              <div className="flex flex-wrap gap-2">
                {securitySettings.ipBlacklist.map((ip) => (
                  <Badge key={ip} variant="destructive" className="gap-1">
                    {ip}
                    <X
                      className="h-3 w-3 cursor-pointer"
                      onClick={() => handleRemoveIP(ip, 'blacklist')}
                    />
                  </Badge>
                ))}
              </div>
              <div className="flex gap-2 mt-1">
                <Input
                  placeholder="Add IP address or CIDR"
                  value={newIP}
                  onChange={(e) => setNewIP(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddIP('blacklist');
                    }
                  }}
                />
                <Button
                  type="button"
                  size="sm"
                  onClick={() => handleAddIP('blacklist')}
                >
                  Add
                </Button>
              </div>
              <p className="text-sm text-slate-500">
                IP addresses blocked from accessing the gateway.
              </p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
