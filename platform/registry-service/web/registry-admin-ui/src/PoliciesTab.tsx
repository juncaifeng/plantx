import React, { useEffect, useMemo, useState } from 'react';
import { Button, Form, Input, Modal, Select, Space, Table, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { IAMServiceClient, type Policy, type Condition, type Permission } from '@plantx/kit-sdk-api/iam';

export function PoliciesTab() {
  const { apiClient } = useKitContext();
  const iamClient = useMemo(() => (apiClient ? new IAMServiceClient(apiClient) : null), [apiClient]);
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [conditions, setConditions] = useState<Condition[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editing, setEditing] = useState<Policy | null>(null);
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  const list = async () => {
    if (!iamClient) return;
    setLoading(true);
    try {
      const [pRes, cRes, permRes] = await Promise.all([
        iamClient.listPolicies(),
        iamClient.listConditions(),
        iamClient.listPermissions(),
      ]);
      setPolicies(pRes.policies);
      setConditions(cRes.conditions);
      setPermissions(permRes.permissions);
    } catch (e) {
      message.error('Failed to load policies');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    list();
  }, [iamClient]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    setIsModalOpen(true);
  };

  const openEdit = (policy: Policy) => {
    setEditing(policy);
    form.setFieldsValue({
      name: policy.name,
      description: policy.description,
      effect: policy.effect,
      priority: policy.priority,
      permissions: policy.permissions,
      conditionIds: policy.conditionIds,
    });
    setIsModalOpen(true);
  };

  const save = async (values: Record<string, unknown>) => {
    if (!iamClient) return;
    setSaving(true);
    try {
      const payload = {
        name: values.name as string,
        description: (values.description as string) ?? '',
        effect: (values.effect as string) ?? 'allow',
        priority: (values.priority as number) ?? 0,
        permissions: (values.permissions as string[]) ?? [],
        conditionIds: (values.conditionIds as string[]) ?? [],
      };
      if (editing) {
        await iamClient.updatePolicy({ id: editing.id, ...payload });
      } else {
        await iamClient.createPolicy(payload);
      }
      message.success('Policy saved');
      form.resetFields();
      setIsModalOpen(false);
      setEditing(null);
      await list();
    } catch (e) {
      message.error('Failed to save policy');
    } finally {
      setSaving(false);
    }
  };

  const deletePolicy = async (policy: Policy) => {
    if (!iamClient) return;
    try {
      await iamClient.deletePolicy({ id: policy.id });
      message.success('Policy deleted');
      await list();
    } catch (e) {
      message.error('Failed to delete policy');
    }
  };

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Effect', dataIndex: 'effect', key: 'effect' },
    { title: 'Priority', dataIndex: 'priority', key: 'priority' },
    { title: 'Permissions', dataIndex: 'permissions', key: 'permissions', render: (v: string[]) => v?.join(', ') },
    { title: 'Conditions', dataIndex: 'conditionIds', key: 'conditionIds', render: (v: string[]) => v?.join(', ') },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, policy: Policy) => (
        <Space>
          <Button type="link" onClick={() => openEdit(policy)}>Edit</Button>
          <Button danger type="link" onClick={() => deletePolicy(policy)}>Delete</Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Button type="primary" onClick={openCreate}>Create Policy</Button>
      <Table loading={loading} rowKey="id" dataSource={policies} columns={columns} />
      <Modal title={editing ? 'Edit Policy' : 'Create Policy'} open={isModalOpen} onCancel={() => setIsModalOpen(false)} footer={null} width={640}>
        <Form form={form} layout="vertical" onFinish={save}>
          <Form.Item label="Name" name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Effect" name="effect" initialValue="allow">
            <Select options={[{ value: 'allow' }, { value: 'deny' }]} />
          </Form.Item>
          <Form.Item label="Priority" name="priority" initialValue={0}>
            <Input type="number" />
          </Form.Item>
          <Form.Item label="Permissions" name="permissions">
            <Select mode="multiple" options={permissions.map((p) => ({ value: p.name, label: p.name }))} />
          </Form.Item>
          <Form.Item label="Conditions" name="conditionIds">
            <Select mode="multiple" options={conditions.map((c) => ({ value: c.id, label: `${c.name} (${c.attributeKey} ${c.operator} ${c.value})` }))} />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <Input.TextArea />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={saving}>Save</Button>
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
