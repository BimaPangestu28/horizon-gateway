class ApiService {
  constructor() {
    this.baseUrl =
      process.env.REACT_APP_API_URL || 'http://localhost:8081/admin';
    this.token = localStorage.getItem('auth_token');
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;

    const headers = {
      'Content-Type': 'application/json',
      ...(this.token && { Authorization: `Bearer ${this.token}` }),
      ...(options.headers || {}),
    };

    const config = {
      ...options,
      headers,
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.error ||
            `API request failed with status ${response.status}`,
        );
      }

      if (response.status === 204) {
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error(`API request failed: ${error.message}`);
      throw error;
    }
  }

  // Authentication
  async login(username, password) {
    const response = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });

    if (response && response.token) {
      this.token = response.token;
      localStorage.setItem('auth_token', response.token);
    }

    return response;
  }

  logout() {
    this.token = null;
    localStorage.removeItem('auth_token');
  }

  // Routes
  async getRoutes() {
    return this.request('/routes');
  }

  async getRoute(id) {
    return this.request(`/routes/${id}`);
  }

  async createRoute(routeData) {
    return this.request('/routes', {
      method: 'POST',
      body: JSON.stringify(routeData),
    });
  }

  async updateRoute(id, routeData) {
    return this.request(`/routes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(routeData),
    });
  }

  async deleteRoute(id) {
    return this.request(`/routes/${id}`, {
      method: 'DELETE',
    });
  }

  // API Keys
  async getApiKeys() {
    return this.request('/auth/api-keys');
  }

  async createApiKey(apiKeyData) {
    return this.request('/auth/api-keys', {
      method: 'POST',
      body: JSON.stringify(apiKeyData),
    });
  }

  async deleteApiKey(routeName, keyName) {
    return this.request(`/auth/api-keys/${routeName}/${keyName}`, {
      method: 'DELETE',
    });
  }

  // JWT Configs
  async getJwtConfigs() {
    return this.request('/auth/jwt');
  }

  async updateJwtConfig(routeName, jwtData) {
    return this.request(`/auth/jwt/${routeName}`, {
      method: 'PUT',
      body: JSON.stringify(jwtData),
    });
  }

  // Rate Limits
  async getRateLimits() {
    return this.request('/rate-limits');
  }

  async getRateLimit(routeName) {
    return this.request(`/rate-limits/${routeName}`);
  }

  async createRateLimit(rateLimitData) {
    return this.request('/rate-limits', {
      method: 'POST',
      body: JSON.stringify(rateLimitData),
    });
  }

  async updateRateLimit(routeName, rateLimitData) {
    return this.request(`/rate-limits/${routeName}`, {
      method: 'PUT',
      body: JSON.stringify(rateLimitData),
    });
  }

  async deleteRateLimit(routeName) {
    return this.request(`/rate-limits/${routeName}`, {
      method: 'DELETE',
    });
  }

  // Circuit Breakers
  async getCircuitBreakers() {
    return this.request('/circuit-breakers');
  }

  async getCircuitBreaker(routeName) {
    return this.request(`/circuit-breakers/${routeName}`);
  }

  async createCircuitBreaker(circuitBreakerData) {
    return this.request('/circuit-breakers', {
      method: 'POST',
      body: JSON.stringify(circuitBreakerData),
    });
  }

  async updateCircuitBreaker(routeName, circuitBreakerData) {
    return this.request(`/circuit-breakers/${routeName}`, {
      method: 'PUT',
      body: JSON.stringify(circuitBreakerData),
    });
  }

  async deleteCircuitBreaker(routeName) {
    return this.request(`/circuit-breakers/${routeName}`, {
      method: 'DELETE',
    });
  }

  async resetCircuitBreaker(routeName) {
    return this.request(`/circuit-breakers/${routeName}/reset`, {
      method: 'POST',
    });
  }

  // Cache
  async getCacheConfigs() {
    return this.request('/cache');
  }

  async getCacheConfig(routeName) {
    return this.request(`/cache/${routeName}`);
  }

  async createCacheConfig(cacheData) {
    return this.request('/cache', {
      method: 'POST',
      body: JSON.stringify(cacheData),
    });
  }

  async updateCacheConfig(routeName, cacheData) {
    return this.request(`/cache/${routeName}`, {
      method: 'PUT',
      body: JSON.stringify(cacheData),
    });
  }

  async deleteCacheConfig(routeName) {
    return this.request(`/cache/${routeName}`, {
      method: 'DELETE',
    });
  }

  async clearCache(routeName) {
    return this.request(`/cache/${routeName}/clear`, {
      method: 'POST',
    });
  }

  // Config Management
  async getConfig() {
    return this.request('/config');
  }

  async updateConfig(configData) {
    return this.request('/config', {
      method: 'PUT',
      body: JSON.stringify(configData),
    });
  }

  async getConfigBackups() {
    return this.request('/config/backups');
  }

  async createConfigBackup() {
    return this.request('/config/backups', {
      method: 'POST',
    });
  }

  async restoreConfigBackup(backupId) {
    return this.request(`/config/backups/${backupId}/restore`, {
      method: 'POST',
    });
  }

  // Metrics
  async getMetrics(params) {
    const queryString = new URLSearchParams(params).toString();
    return this.request(`/metrics?${queryString}`);
  }
}

// Create and export a singleton instance
const apiService = new ApiService();
export default apiService;
