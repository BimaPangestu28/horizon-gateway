class ApiService {
  constructor() {
    this.baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8081/admin';
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
          errorData.error || `API request failed with status ${response.status}`,
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
  
  // Timeout wrapper for fetch to avoid hanging requests
  async requestWithTimeout(endpoint, options = {}, timeout = 3000) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    
    try {
      const response = await this.request(endpoint, {
        ...options,
        signal: controller.signal
      });
      clearTimeout(timeoutId);
      return response;
    } catch (error) {
      clearTimeout(timeoutId);
      throw error;
    }
  }

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
}

const apiService = new ApiService();
export default apiService;