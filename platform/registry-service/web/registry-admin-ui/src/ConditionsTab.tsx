import React, { useEffect, useMemo, useState } from 'react';
import { Button, Form, Input, Modal, Space, Table, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { IAMServiceClient, type Condition } from '@plantx/kit-sdk-api/iam';

export function ConditionsTab() {
  const { apiClient } = useKitContext();
  const iamClient = useMemo(() => (apiClient ? new IAMServiceClient(apiClient) : null), [apiClient]);
  const [conditions, setConditions] = useState<Condition[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editing, setEditing] = useState<Condition | null>(null);
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  const list = async () => {
    if (!iamClient) return;
    setLoading(true);
    try {
      const res = await iamClient.listConditions();
      setConditions(res.conditions);
    } catch (e) {
      message.error('Failed to load conditions');
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

  const openEdit = (cond: Condition) => {
    setEditing(cond);
    form.setFieldsValue({
      name: cond.name,
      attributeKey: cond.attributeKey,
      operator: cond.operator,
      value: cond.value,
      description: cond.description,
    });
    setIsModalOpen(true);
  };

  const save = async (values: Record<string, unknown>) => {
    if (!iamClient) return;
    setSaving(true);
    try {
      if (editing) {
        await iamClient.updateCondition({
          id: editing.id,
          name: values.name as string,
          attributeKey: values.attributeKey as string,
          operator: values.operator as string,
          value: values.value as string,
          description: values.description as string,
        });
      } else {
        await iamClient.createCondition({
          name: values.name as string,
          attributeKey: values.attributeKey as string,
          operator: values.operator as string,
          value: values.value as string,
          description: values.description as string,
        });
      }
      message.success('Condition saved');
      form.resetFields();
      setIsModalOpen(false);
      setEditing(null);
      await list();
    } catch (e) {
      message.error('Failed to save condition');
    } finally {
      setSaving(false);
    }
  };

  const deleteCond = async (cond: Condition) => {
    if (!iamClient) return;
    try {
      await iamClient.deleteCondition({ id: cond.id });
      message.success('Condition deleted');
      await list();
    } catch (e) {
      message.error('Failed to delete condition');
    }
  };

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Attribute', dataIndex: 'attributeKey', key: 'attributeKey' },
    { title: 'Operator', dataIndex: 'operator', key: 'operator' },
    { title: 'Value', dataIndex: 'value', key: 'value' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, cond: Condition) => (
        <Space>
          <Button type="link" onClick={() => openEdit(cond)}>Edit</Button>
          <Button danger type="link" onClick={() => deleteCond(cond)}>Delete</Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Button type="primary" onClick={openCreate}>Create Condition</Button>
      <Table loading={loading} rowKey="id" dataSource={conditions} columns={columns} />
      <Modal title={editing ? 'Edit Condition' : 'Create Condition'} open={isModalOpen} onCancel={() => setIsModalOpen(false)} footer={null}>
        <Form form={form} layout="vertical" onFinish={save}>
          <Form.Item label="Name" name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Attribute Key" name="attributeKey" rules={[{ required: true }]}>
            <Input placeholder="department" />
          </Form.Item>
          <Form.Item label="Operator" name="operator" rules={[{ required: true }]}>
            <Input placeholder="eq | ne | in | not_in" />
          </Form.Item>
          <Form.Item label="Value" name="value" rules={[{ required: true }]}>
            <Input placeholder="sales" />
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
