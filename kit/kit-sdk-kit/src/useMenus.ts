import { useEffect, useMemo, useState } from 'react';
import { RegistryServiceClient, type Menu } from '@plantx/kit-sdk-api/registry';
import { useKitContext } from './index.js';

export interface UseMenusOptions {
  applicationId?: string;
  includeOffline?: boolean;
}

export interface UseMenusResult {
  menus: Menu[];
  loading: boolean;
  error: Error | null;
}

export function useMenus(
  options: UseMenusOptions = {},
  client?: RegistryServiceClient | null
): UseMenusResult {
  const { applicationId, includeOffline } = options;
  const ctx = useKitContext();
  const registryClient = useMemo(
    () => client ?? (ctx.apiClient ? new RegistryServiceClient(ctx.apiClient) : null),
    [client, ctx.apiClient]
  );

  const [menus, setMenus] = useState<Menu[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!registryClient) {
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    const promise = applicationId
      ? registryClient.getApplicationMenus({ applicationId })
      : registryClient.listMenus();

    promise
      .then((data) => {
        if (!cancelled) {
          let list = data.menus ?? [];
          if (!includeOffline) {
            list = list.filter((m) => m.status === 'RESOURCE_STATUS_ONLINE');
          }
          setMenus(list);
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
  }, [registryClient, applicationId, includeOffline]);

  return { menus, loading, error };
}
