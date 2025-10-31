import { useState, useEffect } from 'react';
import hostService from '../services/api/hostService';
import { Host } from '../types';

export const useHosts = () => {
  const [hosts, setHosts] = useState<Host[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchHosts = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await hostService.getAllHosts();
      setHosts(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch hosts');
      setHosts([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHosts();
  }, []);

  return {
    hosts,
    loading,
    error,
    refetch: fetchHosts,
  };
};