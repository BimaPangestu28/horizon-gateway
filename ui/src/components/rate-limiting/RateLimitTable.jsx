import React from 'react';
import { Link } from 'react-router-dom';
import { Edit2, Trash2, Check, X, ExternalLink, Info } from 'react-feather';
import { TooltipProvider, Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

export default function RateLimitTable({ rateLimits = [], onEdit, onDelete }) {
  // Format the rate limit window for display
  const formatRateLimitWindow = (window) => {
    if (window.endsWith('s')) {
      return `${window} (seconds)`;
    } else if (window.endsWith('m')) {
      return `${window} (minutes)`;
    } else if (window.endsWith('h')) {
      return `${window} (hours)`;
    } else if (window.endsWith('d')) {
      return `${window} (days)`;
    }
    return window;
  };

  // Format the rate limit type for display
  const formatRateLimitType = (type) => {
    switch (type) {
      case 'sliding_window':
        return 'Sliding Window';
      case 'fixed_window':
        return 'Fixed Window';
      case 'token_bucket':
        return 'Token Bucket';
      default:
        return type;
    }
  };
  
  // Format the key display value
  const formatKeyDisplay = (key) => {
    switch(key) {
      case 'ip':
        return 'Client IP';
      case 'api_key':
        return 'API Key';
      case 'user_id':
        return 'User ID';
      default:
        if (key.startsWith('header:')) {
          return `Header (${key.substring(7)})`;
        }
        return key;
    }
  };

  return (
    <div className="flex flex-col">
      <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
          <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300">
              <thead className="bg-gray-50">
                <tr>
                  <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">
                    Route
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Type
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Limit
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Window
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Key By
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Status
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Current Usage
                  </th>
                  <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {rateLimits.length > 0 ? (
                  rateLimits.map((limit) => (
                    <tr key={limit.id} className="hover:bg-gray-50">
                      <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                        <Link 
                          to={`/rate-limiting/${limit.id}`}
                          className="text-indigo-600 hover:text-indigo-900 flex items-center"
                        >
                          {limit.global ? (
                            <span className="inline-flex items-center">
                              <span className="mr-2 h-2 w-2 rounded-full bg-purple-400"></span>
                              Global (All Routes)
                            </span>
                          ) : limit.route}
                          <ExternalLink className="ml-1 h-3.5 w-3.5" />
                        </Link>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger className="flex items-center">
                              {formatRateLimitType(limit.type)}
                              <Info className="ml-1 h-3.5 w-3.5 text-gray-400" />
                            </TooltipTrigger>
                            <TooltipContent>
                              {limit.type === 'sliding_window' && 
                                'Smoothly limiting over a continuous time window'}
                              {limit.type === 'fixed_window' && 
                                'Divides time into fixed intervals that reset periodically'}
                              {limit.type === 'token_bucket' && 
                                'Allows traffic bursts while maintaining average rate limits'}
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        <span className="font-medium">{limit.limit}</span> req/{limit.window}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        {formatRateLimitWindow(limit.window)}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        {formatKeyDisplay(limit.key)}
                        {limit.client_exceptions && limit.client_exceptions.length > 0 && (
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger className="ml-1 text-amber-500">
                                <span className="text-xs bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded-full">
                                  {limit.client_exceptions.length}
                                </span>
                              </TooltipTrigger>
                              <TooltipContent>
                                {limit.client_exceptions.length} client exception{limit.client_exceptions.length !== 1 ? 's' : ''} defined
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        )}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm">
                        {limit.enabled ? (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                            <Check className="mr-1 h-3 w-3" />
                            Enabled
                          </span>
                        ) : (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                            <X className="mr-1 h-3 w-3" />
                            Disabled
                          </span>
                        )}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm">
                        <div className="flex flex-col">
                          <span className="text-xs text-gray-500">
                            Active limits: {limit.current_state?.active_limits || 0}
                          </span>
                          <span className="text-xs text-gray-500">
                            Blocked: {limit.current_state?.blocked_clients || 0}
                          </span>
                          <span className="text-xs text-gray-500">
                            Requests (1h): {limit.current_state?.requests_last_hour || 0}
                          </span>
                        </div>
                      </td>
                      <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                        <div className="flex justify-end space-x-2">
                          <button 
                            className="text-indigo-600 hover:text-indigo-900"
                            onClick={() => onEdit(limit)}
                            aria-label={`Edit rate limit for ${limit.route}`}
                          >
                            <Edit2 className="h-4 w-4" />
                          </button>
                          <button 
                            className="text-red-600 hover:text-red-900"
                            onClick={() => onDelete(limit)}
                            aria-label={`Delete rate limit for ${limit.route}`}
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan="8" className="py-8 text-center text-sm text-gray-500">
                      No rate limits found
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}