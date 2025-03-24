import React, { useState } from 'react';
import { Trash2 } from 'react-feather';
import RateLimitForm from './RateLimitForm';

export default function CreateRateLimitModal({ routes, onSubmit, onCancel }) {
  const [formData, setFormData] = useState({
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

  const handleChange = (name, value) => {
    if (name === 'route' && value === 'global') {
      setFormData({ ...formData, route: value, global: true });
    } else {
      setFormData({ ...formData, [name]: value });
    }
  };

  const handleAddClientException = (exception) => {
    setFormData({
      ...formData,
      client_exceptions: [...formData.client_exceptions, exception]
    });
  };

  const handleRemoveClientException = (index) => {
    const exceptions = [...formData.client_exceptions];
    exceptions.splice(index, 1);
    setFormData({ ...formData, client_exceptions: exceptions });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
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
          
          <RateLimitForm
            routes={routes}
            formData={formData}
            onChange={handleChange}
            onAddException={handleAddClientException}
            onRemoveException={handleRemoveClientException}
            onSubmit={handleSubmit}
            onCancel={onCancel}
            isCreating={true}
          />
        </div>
      </div>
    </div>
  );
}