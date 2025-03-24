import React, { useState, useEffect } from 'react';
import { Trash2, HelpCircle, Info } from 'react-feather';
import ClientExceptionList from './ClientExceptionList';
import { Tooltip } from '@/components/ui/tooltip';
import { TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

export default function RateLimitForm({
  routes,
  formData,
  onChange,
  onAddException,
  onRemoveException,
  onSubmit,
  onCancel,
  isCreating = true,
  submitButtonText = "Create Rate Limit",
  submitButtonIcon = null
}) {
  const [clientKey, setClientKey] = useState('');
  const [clientLimit, setClientLimit] = useState('');
  const [errors, setErrors] = useState({});
  const [isFormValid, setIsFormValid] = useState(true);

  // Validate the form whenever formData changes
  useEffect(() => {
    validateForm();
  }, [formData]);

  const validateForm = () => {
    const newErrors = {};
    let valid = true;

    // Route is required
    if (isCreating && !formData.route) {
      newErrors.route = 'Route is required';
      valid = false;
    }

    // Limit must be positive
    if (!formData.limit || formData.limit <= 0) {
      newErrors.limit = 'Limit must be greater than 0';
      valid = false;
    }

    // Window format validation (simple regex for now)
    if (!formData.window || !/^\d+[smhd]$/.test(formData.window)) {
      newErrors.window = 'Window must be in format like 60s, 5m, 1h, 1d';
      valid = false;
    }

    // Response code must be between 400-599
    if (!formData.response_code || formData.response_code < 400 || formData.response_code > 599) {
      newErrors.response_code = 'Status code must be between 400-599';
      valid = false;
    }

    // Response message is required
    if (!formData.response_message) {
      newErrors.response_message = 'Response message is required';
      valid = false;
    }

    setErrors(newErrors);
    setIsFormValid(valid);
    return valid;
  };

  const handleAddException = () => {
    if (!clientKey || !clientLimit) return;
    
    onAddException({
      key: clientKey,
      limit: parseInt(clientLimit)
    });
    
    // Reset form fields
    setClientKey('');
    setClientLimit('');
  };

  const handleFormSubmit = (e) => {
    e.preventDefault();
    if (validateForm()) {
      onSubmit(e);
    }
  };

  // Helper function to render a tooltip with information
  const renderTooltip = (content) => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <button type="button" className="ml-1 text-gray-400 hover:text-gray-500">
            <HelpCircle className="h-4 w-4" />
          </button>
        </TooltipTrigger>
        <TooltipContent>
          <p className="max-w-xs text-sm">{content}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );

  return (
    <form className="mt-5 sm:mt-6" onSubmit={handleFormSubmit} noValidate>
      <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
        {isCreating ? (
          <div className="sm:col-span-3">
            <div className="flex items-center">
              <label htmlFor="route" className="block text-sm font-medium text-gray-700">
                Route
              </label>
              {renderTooltip("Select which API route this rate limit will apply to, or choose 'Global' to apply to all routes")}
            </div>
            <select
              id="route"
              name="route"
              value={formData.route}
              onChange={(e) => onChange('route', e.target.value)}
              className={`mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md ${errors.route ? 'border-red-300' : ''}`}
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
            {errors.route && (
              <p className="mt-1 text-sm text-red-600">{errors.route}</p>
            )}
          </div>
        ) : (
          <div className="sm:col-span-3">
            <label className="block text-sm font-medium text-gray-700">
              Route
            </label>
            <div className="mt-1 block w-full py-2 text-base text-gray-700">
              {formData.global ? 'Global (All Routes)' : formData.route}
            </div>
          </div>
        )}

        <div className="sm:col-span-3">
          <div className="flex items-center">
            <label htmlFor="type" className="block text-sm font-medium text-gray-700">
              Rate Limit Type
            </label>
            {renderTooltip(
              "Sliding Window: Uses a moving time window for more accurate limiting.\n" +
              "Fixed Window: Resets counters at fixed intervals.\n" +
              "Token Bucket: Allows bursts of traffic up to a certain point."
            )}
          </div>
          <select
            id="type"
            name="type"
            value={formData.type}
            onChange={(e) => onChange('type', e.target.value)}
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          >
            <option value="sliding_window">Sliding Window</option>
            <option value="fixed_window">Fixed Window</option>
            <option value="token_bucket">Token Bucket</option>
          </select>
          <p className="mt-1 text-xs text-gray-500">
            {formData.type === 'sliding_window' && 'Smoothest limiting that moves with time for accurate throttling'}
            {formData.type === 'fixed_window' && 'Simple time-based limiting that resets at fixed intervals'}
            {formData.type === 'token_bucket' && 'Allows brief traffic bursts while maintaining average rate limits'}
          </p>
        </div>

        <div className="sm:col-span-3">
          <div className="flex items-center">
            <label htmlFor="limit" className="block text-sm font-medium text-gray-700">
              Request Limit
            </label>
            {renderTooltip("Maximum number of requests allowed within the time window")}
          </div>
          <input
            type="number"
            name="limit"
            id="limit"
            min="1"
            value={formData.limit}
            onChange={(e) => onChange('limit', parseInt(e.target.value))}
            className={`mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md ${errors.limit ? 'border-red-300' : ''}`}
            required
          />
          {errors.limit && (
            <p className="mt-1 text-sm text-red-600">{errors.limit}</p>
          )}
        </div>

        <div className="sm:col-span-3">
          <div className="flex items-center">
            <label htmlFor="window" className="block text-sm font-medium text-gray-700">
              Time Window
            </label>
            {renderTooltip("Time period for rate limiting (e.g., 60s = 60 seconds, 5m = 5 minutes, 1h = 1 hour, 1d = 1 day)")}
          </div>
          <div className="mt-1 flex rounded-md shadow-sm">
            <input
              type="text"
              name="window"
              id="window"
              value={formData.window}
              onChange={(e) => onChange('window', e.target.value)}
              className={`flex-1 block w-full focus:ring-indigo-500 focus:border-indigo-500 min-w-0 rounded-none rounded-l-md sm:text-sm border-gray-300 ${errors.window ? 'border-red-300' : ''}`}
              required
            />
            <span className="inline-flex items-center px-3 rounded-r-md border border-l-0 border-gray-300 bg-gray-50 text-gray-500 sm:text-sm">
              e.g. 60s, 5m, 1h, 1d
            </span>
          </div>
          {errors.window && (
            <p className="mt-1 text-sm text-red-600">{errors.window}</p>
          )}
        </div>

        <div className="sm:col-span-3">
          <div className="flex items-center">
            <label htmlFor="key" className="block text-sm font-medium text-gray-700">
              Rate Limit Key
            </label>
            {renderTooltip("The identifier used to track and limit requests from clients")}
          </div>
          <select
            id="key"
            name="key"
            value={formData.key}
            onChange={(e) => onChange('key', e.target.value)}
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          >
            <option value="ip">Client IP</option>
            <option value="api_key">API Key</option>
            <option value="user_id">User ID</option>
            <option value="header:x-client-id">Header (x-client-id)</option>
            <option value="header:authorization">Header (Authorization)</option>
            <option value="custom">Custom (Advanced)</option>
          </select>
          <p className="mt-1 text-xs text-gray-500">
            {formData.key === 'ip' && 'Limits per client IP address - good for anonymous access'}
            {formData.key === 'api_key' && 'Limits per API key - ideal for authenticated API usage'}
            {formData.key === 'user_id' && 'Limits per user ID - best for logged-in users'}
            {formData.key === 'header:x-client-id' && 'Limits based on the x-client-id header value'}
            {formData.key === 'header:authorization' && 'Limits based on the Authorization header value'}
            {formData.key === 'custom' && 'Advanced: Use a custom limiting strategy'}
          </p>
        </div>

        <div className="sm:col-span-3">
          <div className="flex items-center">
            <label htmlFor="response_code" className="block text-sm font-medium text-gray-700">
              Response Status Code
            </label>
            {renderTooltip("HTTP status code to return when rate limit is exceeded (typically 429 - Too Many Requests)")}
          </div>
          <input
            type="number"
            name="response_code"
            id="response_code"
            min="400"
            max="599"
            value={formData.response_code}
            onChange={(e) => onChange('response_code', parseInt(e.target.value))}
            className={`mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md ${errors.response_code ? 'border-red-300' : ''}`}
          />
          {errors.response_code && (
            <p className="mt-1 text-sm text-red-600">{errors.response_code}</p>
          )}
          {formData.response_code === 429 && (
            <p className="mt-1 text-xs text-green-600">✓ Using standard "Too Many Requests" code</p>
          )}
        </div>

        <div className="sm:col-span-6">
          <div className="flex items-center">
            <label htmlFor="response_message" className="block text-sm font-medium text-gray-700">
              Response Message
            </label>
            {renderTooltip("Error message returned to the client when rate limit is exceeded")}
          </div>
          <input
            type="text"
            name="response_message"
            id="response_message"
            value={formData.response_message}
            onChange={(e) => onChange('response_message', e.target.value)}
            className={`mt-1 block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md ${errors.response_message ? 'border-red-300' : ''}`}
          />
          {errors.response_message && (
            <p className="mt-1 text-sm text-red-600">{errors.response_message}</p>
          )}
        </div>

        <div className="sm:col-span-6">
          <div className="flex items-start">
            <div className="flex items-center h-5">
              <input
                id="include_headers"
                name="include_headers"
                type="checkbox"
                checked={formData.include_headers}
                onChange={(e) => onChange('include_headers', e.target.checked)}
                className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
              />
            </div>
            <div className="ml-3 text-sm">
              <label htmlFor="include_headers" className="font-medium text-gray-700 flex items-center">
                Include Rate Limit Headers
                {renderTooltip("Adds standard headers to responses so clients can track their usage: X-RateLimit-Limit, X-RateLimit-Remaining, and X-RateLimit-Reset")}
              </label>
              <p className="text-gray-500">
                Add X-RateLimit-* headers to responses (recommended for API clients)
              </p>
              {formData.include_headers && (
                <div className="mt-2 text-xs text-gray-500 bg-gray-50 p-2 rounded">
                  <div><code className="font-mono">X-RateLimit-Limit</code>: Shows max requests allowed</div>
                  <div><code className="font-mono">X-RateLimit-Remaining</code>: Shows remaining requests</div>
                  <div><code className="font-mono">X-RateLimit-Reset</code>: Shows when limit resets (epoch time)</div>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="sm:col-span-3">
          <div className="flex items-start">
            <div className="flex items-center h-5">
              <input
                id="enabled"
                name="enabled"
                type="checkbox"
                checked={formData.enabled}
                onChange={(e) => onChange('enabled', e.target.checked)}
                className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
              />
            </div>
            <div className="ml-3 text-sm">
              <label htmlFor="enabled" className="font-medium text-gray-700 flex items-center">
                Enabled
                {renderTooltip("Turns rate limiting on or off without deleting the configuration")}
              </label>
              <p className="text-gray-500">
                Toggle rate limiting on or off
              </p>
            </div>
          </div>
        </div>
        
        <div className="sm:col-span-3">
          <div className="flex justify-end mt-4">
            <button
              type="button"
              onClick={() => validateForm()}
              className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <Info className="h-4 w-4 mr-2" />
              Validate Settings
            </button>
          </div>
        </div>

        <div className="sm:col-span-6 border-t pt-4">
          <div className="flex items-center">
            <h4 className="text-sm font-medium text-gray-900">Client Exceptions</h4>
            {renderTooltip("Allow specific clients to have different rate limits than the default")}
          </div>
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
                  value={clientKey}
                  onChange={(e) => setClientKey(e.target.value)}
                  className="block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                  aria-label="Client identifier"
                />
              </div>
              <div className="w-32">
                <input
                  type="number"
                  id="client-limit"
                  placeholder="Limit"
                  min="1"
                  value={clientLimit}
                  onChange={(e) => setClientLimit(e.target.value)}
                  className="block w-full shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm border-gray-300 rounded-md"
                  aria-label="Custom limit for client"
                />
              </div>
              <button
                type="button"
                onClick={handleAddException}
                className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Add
              </button>
            </div>
            
            {formData.client_exceptions.length > 0 ? (
              <ClientExceptionList 
                exceptions={formData.client_exceptions} 
                onRemove={onRemoveException} 
              />
            ) : (
              <div className="mt-3 text-sm text-gray-500 italic bg-gray-50 p-3 rounded border border-gray-200 text-center">
                No client exceptions added. All clients will use the default rate limit.
              </div>
            )}
            
            {formData.client_exceptions.length > 0 && (
              <div className="mt-2 text-xs text-gray-500">
                <p className="font-medium">How client exceptions work:</p>
                <ul className="list-disc pl-5 mt-1 space-y-1">
                  <li>Exceptions are evaluated before the default rate limit</li>
                  <li>The client key must match exactly what you specified as the "Rate Limit Key" above</li>
                  <li>For IP-based limiting, use the full IP address (e.g., 192.168.1.10)</li>
                </ul>
              </div>
            )}
          </div>
        </div>
      </div>
      
      <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
        <button
          type="submit"
          disabled={!isFormValid}
          className={`w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 text-base font-medium text-white focus:outline-none focus:ring-2 focus:ring-offset-2 sm:col-start-2 sm:text-sm
            ${isFormValid ? 'bg-indigo-600 hover:bg-indigo-700 focus:ring-indigo-500' : 'bg-indigo-300 cursor-not-allowed'}`}
        >
          {submitButtonIcon}
          {submitButtonText}
        </button>
        <button
          type="button"
          className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
          onClick={onCancel}
        >
          Cancel
        </button>
      </div>
    </form>
  );
}