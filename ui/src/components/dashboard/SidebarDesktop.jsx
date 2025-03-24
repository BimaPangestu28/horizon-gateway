import React from 'react';
import { Zap, User, Settings, LogOut } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import SidebarNav from './SidebarNav';

export default function SidebarDesktop({ user }) {
  return (
    <div className="hidden md:flex md:w-72 md:flex-col">
      <div className="flex flex-col flex-1 min-h-0 bg-slate-900 text-white">
        <div className="flex items-center h-16 px-6 border-b border-slate-700">
          <div className="flex items-center gap-2">
            <div className="rounded-md bg-indigo-600 p-1">
              <Zap className="h-6 w-6 text-white" />
            </div>
            <span className="text-xl font-bold">Horizon Gateway</span>
          </div>
        </div>
        
        <SidebarNav />
        
        <div className="p-4 mt-auto border-t border-slate-700">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                className="w-full justify-start text-slate-300 hover:text-white hover:bg-slate-800 px-2"
              >
                <div className="flex items-center">
                  <Avatar className="h-8 w-8 mr-2">
                    <AvatarImage
                      src="https://github.com/shadcn.png"
                      alt={user.name}
                    />
                    <AvatarFallback className="bg-indigo-600">
                      {user.name.split(' ').map(n => n[0]).join('')}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex flex-col items-start">
                    <span className="text-sm font-medium">{user.name}</span>
                    <span className="text-xs text-slate-400">
                      {user.email}
                    </span>
                  </div>
                </div>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
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
        </div>
      </div>
    </div>
  );
}