import ApiService from './ApiService';

class MetricsService {
  async getMetrics(params = {}) {
    try {
      const queryString = new URLSearchParams(params).toString();
      // First try to call the API
      try {
        return await ApiService.request(`/metrics?${queryString}`);
      } catch (error) {
        console.log('Metrics API not available, returning mock data');
        return this.getMockMetrics(params);
      }
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
      throw error;
    }
  }
  
  getMockMetrics(params = {}) {
    const now = new Date();
    const trafficHistory = [];
    const responseTimeHistory = [];
    
    // Generate 24 hours of mock data
    for (let i = 23; i >= 0; i--) {
      const hour = new Date(now);
      hour.setHours(now.getHours() - i);
      const timeStr = hour.getHours() + ':00';
      
      trafficHistory.push({
        time: timeStr,
        requests: Math.floor(Math.random() * 400) + 100,
        errors: Math.floor(Math.random() * 20)
      });
      
      responseTimeHistory.push({
        time: timeStr,
        responseTime: Math.floor(Math.random() * 300) + 70
      });
    }
    
    const topRoutes = [
      { name: '/api/*', requests: 35243, avgResponseTime: 125 },
      { name: '/users/*', requests: 27891, avgResponseTime: 187 },
      { name: '/products/*', requests: 19432, avgResponseTime: 93 },
      { name: '/auth/*', requests: 12985, avgResponseTime: 210 },
      { name: '/search/*', requests: 9872, avgResponseTime: 156 }
    ];
    
    return {
      totalRequests: Math.floor(Math.random() * 1000000) + 500000,
      avgResponseTime: Math.floor(Math.random() * 200) + 80,
      errorRate: (Math.random() * 2),
      activeRoutes: Math.floor(Math.random() * 20) + 5,
      requestsPerSecond: Math.floor(Math.random() * 500) + 100,
      trafficHistory,
      responseTimeHistory,
      topRoutes,
      status: {
        proxyStatus: Math.random() > 0.1 ? 'healthy' : 'degraded',
        adminStatus: Math.random() > 0.05 ? 'healthy' : 'degraded',
        configStatus: Math.random() > 0.2 ? 'up-to-date' : 'modified'
      }
    };
  }
}

const metricsService = new MetricsService();
export default metricsService;