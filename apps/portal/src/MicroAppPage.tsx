import React, { useEffect, useRef } from 'react';
import { Spin } from 'antd';
import { loadMicroApp, type MicroApp } from 'qiankun';
import { useLocation } from 'react-router-dom';
import { useKitContext, type MicroAppManifest } from '@plantx/kit-sdk-kit';
import { useI18n } from './i18n';

type MicroAppPageProps =
  | { manifest: MicroAppManifest; name?: never; entry?: never }
  | { manifest?: never; name: string; entry: string };

export function MicroAppPage(props: MicroAppPageProps) {
  const manifest: MicroAppManifest = props.manifest
    ? props.manifest
    : {
        name: props.name,
        route: '',
        bundleUrl: props.entry.replace(/\/$/, '') + `/${props.name}.js`,
        menuLabelKey: '',
      };

  const containerRef = useRef<HTMLDivElement>(null);
  const microAppRef = useRef<MicroApp | null>(null);
  const ctx = useKitContext();
  const { t, locale } = useI18n();
  const location = useLocation();

  useEffect(() => {
    if (!containerRef.current) return;

    const global = window as any;
    if (!global.process) {
      global.process = { env: { NODE_ENV: 'production' } };
    }

    const microApp = loadMicroApp(
      {
        name: manifest.name,
        entry: { scripts: [manifest.bundleUrl], styles: [] },
        container: containerRef.current,
        props: {
          name: manifest.name,
          user: ctx.user,
          tenant: ctx.tenant,
          permissions: ctx.permissions,
          apiClient: ctx.apiClient,
          locale,
        },
      },
      { sandbox: false },
    );

    microAppRef.current = microApp;

    return () => {
      microAppRef.current = null;
      microApp.unmount().catch((err) => {
        // eslint-disable-next-line no-console
        console.error(`failed to unmount micro-app ${manifest.name}:`, err);
      });
    };
  }, [manifest.name, manifest.bundleUrl, ctx.user, ctx.tenant, ctx.permissions, ctx.apiClient, locale, location.pathname]);

  return (
    <div ref={containerRef} style={{ minHeight: 400 }}>
      <Spin tip={t('microapp.loading', { name: manifest.name })} />
    </div>
  );
}
