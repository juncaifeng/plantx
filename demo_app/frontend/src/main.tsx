import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider } from 'antd';
import { KitProvider } from '@plantx/kit-sdk-kit';
import { createClient } from '@plantx/kit-sdk-api';
import { DemoPage } from './DemoPage';

// Standalone development entry point.
// In production this micro-app is mounted by the PlantX portal via qiankun,
// which provides user, tenant, permissions, and apiClient as props.
const apiClient = createClient({
  baseURL: '/api',
  getToken: () => localStorage.getItem('plantx_token'),
  onUnauthorized: () => {
    localStorage.removeItem('plantx_token');
    window.location.href = '/login';
  },
});

const root = ReactDOM.createRoot(document.getElementById('root')!);
root.render(
  <ConfigProvider>
    <KitProvider
      value={{
        user: {
          id: 'demo-user',
          username: 'demo',
          displayName: 'Demo User',
          permissions: [
          'item:list',
          'item:create',
          'setting:list',
          'setting:create',
          'setting:update',
          'setting:delete',
          'setting:admin',
        ],
        },
        tenant: { id: 'demo-tenant', name: 'Demo Tenant' },
        permissions: [
          'item:list',
          'item:create',
          'setting:list',
          'setting:create',
          'setting:update',
          'setting:delete',
          'setting:admin',
        ],
        apiClient,
      }}
    >
      <DemoPage />
    </KitProvider>
  </ConfigProvider>
);
