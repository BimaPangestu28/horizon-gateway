import * as React from 'react';

import { cn } from '@/lib/utils';

const Progress = React.forwardRef(
  ({ className, value, max = 100, indicatorClassName, ...props }, ref) => {
    const percentage = value != null ? Math.min(Math.max(value, 0), max) : null;

    return (
      <div
        ref={ref}
        className={cn(
          'relative h-4 w-full overflow-hidden rounded-full bg-slate-100',
          className,
        )}
        {...props}
      >
        <div
          className={cn(
            'h-full w-full flex-1 bg-indigo-600 transition-all',
            indicatorClassName,
          )}
          style={{
            width: percentage !== null ? `${(percentage / max) * 100}%` : '0%',
          }}
        />
      </div>
    );
  },
);
Progress.displayName = 'Progress';

export { Progress };
