import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  message,
} from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import {
  RegistryServiceClient,
  type Application,
  type ApplicationStatus,
} from '@plantx/kit-sdk-api/registry';

const STATUS_OPTIONS: { label: string; value: ApplicationStatus }[] = [
  { label: 'Active', value: 'APPLICATION_STATUS_ACTIVE' },
  { label: 'Offline', value: 'APPLICATION_STATUS_OFFLINE' },
];

export function ApplicationsTab() {
  const { apiClient } = useKitContext();
  const registryClient = useMemo(
    () => (apiClient ? new RegistryServiceClient(apiClient) : null),
    [apiClient]
  );

  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<Application | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [form] = Form.useForm();

  const load = async () => {
    if (!registryClient) return;
    setLoading(true);
    try {
      const res = await registryClient.listApplications();
      setApplications(res.applications);
    } catch (e) {
      message.error('Failed to load applications');
      // eslint-disable-next-line no-console
      console.error('failed to list applications', e);
    } finally {
      setLoading(false);
    }
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ status: 'APPLICATION_STATUS_ACTIVE', sortOrder: 0 });
    setIsModalOpen(true);
  };

  const openEdit = (app: Application) => {
    setEditing(app);
    form.setFieldsValue({
      key: app.key,
      name: app.name,
      labelKey: app.labelKey,
      icon: app.icon,
      description: app.description,
      status: app.status,
      sortOrder: app.sortOrder,
    });
    setIsModalOpen(true);
  };

  const save = async (values: Record<string, unknown>) => {
    if (!registryClient) return;
    setSaving(true);
    try {
      if (editing) {
        const updated = await registryClient.updateApplication({
          id: editing.id,
          key: values.key as string,
          name: values.name as string,
          labelKey: values.labelKey as string,
          icon: values.icon as string | undefined,
          description: values.description as string | undefined,
          status: values.status as ApplicationStatus | undefined,
          sortOrder: values.sortOrder as number | undefined,
        });
        setApplications((prev) => prev.map((a) => (a.id === updated.id ? updated : a)));
        message.success('Application updated');
      } else {
        const created = await registryClient.registerApplication({
          key: values.key as string,
          name: values.name as string,
          labelKey: values.labelKey as string,
          icon: values.icon as string | undefined,
          description: values.description as string | undefined,
          status: values.status as ApplicationStatus | undefined,
          sortOrder: values.sortOrder as number | undefined,
        });
        setApplications((prev) => [...prev, created]);
        message.success('Application created');
      }
      form.resetFields();
      setIsModalOpen(false);
      setEditing(null);
    } catch (e) {
      message.error(editing ? 'Failed to update application' : 'Failed to create application');
      // eslint-disable-next-line no-console
      console.error('failed to save application', e);
    } finally {
      setSaving(false);
    }
  };

  const deleteApplication = async (app: Application) => {
    if (!registryClient) return;
    try {
      await registryClient.deleteApplication({ id: app.id });
      setApplications((prev) => prev.filter((a) => a.id !== app.id));
      message.success('Application deleted');
    } catch (e) {
      message.error('Failed to delete application');
      // eslint-disable-next-line no-console
      console.error('failed to delete application', e);
    }
  };

  useEffect(() => {
    load();
  }, [registryClient]);

  const columns = [
    { title: 'Key', dataIndex: 'key', key: 'key' },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Label Key', dataIndex: 'labelKey', key: 'labelKey' },
    { title: 'Icon', dataIndex: 'icon', key: 'icon' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: ApplicationStatus) =>
        status === 'APPLICATION_STATUS_ACTIVE' ? <Tag color="green">Active</Tag> : <Tag>Offline</Tag>,
    },
    { title: 'Sort Order', dataIndex: 'sortOrder', key: 'sortOrder' },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, app: Application) => (
        <Space>
          <Button type="link" onClick={() => openEdit(app)}>
            Edit
          </Button>
          <Button danger type="link" onClick={() => deleteApplication(app)}>
            Delete
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Card title="Applications">
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Button type="primary" onClick={openCreate}>
          Create Application
        </Button>
        <Table
          loading={loading}
          rowKey="id"
          dataSource={applications}
          columns={columns}
          pagination={{ pageSize: 10 }}
        />
      </Space>

      <Modal
        title={editing ? 'Edit Application' : 'Create Application'}
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={save}>
          <Form.Item
            label="Key"
            name="key"
            rules={[{ required: true, message: 'Please enter application key' }]}
          >
            <Input placeholder="order" disabled={!!editing} />
          </Form.Item>
          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter application name' }]}
          >
            <Input placeholder="Order Management" />
          </Form.Item>
          <Form.Item
            label="Label Key"
            name="labelKey"
            rules={[{ required: true, message: 'Please enter label key' }]}
          >
            <Input placeholder="nav.orders" />
          </Form.Item>
          <Form.Item label="Icon" name="icon">
            <Input placeholder="AppstoreOutlined" />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <Input.TextArea rows={3} placeholder="Short description" />
          </Form.Item>
          <Form.Item
            label="Status"
            name="status"
            rules={[{ required: true, message: 'Please select status' }]}
          >
            <Select options={STATUS_OPTIONS} placeholder="Select status" />
          </Form.Item>
          <Form.Item label="Sort Order" name="sortOrder">
            <InputNumber style={{ width: '100%' }} placeholder="0" />
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
