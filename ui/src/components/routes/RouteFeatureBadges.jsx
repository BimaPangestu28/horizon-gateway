import React from 'react';

export default function RouteFeatureBadges({ route }) {
  const hasAuth =
    route.Name?.includes('jwt') || route.Name?.includes('secured');
  const hasCache = route.Name?.includes('cached');
  const hasRateLimit = route.Name?.includes('rate-limited');
  const hasCircuitBreaker = route.Name?.includes('circuit-breaker');
  const hasTransform = route.Name?.includes('transformed');
  const hasIPFilter = route.Name?.includes('ip-filtered');
  const hasLoadBalancer = route.Name?.includes('load-balanced');

  return (
    <div className="flex flex-wrap gap-2">
      {hasAuth && (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
          Auth
        </span>
      )}
      {hasCache && (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
          Cache
        </span>
      )}
      {hasRateLimit && (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
          Rate Limit
        </span>
      )}
      {hasCircuitBreaker && (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
          Circuit Breaker
        </span>
      )}
      {hasTransform && (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
          Transform
        </span>
      )}
      {hasIPFilter && (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
          IP Filter
        </span>
      )}
      {hasLoadBalancer && (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
          Load Balanced
        </span>
      )}
    </div>
  );
}
