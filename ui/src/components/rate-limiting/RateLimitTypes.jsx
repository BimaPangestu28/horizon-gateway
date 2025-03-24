import React from 'react';
import { HelpCircle, Clock, Shield, Cpu } from 'react-feather';

const RateLimitTypes = () => {
  return (
    <div className="bg-white shadow overflow-hidden sm:rounded-lg">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Rate Limiting Strategies
        </h3>
        <p className="mt-1 max-w-2xl text-sm text-gray-500">
          Different approaches to rate limiting with their pros and cons.
        </p>
      </div>
      
      <div className="border-t border-gray-200">
        <dl>
          <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 flex items-center">
              <Clock className="h-5 w-5 text-indigo-500 mr-2" />
              <span>Sliding Window</span>
            </dt>
            <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
              <p>A moving time window approach that provides the most accurate rate limiting by continuously tracking requests over time.</p>
              <div className="mt-2">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 mr-2">
                  Most accurate
                </span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 mr-2">
                  More resource intensive
                </span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  Best for APIs with consistent traffic
                </span>
              </div>
            </dd>
          </div>
          
          <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 flex items-center">
              <Shield className="h-5 w-5 text-indigo-500 mr-2" />
              <span>Fixed Window</span>
            </dt>
            <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
              <p>Divides time into fixed windows (e.g., minutes, hours) and counts requests in each window. Resets counter at the start of each window.</p>
              <div className="mt-2">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 mr-2">
                  Simple implementation
                </span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 mr-2">
                  Potential traffic spikes at window boundaries
                </span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  Good for simple scenarios
                </span>
              </div>
            </dd>
          </div>
          
          <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 flex items-center">
              <Cpu className="h-5 w-5 text-indigo-500 mr-2" />
              <span>Token Bucket</span>
            </dt>
            <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
              <p>Uses a "bucket" of tokens that refills at a constant rate. Each request consumes a token, allowing for rate limiting with the ability to handle traffic bursts.</p>
              <div className="mt-2">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 mr-2">
                  Handles traffic bursts
                </span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 mr-2">
                  Moderate complexity
                </span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  Ideal for APIs with variable traffic
                </span>
              </div>
            </dd>
          </div>
        </dl>
      </div>
    </div>
  );
};

export default RateLimitTypes;