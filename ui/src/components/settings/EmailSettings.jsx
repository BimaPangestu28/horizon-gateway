import React, { useEffect, useState } from 'react';
import { Mail, X } from 'lucide-react';

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
import { useToast } from '@/components/ui/use-toast.jsx';

export default function EmailSettings({ settings, updateSettings }) {
  const { toast } = useToast();

  // Initialize local state with sensible defaults merged with passed settings
  const [emailSettings, setEmailSettings] = useState({
    smtpServer: '',
    smtpPort: 587,
    smtpUsername: '',
    smtpPassword: '',
    emailFrom: 'noreply@horizon-gateway.org',
    enableTLS: true,
    enableAlerts: false,
    alertRecipients: [],
    ...settings, // Override defaults with any passed settings
  });

  // For adding new alert email recipients
  const [newAlertEmail, setNewAlertEmail] = useState('');

  // Sync local state with parent component when settings prop changes
  useEffect(() => {
    if (Object.keys(settings).length > 0) {
      setEmailSettings({
        smtpServer: '',
        smtpPort: 587,
        smtpUsername: '',
        smtpPassword: '',
        emailFrom: 'noreply@horizon-gateway.org',
        enableTLS: true,
        enableAlerts: false,
        alertRecipients: [],
        ...settings,
      });
    }
  }, [settings]);

  // Update local state and propagate changes to parent
  const handleChange = (field, value) => {
    const updatedSettings = {
      ...emailSettings,
      [field]: value,
    };

    setEmailSettings(updatedSettings);
    updateSettings(updatedSettings);
  };

  const handleAddAlertEmail = () => {
    if (!newAlertEmail.trim()) return;

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(newAlertEmail)) {
      toast({
        variant: 'destructive',
        title: 'Invalid email address',
        description: 'Please enter a valid email address.',
      });
      return;
    }

    handleChange('alertRecipients', [
      ...emailSettings.alertRecipients,
      newAlertEmail.trim(),
    ]);
    setNewAlertEmail('');
  };

  const handleRemoveAlertEmail = (email) => {
    handleChange(
      'alertRecipients',
      emailSettings.alertRecipients.filter((item) => item !== email),
    );
  };

  const handleSendTestEmail = () => {
    // In a real implementation, this would call the API
    toast({
      title: 'Test email sent',
      description: 'A test email has been sent to all recipients.',
    });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Mail className="h-5 w-5 mr-2" />
          Email Settings
        </CardTitle>
        <CardDescription>
          Configure email settings for alerts and notifications.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="smtpServer">SMTP Server</Label>
              <Input
                id="smtpServer"
                placeholder="smtp.example.com"
                value={emailSettings.smtpServer}
                onChange={(e) => handleChange('smtpServer', e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="smtpPort">SMTP Port</Label>
              <Input
                id="smtpPort"
                type="number"
                value={emailSettings.smtpPort}
                onChange={(e) =>
                  handleChange('smtpPort', parseInt(e.target.value))
                }
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="smtpUsername">SMTP Username</Label>
              <Input
                id="smtpUsername"
                placeholder="user@example.com"
                value={emailSettings.smtpUsername}
                onChange={(e) => handleChange('smtpUsername', e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="smtpPassword">SMTP Password</Label>
              <Input
                id="smtpPassword"
                type="password"
                value={emailSettings.smtpPassword}
                onChange={(e) => handleChange('smtpPassword', e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="emailFrom">From Email Address</Label>
              <Input
                id="emailFrom"
                placeholder="noreply@horizon-gateway.org"
                value={emailSettings.emailFrom}
                onChange={(e) => handleChange('emailFrom', e.target.value)}
              />
            </div>

            <div className="flex items-center space-x-2">
              <Switch
                id="enableTLSEmail"
                checked={emailSettings.enableTLS}
                onCheckedChange={(checked) =>
                  handleChange('enableTLS', checked)
                }
              />
              <Label htmlFor="enableTLSEmail">Enable TLS for SMTP</Label>
            </div>

            <div className="flex items-center space-x-2">
              <Switch
                id="enableAlerts"
                checked={emailSettings.enableAlerts}
                onCheckedChange={(checked) =>
                  handleChange('enableAlerts', checked)
                }
              />
              <Label htmlFor="enableAlerts">Enable Email Alerts</Label>
            </div>

            {emailSettings.enableAlerts && (
              <div className="space-y-2">
                <Label>Alert Recipients</Label>
                <div className="flex flex-wrap gap-2">
                  {emailSettings.alertRecipients.map((email) => (
                    <Badge key={email} variant="secondary" className="gap-1">
                      {email}
                      <X
                        className="h-3 w-3 cursor-pointer"
                        onClick={() => handleRemoveAlertEmail(email)}
                      />
                    </Badge>
                  ))}
                </div>
                <div className="flex gap-2 mt-1">
                  <Input
                    placeholder="Add email address"
                    value={newAlertEmail}
                    onChange={(e) => setNewAlertEmail(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        e.preventDefault();
                        handleAddAlertEmail();
                      }
                    }}
                  />
                  <Button type="button" size="sm" onClick={handleAddAlertEmail}>
                    Add
                  </Button>
                </div>
                <p className="text-sm text-slate-500">
                  Email addresses that will receive system alerts.
                </p>
              </div>
            )}
          </div>
        </div>

        {emailSettings.enableAlerts && (
          <div className="space-y-4 pt-4">
            <Separator />
            <h3 className="text-md font-semibold">Alert Types</h3>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-center space-x-2">
                <Switch id="alertErrors" checked={true} />
                <Label htmlFor="alertErrors">Error Alerts</Label>
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="alertCircuitBreaker" checked={true} />
                <Label htmlFor="alertCircuitBreaker">
                  Circuit Breaker Events
                </Label>
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="alertRateLimit" checked={true} />
                <Label htmlFor="alertRateLimit">Rate Limit Exceeded</Label>
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="alertConfigChanges" checked={true} />
                <Label htmlFor="alertConfigChanges">
                  Configuration Changes
                </Label>
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="alertServerStart" checked={true} />
                <Label htmlFor="alertServerStart">Server Start/Stop</Label>
              </div>

              <div className="flex items-center space-x-2">
                <Switch id="alertFailedAuth" checked={true} />
                <Label htmlFor="alertFailedAuth">Failed Authentication</Label>
              </div>
            </div>

            <div className="mt-4">
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={handleSendTestEmail}
              >
                Send Test Email
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
