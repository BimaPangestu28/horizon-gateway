import ApiService from './ApiService';

class RoutesService {
  async getRoutes() {
    return ApiService.request('/routes');
  }

  async getRoute(id) {
    return ApiService.request(`/routes/${id}`);
  }

  async createRoute(routeData) {
    return ApiService.request('/routes', {
      method: 'POST',
      body: JSON.stringify(routeData),
    });
  }

  async updateRoute(id, routeData) {
    return ApiService.request(`/routes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(routeData),
    });
  }

  async deleteRoute(id) {
    return ApiService.request(`/routes/${id}`, {
      method: 'DELETE',
    });
  }
}

const routesService = new RoutesService();
export default routesService;