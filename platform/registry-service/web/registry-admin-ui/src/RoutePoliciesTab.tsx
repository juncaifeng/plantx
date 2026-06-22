import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Switch,
  Table,
  Typography,
  message,
} from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import {
  RegistryServiceClient,
  type RoutePolicy,
  type Service,
  type ServiceRoute,
} from '@plantx/kit-sdk-api/registry';

export function RoutePoliciesTab() {
  const { apiClient } = useKitContext();
  const registryClient = useMemo(
    () => (apiClient ? new RegistryServiceClient(apiClient) : null),
    [apiClient]
  );

  const [services, setServices] = useState<Service[]>([]);
  const [policies, setPolicies] = useState<Record<string, RoutePolicy>>({});
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [editingService, setEditingService] = useState<Service | null>(null);
  const [syncRoutes, setSyncRoutes] = useState<ServiceRoute[] | null>(null);
  const [policyForm] = Form.useForm();

  const load = async () => {
    if (!registryClient) return;
    setLoading(true);
    try {
      const [servicesRes, syncRes] = await Promise.all([
        registryClient.listServices(),
        registryClient.syncRoutes(),
      ]);
      setServices(servicesRes.services);
      const map: Record<string, RoutePolicy> = {};
      for (const route of syncRes.routes) {
        map[route.serviceId] = route.policy ?? {};
      }
      setPolicies(map);
    } catch (e) {
      message.error('Failed to load route policies');
      // eslint-disable-next-line no-console
      console.error('failed to load route policies', e);
    } finally {
      setLoading(false);
    }
  };

  const openEdit = (svc: Service) => {
    setEditingService(svc);
    const policy = policies[svc.id] ?? {};
    policyForm.setFieldsValue({
      rateLimitRps: policy.rateLimitRps ?? 0,
      authRequired: policy.authRequired ?? true,
      canaryWeight: policy.canaryWeight ?? 0,
      canaryHost: policy.canaryHost ?? '',
    });
  };

  const savePolicy = async (values: Record<string, unknown>) => {
    if (!registryClient || !editingService) return;
    setSaving(true);
    try {
      const policy: RoutePolicy = {
        rateLimitRps: values.rateLimitRps as number | undefined,
        authRequired: values.authRequired as boolean | undefined,
        canaryWeight: values.canaryWeight as number | undefined,
        canaryHost: (values.canaryHost as string) || undefined,
      };
      const updated = await registryClient.setRoutePolicy({
        serviceId: editingService.id,
        policy,
      });
      setPolicies((prev) => ({ ...prev, [editingService.id]: updated }));
      message.success('Route policy updated');
      setEditingService(null);
      policyForm.resetFields();
    } catch (e) {
      message.error('Failed to update route policy');
      // eslint-disable-next-line no-console
      console.error('failed to update route policy', e);
    } finally {
      setSaving(false);
    }
  };

  const runSync = async () => {
    if (!registryClient) return;
    setSyncing(true);
    try {
      const res = await registryClient.syncRoutes();
      setSyncRoutes(res.routes);
    } catch (e) {
      message.error('Failed to sync routes');
      // eslint-disable-next-line no-console
      console.error('failed to sync routes', e);
    } finally {
      setSyncing(false);
    }
  };

  useEffect(() => {
    load();
  }, [registryClient]);

  const columns = [
    { title: 'Service', dataIndex: 'name', key: 'name' },
    { title: 'REST Prefix', dataIndex: 'restPrefix', key: 'restPrefix' },
    {
      title: 'Rate Limit (rps)',
      key: 'rateLimit',
      render: (_: unknown, svc: Service) => policies[svc.id]?.rateLimitRps || '-',
    },
    {
      title: 'Auth Required',
      key: 'auth',
      render: (_: unknown, svc: Service) =>
        policies[svc.id]?.authRequired === false ? 'No' : 'Yes',
    },
    {
      title: 'Canary',
      key: 'canary',
      render: (_: unknown, svc: Service) => {
        const p = policies[svc.id];
        if (!p?.canaryWeight) return '-';
        return `${p.canaryWeight}% → ${p.canaryHost || svc.name}`;
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, svc: Service) => (
        <Button type="link" onClick={() => openEdit(svc)}>
          Edit Policy
        </Button>
      ),
    },
  ];

  return (
    <Card title="Route Policies">
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Space>
          <Button type="primary" onClick={runSync} loading={syncing}>
            Sync Routes
          </Button>
          <Button onClick={load} loading={loading}>
            Refresh
          </Button>
        </Space>
        <Table loading={loading} rowKey="id" dataSource={services} columns={columns} />
      </Space>

      <Modal
        title={`Route Policy for ${editingService?.name}`}
        open={!!editingService}
        onCancel={() => setEditingService(null)}
        footer={null}
        destroyOnClose
      >
        <Form form={policyForm} layout="vertical" onFinish={savePolicy}>
          <Form.Item label="Rate Limit (requests per second)" name="rateLimitRps">
            <InputNumber style={{ width: '100%' }} min={0} placeholder="0 = unlimited" />
          </Form.Item>
          <Form.Item label="Auth Required" name="authRequired" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label="Canary Weight (%)" name="canaryWeight">
            <InputNumber style={{ width: '100%' }} min={0} max={100} placeholder="0 = no canary" />
          </Form.Item>
          <Form.Item label="Canary Host" name="canaryHost">
            <Input placeholder="e.g. iam-service-canary:8081" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                Save
              </Button>
              <Button onClick={() => setEditingService(null)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Generated Gateway Routes"
        open={syncRoutes !== null}
        onCancel={() => setSyncRoutes(null)}
        footer={null}
        width={800}
      >
        <Typography.Paragraph>
          <pre style={{ maxHeight: 500, overflow: 'auto' }}>
            {JSON.stringify(syncRoutes, null, 2)}
          </pre>
        </Typography.Paragraph>
      </Modal>
    </Card>
  );
}
