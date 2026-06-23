import { useEffect, useMemo, useState } from 'react';
import { useKitContext, type MicroAppManifest } from '@plantx/kit-sdk-kit';
import { RegistryServiceClient, type MicroApp } from '@plantx/kit-sdk-api/registry';

export interface UseMicroAppsOptions {
  applicationId?: string;
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

export function useMicroApps(options: UseMicroAppsOptions = {}): UseMicroAppsResult {
  const { applicationId } = options;
  const ctx = useKitContext();
  const client = useMemo(
    () => (ctx.apiClient ? new RegistryServiceClient(ctx.apiClient) : null),
    [ctx.apiClient]
  );
  const [microApps, setMicroApps] = useState<MicroAppManifest[]>([]);
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
      ? client.getApplicationMicroApps({ applicationId })
      : client.listMicroApps();

    promise
      .then((data) => {
        if (!cancelled) {
          setMicroApps(
            (data.microApps ?? []).filter((m) => m.status === 'ONLINE').map(toManifest)
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
  }, [client, applicationId]);

  return { microApps, loading, error };
}
