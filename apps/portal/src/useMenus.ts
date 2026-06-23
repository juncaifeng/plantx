import { useEffect, useMemo, useState } from 'react';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { RegistryServiceClient, type Menu } from '@plantx/kit-sdk-api/registry';

export interface UseMenusOptions {
  applicationId?: string;
}

export interface UseMenusResult {
  menus: Menu[];
  loading: boolean;
  error: Error | null;
}

export function useMenus(options: UseMenusOptions = {}): UseMenusResult {
  const { applicationId } = options;
  const ctx = useKitContext();
  const client = useMemo(
    () => (ctx.apiClient ? new RegistryServiceClient(ctx.apiClient) : null),
    [ctx.apiClient]
  );
  const [menus, setMenus] = useState<Menu[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!client) {
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    const promise = applicationId
      ? client.getApplicationMenus({ applicationId })
      : client.listMenus();

    promise
      .then((data) => {
        if (!cancelled) {
          setMenus((data.menus ?? []).filter((m) => m.status === 'RESOURCE_STATUS_ONLINE'));
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
  }, [client, applicationId]);

  return { menus, loading, error };
}
