import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { KitProvider, type KitContextValue } from '@plantx/kit-sdk-kit';
import { createClient, type KitApiClient } from '@plantx/kit-sdk-api';
import { decodeJwt, TOKEN_KEY } from './auth';
import { HomePage } from './HomePage';
import { Layout } from './Layout';
import { LoginPage } from './LoginPage';
import { MicroAppPage } from './MicroAppPage';
import { useMicroApps } from './useMicroApps';

function buildApiClient(onUnauthorized?: () => void): KitApiClient {
  return createClient({
    baseURL: '/api',
    getToken: () => localStorage.getItem(TOKEN_KEY),
    onUnauthorized,
  });
}

function buildContext(token: string, onUnauthorized?: () => void): KitContextValue {
  const claims = decodeJwt(token);
  const permissions = claims?.permissions ?? [];
  const roles = claims?.roles ?? [];
  const username = claims?.preferred_username ?? claims?.username ?? claims?.sub ?? 'user';
  const displayName = claims?.name ?? username;

  return {
    user: {
      id: claims?.sub ?? 'unknown',
      username,
      displayName,
      email: claims?.email,
      roles,
      permissions,
    },
    tenant: {
      id: claims?.tenant_id ?? 'unknown',
      name: `Tenant ${claims?.tenant_id ?? 'unknown'}`,
    },
    permissions,
    apiClient: buildApiClient(onUnauthorized),
  };
}

interface AppRoutesProps {
  context: KitContextValue;
  onLogin: (token: string) => void;
  onLogout: () => void;
}

function AppRoutes({ context, onLogin, onLogout }: AppRoutesProps) {
  const { microApps } = useMicroApps();

  return (
    <BrowserRouter>
      <Routes>
        {!context.user ? (
          <Route path="*" element={<LoginPage onLogin={onLogin} />} />
        ) : (
          <Route path="/" element={<Layout onLogout={onLogout} />}>
            <Route index element={<HomePage />} />
            {microApps.map((app) => (
              <Route
                key={app.name}
                path={app.route.replace(/^\//, '')}
                element={<MicroAppPage manifest={app} />}
              />
            ))}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        )}
      </Routes>
    </BrowserRouter>
  );
}

export function App() {
  const [token, setToken] = useState<string | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    setToken(localStorage.getItem(TOKEN_KEY));
    setReady(true);
  }, []);

  const handleLogout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    window.location.href = '/';
  }, []);

  const context = useMemo<KitContextValue>(() => {
    if (!token) return {};
    return buildContext(token, handleLogout);
  }, [token, handleLogout]);

  const handleLogin = useCallback((newToken: string) => {
    localStorage.setItem(TOKEN_KEY, newToken);
    setToken(newToken);
  }, []);

  if (!ready) {
    return null;
  }

  return (
    <KitProvider value={context}>
      <AppRoutes context={context} onLogin={handleLogin} onLogout={handleLogout} />
    </KitProvider>
  );
}
