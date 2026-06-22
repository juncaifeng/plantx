import React, { useEffect, useRef } from 'react';
import { Spin } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { useI18n } from './i18n';

export function OrdersPage() {
  const containerRef = useRef<HTMLDivElement>(null);
  const loadedRef = useRef(false);
  const ctx = useKitContext();
  const { t, locale } = useI18n();

  useEffect(() => {
    if (!containerRef.current || loadedRef.current) return;
    loadedRef.current = true;

    const scriptUrl = '/apps/order-ui/order-ui.js?v=2';

    // Some micro-app bundles reference process.env at runtime
    const global = window as any;
    if (!global.process) {
      global.process = { env: { NODE_ENV: 'production' } };
    }

    const script = document.createElement('script');
    script.src = scriptUrl;
    script.async = true;

    const props = {
      container: containerRef.current,
      name: 'order-ui',
      user: ctx.user,
      tenant: ctx.tenant,
      permissions: ctx.permissions,
      apiClient: ctx.apiClient,
      locale,
    };

    script.onload = () => {
      const global = window as any;
      if (typeof global.bootstrap === 'function') {
        global.bootstrap().then(() => {
          global.mount(props);
        });
      } else if (typeof global.mount === 'function') {
        global.mount(props);
      }
    };

    script.onerror = () => {
      // eslint-disable-next-line no-console
      console.error(`failed to load micro-app script: ${scriptUrl}`);
    };

    document.head.appendChild(script);

    return () => {
      const global = window as any;
      if (typeof global.unmount === 'function') {
        global.unmount();
      }
      script.remove();
      loadedRef.current = false;
    };
  }, [ctx.user, ctx.tenant, ctx.permissions, ctx.apiClient, locale]);

  return (
    <div ref={containerRef} style={{ minHeight: 400 }}>
      <Spin tip={t('orders.loading')} />
    </div>
  );
}
