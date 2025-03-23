import React, { useState } from 'react';
import {
  Settings as SettingsIcon,
  Save,
  RefreshCw,
  Upload,
  Download,
  Check,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/components/ui/use-toast.jsx';
import { ToastAction } from '@/components/ui/toast';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

import ServerSettings from '@/components/settings/ServerSettings';
import LoggingSettings from '@/components/settings/LoggingSettings';
import MonitoringSettings from '@/components/settings/MonitoringSettings';
import AdminSettings from '@/components/settings/AdminSettings';
import StorageSettings from '@/components/settings/StorageSettings';
import SecuritySettings from '@/components/settings/SecuritySettings';
import EmailSettings from '@/components/settings/EmailSettings';

import ApiService from '@/services/ApiService';

export default function Settings() {
  const { toast } = useToast();
  const [saving, setSaving] = useState(false);
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [configImport, setConfigImport] = useState('');

  const [settings, setSettings] = useState({
    server: {},
    logging: {},
    monitoring: {},
    admin: {},
    storage: {},
    security: {},
    email: {},
  });

  const updateSettings = (category, newSettings) => {
    setSettings((prev) => ({
      ...prev,
      [category]: newSettings,
    }));
  };

  const saveSettings = async () => {
    setSaving(true);
    try {
      // In a real implementation, this would call the API
      // await ApiService.saveSettings(settings);

      // Simulate API delay
      await new Promise((resolve) => setTimeout(resolve, 1500));

      toast({
        title: 'Settings saved successfully',
        description: 'Your gateway settings have been updated.',
        action: (
          <ToastAction altText="OK">
            <Check className="h-4 w-4" />
          </ToastAction>
        ),
      });
    } catch (error) {
      console.error('Failed to save settings:', error);
      toast({
        variant: 'destructive',
        title: 'Failed to save settings',
        description:
          error.message || 'There was an error saving your settings.',
        action: <ToastAction altText="Try again">Try Again</ToastAction>,
      });
    } finally {
      setSaving(false);
    }
  };

  const exportSettings = () => {
    const blob = new Blob([JSON.stringify(settings, null, 2)], {
      type: 'application/json',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'horizon-settings.json';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    toast({
      title: 'Settings exported',
      description: 'Your settings have been exported to a JSON file.',
    });
  };

  const importSettings = () => {
    try {
      const importedSettings = JSON.parse(configImport);
      setSettings((prev) => ({
        ...prev,
        ...importedSettings,
      }));

      setIsImportDialogOpen(false);
      setConfigImport('');

      toast({
        title: 'Settings imported',
        description:
          "Your settings have been imported successfully. Don't forget to save them.",
      });
    } catch (error) {
      console.error('Failed to import settings:', error);
      toast({
        variant: 'destructive',
        title: 'Import failed',
        description:
          'The provided JSON is invalid. Please check the format and try again.',
      });
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Settings</h1>
          <p className="text-sm text-slate-500 mt-1">
            Configure your Horizon API Gateway settings.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Dialog
            open={isImportDialogOpen}
            onOpenChange={setIsImportDialogOpen}
          >
            <DialogTrigger asChild>
              <Button variant="outline" size="sm">
                <Upload className="h-4 w-4 mr-1" />
                Import
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Import Settings</DialogTitle>
                <DialogDescription>
                  Paste your Horizon API Gateway settings JSON below.
                </DialogDescription>
              </DialogHeader>
              <div className="py-4">
                <Textarea
                  className="h-60 font-mono text-xs"
                  placeholder="Paste your Horizon settings JSON here..."
                  value={configImport}
                  onChange={(e) => setConfigImport(e.target.value)}
                />
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setIsImportDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button onClick={importSettings}>Import Settings</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          <Button variant="outline" size="sm" onClick={exportSettings}>
            <Download className="h-4 w-4 mr-1" />
            Export
          </Button>

          <Button size="sm" disabled={saving} onClick={saveSettings}>
            {saving ? (
              <>
                <RefreshCw className="h-4 w-4 mr-1 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="h-4 w-4 mr-1" />
                Save Settings
              </>
            )}
          </Button>
        </div>
      </div>

      <Tabs defaultValue="server">
        <TabsList className="grid grid-cols-7 lg:w-auto w-full mb-4">
          <TabsTrigger value="server">Server</TabsTrigger>
          <TabsTrigger value="logging">Logging</TabsTrigger>
          <TabsTrigger value="monitoring">Monitoring</TabsTrigger>
          <TabsTrigger value="admin">Admin</TabsTrigger>
          <TabsTrigger value="storage">Storage</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
          <TabsTrigger value="email">Email</TabsTrigger>
        </TabsList>

        {/* Server Settings */}
        <TabsContent value="server">
          <ServerSettings
            settings={settings.server}
            updateSettings={(newSettings) =>
              updateSettings('server', newSettings)
            }
          />
        </TabsContent>

        {/* Logging Settings */}
        <TabsContent value="logging">
          <LoggingSettings
            settings={settings.logging}
            updateSettings={(newSettings) =>
              updateSettings('logging', newSettings)
            }
          />
        </TabsContent>

        {/* Monitoring Settings */}
        <TabsContent value="monitoring">
          <MonitoringSettings
            settings={settings.monitoring}
            updateSettings={(newSettings) =>
              updateSettings('monitoring', newSettings)
            }
          />
        </TabsContent>

        {/* Admin Settings */}
        <TabsContent value="admin">
          <AdminSettings
            settings={settings.admin}
            updateSettings={(newSettings) =>
              updateSettings('admin', newSettings)
            }
          />
        </TabsContent>

        {/* Storage Settings */}
        <TabsContent value="storage">
          <StorageSettings
            settings={settings.storage}
            updateSettings={(newSettings) =>
              updateSettings('storage', newSettings)
            }
          />
        </TabsContent>

        {/* Security Settings */}
        <TabsContent value="security">
          <SecuritySettings
            settings={settings.security}
            updateSettings={(newSettings) =>
              updateSettings('security', newSettings)
            }
          />
        </TabsContent>

        {/* Email Settings */}
        <TabsContent value="email">
          <EmailSettings
            settings={settings.email}
            updateSettings={(newSettings) =>
              updateSettings('email', newSettings)
            }
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
