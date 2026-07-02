import { useEffect, useMemo, useState } from 'react';
import { RegistryServiceClient, type MicroApp } from '@plantx/kit-sdk-api/registry';
import { useKitContext, type MicroAppManifest, useKitPermission } from './index.js';

export interface UseMicroAppsOptions {
  applicationId?: string;
  includeOffline?: boolean;
}

export interface UseMicroAppsResult {
  microApps: MicroAppManifest[];
  loading: boolean;
  error: Error | null;
}

function toManifest(app: MicroApp): MicroAppManifest {
  return {
    name: app.name,
    route: app.route,
    bundleUrl: app.bundleUrl,
    menuLabelKey: app.menuLabelKey,
    requirePermission: app.requirePermission,
  };
}

export function useMicroApps(
  options: UseMicroAppsOptions = {},
  client?: RegistryServiceClient | null
): UseMicroAppsResult {
  const { applicationId, includeOffline } = options;
  const ctx = useKitContext();
  const registryClient = useMemo(
    () => client ?? (ctx.apiClient ? new RegistryServiceClient(ctx.apiClient) : null),
    [client, ctx.apiClient]
  );

  const [microApps, setMicroApps] = useState<MicroAppManifest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!registryClient) {
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    const promise = applicationId
      ? registryClient.getApplicationMicroApps({ applicationId })
      : registryClient.listMicroApps();

    promise
      .then((data) => {
        if (!cancelled) {
          let list = data.microApps ?? [];
          if (!includeOffline) {
            list = list.filter((m) => m.status === 'RESOURCE_STATUS_ONLINE');
          }
          setMicroApps(list.map(toManifest));
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

  return { microApps, loading, error };
}


