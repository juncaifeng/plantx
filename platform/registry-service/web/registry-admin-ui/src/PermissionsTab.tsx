import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Modal, Space, Table, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import {
  IAMServiceClient,
  type Permission,
} from '@plantx/kit-sdk-api/iam';

export function PermissionsTab() {
  const { apiClient } = useKitContext();
  const iamClient = useMemo(
    () => (apiClient ? new IAMServiceClient(apiClient) : null),
    [apiClient]
  );

  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [form] = Form.useForm();

  const load = async () => {
    if (!iamClient) return;
    setLoading(true);
    try {
      const res = await iamClient.listPermissions();
      setPermissions(res.permissions);
    } catch (e) {
      message.error('Failed to load permissions');
      // eslint-disable-next-line no-console
      console.error('failed to list permissions', e);
    } finally {
      setLoading(false);
    }
  };

  const openCreate = () => {
    form.resetFields();
    setIsModalOpen(true);
  };

  const save = async (values: Record<string, unknown>) => {
    if (!iamClient) return;
    setSaving(true);
    try {
      await iamClient.createPermission({
        name: values.name as string,
        resource: values.resource as string,
        operation: values.operation as string,
        description: values.description as string | undefined,
      });
      message.success('Permission created');
      form.resetFields();
      setIsModalOpen(false);
      await load();
    } catch (e) {
      message.error('Failed to create permission');
      // eslint-disable-next-line no-console
      console.error('failed to create permission', e);
    } finally {
      setSaving(false);
    }
  };

  const deletePermission = async (perm: Permission) => {
    if (!iamClient || !perm.id) return;
    try {
      await iamClient.deletePermission({ id: perm.id });
      message.success('Permission deleted');
      await load();
    } catch (e) {
      message.error('Failed to delete permission');
      // eslint-disable-next-line no-console
      console.error('failed to delete permission', e);
    }
  };

  useEffect(() => {
    load();
  }, [iamClient]);

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Resource', dataIndex: 'resource', key: 'resource' },
    { title: 'Operation', dataIndex: 'operation', key: 'operation' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, perm: Permission) => (
        <Button danger type="link" onClick={() => deletePermission(perm)}>
          Delete
        </Button>
      ),
    },
  ];

  return (
    <Card title="Permissions">
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Button type="primary" onClick={openCreate}>
          Create Permission
        </Button>
        <Table
          loading={loading}
          rowKey="id"
          dataSource={permissions}
          columns={columns}
          pagination={{ pageSize: 10 }}
        />
      </Space>

      <Modal
        title="Create Permission"
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={save}>
          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter permission name' }]}
          >
            <Input placeholder="order:read" />
          </Form.Item>
          <Form.Item
            label="Resource"
            name="resource"
            rules={[{ required: true, message: 'Please enter resource' }]}
          >
            <Input placeholder="order" />
          </Form.Item>
          <Form.Item
            label="Operation"
            name="operation"
            rules={[{ required: true, message: 'Please enter operation' }]}
          >
            <Input placeholder="read" />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <Input.TextArea rows={3} placeholder="Short description" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                Save
              </Button>
              <Button onClick={() => setIsModalOpen(false)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
