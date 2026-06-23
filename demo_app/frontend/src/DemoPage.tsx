import React, { useState } from 'react';
import { Menu, Typography } from 'antd';
import {
  useKitPermission,
  useKitUser,
} from '@plantx/kit-sdk-kit';
import { KitLayout } from '@plantx/kit-ui';
import { ConfigManagement } from './ConfigManagement';
import { DemoHome } from './DemoHome';
import { SystemSettings } from './SystemSettings';

type PageKey = 'home' | 'config' | 'system';

export function DemoPage() {
  const user = useKitUser();
  const [activePage, setActivePage] = useState<PageKey>('home');

  const canListItems = useKitPermission('item:list');
  const canListSettings = useKitPermission('setting:list');
  const canAdminSettings = useKitPermission('setting:admin');

  const menuItems = [
    ...(canListItems ? [{ key: 'home', label: 'Home' }] : []),
    ...(canListSettings ? [{ key: 'config', label: 'Configuration' }] : []),
    ...(canAdminSettings ? [{ key: 'system', label: 'System Settings' }] : []),
  ];

  return (
    <KitLayout title="Demo App" user={{ displayName: user?.displayName ?? 'User' }}>
      <Typography.Title level={4}>Demo Micro-App</Typography.Title>

      {menuItems.length > 0 && (
        <Menu
          mode="horizontal"
          selectedKeys={[activePage]}
          items={menuItems}
          onClick={({ key }) => setActivePage(key as PageKey)}
          style={{ marginBottom: 24 }}
        />
      )}

      {activePage === 'home' && <DemoHome />}
      {activePage === 'config' && <ConfigManagement />}
      {activePage === 'system' && <SystemSettings />}
    </KitLayout>
  );
}
