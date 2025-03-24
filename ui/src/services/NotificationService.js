import ApiService from './ApiService';

class NotificationService {
  async getNotifications() {
    try {
      return await ApiService.request('/notifications');
    } catch (error) {
      console.log('Notifications API not available, returning mock data');
      return this.getMockNotifications();
    }
  }
  
  async markAsRead(notificationId) {
    try {
      return await ApiService.request(`/notifications/${notificationId}/read`, {
        method: 'POST'
      });
    } catch (error) {
      console.log('Notifications API not available, simulating success');
      return { success: true };
    }
  }
  
  async markAllAsRead() {
    try {
      return await ApiService.request('/notifications/read-all', {
        method: 'POST'
      });
    } catch (error) {
      console.log('Notifications API not available, simulating success');
      return { success: true };
    }
  }
  
  getMockNotifications() {
    return [
      {
        id: 1,
        title: 'New error rate spike',
        description: 'Error rate increased by 5% in the last hour',
        time: '5 min ago',
        read: false,
      },
      {
        id: 2,
        title: 'Rate limit reached',
        description: 'API route /users/* hit rate limit',
        time: '20 min ago',
        read: false,
      },
      {
        id: 3,
        title: 'Circuit breaker open',
        description: 'Circuit breaker triggered for /payments/*',
        time: '1 hour ago',
        read: true,
      },
      {
        id: 4,
        title: 'Config changes detected',
        description: 'Configuration was modified by admin@example.com',
        time: '3 hours ago',
        read: true,
      },
      {
        id: 5,
        title: 'New route added',
        description: 'Route /analytics/* was added to configuration',
        time: '1 day ago',
        read: true,
      }
    ];
  }
}

const notificationService = new NotificationService();
export default notificationService;