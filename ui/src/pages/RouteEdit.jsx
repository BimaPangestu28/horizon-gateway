import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Save, ArrowLeft, Trash2 } from 'react-feather';
import ApiService from '../services/ApiService';

export default function RouteEdit() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [route, setRoute] = useState({
    name: '',
    listen_path: '',
    upstream_url: '',
    methods: [],
    strip_path: false,
    headers: {},
    query_params: {},
    host: '',
    priority: 0
  });

  useEffect(() => {
    fetchRoute();
  }, [id]);

  const fetchRoute = async () => {
    setLoading(true);
    try {
      // In a real implementation, this would call the API
      // For now, use sample data based on the route ID
      
      const sampleRoute = {
        id: parseInt(id),
        name: `route-${id}`,
        listen_path: `/api/${id}/*`,
        upstream_url: `http://service-${id}.example.com`,
        methods: ['GET', 'POST', 'PUT'],
        strip_path: true,
        headers: {
          'Content-Type': 'application/json'
        },
        query_params: {
          'version': 'v1'
        },
        host: '',
        priority: 10,
        auth_enabled: id % 2 === 0,
        cache_enabled: id % 3 === 0,
        rate_limit_enabled: id % 2 === 0,
        circuit_breaker_enabled: id % 3 === 0
      };
      
      setRoute(sampleRoute);
      setError(null);
    } catch (err) {
      setError('Failed to fetch route details. Please try again.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    
    setRoute(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
  };

  const handleMethodToggle = (method) => {
    setRoute(prev => {
      const methods = [...prev.methods];
      const index = methods.indexOf(method);
      
      if (index === -1) {
        methods.push(method);
      } else {
        methods.splice(index, 1);
      }
      
      return {
        ...prev,
        methods
      };
    });
  };

  const handleSelectAllMethods = () => {
    setRoute(prev => ({
      ...prev,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD']
    }));
  };

  const handleHeaderChange = (key, value, isNewKey = false) => {
    setRoute(prev => {
      const headers = { ...prev.headers };
      
      if (isNewKey && key) {
        headers[key] = value;
      } else if (key in headers) {
        headers[key] = value;
      }
      
      return {
        ...prev,
        headers
      };
    });
  };

  const handleRemoveHeader = (key) => {
    setRoute(prev => {
      const headers = { ...prev.headers };
      delete headers[key];
      
      return {
        ...prev,
        headers
      };
    });
  };

  const handleQueryParamChange = (key, value, isNewKey = false) => {
    setRoute(prev => {
      const queryParams = { ...prev.query_params };
      
      if (isNewKey && key) {
        queryParams[key] = value;
      } else if (key in queryParams) {
        queryParams[key] = value;
      }
      
      return {
        ...prev,
        query_params: queryParams
      };
    });
  };

  const handleRemoveQueryParam = (key) => {
    setRoute(prev => {
      const queryParams = { ...prev.query_params };
      delete queryParams[key];
      
      return {
        ...prev,
        query_params: queryParams
      };
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    
    try {
      // In a real implementation, this would call the API
      // For now, simulate a save operation
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Navigate back to route details
      navigate(`/routes/${id}`);
    } catch (err) {
      setError('Failed to save route. Please try again.');
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this route?')) {
      return;
    }
    
    try {
      // In a real implementation, this would call the API
      // For now, simulate a delete operation
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Navigate back to routes list
      navigate('/routes');
    } catch (err) {
      console.error(err);
      alert('Failed to delete route. Please try again.');
    }
  };

  return (
    <div>
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Edit Route</h1>
          <p className="mt-2 text-sm text-gray-700">
            Update the configuration for this API route.
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none space-x-3">
          <Link
            to={`/routes/${id}`}
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Link>
          <button
            onClick={handleDelete}
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent rounded-md shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
          >
            <Trash2 className="w-4 h-4 mr-2" />
            Delete
          </button>
        </div>
      </div>
      
      {loading ? (
        <div className="flex justify-center items-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
        </div>
      ) : error ? (
        <div className="mt-6 bg-red-50 p-4 rounded-md">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">{error}</h3>
            </div>
          </div>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="mt-6 space-y-8">
          {/* Basic Information */}
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Basic Information</h2>
            <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
              <div className="sm:col-span-3">
                <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                  Name
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="name"
                    id="name"
                    value={route.name}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    required
                  />
                </div>
              </div>

              <div className="sm:col-span-3">
                <label htmlFor="listen_path" className="block text-sm font-medium text-gray-700">
                  Listen Path
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="listen_path"
                    id="listen_path"
                    value={route.listen_path}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    required
                  />
                  <p className="mt-1 text-xs text-gray-500">Example: /api/* or /service/users/*</p>
                </div>
              </div>

              <div className="sm:col-span-3">
                <label htmlFor="upstream_url" className="block text-sm font-medium text-gray-700">
                  Upstream URL
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="upstream_url"
                    id="upstream_url"
                    value={route.upstream_url}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    required
                  />
                  <p className="mt-1 text-xs text-gray-500">Example: http://service:8000 or http://api.example.com</p>
                </div>
              </div>

              <div className="sm:col-span-3">
                <label htmlFor="host" className="block text-sm font-medium text-gray-700">
                  Host (Optional)
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="host"
                    id="host"
                    value={route.host}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                  />
                  <p className="mt-1 text-xs text-gray-500">Example: api.example.com</p>
                </div>
              </div>

              <div className="sm:col-span-3">
                <label htmlFor="priority" className="block text-sm font-medium text-gray-700">
                  Priority
                </label>
                <div className="mt-1">
                  <input
                    type="number"
                    name="priority"
                    id="priority"
                    value={route.priority}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                  />
                  <p className="mt-1 text-xs text-gray-500">Higher values take precedence when multiple routes match</p>
                </div>
              </div>

              <div className="sm:col-span-3">
                <div className="flex items-center">
                  <input
                    id="strip_path"
                    name="strip_path"
                    type="checkbox"
                    checked={route.strip_path}
                    onChange={handleChange}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                  <label htmlFor="strip_path" className="ml-2 block text-sm text-gray-900">
                    Strip Path
                  </label>
                </div>
                <p className="mt-1 text-xs text-gray-500">Remove the matched part of the path before forwarding</p>
              </div>
            </div>
          </div>

          {/* HTTP Methods */}
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">HTTP Methods</h2>
            <div className="space-y-4">
              <div className="flex flex-wrap gap-4">
                {['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD'].map((method) => (
                  <div key={method} className="flex items-center">
                    <input
                      id={`method-${method}`}
                      type="checkbox"
                      checked={route.methods.includes(method)}
                      onChange={() => handleMethodToggle(method)}
                      className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                    />
                    <label htmlFor={`method-${method}`} className="ml-2 block text-sm text-gray-900">
                      {method}
                    </label>
                  </div>
                ))}
              </div>
              <div>
                <button
                  type="button"
                  onClick={handleSelectAllMethods}
                  className="text-sm text-indigo-600 hover:text-indigo-900"
                >
                  Select All
                </button>
              </div>
            </div>
          </div>

          {/* Request Headers */}
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Required Request Headers</h2>
            <div className="space-y-4">
              {Object.entries(route.headers).map(([key, value]) => (
                <div key={key} className="flex items-center space-x-2">
                  <input
                    type="text"
                    value={key}
                    disabled
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                  />
                  <input
                    type="text"
                    value={value}
                    onChange={(e) => handleHeaderChange(key, e.target.value)}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                  />
                  <button
                    type="button"
                    onClick={() => handleRemoveHeader(key)}
                    className="inline-flex items-center p-1 border border-transparent rounded-full shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
              
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  placeholder="Header Name"
                  id="new-header-key"
                  className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                />
                <input
                  type="text"
                  placeholder="Header Value"
                  id="new-header-value"
                  className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                />
                <button
                  type="button"
                  onClick={() => {
                    const key = document.getElementById('new-header-key').value;
                    const value = document.getElementById('new-header-value').value;
                    handleHeaderChange(key, value, true);
                    document.getElementById('new-header-key').value = '';
                    document.getElementById('new-header-value').value = '';
                  }}
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Add Header
                </button>
              </div>
            </div>
          </div>

          {/* Query Parameters */}
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Required Query Parameters</h2>
            <div className="space-y-4">
              {Object.entries(route.query_params).map(([key, value]) => (
                <div key={key} className="flex items-center space-x-2">
                  <input
                    type="text"
                    value={key}
                    disabled
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                  />
                  <input
                    type="text"
                    value={value}
                    onChange={(e) => handleQueryParamChange(key, e.target.value)}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                  />
                  <button
                    type="button"
                    onClick={() => handleRemoveQueryParam(key)}
                    className="inline-flex items-center p-1 border border-transparent rounded-full shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
              
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  placeholder="Parameter Name"
                  id="new-param-key"
                  className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                />
                <input
                  type="text"
                  placeholder="Parameter Value"
                  id="new-param-value"
                  className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                />
                <button
                  type="button"
                  onClick={() => {
                    const key = document.getElementById('new-param-key').value;
                    const value = document.getElementById('new-param-value').value;
                    handleQueryParamChange(key, value, true);
                    document.getElementById('new-param-key').value = '';
                    document.getElementById('new-param-value').value = '';
                  }}
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Add Parameter
                </button>
              </div>
            </div>
          </div>

          {/* Feature Toggles */}
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Advanced Features</h2>
            <p className="text-sm text-gray-500 mb-4">
              Toggle advanced features for this route. Configuration for each feature can be set in the respective sections.
            </p>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4">
              <div className="border rounded-lg p-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <label htmlFor="auth_enabled" className="text-sm font-medium text-gray-700">Authentication</label>
                  <input
                    type="checkbox"
                    id="auth_enabled"
                    name="auth_enabled"
                    checked={route.auth_enabled}
                    onChange={handleChange}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                </div>
                <p className="mt-1 text-xs text-gray-500">API key, JWT, or OAuth</p>
              </div>
              
              <div className="border rounded-lg p-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <label htmlFor="rate_limit_enabled" className="text-sm font-medium text-gray-700">Rate Limiting</label>
                  <input
                    type="checkbox"
                    id="rate_limit_enabled"
                    name="rate_limit_enabled"
                    checked={route.rate_limit_enabled}
                    onChange={handleChange}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                </div>
                <p className="mt-1 text-xs text-gray-500">Throttle requests by IP or key</p>
              </div>
              
              <div className="border rounded-lg p-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <label htmlFor="cache_enabled" className="text-sm font-medium text-gray-700">Caching</label>
                  <input
                    type="checkbox"
                    id="cache_enabled"
                    name="cache_enabled"
                    checked={route.cache_enabled}
                    onChange={handleChange}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                </div>
                <p className="mt-1 text-xs text-gray-500">Cache responses to improve performance</p>
              </div>
              
              <div className="border rounded-lg p-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <label htmlFor="circuit_breaker_enabled" className="text-sm font-medium text-gray-700">Circuit Breaker</label>
                  <input
                    type="checkbox"
                    id="circuit_breaker_enabled"
                    name="circuit_breaker_enabled"
                    checked={route.circuit_breaker_enabled}
                    onChange={handleChange}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                </div>
                <p className="mt-1 text-xs text-gray-500">Prevent cascading failures</p>
              </div>
            </div>
          </div>

          <div className="flex justify-end">
            <button
              type="button"
              className="mr-3 inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              onClick={() => navigate(`/routes/${id}`)}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              {saving ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                  </svg>
                  Saving...
                </>
              ) : (
                <>
                  <Save className="w-4 h-4 mr-2" />
                  Save Changes
                </>
              )}
            </button>
          </div>
        </form>
      )}
    </div>
  );
}