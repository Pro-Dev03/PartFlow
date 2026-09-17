import { useEffect } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { settingsApi } from '../services/api/endpoints';
import { RegionalProfile } from '../types/regional';
import { getDeviceTimezone, setRegionalProfile } from '../utils/store-time';

export function RegionalProfileLoader() {
  const { data } = useQuery({
    queryKey: ['settings', 'regional'],
    queryFn: () => settingsApi.getRegionalSettings(),
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  const initializeMutation = useMutation({
    mutationFn: (timezone: string) => settingsApi.initializeRegionalSettings(timezone),
    onSuccess: (response) => {
      const profile = response.data?.profile as RegionalProfile | undefined;
      if (profile) setRegionalProfile(profile);
    },
  });

  useEffect(() => {
    const profile = data?.data?.profile as RegionalProfile | undefined;
    if (!profile) return;
    setRegionalProfile(profile);
    if (data?.data?.timezone_persisted === false && !initializeMutation.isPending) {
      initializeMutation.mutate(getDeviceTimezone());
    }
  }, [data]);

  return null;
}
