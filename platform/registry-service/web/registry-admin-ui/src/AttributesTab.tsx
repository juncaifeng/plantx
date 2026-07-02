import React, { useEffect, useMemo, useState } from 'react';
import { Button, Form, Input, Modal, Space, Table, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { IAMServiceClient, type Attribute } from '@plantx/kit-sdk-api/iam';

export function AttributesTab() {
  const { apiClient } = useKitContext();
  const iamClient = useMemo(() => (apiClient ? new IAMServiceClient(apiClient) : null), [apiClient]);
  const [attributes, setAttributes] = useState<Attribute[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editing, setEditing] = useState<Attribute | null>(null);
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  const list = async () => {
    if (!iamClient) return;
    setLoading(true);
    try {
      const res = await iamClient.listAttributes();
      setAttributes(res.attributes);
    } catch (e) {
      message.error('Failed to load attributes');
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

  const openEdit = (attr: Attribute) => {
    setEditing(attr);
    form.setFieldsValue({ key: attr.key, valueType: attr.valueType, description: attr.description });
    setIsModalOpen(true);
  };

  const save = async (values: Record<string, unknown>) => {
    if (!iamClient) return;
    setSaving(true);
    try {
      if (editing) {
        await iamClient.updateAttribute({
          id: editing.id,
          key: values.key as string,
          valueType: values.valueType as string,
          description: values.description as string,
        });
      } else {
        await iamClient.createAttribute({
          key: values.key as string,
          valueType: values.valueType as string,
          description: values.description as string,
        });
      }
      message.success('Attribute saved');
      form.resetFields();
      setIsModalOpen(false);
      setEditing(null);
      await list();
    } catch (e) {
      message.error('Failed to save attribute');
    } finally {
      setSaving(false);
    }
  };

  const deleteAttr = async (attr: Attribute) => {
    if (!iamClient) return;
    try {
      await iamClient.deleteAttribute({ id: attr.id });
      message.success('Attribute deleted');
      await list();
    } catch (e) {
      message.error('Failed to delete attribute');
    }
  };

  const columns = [
    { title: 'Key', dataIndex: 'key', key: 'key' },
    { title: 'Value Type', dataIndex: 'valueType', key: 'valueType' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, attr: Attribute) => (
        <Space>
          <Button type="link" onClick={() => openEdit(attr)}>Edit</Button>
          <Button danger type="link" onClick={() => deleteAttr(attr)}>Delete</Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Button type="primary" onClick={openCreate}>Create Attribute</Button>
      <Table loading={loading} rowKey="id" dataSource={attributes} columns={columns} />
      <Modal title={editing ? 'Edit Attribute' : 'Create Attribute'} open={isModalOpen} onCancel={() => setIsModalOpen(false)} footer={null}>
        <Form form={form} layout="vertical" onFinish={save}>
          <Form.Item label="Key" name="key" rules={[{ required: true }]}>
            <Input placeholder="department" />
          </Form.Item>
          <Form.Item label="Value Type" name="valueType" rules={[{ required: true }]}>
            <Input placeholder="string" />
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
