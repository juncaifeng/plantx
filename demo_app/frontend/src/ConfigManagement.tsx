import React, { useCallback, useEffect, useState } from 'react';
import { Button, Form, Input, Radio, Space, Table, Tag, Typography } from 'antd';
import { useKitContext, useKitPermission, useKitUser } from '@plantx/kit-sdk-kit';

interface Setting {
  id: string;
  tenant_id?: string;
  key: string;
  value: string;
  scope: 'SETTING_SCOPE_UNSPECIFIED' | 'SETTING_SCOPE_GLOBAL' | 'SETTING_SCOPE_TENANT';
  updated_by: string;
  updated_at: number;
}

export function ConfigManagement() {
  const ctx = useKitContext();
  const user = useKitUser();
  const canList = useKitPermission('setting:list');
  const canCreate = useKitPermission('setting:create');
  const canUpdate = useKitPermission('setting:update');
  const canDelete = useKitPermission('setting:delete');

  const [settings, setSettings] = useState<Setting[]>([]);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const fetchSettings = useCallback(async () => {
    if (!canList || !ctx.apiClient) return;
    setLoading(true);
    try {
      const data = await ctx.apiClient.get<{ settings: Setting[] }>('/demo/v1/settings');
      setSettings(data.settings ?? []);
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error('failed to load settings:', err);
    } finally {
      setLoading(false);
    }
  }, [canList, ctx.apiClient]);

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  const handleCreate = async (values: { key: string; value: string; scope: string }) => {
    if (!ctx.apiClient) return;
    try {
      await ctx.apiClient.post<Setting>('/demo/v1/settings', {
        key: values.key,
        value: values.value,
        scope: values.scope,
      });
      form.resetFields();
      await fetchSettings();
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error('failed to create setting:', err);
    }
  };

  const handleDelete = async (id: string) => {
    if (!ctx.apiClient) return;
    try {
      await ctx.apiClient.delete<void>(`/demo/v1/settings/${encodeURIComponent(id)}`);
      await fetchSettings();
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error('failed to delete setting:', err);
    }
  };

  const columns = [
    { title: 'Key', dataIndex: 'key', key: 'key' },
    { title: 'Value', dataIndex: 'value', key: 'value' },
    {
      title: 'Scope',
      dataIndex: 'scope',
      key: 'scope',
      render: (scope: string) => {
        const label = scope.replace('SETTING_SCOPE_', '').toLowerCase();
        const color = label === 'global' ? 'red' : label === 'tenant' ? 'blue' : 'default';
        return <Tag color={color}>{label}</Tag>;
      },
    },
    { title: 'Updated By', dataIndex: 'updated_by', key: 'updated_by' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, record: Setting) => (
        <Space>
          {canUpdate && (
            <Button
              size="small"
              onClick={async () => {
                const value = window.prompt('New value', record.value);
                if (value === null || !ctx.apiClient) return;
                try {
                  await ctx.apiClient.put<Setting>(
                    `/demo/v1/settings/${encodeURIComponent(record.id)}`,
                    { value }
                  );
                  await fetchSettings();
                } catch (err) {
                  // eslint-disable-next-line no-console
                  console.error('failed to update setting:', err);
                }
              }}
            >
              Edit
            </Button>
          )}
          {canDelete && (
            <Button size="small" danger onClick={() => handleDelete(record.id)}>
              Delete
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Typography.Title level={4}>Configuration Management</Typography.Title>
      <Typography.Text type="secondary">
        Current user: {user?.displayName ?? user?.username ?? 'unknown'}
      </Typography.Text>

      {canCreate && (
        <Form
          form={form}
          layout="inline"
          onFinish={handleCreate}
          style={{ marginTop: 16, marginBottom: 16 }}
        >
          <Form.Item name="key" rules={[{ required: true }]}>
            <Input placeholder="Key" />
          </Form.Item>
          <Form.Item name="value" rules={[{ required: true }]}>
            <Input placeholder="Value" />
          </Form.Item>
          <Form.Item name="scope" initialValue="SETTING_SCOPE_TENANT">
            <Radio.Group>
              <Radio.Button value="SETTING_SCOPE_TENANT">Tenant</Radio.Button>
              <Radio.Button value="SETTING_SCOPE_GLOBAL">Global</Radio.Button>
            </Radio.Group>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">
              Create
            </Button>
          </Form.Item>
        </Form>
      )}

      <Table
        rowKey="id"
        loading={loading}
        dataSource={settings}
        columns={columns}
        pagination={false}
      />
    </div>
  );
}
