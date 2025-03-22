import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { PlusCircle, Edit2, Trash2, ExternalLink, Search, Filter, RefreshCw } from 'react-feather';
import ApiService from '../services/ApiService';

export default function RoutesList() {
  const [routes, setRoutes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');

  useEffect(() => {
    fetchRoutes();
  }, []);

  const fetchRoutes = async () => {
    setLoading(true);
    try {
      // In a real implementation, this would call the API
      // For now, use sample data
      const sampleRoutes = [
        {
          id: 1,
          name: 'api-route',
          listen_path: '/api/*',
          upstream_url: 'http://api.example.com',
          methods: ['GET', 'POST'],
          strip_path: true,
          created_at: '2023-01-15T10:30:00Z',
          auth_enabled: true,
          cache_enabled: false,
          rate_limit_enabled: true,
          circuit_breaker_enabled: false
        },
        {
          id: 2, 
          name: 'web-route',
          listen_path: '/web/*',
          upstream_url: 'http://web.example.com',
          methods: ['*'],
          strip_path: false,
          created_at: '2023-01-10T15:45:00Z',
          auth_enabled: false,
          cache_enabled: true,
          rate_limit_enabled: false,
          circuit_breaker_enabled: true
        },
        {
          id: 3,
          name: 'admin-api',
          listen_path: '/admin/*',
          upstream_url: 'http://admin-service:8000',
          methods: ['GET', 'POST', 'PUT', 'DELETE'],
          strip_path: true,
          created_at: '2023-02-05T09:20:00Z',
          auth_enabled: true,
          cache_enabled: false,
          rate_limit_enabled: true,
          circuit_breaker_enabled: false
        },
        {
          id: 4,
          name: 'public-api',
          listen_path: '/public/*',
          upstream_url: 'http://public-service:8000',
          methods: ['GET'],
          strip_path: true,
          created_at: '2023-02-10T14:15:00Z',
          auth_enabled: false,
          cache_enabled: true,
          rate_limit_enabled: false,
          circuit_breaker_enabled: false
        },
        {
          id: 5,
          name: 'monitoring-only',
          listen_path: '/api/monitor/*',
          upstream_url: 'http://service:8000',
          methods: ['*'],
          strip_path: true,
          created_at: '2023-03-01T11:30:00Z',
          auth_enabled: false,
          cache_enabled: false,
          rate_limit_enabled: false,
          circuit_breaker_enabled: false
        }
      ];
      
      setRoutes(sampleRoutes);
      setError(null);
    } catch (err) {
      setError('Failed to fetch routes. Please try again.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const deleteRoute = async (id) => {
    if (!confirm('Are you sure you want to delete this route?')) {
      return;
    }
    
    try {
      // In a real implementation, this would call the API
      setRoutes(routes.filter(route => route.id !== id));
    } catch (err) {
      console.error(err);
      alert('Failed to delete route. Please try again.');
    }
  };

  const filteredRoutes = routes.filter(route => {
    const matchesSearch = route.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                          route.listen_path.toLowerCase().includes(searchTerm.toLowerCase());
    
    if (filterType === 'all') return matchesSearch;
    if (filterType === 'auth') return matchesSearch && route.auth_enabled;
    if (filterType === 'cache') return matchesSearch && route.cache_enabled;
    if (filterType === 'ratelimit') return matchesSearch && route.rate_limit_enabled;
    if (filterType === 'circuitbreaker') return matchesSearch && route.circuit_breaker_enabled;
    
    return matchesSearch;
  });

  return (
    <div>
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Routes</h1>
          <p className="mt-2 text-sm text-gray-700">
            A list of all API routes configured in your gateway.
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <Link
            to="/routes/new"
            className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
          >
            <PlusCircle className="w-4 h-4 mr-2" />
            Add Route
          </Link>
        </div>
      </div>
      
      <div className="flex flex-col sm:flex-row space-y-4 sm:space-y-0 sm:space-x-4 mt-6 mb-6">
        <div className="relative flex-grow">
          <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
            <Search className="w-5 h-5 text-gray-400" />
          </div>
          <input
            type="text"
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Search routes..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        
        <div className="flex space-x-2">
          <div className="relative inline-block text-left">
            <select
              className="block pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
              value={filterType}
              onChange={(e) => setFilterType(e.target.value)}
            >
              <option value="all">All Routes</option>
              <option value="auth">With Auth</option>
              <option value="cache">With Cache</option>
              <option value="ratelimit">With Rate Limit</option>
              <option value="circuitbreaker">With Circuit Breaker</option>
            </select>
          </div>
          
          <button
            type="button"
            className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            onClick={fetchRoutes}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </button>
        </div>
      </div>
      
      {loading ? (
        <div className="flex justify-center items-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
        </div>
      ) : error ? (
        <div className="bg-red-50 p-4 rounded-md">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">{error}</h3>
            </div>
          </div>
        </div>
      ) : (
        <div className="mt-8 flex flex-col">
          <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
            <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
              <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                <table className="min-w-full divide-y divide-gray-300">
                  <thead className="bg-gray-50">
                    <tr>
                      <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">
                        Name
                      </th>
                      <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                        Path
                      </th>
                      <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                        Target
                      </th>
                      <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                        Methods
                      </th>
                      <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                        Features
                      </th>
                      <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                        <span className="sr-only">Actions</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 bg-white">
                    {filteredRoutes.length === 0 ? (
                      <tr>
                        <td colSpan="6" className="py-10 text-center text-sm text-gray-500">
                          No routes found
                        </td>
                      </tr>
                    ) : (
                      filteredRoutes.map((route) => (
                        <tr key={route.id}>
                          <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                            {route.name}
                          </td>
                          <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                            {route.listen_path}
                          </td>
                          <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                            <span className="truncate max-w-xs block">{route.upstream_url}</span>
                          </td>
                          <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                            {route.methods.includes('*') ? 'ALL' : route.methods.join(', ')}
                          </td>
                          <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                            <div className="flex space-x-2">
                              {route.auth_enabled && (
                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                  Auth
                                </span>
                              )}
                              {route.cache_enabled && (
                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                  Cache
                                </span>
                              )}
                              {route.rate_limit_enabled && (
                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                                  Rate Limit
                                </span>
                              )}
                              {route.circuit_breaker_enabled && (
                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                                  Circuit Breaker
                                </span>
                              )}
                            </div>
                          </td>
                          <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                            <div className="flex justify-end space-x-2">
                              <Link
                                to={`/routes/${route.id}`}
                                className="text-indigo-600 hover:text-indigo-900 inline-flex items-center"
                              >
                                <ExternalLink className="w-4 h-4 mr-1" />
                                View
                              </Link>
                              <Link
                                to={`/routes/${route.id}/edit`}
                                className="text-indigo-600 hover:text-indigo-900 inline-flex items-center"
                              >
                                <Edit2 className="w-4 h-4 mr-1" />
                                Edit
                              </Link>
                              <button
                                onClick={() => deleteRoute(route.id)}
                                className="text-red-600 hover:text-red-900 inline-flex items-center"
                              >
                                <Trash2 className="w-4 h-4 mr-1" />
                                Delete
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}