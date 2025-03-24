import React from 'react';
import { 
  Menu, 
  Bell, 
  Search, 
  MoreHorizontal, 
  User, 
  Settings, 
  LogOut 
} from 'lucide-react';
import { Button } from '@/components/ui/button';
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
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export default function DashboardHeader({ setSidebarOpen, notifications, user }) {
  const unreadCount = notifications.filter((n) => !n.read).length;

  return (
    <header className="relative z-10 flex items-center justify-between h-16 px-4 bg-white border-b border-slate-200 shadow-sm">
      <div className="flex items-center flex-1 gap-4">
        <Button
          variant="ghost"
          size="icon"
          className="md:hidden"
          onClick={() => setSidebarOpen(true)}
        >
          <Menu className="h-6 w-6" />
        </Button>
        <div className="relative hidden md:block w-64">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-slate-400" />
          <Input placeholder="Search..." className="pl-8 bg-slate-50" />
        </div>
      </div>
      
      <div className="flex items-center gap-3">
        <NotificationsButton notifications={notifications} unreadCount={unreadCount} />
        <MoreOptionsButton />
        <MobileUserMenu user={user} />
      </div>
    </header>
  );
}

function NotificationsButton({ notifications, unreadCount }) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="outline" size="icon" className="relative">
            <Bell className="h-5 w-5" />
            {unreadCount > 0 && (
              <span className="absolute top-1 right-1 flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
              </span>
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          You have {unreadCount} unread notifications
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

function MoreOptionsButton() {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <MoreHorizontal className="h-5 w-5" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem>View Documentation</DropdownMenuItem>
        <DropdownMenuItem>Check for Updates</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Support</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function MobileUserMenu({ user }) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="flex items-center gap-2 md:hidden"
        >
          <Avatar className="h-8 w-8">
            <AvatarFallback className="bg-indigo-600 text-white">
              {user.name.split(' ').map(n => n[0]).join('')}
            </AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>My Account</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <User className="mr-2 h-4 w-4" />
          <span>Profile</span>
        </DropdownMenuItem>
        <DropdownMenuItem>
          <Settings className="mr-2 h-4 w-4" />
          <span>Settings</span>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <LogOut className="mr-2 h-4 w-4" />
          <span>Sign out</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}