import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
  Home, 
  Layers, 
  PieChart, 
  Shield, 
  Clock, 
  AlertTriangle, 
  Database, 
  Settings 
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';

export default function SidebarNav() {
  return (
    <div className="flex flex-col flex-1 py-4 overflow-y-auto">
      <nav className="flex-1 px-2 space-y-1">
        <div className="px-3 pb-2 text-xs font-semibold text-slate-400 uppercase tracking-wider">
          Main
        </div>
        <NavItem to="/" icon={<Home />} label="Overview" />
        <NavItem to="/routes" icon={<Layers />} label="Routes" />
        <NavItem to="/analytics" icon={<PieChart />} label="Analytics" />

        <Separator className="my-4 bg-slate-700" />

        <div className="px-3 pb-2 text-xs font-semibold text-slate-400 uppercase tracking-wider">
          Security
        </div>
        <NavItem
          to="/authentication"
          icon={<Shield />}
          label="Authentication"
        />
        <NavItem to="/rate-limiting" icon={<Clock />} label="Rate Limiting" />

        <Separator className="my-4 bg-slate-700" />

        <div className="px-3 pb-2 text-xs font-semibold text-slate-400 uppercase tracking-wider">
          Reliability
        </div>
        <NavItem
          to="/circuit-breakers"
          icon={<AlertTriangle />}
          label="Circuit Breakers"
        />
        <NavItem to="/caching" icon={<Database />} label="Caching" />
        <NavItem to="/settings" icon={<Settings />} label="Settings" />
      </nav>
    </div>
  );
}

function NavItem({ to, icon, label }) {
  const location = useLocation();
  const isActive = location.pathname === to;

  return (
    <Link to={to}>
      <Button
        variant="ghost"
        className={`w-full justify-start gap-3 mb-1 ${
          isActive
            ? 'bg-slate-800 text-white hover:bg-slate-700'
            : 'text-slate-300 hover:bg-slate-800 hover:text-white'
        }`}
      >
        <span>{icon}</span>
        <span>{label}</span>
        {isActive && (
          <span className="absolute left-0 rounded-r-md inset-y-1 w-1 bg-indigo-500"></span>
        )}
      </Button>
    </Link>
  );
}