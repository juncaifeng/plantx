import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Layout,
  Menu,
  Modal,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import {
  AppstoreOutlined,
  CloudServerOutlined,
  FileProtectOutlined,
  FilterOutlined,
  MenuOutlined,
  MobileOutlined,
  SafetyOutlined,
  TagOutlined,
  TeamOutlined,
  ToolOutlined,
} from '@ant-design/icons';
import { useKitContext } from '@plantx/kit-sdk-kit';
import {
  RegistryServiceClient,
  type Service,
  type MicroApp,
  type Menu,
  type Route,
} from '@plantx/kit-sdk-api/registry';
import { ApplicationsTab } from './ApplicationsTab';
import { AttributesTab } from './AttributesTab';
import { ConditionsTab } from './ConditionsTab';
import { PermissionsTab } from './PermissionsTab';
import { PoliciesTab } from './PoliciesTab';
import { RolesTab } from './RolesTab';
import { RoutePoliciesTab } from './RoutePoliciesTab';

const { Sider, Content } = Layout;

type Section =
  | 'applications'
  | 'services'
  | 'menus'
  | 'micro-apps'
  | 'permissions'
  | 'roles'
  | 'route-policies'
  | 'attributes'
  | 'conditions'
  | 'policies';

const menuItems = [
  { key: 'applications', label: 'Applications', icon: <AppstoreOutlined />, path: '/admin/registry/applications' },
  { key: 'services', label: 'Services', icon: <CloudServerOutlined />, path: '/admin/registry/services' },
  { key: 'menus', label: 'Menus', icon: <MenuOutlined />, path: '/admin/registry/menus' },
  { key: 'micro-apps', label: 'Micro Apps', icon: <MobileOutlined />, path: '/admin/registry/micro-apps' },
  { key: 'permissions', label: 'Permissions', icon: <SafetyOutlined />, path: '/admin/registry/permissions' },
  { key: 'roles', label: 'Roles', icon: <TeamOutlined />, path: '/admin/registry/roles' },
  { key: 'route-policies', label: 'Route Policies', icon: <ToolOutlined />, path: '/admin/registry/route-policies' },
  { key: 'attributes', label: 'Attributes', icon: <TagOutlined />, path: '/admin/registry/attributes' },
  { key: 'conditions', label: 'Conditions', icon: <FilterOutlined />, path: '/admin/registry/conditions' },
  { key: 'policies', label: 'Policies', icon: <FileProtectOutlined />, path: '/admin/registry/policies' },
];

export function RegistryAdminPage() {
  const { section } = useParams<{ section: string }>();
  const activeSection: Section = (section as Section) ?? 'applications';
  const navigate = useNavigate();

  const { apiClient } = useKitContext();
  const registryClient = useMemo(
    () => (apiClient ? new RegistryServiceClient(apiClient) : null),
    [apiClient]
  );

  const [services, setServices] = useState<Service[]>([]);
  const [microApps, setMicroApps] = useState<MicroApp[]>([]);
  const [menus, setMenus] = useState<Menu[]>([]);
  const [loadingServices, setLoadingServices] = useState(false);
  const [loadingMicroApps, setLoadingMicroApps] = useState(false);
  const [loadingMenus, setLoadingMenus] = useState(false);

  const [selectedRoutes, setSelectedRoutes] = useState<Route[]>([]);
  const [selectedServiceName, setSelectedServiceName] = useState('');
  const [isRoutesModalOpen, setIsRoutesModalOpen] = useState(false);

  const [menuForm] = Form.useForm();
  const [microAppForm] = Form.useForm();
  const [isMenuModalOpen, setIsMenuModalOpen] = useState(false);
  const [isMicroAppModalOpen, setIsMicroAppModalOpen] = useState(false);
  const [editingMenu, setEditingMenu] = useState<Menu | null>(null);
  const [editingMicroApp, setEditingMicroApp] = useState<MicroApp | null>(null);
  const [saving, setSaving] = useState(false);

  const listServices = async () => {
    if (!registryClient) return;
    setLoadingServices(true);
    try {
      const res = await registryClient.listServices();
      setServices(res.services);
    } catch (e) {
      message.error('Failed to load services');
      // eslint-disable-next-line no-console
      console.error('failed to list services', e);
    } finally {
      setLoadingServices(false);
    }
  };

  const listMicroApps = async () => {
    if (!registryClient) return;
    setLoadingMicroApps(true);
    try {
      const res = await registryClient.listMicroApps();
      setMicroApps(res.microApps);
    } catch (e) {
      message.error('Failed to load micro-apps');
      // eslint-disable-next-line no-console
      console.error('failed to list micro-apps', e);
    } finally {
      setLoadingMicroApps(false);
    }
  };

  const listMenus = async () => {
    if (!registryClient) return;
    setLoadingMenus(true);
    try {
      const res = await registryClient.listMenus();
      setMenus(res.menus);
    } catch (e) {
      message.error('Failed to load menus');
      // eslint-disable-next-line no-console
      console.error('failed to load menus', e);
    } finally {
      setLoadingMenus(false);
    }
  };

  const deregisterService = async (svc: Service) => {
    if (!registryClient) return;
    try {
      await registryClient.deregisterService({ id: svc.id });
      setServices((prev) => prev.filter((s) => s.id !== svc.id));
      message.success('Service deregistered');
    } catch (e) {
      message.error('Failed to deregister service');
      // eslint-disable-next-line no-console
      console.error('failed to deregister service', e);
    }
  };

  const showRoutes = (svc: Service) => {
    const routes = svc.routes?.length ? svc.routes : [{ path: svc.restPrefix, method: '*' }];
    setSelectedRoutes(routes);
    setSelectedServiceName(svc.name);
    setIsRoutesModalOpen(true);
  };

  const openCreateMenu = () => {
    setEditingMenu(null);
    menuForm.resetFields();
    setIsMenuModalOpen(true);
  };

  const openEditMenu = (menu: Menu) => {
    setEditingMenu(menu);
    menuForm.setFieldsValue({
      labelKey: menu.labelKey,
      route: menu.route,
      icon: menu.icon,
      parentId: menu.parentId,
      sortOrder: menu.sortOrder,
      microAppName: menu.microAppName,
      requirePermission: menu.requirePermission,
    });
    setIsMenuModalOpen(true);
  };

  const saveMenu = async (values: Record<string, unknown>) => {
    if (!registryClient) return;
    setSaving(true);
    try {
      if (editingMenu) {
        const updated = await registryClient.updateMenu({
          id: editingMenu.id,
          labelKey: values.labelKey as string,
          route: values.route as string | undefined,
          icon: values.icon as string | undefined,
          parentId: values.parentId as string | undefined,
          sortOrder: values.sortOrder as number | undefined,
          microAppName: values.microAppName as string | undefined,
          requirePermission: values.requirePermission as string | undefined,
        });
        setMenus((prev) => prev.map((m) => (m.id === updated.id ? updated : m)));
        message.success('Menu updated');
      } else {
        const created = await registryClient.createMenu({
          labelKey: values.labelKey as string,
          route: values.route as string | undefined,
          icon: values.icon as string | undefined,
          parentId: values.parentId as string | undefined,
          sortOrder: values.sortOrder as number | undefined,
          microAppName: values.microAppName as string | undefined,
          requirePermission: values.requirePermission as string | undefined,
        });
        setMenus((prev) => [...prev, created]);
        message.success('Menu created');
      }
      menuForm.resetFields();
      setIsMenuModalOpen(false);
      setEditingMenu(null);
    } catch (e) {
      message.error(editingMenu ? 'Failed to update menu' : 'Failed to create menu');
      // eslint-disable-next-line no-console
      console.error('failed to save menu', e);
    } finally {
      setSaving(false);
    }
  };

  const deleteMenu = async (menu: Menu) => {
    if (!registryClient) return;
    try {
      await registryClient.deleteMenu({ id: menu.id });
      setMenus((prev) => prev.filter((m) => m.id !== menu.id));
      message.success('Menu deleted');
    } catch (e) {
      message.error('Failed to delete menu');
      // eslint-disable-next-line no-console
      console.error('failed to delete menu', e);
    }
  };

  const reorderMenus = async (items: Menu[]) => {
    if (!registryClient) return;
    try {
      const res = await registryClient.reorderMenus({
        items: items.map((m, idx) => ({ id: m.id, sortOrder: idx + 1 })),
      });
      setMenus(res.menus);
      message.success('Menus reordered');
    } catch (e) {
      message.error('Failed to reorder menus');
      // eslint-disable-next-line no-console
      console.error('failed to reorder menus', e);
    }
  };

  const moveMenu = (index: number, direction: -1 | 1) => {
    const nextIndex = index + direction;
    if (nextIndex < 0 || nextIndex >= menus.length) return;
    const next = [...menus];
    const [removed] = next.splice(index, 1);
    next.splice(nextIndex, 0, removed);
    setMenus(next);
    void reorderMenus(next);
  };

  const openEditMicroApp = (app: MicroApp) => {
    setEditingMicroApp(app);
    microAppForm.setFieldsValue({
      name: app.name,
      route: app.route,
      bundleUrl: app.bundleUrl,
      menuLabelKey: app.menuLabelKey,
      requirePermission: app.requirePermission,
    });
    setIsMicroAppModalOpen(true);
  };

  const saveMicroApp = async (values: Record<string, unknown>) => {
    if (!registryClient || !editingMicroApp) return;
    setSaving(true);
    try {
      const updated = await registryClient.updateMicroApp({
        name: editingMicroApp.name,
        route: values.route as string,
        bundleUrl: values.bundleUrl as string,
        menuLabelKey: values.menuLabelKey as string,
        requirePermission: values.requirePermission as string,
      });
      setMicroApps((prev) => prev.map((a) => (a.name === updated.name ? updated : a)));
      message.success('Micro-app updated');
      microAppForm.resetFields();
      setIsMicroAppModalOpen(false);
      setEditingMicroApp(null);
    } catch (e) {
      message.error('Failed to update micro-app');
      // eslint-disable-next-line no-console
      console.error('failed to update micro-app', e);
    } finally {
      setSaving(false);
    }
  };

  const deleteMicroApp = async (app: MicroApp) => {
    if (!registryClient) return;
    try {
      await registryClient.deleteMicroApp({ name: app.name });
      setMicroApps((prev) => prev.filter((a) => a.name !== app.name));
      message.success('Micro-app deleted');
    } catch (e) {
      message.error('Failed to delete micro-app');
      // eslint-disable-next-line no-console
      console.error('failed to delete micro-app', e);
    }
  };

  useEffect(() => {
    listServices();
    listMicroApps();
    listMenus();
  }, [registryClient]);

  const serviceColumns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'gRPC Host', dataIndex: 'grpcHost', key: 'grpcHost' },
    { title: 'REST Prefix', dataIndex: 'restPrefix', key: 'restPrefix' },
    {
      title: 'Routes',
      key: 'routes',
      render: (_: unknown, svc: Service) => (
        <Button type="link" onClick={() => showRoutes(svc)}>
          View
        </Button>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, svc: Service) => (
        <Button danger onClick={() => deregisterService(svc)}>
          Deregister
        </Button>
      ),
    },
  ];

  const menuColumns = [
    { title: 'Label Key', dataIndex: 'labelKey', key: 'labelKey' },
    { title: 'Route', dataIndex: 'route', key: 'route' },
    { title: 'Icon', dataIndex: 'icon', key: 'icon' },
    { title: 'Parent', dataIndex: 'parentId', key: 'parentId' },
    { title: 'Sort Order', dataIndex: 'sortOrder', key: 'sortOrder' },
    { title: 'Micro App', dataIndex: 'microAppName', key: 'microAppName' },
    { title: 'Permission', dataIndex: 'requirePermission', key: 'requirePermission' },
    {
      title: 'Reorder',
      key: 'reorder',
      render: (_: unknown, _menu: Menu, index: number) => (
        <Space>
          <Button size="small" onClick={() => moveMenu(index, -1)} disabled={index === 0}>
            ↑
          </Button>
          <Button size="small" onClick={() => moveMenu(index, 1)} disabled={index === menus.length - 1}>
            ↓
          </Button>
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, menu: Menu) => (
        <Space>
          <Button type="link" onClick={() => openEditMenu(menu)}>
            Edit
          </Button>
          <Button danger type="link" onClick={() => deleteMenu(menu)}>
            Delete
          </Button>
        </Space>
      ),
    },
  ];

  const microAppColumns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Route', dataIndex: 'route', key: 'route' },
    { title: 'Bundle URL', dataIndex: 'bundleUrl', key: 'bundleUrl' },
    { title: 'Menu Label Key', dataIndex: 'menuLabelKey', key: 'menuLabelKey' },
    { title: 'Permission', dataIndex: 'requirePermission', key: 'requirePermission' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, app: MicroApp) => (
        <Space>
          <Button type="link" onClick={() => openEditMicroApp(app)}>
            Edit
          </Button>
          <Button danger type="link" onClick={() => deleteMicroApp(app)}>
            Delete
          </Button>
        </Space>
      ),
    },
  ];

  const servicesContent = (
    <Table loading={loadingServices} rowKey="id" dataSource={services} columns={serviceColumns} />
  );

  const menusContent = (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Button type="primary" onClick={openCreateMenu}>
        Create Menu
      </Button>
      <Table loading={loadingMenus} rowKey="id" dataSource={menus} columns={menuColumns} />
    </Space>
  );

  const microAppsContent = (
    <Table loading={loadingMicroApps} rowKey="name" dataSource={microApps} columns={microAppColumns} />
  );

  const renderContent = () => {
    switch (activeSection) {
      case 'applications':
        return <ApplicationsTab />;
      case 'services':
        return servicesContent;
      case 'menus':
        return menusContent;
      case 'micro-apps':
        return microAppsContent;
      case 'permissions':
        return <PermissionsTab />;
      case 'roles':
        return <RolesTab />;
      case 'route-policies':
        return <RoutePoliciesTab />;
      case 'attributes':
        return <AttributesTab />;
      case 'conditions':
        return <ConditionsTab />;
      case 'policies':
        return <PoliciesTab />;
      default:
        return <ApplicationsTab />;
    }
  };

  const activeMenuItem = menuItems.find((item) => item.key === activeSection);

  return (
    <Layout style={{ minHeight: 'calc(100vh - 64px)' }}>
      <Sider width={200} style={{ background: '#fff' }}>
        <Menu
          mode="inline"
          selectedKeys={[activeSection]}
          items={menuItems.map((item) => ({
            key: item.key,
            icon: item.icon,
            label: item.label,
          }))}
          onClick={({ key }) => {
            const item = menuItems.find((i) => i.key === key);
            if (item) {
              navigate(item.path);
            }
          }}
        />
      </Sider>
      <Content style={{ padding: 24 }}>
        <Card title={activeMenuItem?.label ?? 'Registry Management'}>{renderContent()}</Card>

        <Modal
          title={`Routes for ${selectedServiceName}`}
          open={isRoutesModalOpen}
          onCancel={() => setIsRoutesModalOpen(false)}
          footer={null}
        >
          {selectedRoutes.map((route, idx) => (
            <Typography.Paragraph key={idx}>
              <Typography.Text code>{route.method}</Typography.Text>{' '}
              <Typography.Text>{route.path}</Typography.Text>
            </Typography.Paragraph>
          ))}
        </Modal>

        <Modal
          title={editingMenu ? 'Edit Menu' : 'Create Menu'}
          open={isMenuModalOpen}
          onCancel={() => setIsMenuModalOpen(false)}
          footer={null}
          destroyOnClose
        >
          <Form form={menuForm} layout="vertical" onFinish={saveMenu}>
            <Form.Item
              label="Label Key"
              name="labelKey"
              rules={[{ required: true, message: 'Please enter label key' }]}
            >
              <Input placeholder="nav.orders" />
            </Form.Item>
            <Form.Item label="Route" name="route">
              <Input placeholder="/order" />
            </Form.Item>
            <Form.Item label="Icon" name="icon">
              <Input placeholder="AppstoreOutlined" />
            </Form.Item>
            <Form.Item label="Parent ID" name="parentId">
              <Input placeholder="parent-menu-id" />
            </Form.Item>
            <Form.Item label="Sort Order" name="sortOrder">
              <InputNumber style={{ width: '100%' }} placeholder="1" />
            </Form.Item>
            <Form.Item label="Micro App Name" name="microAppName">
              <Input placeholder="order-ui" />
            </Form.Item>
            <Form.Item label="Required Permission" name="requirePermission">
              <Input placeholder="order:read" />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={saving}>
                  Save
                </Button>
                <Button onClick={() => setIsMenuModalOpen(false)}>Cancel</Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>

        <Modal
          title="Edit Micro App"
          open={isMicroAppModalOpen}
          onCancel={() => setIsMicroAppModalOpen(false)}
          footer={null}
          destroyOnClose
        >
          <Form form={microAppForm} layout="vertical" onFinish={saveMicroApp}>
            <Form.Item label="Name" name="name">
              <Input disabled />
            </Form.Item>
            <Form.Item
              label="Route"
              name="route"
              rules={[{ required: true, message: 'Please enter route' }]}
            >
              <Input placeholder="/admin/registry" />
            </Form.Item>
            <Form.Item
              label="Bundle URL"
              name="bundleUrl"
              rules={[{ required: true, message: 'Please enter bundle URL' }]}
            >
              <Input placeholder="/apps/registry-admin-ui/registry-admin-ui.js" />
            </Form.Item>
            <Form.Item
              label="Menu Label Key"
              name="menuLabelKey"
              rules={[{ required: true, message: 'Please enter menu label key' }]}
            >
              <Input placeholder="nav.registry" />
            </Form.Item>
            <Form.Item
              label="Required Permission"
              name="requirePermission"
              rules={[{ required: true, message: 'Please enter required permission' }]}
            >
              <Input placeholder="registry:read" />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={saving}>
                  Save
                </Button>
                <Button onClick={() => setIsMicroAppModalOpen(false)}>Cancel</Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>
      </Content>
    </Layout>
  );
}
