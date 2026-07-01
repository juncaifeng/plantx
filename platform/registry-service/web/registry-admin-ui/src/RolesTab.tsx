import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Checkbox, Form, Input, Modal, Space, Table, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import {
  IAMServiceClient,
  type Role,
  type Permission,
} from '@plantx/kit-sdk-api/iam';

export function RolesTab() {
  const { apiClient } = useKitContext();
  const iamClient = useMemo(
    () => (apiClient ? new IAMServiceClient(apiClient) : null),
    [apiClient]
  );

  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editing, setEditing] = useState<Role | null>(null);
  const [form] = Form.useForm();

  const load = async () => {
    if (!iamClient) return;
    setLoading(true);
    try {
      const [rolesRes, permsRes] = await Promise.all([
        iamClient.listRoles(),
        iamClient.listPermissions(),
      ]);
      setRoles(rolesRes.roles);
      setPermissions(permsRes.permissions);
    } catch (e) {
      message.error('Failed to load roles');
      // eslint-disable-next-line no-console
      console.error('failed to load roles', e);
    } finally {
      setLoading(false);
    }
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ permissions: [] });
    setIsModalOpen(true);
  };

  const openEdit = (role: Role) => {
    setEditing(role);
    form.resetFields();
    form.setFieldsValue({
      name: role.name,
      description: role.description,
      permissions: role.permissions || [],
    });
    setIsModalOpen(true);
  };

  const save = async (values: Record<string, unknown>) => {
    if (!iamClient) return;
    setSaving(true);
    try {
      const selectedPermissions = (values.permissions as string[]) || [];
      if (editing) {
        await iamClient.updateRole({
          id: editing.id,
          name: values.name as string,
          description: values.description as string | undefined,
          permissions: selectedPermissions,
        });
        message.success('Role updated');
      } else {
        await iamClient.createRole({
          name: values.name as string,
          description: values.description as string | undefined,
          permissions: selectedPermissions,
        });
        message.success('Role created');
      }
      form.resetFields();
      setIsModalOpen(false);
      setEditing(null);
      await load();
    } catch (e) {
      message.error(editing ? 'Failed to update role' : 'Failed to create role');
      // eslint-disable-next-line no-console
      console.error('failed to save role', e);
    } finally {
      setSaving(false);
    }
  };

  const deleteRole = async (role: Role) => {
    if (!iamClient || !role.id) return;
    try {
      await iamClient.deleteRole({ id: role.id });
      message.success('Role deleted');
      await load();
    } catch (e) {
      message.error('Failed to delete role');
      // eslint-disable-next-line no-console
      console.error('failed to delete role', e);
    }
  };

  useEffect(() => {
    load();
  }, [iamClient]);

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Permissions',
      dataIndex: 'permissions',
      key: 'permissions',
      render: (perms: string[]) => perms?.join(', ') || '-',
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, role: Role) => (
        <Space>
          <Button type="link" onClick={() => openEdit(role)}>
            Edit
          </Button>
          <Button danger type="link" onClick={() => deleteRole(role)}>
            Delete
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Card title="Roles">
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Button type="primary" onClick={openCreate}>
          Create Role
        </Button>
        <Table
          loading={loading}
          rowKey="id"
          dataSource={roles}
          columns={columns}
          pagination={{ pageSize: 10 }}
        />
      </Space>

      <Modal
        title={editing ? 'Edit Role' : 'Create Role'}
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={save}>
          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter role name' }]}
          >
            <Input placeholder="order-manager" />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <Input.TextArea rows={3} placeholder="Short description" />
          </Form.Item>
          <Form.Item label="Permissions" name="permissions">
            <Checkbox.Group options={permissions.map((p) => ({ label: p.name, value: p.name }))} />
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
