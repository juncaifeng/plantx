import React, { useEffect, useMemo, useState } from 'react';
import { Button, Form, Input, Modal, Space, Table, Tag, Typography, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { Tenant, TenantServiceClient } from '@plantx/kit-sdk-api/tenant';

export function TenantAdminPage() {
  const { apiClient } = useKitContext();
  const tenantClient = useMemo(
    () => (apiClient ? new TenantServiceClient(apiClient) : null),
    [apiClient]
  );
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);

  const list = async () => {
    if (!tenantClient) return;
    setLoading(true);
    try {
      const res = await tenantClient.listTenants();
      setTenants(res.tenants);
    } catch (e) {
      message.error('Failed to load tenants');
      // eslint-disable-next-line no-console
      console.error('failed to list tenants', e);
    } finally {
      setLoading(false);
    }
  };

  const create = async (values: { name: string }) => {
    if (!tenantClient) return;
    setSubmitting(true);
    try {
      const created = await tenantClient.createTenant({ name: values.name.trim() });
      setTenants((prev: Tenant[]) => [created, ...prev]);
      form.resetFields();
      setModalOpen(false);
      message.success('Tenant created');
    } catch (e) {
      message.error('Failed to create tenant');
      // eslint-disable-next-line no-console
      console.error('failed to create tenant', e);
    } finally {
      setSubmitting(false);
    }
  };

  useEffect(() => {
    list();
  }, [tenantClient]);

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      render: (id: string) => <Typography.Text type="secondary" copyable>{id}</Typography.Text>,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color={status === 'active' ? 'green' : 'default'}>{status}</Tag>,
    },
    {
      title: 'Created At',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (ts: number) => new Date(ts * 1000).toLocaleString(),
    },
  ];

  return (
    <>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Space style={{ justifyContent: 'space-between', width: '100%' }}>
          <Typography.Title level={4} style={{ margin: 0 }}>
            Tenants
          </Typography.Title>
          <Button type="primary" onClick={() => setModalOpen(true)}>
            Create Tenant
          </Button>
        </Space>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={tenants}
          columns={columns}
          pagination={{ pageSize: 10 }}
        />
      </Space>

      <Modal
        title="Create Tenant"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        confirmLoading={submitting}
      >
        <Form form={form} layout="vertical" onFinish={create}>
          <Form.Item
            name="name"
            label="Tenant Name"
            rules={[{ required: true, message: 'Please enter tenant name' }]}
          >
            <Input placeholder="Tenant A" autoFocus />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
