import ApiService from './ApiService';

class RoutesService {
  async getRoutes() {
    return ApiService.request('/routes');
  }

  async getRoute(name) {
    return ApiService.request(`/routes/${name}`);
  }

  async createRoute(routeData) {
    return ApiService.request('/routes', {
      method: 'POST',
      body: JSON.stringify(routeData),
    });
  }

  async updateRoute(name, routeData) {
    return ApiService.request(`/routes/${name}`, {
      method: 'PUT',
      body: JSON.stringify(routeData),
    });
  }

  async deleteRoute(name) {
    return ApiService.request(`/routes/${name}`, {
      method: 'DELETE',
    });
  }
}

const routesService = new RoutesService();
export default routesService;
