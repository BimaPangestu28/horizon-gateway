import React, { useEffect, useState } from 'react';
import { Database } from 'lucide-react';

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

export default function StorageSettings({ settings, updateSettings }) {
  // Initialize local state with sensible defaults merged with passed settings
  const [storageSettings, setStorageSettings] = useState({
    configStorageType: 'file',
    configFilePath: 'config.yaml',
    enableConfigBackups: true,
    maxConfigBackups: 10,
    dataStorageType: 'badger',
    dataStoragePath: './data',
    enableRedis: false,
    redisAddress: 'localhost:6379',
    redisPassword: '',
    redisDB: 0,
    ...settings, // Override defaults with any passed settings
  });

  // Sync local state with parent component when settings prop changes
  useEffect(() => {
    if (Object.keys(settings).length > 0) {
      setStorageSettings({
        configStorageType: 'file',
        configFilePath: 'config.yaml',
        enableConfigBackups: true,
        maxConfigBackups: 10,
        dataStorageType: 'badger',
        dataStoragePath: './data',
        enableRedis: false,
        redisAddress: 'localhost:6379',
        redisPassword: '',
        redisDB: 0,
        ...settings,
      });
    }
  }, [settings]);

  // Update local state and propagate changes to parent
  const handleChange = (field, value) => {
    const updatedSettings = {
      ...storageSettings,
      [field]: value,
    };

    setStorageSettings(updatedSettings);
    updateSettings(updatedSettings);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Database className="h-5 w-5 mr-2" />
          Storage Settings
        </CardTitle>
        <CardDescription>
          Configure data storage and persistence options.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="configStorageType">
                Configuration Storage Type
              </Label>
              <Select
                value={storageSettings.configStorageType}
                onValueChange={(value) =>
                  handleChange('configStorageType', value)
                }
              >
                <SelectTrigger id="configStorageType">
                  <SelectValue placeholder="Select storage type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="file">File</SelectItem>
                  <SelectItem value="database">Database</SelectItem>
                  <SelectItem value="etcd">Etcd</SelectItem>
                  <SelectItem value="consul">Consul</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-slate-500">
                Where to store gateway configuration.
              </p>
            </div>

            {storageSettings.configStorageType === 'file' && (
              <div className="space-y-2">
                <Label htmlFor="configFilePath">Configuration File Path</Label>
                <Input
                  id="configFilePath"
                  placeholder="config.yaml"
                  value={storageSettings.configFilePath}
                  onChange={(e) =>
                    handleChange('configFilePath', e.target.value)
                  }
                />
              </div>
            )}

            <div className="flex items-center space-x-2">
              <Switch
                id="enableConfigBackups"
                checked={storageSettings.enableConfigBackups}
                onCheckedChange={(checked) =>
                  handleChange('enableConfigBackups', checked)
                }
              />
              <Label htmlFor="enableConfigBackups">
                Enable Configuration Backups
              </Label>
            </div>

            {storageSettings.enableConfigBackups && (
              <div className="space-y-2 pl-6">
                <Label htmlFor="maxConfigBackups">Maximum Backup Files</Label>
                <Input
                  id="maxConfigBackups"
                  type="number"
                  value={storageSettings.maxConfigBackups}
                  onChange={(e) =>
                    handleChange('maxConfigBackups', parseInt(e.target.value))
                  }
                />
                <p className="text-sm text-slate-500">
                  Maximum number of configuration backups to keep.
                </p>
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="dataStorageType">Data Storage Type</Label>
              <Select
                value={storageSettings.dataStorageType}
                onValueChange={(value) =>
                  handleChange('dataStorageType', value)
                }
              >
                <SelectTrigger id="dataStorageType">
                  <SelectValue placeholder="Select data storage type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="badger">Badger DB (Embedded)</SelectItem>
                  <SelectItem value="postgres">PostgreSQL</SelectItem>
                  <SelectItem value="mysql">MySQL</SelectItem>
                  <SelectItem value="redis">Redis</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-slate-500">
                Where to store operational data like keys, analytics, etc.
              </p>
            </div>

            {storageSettings.dataStorageType === 'badger' && (
              <div className="space-y-2">
                <Label htmlFor="dataStoragePath">Data Directory Path</Label>
                <Input
                  id="dataStoragePath"
                  placeholder="./data"
                  value={storageSettings.dataStoragePath}
                  onChange={(e) =>
                    handleChange('dataStoragePath', e.target.value)
                  }
                />
              </div>
            )}

            <div className="flex items-center space-x-2">
              <Switch
                id="enableRedis"
                checked={storageSettings.enableRedis}
                onCheckedChange={(checked) =>
                  handleChange('enableRedis', checked)
                }
              />
              <Label htmlFor="enableRedis">
                Enable Redis for Caching/Rate Limiting
              </Label>
            </div>

            {storageSettings.enableRedis && (
              <div className="space-y-4 pl-6">
                <div className="space-y-2">
                  <Label htmlFor="redisAddress">Redis Server Address</Label>
                  <Input
                    id="redisAddress"
                    placeholder="localhost:6379"
                    value={storageSettings.redisAddress}
                    onChange={(e) =>
                      handleChange('redisAddress', e.target.value)
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="redisPassword">Redis Password</Label>
                  <Input
                    id="redisPassword"
                    type="password"
                    placeholder="(optional)"
                    value={storageSettings.redisPassword}
                    onChange={(e) =>
                      handleChange('redisPassword', e.target.value)
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="redisDB">Redis Database Number</Label>
                  <Input
                    id="redisDB"
                    type="number"
                    min="0"
                    max="15"
                    value={storageSettings.redisDB}
                    onChange={(e) =>
                      handleChange('redisDB', parseInt(e.target.value))
                    }
                  />
                </div>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
