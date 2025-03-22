import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeft, Save, XCircle } from 'react-feather';
import ApiService from '../services/ApiService';

export default function RouteEdit() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [formValues, setFormValues] = useState({
    name: '',
    listen_path: '',
    upstream_url: '',
    methods: [],
    strip_path: false,
    priority: 0,
    auth_enabled: false,
    auth_type: 'api_key',
    rate_limit_enabled: false,
    rate_limit: {
      limit: 100,
      window: '60s',
      type: 'sliding_window'
    },
    circuit_breaker_enabled: false,
    circuit_breaker: {
      error_threshold: 50,
      min_requests: 20,
      window: '60s',
      type: 'error'
    },
    cache_enabled: false,
    cache: {
      ttl: '5m',
      type: 'memory'
    }
  });

  useEffect(() => {
    fetchRouteDetails();
  }, [id]);

  const fetchRouteDetails = async () => {
    setLoading(true);
    try {
      // In a real app, we'd call the API
      // For now, generate fake data based on ID
      const mockRoute = {
        id: id,
        name: `route-${id}`,
        listen_path: `/api/${id}/*`,
        upstream_url: `http://service-${id}.internal:8000`,
        methods: ["GET", "POST", "PUT", "DELETE"],
        strip_path: true,
        priority: 10,
        auth_enabled: Math.random() > 0.5,
        auth_type: Math.random() > 0.5 ? "api_key" : "jwt",
        cache_enabled: Math.random() > 0.5,
        rate_limit_enabled: Math.random() > 0.5,
        rate_limit: {
          limit: 100,
          window: "60s",
          type: "sliding_window"
        },
        circuit_breaker_enabled: Math.random() > 0.5,
        circuit_breaker: {
          error_threshold: 50,
          min_requests: 20,
          window: "60s",
          type: "error"
        },
        cache: {
          ttl: "5m",
          type: "memory"
        }
      };
      
      setFormValues(mockRoute);
      setError(null);
    } catch (err) {
      console.error(err);
      setError("Failed to load route details. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    
    if (type === 'checkbox') {
      setFormValues({ ...formValues, [name]: checked });
    } else {
      setFormValues({ ...formValues, [name]: value });
    }
  };

  const handleMethodChange = (method) => {
    const currentMethods = [...formValues.methods];
    
    if (currentMethods.includes(method)) {
      const updatedMethods = currentMethods.filter(m => m !== method);
      setFormValues({ ...formValues, methods: updatedMethods });
    } else {
      setFormValues({ ...formValues, methods: [...currentMethods, method] });
    }
  };

  const handleAllMethods = () => {
    if (formValues.methods.length === HTTP_METHODS.length) {
      setFormValues({ ...formValues, methods: [] });
    } else {
      setFormValues({ ...formValues, methods: [...HTTP_METHODS] });
    }
  };

  const handleNestedChange = (object, field, value) => {
    setFormValues({
      ...formValues,
      [object]: {
        ...formValues[object],
        [field]: value
      }
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    
    try {
      // In a real app, we'd call the API
      // For now, just simulate a delay
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      navigate(`/routes/${id}`);
    } catch (err) {
      console.error(err);
      setError("Failed to save route. Please try again.");
      setSaving(false);
    }
  };

  const HTTP_METHODS = ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"];

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center">
          <Link to={`/routes/${id}`} className="mr-4 text-gray-500 hover:text-gray-700">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <h1 className="text-2xl font-semibold text-gray-900">Edit Route: {formValues.name}</h1>
        </div>
      </div>
      
      {error && (
        <div className="mb-4 bg-red-50 p-4 rounded-md">
          <div className="flex">
            <div className="flex-shrink-0">
              <XCircle className="h-5 w-5 text-red-400" aria-hidden="true" />
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">{error}</h3>
            </div>
          </div>
        </div>
      )}
      
      <form onSubmit={handleSubmit} className="space-y-8 divide-y divide-gray-200">
        <div className="space-y-8 divide-y divide-gray-200">
          {/* Basic Details Section */}
          <div>
            <div>
              <h3 className="text-lg leading-6 font-medium text-gray-900">Basic Details</h3>
              <p className="mt-1 text-sm text-gray-500">
                Configure the core route settings.
              </p>
            </div>

            <div className="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
              {/* Name */}
              <div className="sm:col-span-3">
                <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                  Route Name
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="name"
                    id="name"
                    value={formValues.name}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    required
                  />
                </div>
              </div>

              {/* Priority */}
              <div className="sm:col-span-3">
                <label htmlFor="priority" className="block text-sm font-medium text-gray-700">
                  Priority
                </label>
                <div className="mt-1">
                  <input
                    type="number"
                    name="priority"
                    id="priority"
                    min="0"
                    max="100"
                    value={formValues.priority}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                  />
                </div>
                <p className="mt-2 text-sm text-gray-500">
                  Higher values indicate higher priority when multiple routes match.
                </p>
              </div>

              {/* Listen Path */}
              <div className="sm:col-span-6">
                <label htmlFor="listen_path" className="block text-sm font-medium text-gray-700">
                  Listen Path
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="listen_path"
                    id="listen_path"
                    value={formValues.listen_path}
                    onChange={handleChange}
                    placeholder="/api/*"
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    required
                  />
                </div>
                <p className="mt-2 text-sm text-gray-500">
                  The path pattern to match incoming requests. Use * for wildcards.
                </p>
              </div>

              {/* Upstream URL */}
              <div className="sm:col-span-6">
                <label htmlFor="upstream_url" className="block text-sm font-medium text-gray-700">
                  Upstream URL
                </label>
                <div className="mt-1">
                  <input
                    type="text"
                    name="upstream_url"
                    id="upstream_url"
                    value={formValues.upstream_url}
                    onChange={handleChange}
                    placeholder="http://example.com"
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    required
                  />
                </div>
                <p className="mt-2 text-sm text-gray-500">
                  The target URL where requests will be proxied to.
                </p>
              </div>

              {/* HTTP Methods */}
              <div className="sm:col-span-6">
                <fieldset>
                  <legend className="text-sm font-medium text-gray-700">HTTP Methods</legend>
                  <div className="mt-2 space-y-4">
                    <div className="flex items-center">
                      <input
                        id="all-methods"
                        name="all-methods"
                        type="checkbox"
                        className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                        checked={formValues.methods.length === HTTP_METHODS.length}
                        onChange={handleAllMethods}
                      />
                      <label htmlFor="all-methods" className="ml-3 text-sm font-medium text-gray-700">
                        All Methods
                      </label>
                    </div>
                    
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-y-3">
                      {HTTP_METHODS.map((method) => (
                        <div key={method} className="flex items-center">
                          <input
                            id={`method-${method}`}
                            name={`method-${method}`}
                            type="checkbox"
                            className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                            checked={formValues.methods.includes(method)}
                            onChange={() => handleMethodChange(method)}
                          />
                          <label htmlFor={`method-${method}`} className="ml-3 text-sm font-medium text-gray-700">
                            {method}
                          </label>
                        </div>
                      ))}
                    </div>
                  </div>
                </fieldset>
              </div>

              {/* Strip Path */}
              <div className="sm:col-span-6">
                <div className="flex items-start">
                  <div className="flex items-center h-5">
                    <input
                      id="strip_path"
                      name="strip_path"
                      type="checkbox"
                      checked={formValues.strip_path}
                      onChange={handleChange}
                      className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                    />
                  </div>
                  <div className="ml-3 text-sm">
                    <label htmlFor="strip_path" className="font-medium text-gray-700">
                      Strip Path
                    </label>
                    <p className="text-gray-500">
                      If enabled, the matched URL prefix will be removed before forwarding the request to the upstream URL.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Authentication Section */}
          <div className="pt-8">
            <div>
              <h3 className="text-lg leading-6 font-medium text-gray-900">Authentication</h3>
              <p className="mt-1 text-sm text-gray-500">
                Configure how this route authenticates incoming requests.
              </p>
            </div>

            <div className="mt-6">
              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id="auth_enabled"
                    name="auth_enabled"
                    type="checkbox"
                    checked={formValues.auth_enabled}
                    onChange={handleChange}
                    className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 text-sm">
                  <label htmlFor="auth_enabled" className="font-medium text-gray-700">
                    Enable Authentication
                  </label>
                  <p className="text-gray-500">
                    Require authentication for access to this route.
                  </p>
                </div>
              </div>

              {formValues.auth_enabled && (
                <div className="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div className="sm:col-span-3">
                    <label htmlFor="auth_type" className="block text-sm font-medium text-gray-700">
                      Authentication Type
                    </label>
                    <select
                      id="auth_type"
                      name="auth_type"
                      value={formValues.auth_type}
                      onChange={handleChange}
                      className="mt-1 block w-full py-2 px-3 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    >
                      <option value="api_key">API Key</option>
                      <option value="jwt">JWT</option>
                      <option value="oauth2">OAuth 2.0</option>
                    </select>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Rate Limiting Section */}
          <div className="pt-8">
            <div>
              <h3 className="text-lg leading-6 font-medium text-gray-900">Rate Limiting</h3>
              <p className="mt-1 text-sm text-gray-500">
                Configure rate limiting for this route.
              </p>
            </div>

            <div className="mt-6">
              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id="rate_limit_enabled"
                    name="rate_limit_enabled"
                    type="checkbox"
                    checked={formValues.rate_limit_enabled}
                    onChange={handleChange}
                    className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 text-sm">
                  <label htmlFor="rate_limit_enabled" className="font-medium text-gray-700">
                    Enable Rate Limiting
                  </label>
                  <p className="text-gray-500">
                    Limit the number of requests that can be made to this route.
                  </p>
                </div>
              </div>

              {formValues.rate_limit_enabled && (
                <div className="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div className="sm:col-span-2">
                    <label htmlFor="rate_limit_type" className="block text-sm font-medium text-gray-700">
                      Type
                    </label>
                    <select
                      id="rate_limit_type"
                      value={formValues.rate_limit.type}
                      onChange={(e) => handleNestedChange('rate_limit', 'type', e.target.value)}
                      className="mt-1 block w-full py-2 px-3 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    >
                      <option value="fixed_window">Fixed Window</option>
                      <option value="sliding_window">Sliding Window</option>
                      <option value="token_bucket">Token Bucket</option>
                    </select>
                  </div>

                  <div className="sm:col-span-2">
                    <label htmlFor="rate_limit_limit" className="block text-sm font-medium text-gray-700">
                      Limit
                    </label>
                    <div className="mt-1">
                      <input
                        type="number"
                        id="rate_limit_limit"
                        min="1"
                        value={formValues.rate_limit.limit}
                        onChange={(e) => handleNestedChange('rate_limit', 'limit', parseInt(e.target.value))}
                        className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      />
                    </div>
                  </div>

                  <div className="sm:col-span-2">
                    <label htmlFor="rate_limit_window" className="block text-sm font-medium text-gray-700">
                      Window
                    </label>
                    <div className="mt-1">
                      <input
                        type="text"
                        id="rate_limit_window"
                        value={formValues.rate_limit.window}
                        onChange={(e) => handleNestedChange('rate_limit', 'window', e.target.value)}
                        placeholder="1m, 60s, 1h"
                        className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Circuit Breaker Section */}
          <div className="pt-8">
            <div>
              <h3 className="text-lg leading-6 font-medium text-gray-900">Circuit Breaker</h3>
              <p className="mt-1 text-sm text-gray-500">
                Configure circuit breaking for this route.
              </p>
            </div>

            <div className="mt-6">
              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id="circuit_breaker_enabled"
                    name="circuit_breaker_enabled"
                    type="checkbox"
                    checked={formValues.circuit_breaker_enabled}
                    onChange={handleChange}
                    className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 text-sm">
                  <label htmlFor="circuit_breaker_enabled" className="font-medium text-gray-700">
                    Enable Circuit Breaker
                  </label>
                  <p className="text-gray-500">
                    Automatically stop forwarding requests when the upstream service is failing.
                  </p>
                </div>
              </div>

              {formValues.circuit_breaker_enabled && (
                <div className="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div className="sm:col-span-2">
                    <label htmlFor="circuit_breaker_type" className="block text-sm font-medium text-gray-700">
                      Type
                    </label>
                    <select
                      id="circuit_breaker_type"
                      value={formValues.circuit_breaker.type}
                      onChange={(e) => handleNestedChange('circuit_breaker', 'type', e.target.value)}
                      className="mt-1 block w-full py-2 px-3 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    >
                      <option value="error">Error Rate</option>
                      <option value="timeout">Timeout</option>
                      <option value="concurrency">Concurrency</option>
                      <option value="hybrid">Hybrid</option>
                    </select>
                  </div>

                  <div className="sm:col-span-2">
                    <label htmlFor="circuit_breaker_error_threshold" className="block text-sm font-medium text-gray-700">
                      Error Threshold (%)
                    </label>
                    <div className="mt-1">
                      <input
                        type="number"
                        id="circuit_breaker_error_threshold"
                        min="0"
                        max="100"
                        value={formValues.circuit_breaker.error_threshold}
                        onChange={(e) => handleNestedChange('circuit_breaker', 'error_threshold', parseInt(e.target.value))}
                        className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      />
                    </div>
                  </div>

                  <div className="sm:col-span-2">
                    <label htmlFor="circuit_breaker_min_requests" className="block text-sm font-medium text-gray-700">
                      Minimum Requests
                    </label>
                    <div className="mt-1">
                      <input
                        type="number"
                        id="circuit_breaker_min_requests"
                        min="1"
                        value={formValues.circuit_breaker.min_requests}
                        onChange={(e) => handleNestedChange('circuit_breaker', 'min_requests', parseInt(e.target.value))}
                        className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      />
                    </div>
                  </div>

                  <div className="sm:col-span-2">
                    <label htmlFor="circuit_breaker_window" className="block text-sm font-medium text-gray-700">
                      Window
                    </label>
                    <div className="mt-1">
                      <input
                        type="text"
                        id="circuit_breaker_window"
                        value={formValues.circuit_breaker.window}
                        onChange={(e) => handleNestedChange('circuit_breaker', 'window', e.target.value)}
                        placeholder="1m, 60s, 1h"
                        className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Caching Section */}
          <div className="pt-8">
            <div>
              <h3 className="text-lg leading-6 font-medium text-gray-900">Caching</h3>
              <p className="mt-1 text-sm text-gray-500">
                Configure response caching for this route.
              </p>
            </div>

            <div className="mt-6">
              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id="cache_enabled"
                    name="cache_enabled"
                    type="checkbox"
                    checked={formValues.cache_enabled}
                    onChange={handleChange}
                    className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 text-sm">
                  <label htmlFor="cache_enabled" className="font-medium text-gray-700">
                    Enable Caching
                  </label>
                  <p className="text-gray-500">
                    Cache responses to improve performance.
                  </p>
                </div>
              </div>

              {formValues.cache_enabled && (
                <div className="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div className="sm:col-span-3">
                    <label htmlFor="cache_type" className="block text-sm font-medium text-gray-700">
                      Cache Type
                    </label>
                    <select
                      id="cache_type"
                      value={formValues.cache.type}
                      onChange={(e) => handleNestedChange('cache', 'type', e.target.value)}
                      className="mt-1 block w-full py-2 px-3 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    >
                      <option value="memory">Memory</option>
                      <option value="redis">Redis</option>
                    </select>
                  </div>

                  <div className="sm:col-span-3">
                    <label htmlFor="cache_ttl" className="block text-sm font-medium text-gray-700">
                      TTL (Time to Live)
                    </label>
                    <div className="mt-1">
                      <input
                        type="text"
                        id="cache_ttl"
                        value={formValues.cache.ttl}
                        onChange={(e) => handleNestedChange('cache', 'ttl', e.target.value)}
                        placeholder="5m, 1h, 24h"
                        className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="pt-5">
          <div className="flex justify-end">
            <Link
              to={`/routes/${id}`}
              className="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Cancel
            </Link>
            <button
              type="submit"
              disabled={saving}
              className="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              {saving ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 -ml-1 h-4 w-4" />
                  Save
                </>
              )}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}