import React, { useState, useEffect } from 'react';
import { Plus, RefreshCw, Search } from 'react-feather';
import { useToast } from '@/components/ui/use-toast';
import RateLimitTable from '@/components/rate-limiting/RateLimitTable';
import RateLimitTableSkeleton from '@/components/rate-limiting/RateLimitTableSkeleton';
import CreateRateLimitModal from '@/components/rate-limiting/CreateRateLimitModal';
import EditRateLimitModal from '@/components/rate-limiting/EditRateLimitModal';
import DeleteRateLimitModal from '@/components/rate-limiting/DeleteRateLimitModal';
import RateLimitService from '@/services/RateLimitService';
import RoutesService from '@/services/RoutesService';

export default function RateLimiting() {
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [rateLimits, setRateLimits] = useState([]);
  const [routes, setRoutes] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deleteLimit, setDeleteLimit] = useState(null);
  const [editLimit, setEditLimit] = useState(null);
  
  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      await Promise.all([
        fetchRateLimitData(),
        fetchRoutes()
      ]);
    } catch (err) {
      console.error('Failed to fetch data:', err);
      setError('Failed to fetch rate limiting data');
      toast({
        title: "Error fetching data",
        description: "Could not retrieve rate limiting configuration",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchRateLimitData = async () => {
    try {
      const data = await RateLimitService.getRateLimits();
      // Ensure data is always an array
      setRateLimits(Array.isArray(data) ? data : []);
      setError(null);
    } catch (err) {
      console.error('Error fetching rate limits:', err);
      setRateLimits([]);
      throw err;
    }
  };

  const fetchRoutes = async () => {
    try {
      const routesData = await RoutesService.getRoutes();
      setRoutes(routesData);
    } catch (err) {
      console.error('Error fetching routes:', err);
      throw err;
    }
  };

  const handleCreateRateLimit = async (newRateLimit) => {
    try {
      const createdLimit = await RateLimitService.createRateLimit(newRateLimit);
      setRateLimits([...rateLimits, createdLimit]);
      setShowCreateModal(false);
      toast({
        title: "Rate limit created",
        description: "New rate limit rule has been created successfully",
      });
    } catch (err) {
      console.error('Error creating rate limit:', err);
      toast({
        title: "Failed to create rate limit",
        description: err.message || "An error occurred",
        variant: "destructive",
      });
    }
  };

  const handleUpdateRateLimit = async (updatedLimit) => {
    try {
      await RateLimitService.updateRateLimit(updatedLimit.id, updatedLimit);
      const updatedLimits = rateLimits.map(limit => 
        limit.id === updatedLimit.id ? updatedLimit : limit
      );
      setRateLimits(updatedLimits);
      setShowEditModal(false);
      setEditLimit(null);
      toast({
        title: "Rate limit updated",
        description: "Rate limit settings have been updated successfully",
      });
    } catch (err) {
      console.error('Error updating rate limit:', err);
      toast({
        title: "Failed to update rate limit",
        description: err.message || "An error occurred",
        variant: "destructive",
      });
    }
  };

  const handleDeleteRateLimit = async () => {
    if (!deleteLimit) return;
    
    try {
      await RateLimitService.deleteRateLimit(deleteLimit.id);
      setRateLimits(rateLimits.filter(limit => limit.id !== deleteLimit.id));
      setShowDeleteModal(false);
      setDeleteLimit(null);
      toast({
        title: "Rate limit deleted",
        description: "The rate limit rule has been removed",
      });
    } catch (err) {
      console.error('Error deleting rate limit:', err);
      toast({
        title: "Failed to delete rate limit",
        description: err.message || "An error occurred",
        variant: "destructive",
      });
    }
  };

  const handleEditClick = (limit) => {
    setEditLimit(limit);
    setShowEditModal(true);
  };

  const handleDeleteClick = (limit) => {
    setDeleteLimit(limit);
    setShowDeleteModal(true);
  };

  const filterRateLimits = (item) => {
    if (!searchTerm) return true;
    
    const searchLower = searchTerm.toLowerCase();
    
    return (
      item.route.toLowerCase().includes(searchLower) ||
      item.type.toLowerCase().includes(searchLower) ||
      item.key.toLowerCase().includes(searchLower)
    );
  };

  return (
    <div>
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Rate Limiting</h1>
          <p className="mt-2 text-sm text-gray-700">
            Configure rate limits to control the number of requests clients can make to your API.
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            type="button"
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
          >
            <Plus className="mr-2 h-4 w-4" />
            New Rate Limit
          </button>
        </div>
      </div>

      {/* Search and refresh bar */}
      <div className="mt-6 mb-6 flex">
        <div className="relative flex-grow">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search className="h-5 w-5 text-gray-400" />
          </div>
          <input
            type="text"
            className="focus:ring-indigo-500 focus:border-indigo-500 block w-full pl-10 sm:text-sm border-gray-300 rounded-md"
            placeholder="Search rate limits..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <button
          type="button"
          onClick={fetchData}
          className="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </button>
      </div>

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
        <RateLimitTableSkeleton />
      ) : (
        <RateLimitTable 
          rateLimits={rateLimits.filter(filterRateLimits)} 
          onEdit={handleEditClick} 
          onDelete={handleDeleteClick} 
        />
      )}

      {showCreateModal && (
        <CreateRateLimitModal
          routes={routes}
          onSubmit={handleCreateRateLimit}
          onCancel={() => setShowCreateModal(false)}
        />
      )}

      {showEditModal && editLimit && (
        <EditRateLimitModal
          rateLimit={editLimit}
          routes={routes}
          onSubmit={handleUpdateRateLimit}
          onCancel={() => {
            setShowEditModal(false);
            setEditLimit(null);
          }}
        />
      )}

      {showDeleteModal && deleteLimit && (
        <DeleteRateLimitModal
          rateLimit={deleteLimit}
          onDelete={handleDeleteRateLimit}
          onCancel={() => {
            setShowDeleteModal(false);
            setDeleteLimit(null);
          }}
        />
      )}
    </div>
  );
}