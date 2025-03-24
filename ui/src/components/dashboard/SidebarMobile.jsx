import React from 'react';
import { Zap, Menu } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import SidebarNav from './SidebarNav';

export default function SidebarMobile({ open, onOpenChange }) {
  return (
    <div className="md:hidden">
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="px-4 border-r border-slate-200 md:hidden"
          >
            <Menu className="w-6 h-6" />
            <span className="sr-only">Open sidebar</span>
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="p-0 w-72 border-r-0">
          <div className="flex flex-col h-full bg-white">
            <div className="flex items-center h-16 px-4 border-b border-slate-200 bg-slate-900 text-white">
              <div className="flex items-center gap-2">
                <div className="rounded-md bg-indigo-600 p-1">
                  <Zap className="h-6 w-6 text-white" />
                </div>
                <span className="text-xl font-bold">Horizon Gateway</span>
              </div>
            </div>
            <SidebarNav />
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}