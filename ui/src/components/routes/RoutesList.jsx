import React, { useState } from 'react';
import { useRoutes } from '../../hooks/useRoutes';
import RoutesHeader from './RoutesHeader';
import RoutesSearchFilter from './RoutesSearchFilter';
import RoutesTable from './RoutesTable';
import DeleteRouteModal from './DeleteRouteModal';
import LoadingSpinner from '../common/LoadingSpinner';

export default function RoutesList() {
  const { routes, loading, error, fetchRoutes, deleteRoute } = useRoutes();
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [deleteRouteData, setDeleteRouteData] = useState(null);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

  const handleDeleteConfirm = (route) => {
    setDeleteRouteData(route);
    setIsDeleteModalOpen(true);
  };

  const handleDelete = async () => {
    if (!deleteRouteData) return;
    
    const success = await deleteRoute(deleteRouteData.name);
    if (success) {
      setIsDeleteModalOpen(false);
      setDeleteRouteData(null);
    }
  };

  const filteredRoutes = routes.filter(route => {
    const matchesSearch = route.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                          route.listen_path.toLowerCase().includes(searchTerm.toLowerCase());
    
    if (filterType === 'all') return matchesSearch;
    if (filterType === 'auth') return matchesSearch && route.auth?.enabled;
    if (filterType === 'cache') return matchesSearch && route.caching?.enabled;
    if (filterType === 'ratelimit') return matchesSearch && route.rate_limiting?.enabled;
    if (filterType === 'circuitbreaker') return matchesSearch && route.circuit_breaker?.enabled;
    
    return matchesSearch;
  });

  return (
    <div>
      <RoutesHeader />
      
      <RoutesSearchFilter 
        searchTerm={searchTerm}
        setSearchTerm={setSearchTerm}
        filterType={filterType}
        setFilterType={setFilterType}
        onRefresh={fetchRoutes}
      />
      
      {error && (
        <div className="mb-4 bg-red-50 p-4 rounded-md">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">{error}</h3>
            </div>
          </div>
        </div>
      )}
      
      {loading ? (
        <LoadingSpinner />
      ) : (
        <RoutesTable 
          routes={filteredRoutes} 
          onDelete={handleDeleteConfirm} 
        />
      )}

      <DeleteRouteModal 
        isOpen={isDeleteModalOpen}
        route={deleteRouteData}
        onDelete={handleDelete}
        onClose={() => setIsDeleteModalOpen(false)}
      />
    </div>
  );
}