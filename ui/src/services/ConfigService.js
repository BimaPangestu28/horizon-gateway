import ApiService from './ApiService';

class ConfigService {
  async getConfig() {
    return ApiService.request('/config');
  }

  async updateConfig(configData) {
    return ApiService.request('/config', {
      method: 'PUT',
      body: JSON.stringify(configData),
    });
  }

  async getConfigBackups() {
    return ApiService.request('/config/backups');
  }

  async createConfigBackup() {
    return ApiService.request('/config/backups', {
      method: 'POST',
    });
  }

  async restoreConfigBackup(backupId) {
    return ApiService.request(`/config/backups/${backupId}/restore`, {
      method: 'POST',
    });
  }
}

const configService = new ConfigService();
export default configService;