import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Checkbox,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import {
  IAMServiceClient,
  type Permission,
  type Role,
  type User,
} from '@plantx/kit-sdk-api/iam';

export function IAMAdminPage() {
  const { apiClient } = useKitContext();
  const iamClient = useMemo(() => (apiClient ? new IAMServiceClient(apiClient) : null), [apiClient]);

  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);

  const [loadingUsers, setLoadingUsers] = useState(false);
  const [loadingRoles, setLoadingRoles] = useState(false);
  const [loadingPermissions, setLoadingPermissions] = useState(false);

  const [creatingUser, setCreatingUser] = useState(false);
  const [isCreateUserModalOpen, setIsCreateUserModalOpen] = useState(false);
  const [userForm] = Form.useForm();

  const [permissionModalOpen, setPermissionModalOpen] = useState(false);
  const [savingPermission, setSavingPermission] = useState(false);
  const [permissionForm] = Form.useForm();

  const [roleModalOpen, setRoleModalOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [savingRole, setSavingRole] = useState(false);
  const [roleForm] = Form.useForm();

  const listUsers = async () => {
    if (!iamClient) return;
    setLoadingUsers(true);
    try {
      const res = await iamClient.listUsers();
      setUsers(res.users);
    } catch (e) {
      message.error('Failed to load users');
      // eslint-disable-next-line no-console
      console.error('failed to list users', e);
    } finally {
      setLoadingUsers(false);
    }
  };

  const listRoles = async () => {
    if (!iamClient) return;
    setLoadingRoles(true);
    try {
      const res = await iamClient.listRoles();
      setRoles(res.roles);
    } catch (e) {
      message.error('Failed to load roles');
      // eslint-disable-next-line no-console
      console.error('failed to list roles', e);
    } finally {
      setLoadingRoles(false);
    }
  };

  const listPermissions = async () => {
    if (!iamClient) return;
    setLoadingPermissions(true);
    try {
      const res = await iamClient.listPermissions();
      setPermissions(res.permissions);
    } catch (e) {
      message.error('Failed to load permissions');
      // eslint-disable-next-line no-console
      console.error('failed to list permissions', e);
    } finally {
      setLoadingPermissions(false);
    }
  };

  const createUser = async (values: { username: string; tenant_id: string; role_ids: string[] }) => {
    if (!iamClient) return;
    setCreatingUser(true);
    try {
      const created = await iamClient.createUser({
        username: values.username.trim(),
        tenant_id: values.tenant_id.trim(),
        role_ids: values.role_ids ?? [],
      });
      setUsers((prev: User[]) => [created, ...prev]);
      userForm.resetFields();
      setIsCreateUserModalOpen(false);
      message.success('User created');
    } catch (e) {
      message.error('Failed to create user');
      // eslint-disable-next-line no-console
      console.error('failed to create user', e);
    } finally {
      setCreatingUser(false);
    }
  };

  const createPermission = async (values: {
    name: string;
    resource: string;
    operation: string;
    description?: string;
  }) => {
    if (!iamClient) return;
    setSavingPermission(true);
    try {
      const created = await iamClient.createPermission({
        name: values.name.trim(),
        resource: values.resource.trim(),
        operation: values.operation.trim(),
        description: values.description?.trim(),
      });
      setPermissions((prev: Permission[]) => [created, ...prev]);
      permissionForm.resetFields();
      setPermissionModalOpen(false);
      message.success('Permission created');
    } catch (e) {
      message.error('Failed to create permission');
      // eslint-disable-next-line no-console
      console.error('failed to create permission', e);
    } finally {
      setSavingPermission(false);
    }
  };

  const deletePermission = async (id: string) => {
    if (!iamClient) return;
    try {
      await iamClient.deletePermission({ id });
      setPermissions((prev: Permission[]) => prev.filter((p) => p.id !== id));
      message.success('Permission deleted');
    } catch (e) {
      message.error('Failed to delete permission');
      // eslint-disable-next-line no-console
      console.error('failed to delete permission', e);
    }
  };

  const openCreateRoleModal = () => {
    setEditingRole(null);
    roleForm.resetFields();
    setRoleModalOpen(true);
  };

  const openEditRoleModal = (role: Role) => {
    setEditingRole(role);
    roleForm.setFieldsValue({
      name: role.name,
      description: role.description,
      permissions: role.permissions ?? [],
    });
    setRoleModalOpen(true);
  };

  const closeRoleModal = () => {
    setRoleModalOpen(false);
    setEditingRole(null);
    roleForm.resetFields();
  };

  const saveRole = async (values: {
    name: string;
    description?: string;
    permissions?: string[];
  }) => {
    if (!iamClient) return;
    setSavingRole(true);
    try {
      const payload = {
        name: values.name.trim(),
        description: values.description?.trim(),
        permissions: values.permissions ?? [],
      };
      if (editingRole) {
        const updated = await iamClient.updateRole({ id: editingRole.id, ...payload });
        setRoles((prev: Role[]) => prev.map((r) => (r.id === updated.id ? updated : r)));
        message.success('Role updated');
      } else {
        const created = await iamClient.createRole(payload);
        setRoles((prev: Role[]) => [created, ...prev]);
        message.success('Role created');
      }
      closeRoleModal();
    } catch (e) {
      message.error(editingRole ? 'Failed to update role' : 'Failed to create role');
      // eslint-disable-next-line no-console
      console.error('failed to save role', e);
    } finally {
      setSavingRole(false);
    }
  };

  const deleteRole = async (id: string) => {
    if (!iamClient) return;
    try {
      await iamClient.deleteRole({ id });
      setRoles((prev: Role[]) => prev.filter((r) => r.id !== id));
      message.success('Role deleted');
    } catch (e) {
      message.error('Failed to delete role');
      // eslint-disable-next-line no-console
      console.error('failed to delete role', e);
    }
  };

  const roleName = (id?: string) => roles.find((r: Role) => r.id === id)?.name ?? id;
  const permissionLabel = (id?: string) => {
    const p = permissions.find((perm) => perm.id === id);
    return p ? `${p.resource}:${p.operation}` : id;
  };

  useEffect(() => {
    listUsers();
    listRoles();
    listPermissions();
  }, [iamClient]);

  const userColumns = [
    { title: 'Username', dataIndex: 'username', key: 'username' },
    { title: 'Tenant', dataIndex: 'tenant_id', key: 'tenant_id' },
    {
      title: 'Roles',
      key: 'roles',
      render: (_: unknown, user: User) => (
        <Space size="small" wrap>
          {(user.role_ids ?? []).map((id) => (
            <Tag key={id}>{roleName(id)}</Tag>
          ))}
        </Space>
      ),
    },
  ];

  const permissionColumns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Resource', dataIndex: 'resource', key: 'resource' },
    { title: 'Operation', dataIndex: 'operation', key: 'operation' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, perm: Permission) => (
        <Popconfirm
          title="Delete permission?"
          description={`This will permanently delete ${perm.name}.`}
          onConfirm={() => deletePermission(perm.id)}
          okText="Delete"
          cancelText="Cancel"
        >
          <Button type="link" danger>
            Delete
          </Button>
        </Popconfirm>
      ),
    },
  ];

  const roleColumns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      render: (desc?: string) => desc || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: 'Permissions',
      key: 'permissions',
      render: (_: unknown, role: Role) => (
        <Space size="small" wrap>
          {(role.permissions ?? []).map((id) => (
            <Tag key={id} color="blue">
              {permissionLabel(id)}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, role: Role) => (
        <Space size="small">
          <Button type="link" onClick={() => openEditRoleModal(role)}>
            Edit
          </Button>
          <Popconfirm
            title="Delete role?"
            description={`This will permanently delete ${role.name}.`}
            onConfirm={() => deleteRole(role.id)}
            okText="Delete"
            cancelText="Cancel"
          >
            <Button type="link" danger>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const userTab = (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Button type="primary" onClick={() => setIsCreateUserModalOpen(true)}>
        Create User
      </Button>
      <Table loading={loadingUsers} rowKey="id" dataSource={users} columns={userColumns} />
    </Space>
  );

  const permissionTab = (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Button type="primary" onClick={() => setPermissionModalOpen(true)}>
        Create Permission
      </Button>
      <Table
        loading={loadingPermissions}
        rowKey="id"
        dataSource={permissions}
        columns={permissionColumns}
      />
    </Space>
  );

  const roleTab = (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Button type="primary" onClick={openCreateRoleModal}>
        Create Role
      </Button>
      <Table loading={loadingRoles} rowKey="id" dataSource={roles} columns={roleColumns} />
    </Space>
  );

  return (
    <Card title="IAM">
      <Tabs
        items={[
          { key: 'users', label: 'Users', children: userTab },
          { key: 'roles', label: 'Roles', children: roleTab },
          { key: 'permissions', label: 'Permissions', children: permissionTab },
        ]}
      />

      <Modal
        title="Create User"
        open={isCreateUserModalOpen}
        onCancel={() => setIsCreateUserModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={userForm} layout="vertical" onFinish={createUser}>
          <Form.Item
            label="Username"
            name="username"
            rules={[{ required: true, message: 'Please enter username' }]}
          >
            <Input placeholder="bob" />
          </Form.Item>
          <Form.Item
            label="Tenant ID"
            name="tenant_id"
            rules={[{ required: true, message: 'Please enter tenant ID' }]}
          >
            <Input placeholder="t_001" />
          </Form.Item>
          <Form.Item label="Roles" name="role_ids">
            <Select
              mode="multiple"
              placeholder="Select roles"
              options={roles.map((r: Role) => ({ label: r.name, value: r.id }))}
              allowClear
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={creatingUser}>
                Create
              </Button>
              <Button onClick={() => setIsCreateUserModalOpen(false)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Create Permission"
        open={permissionModalOpen}
        onCancel={() => setPermissionModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={permissionForm} layout="vertical" onFinish={createPermission}>
          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter permission name' }]}
          >
            <Input placeholder="Create Order" />
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
            <Input placeholder="create" />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <Input.TextArea rows={2} placeholder="Optional description" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={savingPermission}>
                Create
              </Button>
              <Button onClick={() => setPermissionModalOpen(false)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingRole ? 'Edit Role' : 'Create Role'}
        open={roleModalOpen}
        onCancel={closeRoleModal}
        footer={null}
        destroyOnClose
      >
        <Form form={roleForm} layout="vertical" onFinish={saveRole}>
          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter role name' }]}
          >
            <Input placeholder="Order Manager" />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <Input.TextArea rows={2} placeholder="Optional description" />
          </Form.Item>
          <Form.Item label="Permissions" name="permissions">
            <Checkbox.Group
              options={permissions.map((p: Permission) => ({
                label: `${p.name} (${p.resource}:${p.operation})`,
                value: p.id,
              }))}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={savingRole}>
                {editingRole ? 'Update' : 'Create'}
              </Button>
              <Button onClick={closeRoleModal}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
