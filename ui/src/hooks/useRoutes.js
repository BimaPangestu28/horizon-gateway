import { useState, useEffect } from 'react';
import RoutesService from '../services/RoutesService';
import { useToast } from '@/components/ui/use-toast';

export function useRoutes() {
  const [routes, setRoutes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { toast } = useToast();

  async function fetchRoutes() {
    setLoading(true);
    try {
      const response = await RoutesService.getRoutes();
      
      const routesArray = response.routes || [];
      setRoutes(routesArray);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch routes:', err);
      setError('Failed to fetch routes. Please try again.');
      toast({
        title: 'Error',
        description: 'Failed to fetch routes. Please try again.',
        variant: 'destructive'
      });
      setRoutes([]);
    } finally {
      setLoading(false);
    }
  }

  async function deleteRoute(routeName) {
    try {
      await RoutesService.deleteRoute(routeName);
      setRoutes(routes.filter(route => route.Name !== routeName));
      toast({
        title: 'Success',
        description: `Route "${routeName}" has been deleted.`
      });
      return true;
    } catch (err) {
      console.error('Failed to delete route:', err);
      toast({
        title: 'Error',
        description: 'Failed to delete route. Please try again.',
        variant: 'destructive'
      });
      return false;
    }
  }

  useEffect(() => {
    fetchRoutes();
  }, []);

  return {
    routes,
    loading,
    error,
    fetchRoutes,
    deleteRoute
  };
}