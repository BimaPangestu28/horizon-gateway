import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Edit2, Trash2, ArrowLeft, RefreshCw, AlertTriangle, Play, Pause } from 'react-feather';
import { useToast } from '@/components/ui/use-toast';
import RateLimitService from '@/services/RateLimitService';
import RateLimitAnalytics from '@/components/rate-limiting/RateLimitAnalytics';
import RateLimitTypes from '@/components/rate-limiting/RateLimitTypes';
import EditRateLimitModal from '@/components/rate-limiting/EditRateLimitModal';
import DeleteRateLimitModal from '@/components/rate-limiting/DeleteRateLimitModal';
import RoutesService from '@/services/RoutesService';
import LoadingSpinner from '@/components/ui/LoadingSpinner';

export default function RateLimitDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [rateLimit, setRateLimit] = useState(null);
  const [routes, setRoutes] = useState([]);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showTypeInfo, setShowTypeInfo] = useState(false);
  
  useEffect(() => {
    fetchData();
  }, [id]);
  
  const fetchData = async () => {
    setLoading(true);
    try {
      const [rateLimitData, routesData] = await Promise.all([
        RateLimitService.getRateLimit(id),
        RoutesService.getRoutes()
      ]);
      
      setRateLimit(rateLimitData);
      setRoutes(routesData);
      setError(null);
    } catch (err) {
      console.error('Error fetching rate limit:', err);
      setError('Failed to load rate limit data');
      toast({
        title: "Error loading rate limit",
        description: err.message || "Could not retrieve rate limit details",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };
  
  const handleUpdateRateLimit = async (updatedLimit) => {
    try {
      await RateLimitService.updateRateLimit(id, updatedLimit);
      setRateLimit(updatedLimit);
      setShowEditModal(false);
      toast({
        title: "Rate limit updated",
        description: "Rate limit settings have been updated successfully",
      });
    } catch (err) {
      console.error('Error updating rate limit:', err);
      toast({
        title: "Failed to update rate limit",
        description: err.message || "An error occurred",
        variant: "destructive",
      });
    }
  };
  
  const handleDeleteRateLimit = async () => {
    try {
      await RateLimitService.deleteRateLimit(id);
      toast({
        title: "Rate limit deleted",
        description: "The rate limit rule has been removed",
      });
      navigate('/rate-limiting');
    } catch (err) {
      console.error('Error deleting rate limit:', err);
      toast({
        title: "Failed to delete rate limit",
        description: err.message || "An error occurred",
        variant: "destructive",
      });
    }
  };
  
  const toggleRateLimit = async () => {
    try {
      const updatedLimit = { ...rateLimit, enabled: !rateLimit.enabled };
      await RateLimitService.updateRateLimit(id, updatedLimit);
      setRateLimit(updatedLimit);
      toast({
        title: updatedLimit.enabled ? "Rate limit enabled" : "Rate limit disabled",
        description: updatedLimit.enabled 
          ? "Rate limit is now active and will be applied to requests" 
          : "Rate limit is now inactive and will not affect requests",
      });
    } catch (err) {
      console.error('Error toggling rate limit:', err);
      toast({
        title: "Failed to update rate limit",
        description: err.message || "An error occurred",
        variant: "destructive",
      });
    }
  };
  
  // Format functions from the table component
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
  
  if (loading) {
    return <LoadingSpinner />;
  }
  
  if (error || !rateLimit) {
    return (
      <div className="bg-red-50 p-4 rounded-md">
        <div className="flex">
          <div className="flex-shrink-0">
            <AlertTriangle className="h-5 w-5 text-red-400" aria-hidden="true" />
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">
              {error || "Rate limit not found"}
            </h3>
            <div className="mt-4">
              <button
                type="button"
                className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                onClick={() => navigate('/rate-limiting')}
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Rate Limits
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }
  
  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div className="flex items-center">
          <button
            type="button"
            className="mr-4 inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            onClick={() => navigate('/rate-limiting')}
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </button>
          <h1 className="text-2xl font-semibold text-gray-900">
            {rateLimit.global ? 'Global Rate Limit' : `Rate Limit: ${rateLimit.route}`}
          </h1>
          {rateLimit.enabled ? (
            <span className="ml-3 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
              Enabled
            </span>
          ) : (
            <span className="ml-3 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
              Disabled
            </span>
          )}
        </div>
        
        <div className="flex space-x-2">
          <button
            type="button"
            onClick={toggleRateLimit}
            className={`inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 ${
              rateLimit.enabled 
                ? 'text-yellow-700 bg-yellow-100 hover:bg-yellow-200 focus:ring-yellow-500' 
                : 'text-green-700 bg-green-100 hover:bg-green-200 focus:ring-green-500'
            }`}
          >
            {rateLimit.enabled ? (
              <>
                <Pause className="mr-2 h-4 w-4" />
                Disable
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                Enable
              </>
            )}
          </button>
          
          <button
            type="button"
            onClick={() => setShowEditModal(true)}
            className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <Edit2 className="mr-2 h-4 w-4" />
            Edit
          </button>
          
          <button
            type="button"
            onClick={() => setShowDeleteModal(true)}
            className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </button>
          
          <button
            type="button"
            onClick={fetchData}
            className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </button>
        </div>
      </div>
      
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Rate Limit Details
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                Configuration and current status
              </p>
            </div>
            
            <div className="border-t border-gray-200">
              <dl>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Route</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {rateLimit.global ? (
                      <span className="inline-flex items-center">
                        <span className="mr-2 h-2 w-2 rounded-full bg-purple-400"></span>
                        Global (All Routes)
                      </span>
                    ) : rateLimit.route}
                  </dd>
                </div>
                
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Type</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    <div className="flex items-center">
                      {formatRateLimitType(rateLimit.type)}
                      <button
                        type="button"
                        className="ml-2 inline-flex items-center px-2 py-1 border border-transparent text-xs font-medium rounded-md text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                        onClick={() => setShowTypeInfo(!showTypeInfo)}
                      >
                        Info
                      </button>
                    </div>
                  </dd>
                </div>
                
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Limit</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {rateLimit.limit} requests per {formatRateLimitWindow(rateLimit.window)}
                  </dd>
                </div>
                
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Key By</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {rateLimit.key === 'ip' ? 'Client IP Address' : 
                     rateLimit.key === 'api_key' ? 'API Key' : 
                     rateLimit.key === 'user_id' ? 'User ID' : 
                     rateLimit.key}
                  </dd>
                </div>
                
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Response</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {rateLimit.response_code} - {rateLimit.response_message}
                  </dd>
                </div>
                
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Headers</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {rateLimit.include_headers ? 'Included' : 'Not included'}
                    {rateLimit.include_headers && (
                      <div className="mt-2 text-xs text-gray-500 bg-gray-50 p-2 rounded">
                        <div><code className="font-mono">X-RateLimit-Limit</code>: {rateLimit.limit}</div>
                        <div><code className="font-mono">X-RateLimit-Remaining</code>: [Dynamic]</div>
                        <div><code className="font-mono">X-RateLimit-Reset</code>: [Dynamic]</div>
                      </div>
                    )}
                  </dd>
                </div>
                
                {rateLimit.client_exceptions.length > 0 && (
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-medium text-gray-500">Client Exceptions</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                      <div className="border border-gray-200 rounded-md overflow-hidden">
                        <table className="min-w-full divide-y divide-gray-200">
                          <thead className="bg-gray-50">
                            <tr>
                              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Client Key
                              </th>
                              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Custom Limit
                              </th>
                            </tr>
                          </thead>
                          <tbody className="bg-white divide-y divide-gray-200">
                            {rateLimit.client_exceptions.map((exception, index) => (
                              <tr key={index}>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                  {exception.key}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                  {exception.limit}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </dd>
                  </div>
                )}
                
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Current Status</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    <div className="flex items-center">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        rateLimit.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                      }`}>
                        {rateLimit.enabled ? 'Active' : 'Inactive'}
                      </span>
                      {rateLimit.enabled && (
                        <span className="ml-4">
                          Active limits: {rateLimit.current_state?.active_limits || 0} | 
                          Blocked clients: {rateLimit.current_state?.blocked_clients || 0}
                        </span>
                      )}
                    </div>
                  </dd>
                </div>
              </dl>
            </div>
          </div>
          
          <RateLimitAnalytics rateLimit={rateLimit} />
        </div>
        
        <div className="lg:col-span-1">
          {showTypeInfo && <RateLimitTypes />}
          
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mt-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Best Practices
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                Tips for effective rate limiting
              </p>
            </div>
            
            <div className="border-t border-gray-200 px-4 py-5 sm:px-6">
              <ul className="list-disc pl-5 space-y-2 text-sm text-gray-700">
                <li>Set reasonable limits that protect your API without blocking legitimate traffic</li>
                <li>Consider different limits for authenticated vs. unauthenticated users</li>
                <li>Add rate limit headers so clients can properly adapt to your limits</li>
                <li>Use client exceptions for trusted partners or important services</li>
                <li>Monitor your rate limits regularly to ensure they're working as intended</li>
                <li>Choose time windows appropriate for your traffic patterns (shorter for high-volume endpoints)</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
      
      {showEditModal && (
        <EditRateLimitModal
          rateLimit={rateLimit}
          routes={routes}
          onSubmit={handleUpdateRateLimit}
          onCancel={() => setShowEditModal(false)}
        />
      )}
      
      {showDeleteModal && (
        <DeleteRateLimitModal
          rateLimit={rateLimit}
          onDelete={handleDeleteRateLimit}
          onCancel={() => setShowDeleteModal(false)}
        />
      )}
    </div>
  );
}