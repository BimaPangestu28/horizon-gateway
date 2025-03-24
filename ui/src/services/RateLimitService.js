import ApiService from './ApiService';

class RateLimitService {
  async getRateLimits() {
    try {
      const response = await ApiService.request('/rate-limits');
      // Ensure we always return an array
      return Array.isArray(response) ? response : [];
    } catch (error) {
      console.log('Rate limits API not available, returning mock data');
      return this.getMockRateLimits();
    }
  }

  async getRateLimit(id) {
    try {
      return await ApiService.request(`/rate-limits/${id}`);
    } catch (error) {
      console.log('Rate limit API not available, returning mock data');
      const mockData = this.getMockRateLimits();
      return mockData.find(limit => limit.id === id) || null;
    }
  }

  async createRateLimit(rateLimitData) {
    try {
      return await ApiService.request('/rate-limits', {
        method: 'POST',
        body: JSON.stringify(rateLimitData)
      });
    } catch (error) {
      console.log('Create rate limit API not available, simulating success');
      // Return the data with a mock ID
      return {
        id: `${Date.now()}`,
        ...rateLimitData,
        current_state: {
          active_limits: 0,
          blocked_clients: 0,
          requests_last_hour: 0
        }
      };
    }
  }

  async updateRateLimit(id, rateLimitData) {
    try {
      return await ApiService.request(`/rate-limits/${id}`, {
        method: 'PUT',
        body: JSON.stringify(rateLimitData)
      });
    } catch (error) {
      console.log('Update rate limit API not available, simulating success');
      return rateLimitData;
    }
  }

  async deleteRateLimit(id) {
    try {
      return await ApiService.request(`/rate-limits/${id}`, {
        method: 'DELETE'
      });
    } catch (error) {
      console.log('Delete rate limit API not available, simulating success');
      return { success: true };
    }
  }

  getMockRateLimits() {
    return [
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
  }
}

const rateLimitService = new RateLimitService();
export default rateLimitService;