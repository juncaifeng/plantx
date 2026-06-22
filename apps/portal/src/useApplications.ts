import { useEffect, useMemo, useState } from 'react';
import { useKitContext } from '@plantx/kit-sdk-kit';
import {
  RegistryServiceClient,
  type Application,
  type ApplicationStatus,
} from '@plantx/kit-sdk-api/registry';

export interface UseApplicationsResult {
  applications: Application[];
  activeApplications: Application[];
  loading: boolean;
  error: Error | null;
}

export function isApplicationActive(app: Application): boolean {
  return (app.status as ApplicationStatus) === 'APPLICATION_STATUS_ACTIVE';
}

export function useApplications(): UseApplicationsResult {
  const ctx = useKitContext();
  const client = useMemo(
    () => (ctx.apiClient ? new RegistryServiceClient(ctx.apiClient) : null),
    [ctx.apiClient]
  );

  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!client) {
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    client
      .listApplications()
      .then((data) => {
        if (!cancelled) {
          const apps = data.applications ?? [];
          setApplications(
            apps.slice().sort((a, b) => {
              const orderDiff = (a.sortOrder ?? 0) - (b.sortOrder ?? 0);
              return orderDiff !== 0 ? orderDiff : a.name.localeCompare(b.name);
            })
          );
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [client]);

  const activeApplications = useMemo(
    () => applications.filter(isApplicationActive),
    [applications]
  );

  return { applications, activeApplications, loading, error };
}
