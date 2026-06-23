import { useMemo } from 'react';
import { RegistryServiceClient } from '@plantx/kit-sdk-api/registry';
import { useKitContext } from './index.js';

export function useRegistryClient(): RegistryServiceClient | null {
  const ctx = useKitContext();
  return useMemo(
    () => (ctx.apiClient ? new RegistryServiceClient(ctx.apiClient) : null),
    [ctx.apiClient]
  );
}
