import { useEffect, useMemo, useState } from 'react';
import {
  RegistryServiceClient,
  type Application,
  type ApplicationStatus,
} from '@plantx/kit-sdk-api/registry';
import { useKitContext } from './index.js';

export interface UseApplicationsResult {
  applications: Application[];
  activeApplications: Application[];
  loading: boolean;
  error: Error | null;
}

export function isApplicationActive(app: Application): boolean {
  return (app.status as ApplicationStatus) === 'APPLICATION_STATUS_ACTIVE';
}

export function useApplications(client?: RegistryServiceClient | null): UseApplicationsResult {
  const ctx = useKitContext();
  const registryClient = useMemo(
    () => client ?? (ctx.apiClient ? new RegistryServiceClient(ctx.apiClient) : null),
    [client, ctx.apiClient]
  );

  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!registryClient) {
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    registryClient
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
  }, [registryClient]);

  const activeApplications = useMemo(
    () => applications.filter(isApplicationActive),
    [applications]
  );

  return { applications, activeApplications, loading, error };
}
