import React, { useState, useEffect } from 'react';
import { 
  Plus, Trash2, Edit2, Clock, Save, RefreshCw, Search, 
  AlertTriangle, Check, X, Settings, Activity
} from 'react-feather';
import ApiService from '../services/ApiService';

export default function RateLimiting() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [rateLimits, setRateLimits] = useState([]);
  const [routes, setRoutes] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deleteLimit, setDeleteLimit] = useState(null);
  const [editLimit, setEditLimit] = useState(null);
  const [newRateLimit, setNewRateLimit] = useState({
    route: '',
    type: 'sliding_window',
    limit: 100,
    window: '60s',
    key: 'ip',
    response_code: 429,
    response_message: 'Rate limit exceeded',
    include_headers: true,
    enabled: true,
    global: false,
    client_exceptions: []
  });

  useEffect(() => {
    fetchRateLimitData();
    fetchRoutes();
  }, []);

  const fetchRateLimitData = async () => {
    setLoading(true);
    try {
      // In a real implementation, this would be a call to the API
      // For now, use sample data
      const sampleRateLimits = [
        {
          id: '1',
          route: 'api-route',
          type: 'sliding_window',
          limit: 300,
          window: '60s',
          key: 'ip',
          response_code: 429,
          response_message: 'Rate limit exceeded. Please try again later.',
          include_headers: true,
          enabled: true,
          global: false,
          client_exceptions: [
            { key: '192.168.1.10', limit: 1000 },
            { key: 'trusted-client', limit: 500 }
          ],
          current_state: {
            active_limits: 3,
            blocked_clients: 0,
            requests_last_hour: 543
          }
        },
        {
          id: '2',
          route: 'partner-api',
          type: 'token_bucket',
          limit: 100,
          window: '60s',
          key: 'api_key',
          response_code: 429,
          response_message: 'API rate limit exceeded',
          include_headers: true,
          enabled: true,
          global: false,
          client_exceptions: [],
          current_state: {
            active_limits: 5,
            blocked_clients: 1,
            requests_last_hour: 876
          }
        },
        {
          id: '3',
          route: 'public-api',
          type: 'fixed_window',
          limit: 50,
          window: '30s',
          key: 'ip',
          response_code: 429,
          response_message: 'Too many requests',
          include_headers: true,
          enabled: true,
          global: false,
          client_exceptions: [
            { key: 'trusted-partner', limit: 1000 }
          ],
          current_state: {
            active_limits: 12,
            blocked_clients: 3,
            requests_last_hour: 2134
          }
        },
        {
          id: '4',
          route: 'global',
          type: 'sliding_window',
          limit: 500,
          window: '60s',
          key: 'ip',
          response_code: 429,
          response_message: 'Global rate limit exceeded',
          include_headers: true,
          enabled: true,
          global: true,
          client_exceptions: [],
          current_state: {
            active_limits: 25,
            blocked_clients: 2,
            requests_last_hour: 3245
          }
        }
      ];
      
      setRateLimits(sampleRateLimits);
      setError(null);
    } catch (err) {
      console.error(err);
      setError('Failed to fetch rate limit data. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const fetchRoutes = async () => {
    try {
      // In a real implementation, this would call the API
      // For now, use sample data
      const sampleRoutes = [
        { id: '1', name: 'api-route', listen_path: '/api/*' },
        { id: '2', name: 'partner-api', listen_path: '/partner/*' },
        { id: '3', name: 'internal-route', listen_path: '/internal/*' },
        { id: '4', name: 'public-api', listen_path: '/public/*' }
      ];
      setRoutes(sampleRoutes);
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateRateLimit = async (e) => {
    e.preventDefault();
    
    try {
      // In a real implementation, this would call the API
      // For now, simulate API call
      const newLimit = {
        id: `${rateLimits.length + 1}`,
        ...newRateLimit,
        current_state: {
          active_limits: 0,
          blocked_clients: 0,
          requests_last_hour: 0
        }
      };
      
      setRateLimits([...rateLimits, newLimit]);
      setShowCreateModal(false);
      setNewRateLimit({
        route: '',
        type: 'sliding_window',
        limit: 100,
        window: '60s',
        key: 'ip',
        response_code: 429,
        response_message: 'Rate limit exceeded',
        include_headers: true,
        enabled: true,
        global: false,
        client_exceptions: []
      });
      
    } catch (err) {
      console.error(err);
      setError('Failed to create rate limit. Please try again.');
    }
  };

  const handleEditRateLimit = async (e) => {
    e.preventDefault();
    
    try {
      // In a real implementation, this would call the API
      // For now, simulate API call
      const updatedLimits = rateLimits.map(limit => 
        limit.id === editLimit.id ? editLimit : limit
      );
      
      setRateLimits(updatedLimits);
      setShowEditModal(false);
      setEditLimit(null);
      
    } catch (err) {
      console.error(err);
      setError('Failed to update rate limit. Please try again.');
    }
  };

  const handleDeleteConfirm = (limit) => {
    setDeleteLimit(limit);
    setShowDeleteModal(true);
  };

  const handleDelete = async () => {
    if (!deleteLimit) return;
    
    try {
      // In a real implementation, this would call the API
      setRateLimits(rateLimits.filter(limit => limit.id !== deleteLimit.id));
      setShowDeleteModal(false);
      setDeleteLimit(null);
    } catch (err) {
      console.error(err);
      setError('Failed to delete rate limit. Please try again.');
    }
  };

  const filterBySearch = (item) => {
    if (!searchTerm) return true;
    
    const searchLower = searchTerm.toLowerCase();
    
    return (
      item.route.toLowerCase().includes(searchLower) ||
      item.type.toLowerCase().includes(searchLower) ||
      item.key.toLowerCase().includes(searchLower)
    );
  };

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

  const addClientException = () => {
    const clientKey = document.getElementById('client-key').value;
    const clientLimit = parseInt(document.getElementById('client-limit').value);
    
    if (!clientKey || !clientLimit) return;
    
    if (showCreateModal) {
      setNewRateLimit({
        ...newRateLimit,
        client_exceptions: [
          ...newRateLimit.client_exceptions, 
          { key: clientKey, limit: clientLimit }
        ]
      });
    } else if (showEditModal) {
      setEditLimit({
        ...editLimit,
        client_exceptions: [
          ...editLimit.client_exceptions, 
          { key: clientKey, limit: clientLimit }
        ]
      });
    }
    
    document.getElementById('client-key').value = '';
    document.getElementById('client-limit').value = '';
  };

  const removeClientException = (index) => {
    if (showCreateModal) {
      const exceptions = [...newRateLimit.client_exceptions];
      exceptions.splice(index, 1);
      setNewRateLimit({...newRateLimit, client_exceptions: exceptions});
    } else if (showEditModal) {
      const exceptions = [...editLimit.client_exceptions];
      exceptions.splice(index, 1);
      setEditLimit({...editLimit, client_exceptions: exceptions});
    }
  };

  return (
    <div>
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Rate Limiting</h1>
          <p className="mt-2 text-sm text-gray-700">
            Configure rate limits to control the number of requests clients can make to your API.
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            type="button"
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
          >
            <Plus className="mr-2 h-4 w-4" />
            New Rate Limit
          </button>
        </div>
      </div>

      {/* Search and refresh bar */}
      <div className="mt-6 mb-6 flex">
        <div className="relative flex-grow">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search className="h-5 w-5 text-gray-400" />
          </div>
          <input
            type="text"
            className="focus:ring-indigo-500 focus:border-indigo-500 block w-full pl-10 sm:text-sm border-gray-300 rounded-md"
            placeholder="Search rate limits..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <button
          type="button"
          onClick={fetchRateLimitData}
          className="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </button>
      </div>

      {error && (
        <div className="mb-4 bg-red-50 p-4 rounded-md">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">{error}</h3>
            </div>
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center items-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
        </div>
      ) : (
        <div className="mt-6 flex flex-col">
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
                    {rateLimits.filter(filterBySearch).map((limit) => (
                      <tr key={limit.id}>
                        <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                          {limit.global ? (
                            <span className="inline-flex items-center">
                              <span className="mr-2 h-2 w-2 rounded-full bg-purple-400"></span>
                              Global (All Routes)
                            </span>
                          ) : limit.route}
                        </td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                          {formatRateLimitType(limit.type)}
                        </td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                          {limit.limit} req/{limit.window}
                        </td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                          {formatRateLimitWindow(limit.window)}
                        </td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                          {limit.key === 'ip' ? 'Client IP' : limit.key === 'api_key' ? 'API Key' : limit.key}
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
                              Active limits: {limit.current_state.active_limits}
                            </span>
                            <span className="text-xs text-gray-500">
                              Blocked: {limit.current_state.blocked_clients}
                            </span>
                            <span className="text-xs text-gray-500">
                              Requests (1h): {limit.current_state.requests_last_hour}
                            </span>
                          </div>
                        </td>
                        <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                          <div className="flex justify-end space-x-2">
                            <button 
                              className="text-indigo-600 hover:text-indigo-900"
                              onClick={() => {
                                setEditLimit(limit);
                                setShowEditModal(true);
                              }}
                            >
                              <Edit2 className="h-4 w-4" />
                            </button>
                            <button 
                              className="text-red-600 hover:text-red-900"
                              onClick={() => handleDeleteConfirm(limit)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}

                    {rateLimits.filter(filterBySearch).length === 0 && (
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
      )}

      {/* Create Rate Limit Modal */}
      {showCreateModal && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity" aria-hidden="true">
              <div className="absolute inset-0 bg-gray-500 opacity-75"></div>
            </div>

            <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

            <div className="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-2xl sm:w-full sm:p-6">
              <div>
                <div className="mt-3 text-center sm:mt-5">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Create New Rate Limit</h3>
                  <div className="mt-2">
                    <p className="text-sm text-gray-500">
                      Configure a new rate limit rule to control API traffic.
                    </p>
                  </div>
                </div>
              </div>
              
              <form className="mt-5 sm:mt-6" onSubmit={handleCreateRateLimit}>
                <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div className="sm:col-span-3">
                    <label htmlFor="route" className="block text-sm font-medium text-gray-700">
                      Route
                    </label>
                    <select
                      id="route"
                      name="route"
                      value={newRateLimit.route}
                      onChange={(e) => setNewRateLimit({...newRateLimit, route: e.target.value, global: e.target.value === 'global'})}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                      required
                    >
                      <option value="">Select a route</option>
                      <option value="global">Global (All Routes)</option>
                      {routes.map((route) => (
                        <option key={route.id} value={route.name}>
                          {route.name} ({route.listen_path})
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="type" className="block text-sm font-medium text-gray-700">
                      Rate Limit Type
                    </label>
                    <select
                      id="type"
                      name="type"
                      value={newRateLimit.type}
                      onChange={(e) => setNewRateLimit({...newRateLimit, type: e.target.value})}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                    >
                      <option value="sliding_window">Sliding Window</option>
                      <option value="fixed_window">Fixed Window</option>
                      <option value="token_bucket">Token Bucket</option>
                    </select>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="limit" className="block text-sm font-medium text-gray-700">
                      Request Limit
                    </label>
                    <input
                      type="number"
                      name="limit"
                      id="limit"
                      min="1"
                      value={newRateLimit.limit}
                      onChange={(e) => setNewRateLimit({...newRateLimit, limit: parseInt(e.target.value)})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                      required
                    />
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="window" className="block text-sm font-medium text-gray-700">
                      Time Window
                    </label>
                    <div className="mt-1 flex rounded-md shadow-sm">
                      <input
                        type="text"
                        name="window"
                        id="window"
                        value={newRateLimit.window}
                        onChange={(e) => setNewRateLimit({...newRateLimit, window: e.target.value})}
                        className="flex-1 block w-full focus:ring-indigo-500 focus:border-indigo-500 min-w-0 rounded-none rounded-l-md sm:text-sm border-gray-300"
                        required
                      />
                      <span className="inline-flex items-center px-3 rounded-r-md border border-l-0 border-gray-300 bg-gray-50 text-gray-500 sm:text-sm">
                        e.g. 60s, 5m, 1h, 1d
                      </span>
                    </div>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="key" className="block text-sm font-medium text-gray-700">
                      Rate Limit Key
                    </label>
                    <select
                      id="key"
                      name="key"
                      value={newRateLimit.key}
                      onChange={(e) => setNewRateLimit({...newRateLimit, key: e.target.value})}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                    >
                      <option value="ip">Client IP</option>
                      <option value="api_key">API Key</option>
                      <option value="user_id">User ID</option>
                      <option value="header:x-client-id">Header (x-client-id)</option>
                    </select>
                    <p className="mt-1 text-xs text-gray-500">
                      What to use as the unique identifier for rate limiting
                    </p>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="response_code" className="block text-sm font-medium text-gray-700">
                      Response Status Code
                    </label>
                    <input
                      type="number"
                      name="response_code"
                      id="response_code"
                      min="400"
                      max="599"
                      value={newRateLimit.response_code}
                      onChange={(e) => setNewRateLimit({...newRateLimit, response_code: parseInt(e.target.value)})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                    />
                  </div>

                  <div className="sm:col-span-6">
                    <label htmlFor="response_message" className="block text-sm font-medium text-gray-700">
                      Response Message
                    </label>
                    <input
                      type="text"
                      name="response_message"
                      id="response_message"
                      value={newRateLimit.response_message}
                      onChange={(e) => setNewRateLimit({...newRateLimit, response_message: e.target.value})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                    />
                  </div>

                  <div className="sm:col-span-6">
                    <div className="flex items-start">
                      <div className="flex items-center h-5">
                        <input
                          id="include_headers"
                          name="include_headers"
                          type="checkbox"
                          checked={newRateLimit.include_headers}
                          onChange={(e) => setNewRateLimit({...newRateLimit, include_headers: e.target.checked})}
                          className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                        />
                      </div>
                      <div className="ml-3 text-sm">
                        <label htmlFor="include_headers" className="font-medium text-gray-700">
                          Include Rate Limit Headers
                        </label>
                        <p className="text-gray-500">
                          Add X-RateLimit-* headers to responses
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="sm:col-span-6">
                    <div className="flex items-start">
                      <div className="flex items-center h-5">
                        <input
                          id="enabled"
                          name="enabled"
                          type="checkbox"
                          checked={newRateLimit.enabled}
                          onChange={(e) => setNewRateLimit({...newRateLimit, enabled: e.target.checked})}
                          className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                        />
                      </div>
                      <div className="ml-3 text-sm">
                        <label htmlFor="enabled" className="font-medium text-gray-700">
                          Enabled
                        </label>
                        <p className="text-gray-500">
                          Toggle rate limiting on or off
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="sm:col-span-6 border-t pt-4">
                    <h4 className="text-sm font-medium text-gray-900">Client Exceptions</h4>
                    <p className="mt-1 text-sm text-gray-500">
                      Specific clients that have different rate limits
                    </p>
                    
                    <div className="mt-2">
                      <div className="flex space-x-2">
                        <div className="flex-1">
                          <input
                            type="text"
                            id="client-key"
                            placeholder="Client key (IP, API Key, etc.)"
                            className="block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                          />
                        </div>
                        <button
                          type="button"
                          onClick={addClientException}
                          className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                        >
                          Add
                        </button>
                      </div>
                      
                      {newRateLimit.client_exceptions.length > 0 && (
                        <div className="mt-3">
                          <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                              <tr>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Client Key
                                </th>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Limit
                                </th>
                                <th scope="col" className="relative px-6 py-3">
                                  <span className="sr-only">Actions</span>
                                </th>
                              </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                              {newRateLimit.client_exceptions.map((exception, index) => (
                                <tr key={index}>
                                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    {exception.key}
                                  </td>
                                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    {exception.limit}
                                  </td>
                                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <button
                                      type="button"
                                      onClick={() => removeClientException(index)}
                                      className="text-red-600 hover:text-red-900"
                                    >
                                      <Trash2 className="h-4 w-4" />
                                    </button>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                
                <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                  <button
                    type="submit"
                    className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                  >
                    Create Rate Limit
                  </button>
                  <button
                    type="button"
                    className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
                    onClick={() => setShowCreateModal(false)}
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Edit Rate Limit Modal */}
      {showEditModal && editLimit && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity" aria-hidden="true">
              <div className="absolute inset-0 bg-gray-500 opacity-75"></div>
            </div>

            <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

            <div className="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-2xl sm:w-full sm:p-6">
              <div>
                <div className="mt-3 text-center sm:mt-5">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Edit Rate Limit</h3>
                  <div className="mt-2">
                    <p className="text-sm text-gray-500">
                      Update rate limit configuration for this route.
                    </p>
                  </div>
                </div>
              </div>
              
              <form className="mt-5 sm:mt-6" onSubmit={handleEditRateLimit}>
                <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div className="sm:col-span-3">
                    <label className="block text-sm font-medium text-gray-700">
                      Route
                    </label>
                    <div className="mt-1 block w-full py-2 text-base text-gray-700">
                      {editLimit.global ? 'Global (All Routes)' : editLimit.route}
                    </div>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="edit-type" className="block text-sm font-medium text-gray-700">
                      Rate Limit Type
                    </label>
                    <select
                      id="edit-type"
                      name="edit-type"
                      value={editLimit.type}
                      onChange={(e) => setEditLimit({...editLimit, type: e.target.value})}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                    >
                      <option value="sliding_window">Sliding Window</option>
                      <option value="fixed_window">Fixed Window</option>
                      <option value="token_bucket">Token Bucket</option>
                    </select>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="edit-limit" className="block text-sm font-medium text-gray-700">
                      Request Limit
                    </label>
                    <input
                      type="number"
                      name="edit-limit"
                      id="edit-limit"
                      min="1"
                      value={editLimit.limit}
                      onChange={(e) => setEditLimit({...editLimit, limit: parseInt(e.target.value)})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                      required
                    />
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="edit-window" className="block text-sm font-medium text-gray-700">
                      Time Window
                    </label>
                    <div className="mt-1 flex rounded-md shadow-sm">
                      <input
                        type="text"
                        name="edit-window"
                        id="edit-window"
                        value={editLimit.window}
                        onChange={(e) => setEditLimit({...editLimit, window: e.target.value})}
                        className="flex-1 block w-full focus:ring-indigo-500 focus:border-indigo-500 min-w-0 rounded-none rounded-l-md sm:text-sm border-gray-300"
                        required
                      />
                      <span className="inline-flex items-center px-3 rounded-r-md border border-l-0 border-gray-300 bg-gray-50 text-gray-500 sm:text-sm">
                        e.g. 60s, 5m, 1h, 1d
                      </span>
                    </div>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="edit-key" className="block text-sm font-medium text-gray-700">
                      Rate Limit Key
                    </label>
                    <select
                      id="edit-key"
                      name="edit-key"
                      value={editLimit.key}
                      onChange={(e) => setEditLimit({...editLimit, key: e.target.value})}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                    >
                      <option value="ip">Client IP</option>
                      <option value="api_key">API Key</option>
                      <option value="user_id">User ID</option>
                      <option value="header:x-client-id">Header (x-client-id)</option>
                    </select>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="edit-response_code" className="block text-sm font-medium text-gray-700">
                      Response Status Code
                    </label>
                    <input
                      type="number"
                      name="edit-response_code"
                      id="edit-response_code"
                      min="400"
                      max="599"
                      value={editLimit.response_code}
                      onChange={(e) => setEditLimit({...editLimit, response_code: parseInt(e.target.value)})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                    />
                  </div>

                  <div className="sm:col-span-6">
                    <label htmlFor="edit-response_message" className="block text-sm font-medium text-gray-700">
                      Response Message
                    </label>
                    <input
                      type="text"
                      name="edit-response_message"
                      id="edit-response_message"
                      value={editLimit.response_message}
                      onChange={(e) => setEditLimit({...editLimit, response_message: e.target.value})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                    />
                  </div>

                  <div className="sm:col-span-6">
                    <div className="flex items-start">
                      <div className="flex items-center h-5">
                        <input
                          id="edit-include_headers"
                          name="edit-include_headers"
                          type="checkbox"
                          checked={editLimit.include_headers}
                          onChange={(e) => setEditLimit({...editLimit, include_headers: e.target.checked})}
                          className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                        />
                      </div>
                      <div className="ml-3 text-sm">
                        <label htmlFor="edit-include_headers" className="font-medium text-gray-700">
                          Include Rate Limit Headers
                        </label>
                        <p className="text-gray-500">
                          Add X-RateLimit-* headers to responses
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="sm:col-span-6">
                    <div className="flex items-start">
                      <div className="flex items-center h-5">
                        <input
                          id="edit-enabled"
                          name="edit-enabled"
                          type="checkbox"
                          checked={editLimit.enabled}
                          onChange={(e) => setEditLimit({...editLimit, enabled: e.target.checked})}
                          className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                        />
                      </div>
                      <div className="ml-3 text-sm">
                        <label htmlFor="edit-enabled" className="font-medium text-gray-700">
                          Enabled
                        </label>
                        <p className="text-gray-500">
                          Toggle rate limiting on or off
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="sm:col-span-6 border-t pt-4">
                    <h4 className="text-sm font-medium text-gray-900">Client Exceptions</h4>
                    <p className="mt-1 text-sm text-gray-500">
                      Specific clients that have different rate limits
                    </p>
                    
                    <div className="mt-2">
                      <div className="flex space-x-2">
                        <div className="flex-1">
                          <input
                            type="text"
                            id="client-key"
                            placeholder="Client key (IP, API Key, etc.)"
                            className="block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                          />
                        </div>
                        <div className="w-32">
                          <input
                            type="number"
                            id="client-limit"
                            placeholder="Limit"
                            min="1"
                            className="block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                          />
                        </div>
                        <button
                          type="button"
                          onClick={addClientException}
                          className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                        >
                          Add
                        </button>
                      </div>
                      
                      {editLimit.client_exceptions.length > 0 && (
                        <div className="mt-3">
                          <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                              <tr>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Client Key
                                </th>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Limit
                                </th>
                                <th scope="col" className="relative px-6 py-3">
                                  <span className="sr-only">Actions</span>
                                </th>
                              </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                              {editLimit.client_exceptions.map((exception, index) => (
                                <tr key={index}>
                                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    {exception.key}
                                  </td>
                                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    {exception.limit}
                                  </td>
                                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <button
                                      type="button"
                                      onClick={() => removeClientException(index)}
                                      className="text-red-600 hover:text-red-900"
                                    >
                                      <Trash2 className="h-4 w-4" />
                                    </button>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                
                <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                  <button
                    type="submit"
                    className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                  >
                    <Save className="mr-2 h-4 w-4" />
                    Save Changes
                  </button>
                  <button
                    type="button"
                    className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
                    onClick={() => {
                      setShowEditModal(false);
                      setEditLimit(null);
                    }}
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteModal && deleteLimit && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity" aria-hidden="true">
              <div className="absolute inset-0 bg-gray-500 opacity-75"></div>
            </div>

            <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

            <div className="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
              <div>
                <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
                  <AlertTriangle className="h-6 w-6 text-red-600" aria-hidden="true" />
                </div>
                <div className="mt-3 text-center sm:mt-5">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">
                    Delete Rate Limit
                  </h3>
                  <div className="mt-2">
                    <p className="text-sm text-gray-500">
                      Are you sure you want to delete the rate limit for
                      {deleteLimit.global ? " all routes" : ` route "${deleteLimit.route}"`}?
                      This action cannot be undone.
                    </p>
                  </div>
                </div>
              </div>
              <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:col-start-2 sm:text-sm"
                  onClick={handleDelete}
                >
                  Delete
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
                  onClick={() => setShowDeleteModal(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}