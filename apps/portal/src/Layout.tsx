import React, { useEffect, useMemo, useState } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { Button, Layout as AntLayout, Menu, Space, Spin, Typography } from 'antd';
import { MenuOutlined, AppstoreOutlined } from '@ant-design/icons';
import { useKitContext, type MicroAppManifest } from '@plantx/kit-sdk-kit';
import type { Menu as RegistryMenu, Application } from '@plantx/kit-sdk-api/registry';
import { useI18n, type I18nContextValue } from './i18n';
import { LocaleSwitch } from './components/LocaleSwitch';
import { ProductDrawer } from './components/ProductDrawer';
import { getStoredApplicationId, storeApplicationId } from './applicationStorage';
import { useApplications } from './useApplications';
import { useMicroApps } from './useMicroApps';
import { useMenus } from './useMenus';

interface LayoutProps {
  onLogout: () => void;
}

const { Header, Content, Sider } = AntLayout;

function menuToItem(menu: RegistryMenu, children: any[], t: I18nContextValue['t']): any {
  const item: any = {
    key: menu.route ?? menu.id,
    label: menu.route ? (
      <Link to={menu.route}>{t(menu.labelKey as any)}</Link>
    ) : (
      t(menu.labelKey as any)
    ),
  };
  if (children.length > 0) {
    item.children = children;
  }
  return item;
}

function buildRegistryMenuItems(
  menus: RegistryMenu[],
  isAdmin: boolean,
  allPermissions: Set<string>,
  t: I18nContextValue['t']
): any[] {
  const visible = menus.filter((menu) => {
    if (!menu.requirePermission) return true;
    return isAdmin || allPermissions.has(menu.requirePermission);
  });

  const visibleIds = new Set(visible.map((m) => m.id));
  const childrenByParent = new Map<string, RegistryMenu[]>();

  visible.forEach((menu) => {
    const parentId = menu.parentId && visibleIds.has(menu.parentId) ? menu.parentId : undefined;
    if (parentId) {
      const list = childrenByParent.get(parentId) ?? [];
      list.push(menu);
      childrenByParent.set(parentId, list);
    }
  });

  function buildChildren(parentId: string): any[] {
    const children = (childrenByParent.get(parentId) ?? []).sort(
      (a, b) => a.sortOrder - b.sortOrder
    );
    return children.map((child) => menuToItem(child, buildChildren(child.id), t));
  }

  const topLevel = visible
    .filter((menu) => !menu.parentId || !visibleIds.has(menu.parentId))
    .sort((a, b) => a.sortOrder - b.sortOrder);

  const hasExplicitAdminParent = topLevel.some((menu) => menu.route === '/admin');

  const adminGroupMenus: RegistryMenu[] = [];
  const regularTopLevel: any[] = [];

  topLevel.forEach((menu) => {
    const children = buildChildren(menu.id);
    if (!hasExplicitAdminParent && menu.route?.startsWith('/admin/')) {
      adminGroupMenus.push(menu);
    } else {
      regularTopLevel.push(menuToItem(menu, children, t));
    }
  });

  if (!hasExplicitAdminParent && adminGroupMenus.length > 0) {
    regularTopLevel.push({
      key: '/admin',
      label: t('nav.admin'),
      children: adminGroupMenus
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map((menu) => menuToItem(menu, buildChildren(menu.id), t)),
    });
  }

  return regularTopLevel;
}

export function Layout({ onLogout }: LayoutProps) {
  const ctx = useKitContext();
  const location = useLocation();
  const { t } = useI18n();

  const [selectedApplicationId, setSelectedApplicationId] = useState<string | null>(
    getStoredApplicationId
  );
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { activeApplications, loading: applicationsLoading, error: applicationsError } = useApplications();

  const selectedApplication = useMemo<Application | undefined>(() => {
    if (!selectedApplicationId) return undefined;
    return activeApplications.find((app) => app.id === selectedApplicationId);
  }, [activeApplications, selectedApplicationId]);

  const handleSelectApplication = (application: Application) => {
    setSelectedApplicationId(application.id || null);
    storeApplicationId(application.id || null);
  };

  const isAdmin =
    (ctx.user?.roles?.includes('admin') ?? false) ||
    (ctx.permissions?.includes('platform:admin') ?? false);

  const allPermissions = new Set([
    ...(ctx.permissions ?? []),
    ...(ctx.user?.permissions ?? []),
  ]);

  const { microApps, loading: microAppsLoading, error: microAppsError } = useMicroApps({
    applicationId: selectedApplication?.id,
  });
  const { menus, loading: menusLoading, error: menusError } = useMenus({
    applicationId: selectedApplication?.id,
  });

  useEffect(() => {
    if (microAppsError) {
      // eslint-disable-next-line no-console
      console.error('Failed to load micro-apps:', microAppsError);
    }
  }, [microAppsError]);

  useEffect(() => {
    if (menusError) {
      // eslint-disable-next-line no-console
      console.error('Failed to load menus:', menusError);
    }
  }, [menusError]);

  const shouldUseMenus = menus.length > 0 && !menusError;

  const menuItems = useMemo(() => {
    const items: any[] = [
      { key: '/', label: <Link to="/">{t('nav.home')}</Link> },
    ];

    if (shouldUseMenus) {
      items.push(...buildRegistryMenuItems(menus, isAdmin, allPermissions, t));
      return items;
    }

    const adminApps: MicroAppManifest[] = [];

    microApps.forEach((app) => {
      if (!isAdmin && app.requirePermission && !allPermissions.has(app.requirePermission)) {
        return;
      }
      if (app.route.startsWith('/admin/')) {
        adminApps.push(app);
      } else {
        items.push({
          key: app.route,
          label: <Link to={app.route}>{t(app.menuLabelKey as any)}</Link>,
        });
      }
    });

    if (adminApps.length > 0) {
      items.push({
        key: '/admin',
        label: t('nav.admin'),
        children: adminApps.map((app) => ({
          key: app.route,
          label: <Link to={app.route}>{t(app.menuLabelKey as any)}</Link>,
        })),
      });
    }

    return items;
  }, [shouldUseMenus, menus, microApps, isAdmin, allPermissions, t]);

  const hasAdminSubmenu = menuItems.some((item) => item.key === '/admin');
  const selectedKeys = [location.pathname];
  const openKeys = hasAdminSubmenu && location.pathname.startsWith('/admin') ? ['/admin'] : [];

  const loading = menusLoading || (!shouldUseMenus && microAppsLoading);

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider theme="dark" width={240}>
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 16px',
            color: '#fff',
            fontSize: 16,
            fontWeight: 600,
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
          }}
        >
          <Space>
            <AppstoreOutlined />
            <span>{selectedApplication?.name ?? t('product.all')}</span>
          </Space>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          defaultOpenKeys={openKeys}
          items={menuItems}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <AntLayout>
        <Header
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 24px',
            background: '#fff',
          }}
        >
          <Space align="center">
            <Button
              type="text"
              icon={<MenuOutlined />}
              onClick={() => setDrawerOpen(true)}
              title={t('product.switcher')}
            />
            <Typography.Title level={4} style={{ margin: 0 }}>
              PlantX Portal
            </Typography.Title>
          </Space>
          <Space align="center">
            {loading && <Spin size="small" />}
            <Typography.Text>
              {ctx.user?.displayName ?? ctx.user?.username}
            </Typography.Text>
            <LocaleSwitch />
            <Button type="default" onClick={onLogout}>
              {t('header.logout')}
            </Button>
          </Space>
        </Header>
        <Content style={{ padding: 24, background: '#f0f2f5' }}>
          <Outlet />
        </Content>
      </AntLayout>
      <ProductDrawer
        applications={activeApplications}
        loading={applicationsLoading}
        error={applicationsError}
        value={selectedApplication}
        onChange={handleSelectApplication}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      />
    </AntLayout>
  );
}
