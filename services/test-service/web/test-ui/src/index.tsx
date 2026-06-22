import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider, Layout as AntLayout, Typography } from 'antd';
import { KitProvider, type KitContextValue } from '@plantx/kit-sdk-kit';
import type { KitApiClient } from '@plantx/kit-sdk-api';
import { TestPage } from './TestPage';

interface QiankunProps {
  container?: HTMLElement;
  name: string;
  user?: KitContextValue['user'];
  tenant?: KitContextValue['tenant'];
  permissions?: string[];
  apiClient?: KitApiClient;
  locale?: string;
}

const { Header, Content } = AntLayout;

function render(props: QiankunProps) {
  const container = props.container ?? document.getElementById('root') ?? document.body;
  const context: KitContextValue = {
    user: props.user,
    tenant: props.tenant,
    permissions: props.permissions,
    apiClient: props.apiClient,
  };
  const root = ReactDOM.createRoot(container);
  root.render(
    <ConfigProvider>
      <KitProvider value={context}>
        <AntLayout style={{ minHeight: '100%' }}>
          <Header style={{ display: 'flex', alignItems: 'center' }}>
            <Typography.Title level={5} style={{ color: '#fff', margin: 0 }}>
              Test Service
            </Typography.Title>
          </Header>
          <Content style={{ padding: 24 }}>
            <TestPage />
          </Content>
        </AntLayout>
      </KitProvider>
    </ConfigProvider>
  );
  return root;
}

export async function bootstrap() {
  // eslint-disable-next-line no-console
  console.log('test-ui bootstrapped');
}

let rootInstance: ReturnType<typeof render> | null = null;

export async function mount(props: QiankunProps) {
  rootInstance = render(props);
}

export async function unmount() {
  rootInstance?.unmount();
  rootInstance = null;
}

if (typeof window !== 'undefined') {
  (window as any).bootstrap = bootstrap;
  (window as any).mount = mount;
  (window as any).unmount = unmount;
  (window as any)['test-ui'] = { bootstrap, mount, unmount };
}
