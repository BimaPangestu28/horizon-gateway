import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { 
  Plus, Trash2, Edit2, Key, Shield, RefreshCw, Eye, EyeOff, Clock, Database, 
  Download, Upload, Search
} from 'react-feather';
import ApiService from '../services/ApiService';

export default function Authentication() {
  const [activeTab, setActiveTab] = useState('api-keys');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [apiKeys, setApiKeys] = useState([]);
  const [jwtConfigs, setJwtConfigs] = useState([]);
  const [oauthConfigs, setOauthConfigs] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newApiKey, setNewApiKey] = useState({
    name: '',
    route: '',
    scopes: [],
    rate_limit: 0,
    expires_at: ''
  });
  const [routes, setRoutes] = useState([]);
  const [isKeyVisible, setIsKeyVisible] = useState({});
  const [deleteItem, setDeleteItem] = useState(null);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  useEffect(() => {
    fetchAuthData();
    fetchRoutes();
  }, [activeTab]);

  const fetchAuthData = async () => {
    setLoading(true);
    try {
      // In a real implementation, these would be API calls
      // For now, use sample data
      if (activeTab === 'api-keys') {
        const sampleApiKeys = [
          {
            id: '1',
            name: 'mobile-app-key',
            key: 'ak_12345abcdefghijklmnopqrstuvwxyz',
            route: 'api-route',
            created_at: '2023-01-15T10:30:00Z',
            expires_at: '2023-12-31T23:59:59Z',
            scopes: ['read', 'write'],
            rate_limit: 100,
            status: 'active',
            last_used: '2023-03-10T14:22:10Z'
          },
          {
            id: '2',
            name: 'partner-integration',
            key: 'ak_67890abcdefghijklmnopqrstuvwxyz',
            route: 'partner-api',
            created_at: '2023-02-01T09:15:00Z',
            expires_at: '2023-12-31T23:59:59Z',
            scopes: ['read'],
            rate_limit: 50,
            status: 'active',
            last_used: '2023-03-09T11:45:22Z'
          },
          {
            id: '3',
            name: 'internal-service',
            key: 'ak_54321abcdefghijklmnopqrstuvwxyz',
            route: 'internal-route',
            created_at: '2023-01-20T11:45:00Z',
            expires_at: '2023-06-30T23:59:59Z',
            scopes: ['read', 'write', 'admin'],
            rate_limit: 200,
            status: 'active',
            last_used: '2023-03-10T16:08:45Z'
          }
        ];
        setApiKeys(sampleApiKeys);
      } else if (activeTab === 'jwt') {
        const sampleJwtConfigs = [
          {
            id: '1',
            route: 'api-route',
            issuer: 'https://auth.example.com',
            audience: 'horizon-gateway',
            secret_type: 'HMAC',
            algorithm: 'HS256',
            public_key_url: '',
            claim_mappings: {
              sub: 'user_id',
              scope: 'permissions'
            },
            required_claims: ['sub', 'exp'],
            token_location: 'header',
            header_name: 'Authorization',
            header_prefix: 'Bearer',
            cookie_name: '',
            status: 'active'
          },
          {
            id: '2',
            route: 'partner-api',
            issuer: 'https://partner.example.com',
            audience: 'partner-app',
            secret_type: 'RSA',
            algorithm: 'RS256',
            public_key_url: 'https://partner.example.com/.well-known/jwks.json',
            claim_mappings: {
              sub: 'user_id',
              role: 'user_role'
            },
            required_claims: ['sub', 'exp', 'role'],
            token_location: 'header',
            header_name: 'Authorization',
            header_prefix: 'Bearer',
            cookie_name: '',
            status: 'active'
          }
        ];
        setJwtConfigs(sampleJwtConfigs);
      } else if (activeTab === 'oauth') {
        const sampleOAuthConfigs = [
          {
            id: '1',
            route: 'api-route',
            provider_type: 'generic',
            client_id: 'client_12345',
            client_secret: 'secret_abcdef',
            token_url: 'https://auth.example.com/oauth/token',
            authorize_url: 'https://auth.example.com/oauth/authorize',
            userinfo_url: 'https://auth.example.com/userinfo',
            scopes: ['profile', 'email'],
            callback_url: 'https://gateway.example.com/oauth/callback',
            token_location: 'header',
            header_name: 'Authorization',
            header_prefix: 'Bearer',
            status: 'active'
          },
          {
            id: '2',
            route: 'partner-api',
            provider_type: 'google',
            client_id: 'google_client_12345',
            client_secret: 'google_secret_abcdef',
            token_url: '',
            authorize_url: '',
            userinfo_url: '',
            scopes: ['profile', 'email'],
            callback_url: 'https://gateway.example.com/oauth/callback/google',
            token_location: 'header',
            header_name: 'Authorization',
            header_prefix: 'Bearer',
            status: 'active'
          }
        ];
        setOauthConfigs(sampleOAuthConfigs);
      }
      setError(null);
    } catch (err) {
      console.error(err);
      setError('Failed to fetch authentication data. Please try again.');
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

  const handleCreateApiKey = async (e) => {
    e.preventDefault();
    
    try {
      // In a real implementation, this would call the API
      // For now, simulate API call
      const newKey = {
        id: `${apiKeys.length + 1}`,
        name: newApiKey.name,
        key: `ak_${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`,
        route: newApiKey.route,
        created_at: new Date().toISOString(),
        expires_at: newApiKey.expires_at || new Date(new Date().setFullYear(new Date().getFullYear() + 1)).toISOString(),
        scopes: newApiKey.scopes,
        rate_limit: newApiKey.rate_limit || 100,
        status: 'active',
        last_used: null
      };
      
      setApiKeys([...apiKeys, newKey]);
      setShowCreateModal(false);
      setNewApiKey({
        name: '',
        route: '',
        scopes: [],
        rate_limit: 0,
        expires_at: ''
      });
      
      // Show the newly created key
      setIsKeyVisible({...isKeyVisible, [newKey.id]: true});
      
    } catch (err) {
      console.error(err);
      setError('Failed to create API key. Please try again.');
    }
  };

  const toggleKeyVisibility = (id) => {
    setIsKeyVisible({...isKeyVisible, [id]: !isKeyVisible[id]});
  };

  const handleDeleteConfirm = (item) => {
    setDeleteItem(item);
    setShowDeleteModal(true);
  };

  const handleDelete = async () => {
    if (!deleteItem) return;
    
    try {
      if (activeTab === 'api-keys') {
        setApiKeys(apiKeys.filter(key => key.id !== deleteItem.id));
      } else if (activeTab === 'jwt') {
        setJwtConfigs(jwtConfigs.filter(config => config.id !== deleteItem.id));
      } else if (activeTab === 'oauth') {
        setOauthConfigs(oauthConfigs.filter(config => config.id !== deleteItem.id));
      }
      
      setShowDeleteModal(false);
      setDeleteItem(null);
    } catch (err) {
      console.error(err);
      setError('Failed to delete item. Please try again.');
    }
  };

  const filterBySearch = (item) => {
    if (!searchTerm) return true;
    
    const searchLower = searchTerm.toLowerCase();
    
    if (activeTab === 'api-keys') {
      return (
        item.name.toLowerCase().includes(searchLower) ||
        item.route.toLowerCase().includes(searchLower)
      );
    } else if (activeTab === 'jwt') {
      return (
        item.route.toLowerCase().includes(searchLower) ||
        item.issuer.toLowerCase().includes(searchLower)
      );
    } else if (activeTab === 'oauth') {
      return (
        item.route.toLowerCase().includes(searchLower) ||
        item.provider_type.toLowerCase().includes(searchLower)
      );
    }
    
    return true;
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  const handleScopeToggle = (scope) => {
    const currentScopes = [...newApiKey.scopes];
    
    if (currentScopes.includes(scope)) {
      setNewApiKey({
        ...newApiKey,
        scopes: currentScopes.filter(s => s !== scope)
      });
    } else {
      setNewApiKey({
        ...newApiKey,
        scopes: [...currentScopes, scope]
      });
    }
  };

  return (
    <div>
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Authentication</h1>
          <p className="mt-2 text-sm text-gray-700">
            Manage authentication methods for your API routes.
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          {activeTab === 'api-keys' && (
            <button
              type="button"
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
            >
              <Plus className="mr-2 h-4 w-4" />
              New API Key
            </button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="mt-6 border-b border-gray-200">
        <div className="sm:flex sm:items-baseline">
          <div className="mt-4 sm:mt-0">
            <nav className="-mb-px flex space-x-8">
              <button
                className={`${
                  activeTab === 'api-keys'
                    ? 'border-indigo-500 text-indigo-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                } whitespace-nowrap pb-4 px-1 border-b-2 font-medium text-sm focus:outline-none flex items-center`}
                onClick={() => setActiveTab('api-keys')}
              >
                <Key className="mr-2 h-4 w-4" />
                API Keys
              </button>
              <button
                className={`${
                  activeTab === 'jwt'
                    ? 'border-indigo-500 text-indigo-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                } whitespace-nowrap pb-4 px-1 border-b-2 font-medium text-sm focus:outline-none flex items-center`}
                onClick={() => setActiveTab('jwt')}
              >
                <Shield className="mr-2 h-4 w-4" />
                JWT Authentication
              </button>
              <button
                className={`${
                  activeTab === 'oauth'
                    ? 'border-indigo-500 text-indigo-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                } whitespace-nowrap pb-4 px-1 border-b-2 font-medium text-sm focus:outline-none flex items-center`}
                onClick={() => setActiveTab('oauth')}
              >
                <Database className="mr-2 h-4 w-4" />
                OAuth 2.0
              </button>
            </nav>
          </div>
        </div>
      </div>

      {/* Search bar */}
      <div className="mt-6 mb-6 flex">
        <div className="relative flex-grow">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search className="h-5 w-5 text-gray-400" />
          </div>
          <input
            type="text"
            className="focus:ring-indigo-500 focus:border-indigo-500 block w-full pl-10 sm:text-sm border-gray-300 rounded-md"
            placeholder={`Search ${activeTab === 'api-keys' ? 'API keys' : activeTab === 'jwt' ? 'JWT configs' : 'OAuth configs'}...`}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <button
          type="button"
          onClick={fetchAuthData}
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
        <>
          {/* API Keys Tab Content */}
          {activeTab === 'api-keys' && (
            <div className="mt-6 flex flex-col">
              <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
                <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
                  <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                    <table className="min-w-full divide-y divide-gray-300">
                      <thead className="bg-gray-50">
                        <tr>
                          <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">Name</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">API Key</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Route</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Scopes</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Expires At</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Rate Limit</th>
                          <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                            <span className="sr-only">Actions</span>
                          </th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-200 bg-white">
                        {apiKeys.filter(filterBySearch).map((key) => (
                          <tr key={key.id}>
                            <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{key.name}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              <div className="flex items-center">
                                <span className="font-mono">
                                  {isKeyVisible[key.id] ? key.key : '••••••••••••••••••••••••••••••••••'}
                                </span>
                                <button
                                  onClick={() => toggleKeyVisibility(key.id)}
                                  className="ml-2 text-indigo-600 hover:text-indigo-900"
                                >
                                  {isKeyVisible[key.id] ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                              </div>
                            </td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{key.route}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              <div className="flex flex-wrap gap-1">
                                {key.scopes.map((scope) => (
                                  <span key={scope} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                    {scope}
                                  </span>
                                ))}
                              </div>
                            </td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{formatDate(key.expires_at)}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{key.rate_limit} req/min</td>
                            <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                              <div className="flex justify-end space-x-2">
                                <button className="text-indigo-600 hover:text-indigo-900">
                                  <Edit2 className="h-4 w-4" />
                                </button>
                                <button 
                                  className="text-red-600 hover:text-red-900"
                                  onClick={() => handleDeleteConfirm(key)}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* JWT Authentication Tab Content */}
          {activeTab === 'jwt' && (
            <div className="mt-6 flex flex-col">
              <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
                <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
                  <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                    <table className="min-w-full divide-y divide-gray-300">
                      <thead className="bg-gray-50">
                        <tr>
                          <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">Route</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Issuer</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Algorithm</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Secret Type</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Token Location</th>
                          <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                            <span className="sr-only">Actions</span>
                          </th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-200 bg-white">
                        {jwtConfigs.filter(filterBySearch).map((config) => (
                          <tr key={config.id}>
                            <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{config.route}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{config.issuer}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{config.algorithm}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{config.secret_type}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              {config.token_location === 'header' 
                                ? `Header (${config.header_name})` 
                                : `Cookie (${config.cookie_name})`}
                            </td>
                            <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                              <div className="flex justify-end space-x-2">
                                <button className="text-indigo-600 hover:text-indigo-900">
                                  <Edit2 className="h-4 w-4" />
                                </button>
                                <button 
                                  className="text-red-600 hover:text-red-900"
                                  onClick={() => handleDeleteConfirm(config)}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* OAuth 2.0 Tab Content */}
          {activeTab === 'oauth' && (
            <div className="mt-6 flex flex-col">
              <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
                <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
                  <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                    <table className="min-w-full divide-y divide-gray-300">
                      <thead className="bg-gray-50">
                        <tr>
                          <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">Route</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Provider</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Client ID</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Scopes</th>
                          <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Callback URL</th>
                          <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                            <span className="sr-only">Actions</span>
                          </th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-200 bg-white">
                        {oauthConfigs.filter(filterBySearch).map((config) => (
                          <tr key={config.id}>
                            <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{config.route}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{config.provider_type}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{config.client_id}</td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              <div className="flex flex-wrap gap-1">
                                {config.scopes.map((scope) => (
                                  <span key={scope} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                    {scope}
                                  </span>
                                ))}
                              </div>
                            </td>
                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{config.callback_url}</td>
                            <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                              <div className="flex justify-end space-x-2">
                                <button className="text-indigo-600 hover:text-indigo-900">
                                  <Edit2 className="h-4 w-4" />
                                </button>
                                <button 
                                  className="text-red-600 hover:text-red-900"
                                  onClick={() => handleDeleteConfirm(config)}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          )}
        </>
      )}

      {/* Create API Key Modal */}
      {showCreateModal && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity" aria-hidden="true">
              <div className="absolute inset-0 bg-gray-500 opacity-75"></div>
            </div>

            <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

            <div className="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
              <div>
                <div className="mt-3 text-center sm:mt-5">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Create New API Key</h3>
                  <div className="mt-2">
                    <p className="text-sm text-gray-500">
                      Create a new API key for authentication to your routes.
                    </p>
                  </div>
                </div>
              </div>
              
              <form className="mt-5 sm:mt-6" onSubmit={handleCreateApiKey}>
                <div className="space-y-4">
                  <div>
                    <label htmlFor="key-name" className="block text-sm font-medium text-gray-700">
                      Key Name
                    </label>
                    <input
                      type="text"
                      name="key-name"
                      id="key-name"
                      value={newApiKey.name}
                      onChange={(e) => setNewApiKey({...newApiKey, name: e.target.value})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                      placeholder="mobile-app-key"
                      required
                    />
                  </div>
                  
                  <div>
                    <label htmlFor="route" className="block text-sm font-medium text-gray-700">
                      Route
                    </label>
                    <select
                      id="route"
                      name="route"
                      value={newApiKey.route}
                      onChange={(e) => setNewApiKey({...newApiKey, route: e.target.value})}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                      required
                    >
                      <option value="">Select a route</option>
                      {routes.map((route) => (
                        <option key={route.id} value={route.name}>
                          {route.name} ({route.listen_path})
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Scopes
                    </label>
                    <div className="mt-2 space-y-2">
                      <div className="flex items-center">
                        <input
                          id="scope-read"
                          name="scope-read"
                          type="checkbox"
                          checked={newApiKey.scopes.includes('read')}
                          onChange={() => handleScopeToggle('read')}
                          className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                        />
                        <label htmlFor="scope-read" className="ml-2 block text-sm text-gray-900">
                          read
                        </label>
                      </div>
                      
                      <div className="flex items-center">
                        <input
                          id="scope-write"
                          name="scope-write"
                          type="checkbox"
                          checked={newApiKey.scopes.includes('write')}
                          onChange={() => handleScopeToggle('write')}
                          className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                        />
                        <label htmlFor="scope-write" className="ml-2 block text-sm text-gray-900">
                          write
                        </label>
                      </div>
                      
                      <div className="flex items-center">
                        <input
                          id="scope-admin"
                          name="scope-admin"
                          type="checkbox"
                          checked={newApiKey.scopes.includes('admin')}
                          onChange={() => handleScopeToggle('admin')}
                          className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                        />
                        <label htmlFor="scope-admin" className="ml-2 block text-sm text-gray-900">
                          admin
                        </label>
                      </div>
                    </div>
                  </div>
                  
                  <div>
                    <label htmlFor="rate-limit" className="block text-sm font-medium text-gray-700">
                      Rate Limit (requests per minute)
                    </label>
                    <input
                      type="number"
                      name="rate-limit"
                      id="rate-limit"
                      min="0"
                      value={newApiKey.rate_limit}
                      onChange={(e) => setNewApiKey({...newApiKey, rate_limit: parseInt(e.target.value)})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                      placeholder="100"
                    />
                  </div>
                  
                  <div>
                    <label htmlFor="expires-at" className="block text-sm font-medium text-gray-700">
                      Expiration Date (optional)
                    </label>
                    <input
                      type="date"
                      name="expires-at"
                      id="expires-at"
                      value={newApiKey.expires_at ? new Date(newApiKey.expires_at).toISOString().split('T')[0] : ''}
                      onChange={(e) => setNewApiKey({...newApiKey, expires_at: e.target.value})}
                      className="mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                    />
                  </div>
                </div>
                
                <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                  <button
                    type="submit"
                    className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                  >
                    Create Key
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

      {/* Delete Confirmation Modal */}
      {showDeleteModal && deleteItem && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity" aria-hidden="true">
              <div className="absolute inset-0 bg-gray-500 opacity-75"></div>
            </div>

            <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

            <div className="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
              <div>
                <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
                  <Trash2 className="h-6 w-6 text-red-600" aria-hidden="true" />
                </div>
                <div className="mt-3 text-center sm:mt-5">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">
                    Delete {activeTab === 'api-keys' ? 'API Key' : activeTab === 'jwt' ? 'JWT Config' : 'OAuth Config'}
                  </h3>
                  <div className="mt-2">
                    <p className="text-sm text-gray-500">
                      Are you sure you want to delete {activeTab === 'api-keys' ? `the API key "${deleteItem.name}"` : activeTab === 'jwt' ? `the JWT config for route "${deleteItem.route}"` : `the OAuth config for route "${deleteItem.route}"`}? This action cannot be undone.
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