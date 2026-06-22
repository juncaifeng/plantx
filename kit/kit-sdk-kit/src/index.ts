import React, { createContext, useContext } from 'react';
import type { KitApiClient } from '@plantx/kit-sdk-api';


export interface KitUser {
  id: string;
  username?: string;
  displayName?: string;
  email?: string;
  roles?: string[];
  permissions?: string[];
}

export interface KitTenant {
  id: string;
  name?: string;
}

export interface KitContextValue {
  user?: KitUser;
  tenant?: KitTenant;
  permissions?: string[];
  apiClient?: KitApiClient;
}

export const KitContext = createContext<KitContextValue>({});

export interface KitProviderProps {
  value: KitContextValue;
  children?: React.ReactNode;
}

export function KitProvider({ value, children }: KitProviderProps) {
  return React.createElement(KitContext.Provider, { value }, children);
}

export function useKitUser(): KitUser | undefined {
  return useContext(KitContext).user;
}

export function useKitTenant(): KitTenant | undefined {
  return useContext(KitContext).tenant;
}

export function useKitPermission(permission: string): boolean {
  const ctx = useContext(KitContext);
  return (ctx.permissions ?? []).includes(permission) || (ctx.user?.permissions ?? []).includes(permission);
}

export function useKitContext(): KitContextValue {
  return useContext(KitContext);
}

export interface MicroAppManifest {
  name: string;
  route: string;
  bundleUrl: string;
  menuLabelKey: string;
  requirePermission?: string;
}

export function useMicroApps(manifests: MicroAppManifest[]): MicroAppManifest[] {
  // Static manifest for now; reserved for future server-side discovery.
  return manifests;
}
