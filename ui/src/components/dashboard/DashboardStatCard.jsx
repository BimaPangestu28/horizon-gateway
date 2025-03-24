import React from 'react';
import {
  Card,
  CardContent,
  CardHeader,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

export default function DashboardStatCard({
  title,
  value,
  icon,
  description,
  trend,
  trendValue,
  color,
}) {
  const colors = {
    blue: {
      bg: 'bg-blue-50',
      text: 'text-blue-600',
      icon: 'bg-blue-100',
      light: 'text-blue-500',
      trend: { up: 'text-green-600', down: 'text-red-600' },
    },
    green: {
      bg: 'bg-emerald-50',
      text: 'text-emerald-600',
      icon: 'bg-emerald-100',
      light: 'text-emerald-500',
      trend: { up: 'text-emerald-600', down: 'text-emerald-600' },
    },
    amber: {
      bg: 'bg-amber-50',
      text: 'text-amber-600',
      icon: 'bg-amber-100',
      light: 'text-amber-500',
      trend: { up: 'text-red-600', down: 'text-green-600' },
    },
    red: {
      bg: 'bg-rose-50',
      text: 'text-rose-600',
      icon: 'bg-rose-100',
      light: 'text-rose-500',
      trend: { up: 'text-red-600', down: 'text-green-600' },
    },
    purple: {
      bg: 'bg-violet-50',
      text: 'text-violet-600',
      icon: 'bg-violet-100',
      light: 'text-violet-500',
      trend: { up: 'text-green-600', down: 'text-red-600' },
    },
  }[color || 'blue'];

  return (
    <Card className="overflow-hidden">
      <CardHeader className="p-0">
        <div
          className={`h-1 w-full ${colors.text.replace('text', 'bg')}`}
        ></div>
      </CardHeader>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">{title}</p>
            <div className="flex items-baseline mt-1">
              <p className="text-2xl font-semibold">{value}</p>
              <p className={`ml-2 text-xs font-medium ${colors.light}`}>
                {description}
              </p>
            </div>
          </div>
          <div className={`p-2 rounded-full ${colors.icon}`}>{icon}</div>
        </div>
        <div className="flex items-center mt-4">
          <Badge
            variant="outline"
            className={`gap-1 ${
              trend === 'up' ? colors.trend.up : colors.trend.down
            }`}
          >
            {trend === 'up' ? <TrendUpIcon /> : <TrendDownIcon />}
            {trendValue}
          </Badge>
          <span className="ml-2 text-xs text-slate-500">vs last period</span>
        </div>
      </CardContent>
    </Card>
  );
}

function TrendUpIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 20 20"
      fill="currentColor"
      className="w-3 h-3"
    >
      <path
        fillRule="evenodd"
        d="M12.577 4.878a.75.75 0 01.919-.53l4.78 1.281a.75.75 0 01.531.919l-1.281 4.78a.75.75 0 01-1.449-.387l.81-3.022a19.407 19.407 0 00-5.594 5.203.75.75 0 01-1.139.093L7 10.06l-4.72 4.72a.75.75 0 01-1.06-1.061l5.25-5.25a.75.75 0 011.06 0l3.074 3.073a20.923 20.923 0 015.545-4.931l-3.042-.815a.75.75 0 01-.53-.919z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function TrendDownIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 20 20"
      fill="currentColor"
      className="w-3 h-3"
    >
      <path
        fillRule="evenodd"
        d="M1.22 5.222a.75.75 0 011.06 0L7 9.94l3.172-3.172a.75.75 0 011.06 0l3.1 3.1a20.98 20.98 0 015.9-4.964.75.75 0 01.694 1.325 19.494 19.494 0 00-5.506 4.625l3.037.9a.75.75 0 01-.23 1.471l-4.8-1.425a.75.75 0 01-.535-.767l.2-4.8a.75.75 0 011.467-.246l.7 3.36a21.47 21.47 0 015.393-4.761.75.75 0 01.818 1.257 20.97 20.97 0 00-5.168 4.561l2.935.87a.75.75 0 01-.231 1.47l-4.8-1.425a.75.75 0 01-.535-.767l.2-4.783.002-.045a.75.75 0 011.466-.246l.28 1.11a20.87 20.87 0 016.229-3.581.75.75 0 01.599 1.374 19.548 19.548 0 00-4.653 2.9l.292 1.76a.75.75 0 01-.713.883l-4.8.4a.75.75 0 01-.817-.744v-.049l.64-4.759a.75.75 0 011.5.201l-.205 1.533a22.704 22.704 0 016.258-3.087.75.75 0 11.548 1.391 20.384 20.384 0 00-7.786 4.998l1.943 1.95a.75.75 0 11-1.06 1.06l-5.25-5.25a.75.75 0 010-1.06l5.25-5.25a.75.75 0 011.06 0l3.074 3.073a20.923 20.923 0 015.545-4.93l-3.042-.815a.75.75 0 01-.53-.919z"
        clipRule="evenodd"
      />
    </svg>
  );
}