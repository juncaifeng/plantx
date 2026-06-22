import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Modal, Space, Table, Typography, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { GatewayServiceClient, type Service, type Route } from '@plantx/kit-sdk-api/gateway';

export function GatewayAdminPage() {
  const { apiClient } = useKitContext();
  const gatewayClient = useMemo(() => (apiClient ? new GatewayServiceClient(apiClient) : null), [apiClient]);
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState(false);
  const [registering, setRegistering] = useState(false);
  const [isRegisterModalOpen, setIsRegisterModalOpen] = useState(false);
  const [isRoutesModalOpen, setIsRoutesModalOpen] = useState(false);
  const [selectedRoutes, setSelectedRoutes] = useState<Route[]>([]);
  const [selectedServiceName, setSelectedServiceName] = useState('');
  const [form] = Form.useForm();

  const list = async () => {
    if (!gatewayClient) return;
    setLoading(true);
    try {
      const res = await gatewayClient.listServices();
      setServices(res.services);
    } catch (e) {
      message.error('Failed to load services');
      // eslint-disable-next-line no-console
      console.error('failed to list services', e);
    } finally {
      setLoading(false);
    }
  };

  const register = async (values: { name: string; grpc_host: string; rest_prefix: string }) => {
    if (!gatewayClient) return;
    setRegistering(true);
    try {
      const created = await gatewayClient.registerService({
        name: values.name.trim(),
        grpc_host: values.grpc_host.trim(),
        rest_prefix: values.rest_prefix.trim(),
      });
      setServices((prev) => [created, ...prev]);
      form.resetFields();
      setIsRegisterModalOpen(false);
      message.success('Service registered');
    } catch (e) {
      message.error('Failed to register service');
      // eslint-disable-next-line no-console
      console.error('failed to register service', e);
    } finally {
      setRegistering(false);
    }
  };

  const showRoutes = (svc: Service) => {
    const routes = svc.routes?.length ? svc.routes : [{ path: svc.rest_prefix, method: '*' }];
    setSelectedRoutes(routes);
    setSelectedServiceName(svc.name);
    setIsRoutesModalOpen(true);
  };

  useEffect(() => {
    list();
  }, [gatewayClient]);

  return (
    <Card title="Service Registry">
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Button type="primary" onClick={() => setIsRegisterModalOpen(true)}>
          Register Service
        </Button>

        <Table
          loading={loading}
          rowKey="id"
          dataSource={services}
          columns={[
            { title: 'Name', dataIndex: 'name', key: 'name' },
            { title: 'gRPC Host', dataIndex: 'grpc_host', key: 'grpc_host' },
            { title: 'REST Prefix', dataIndex: 'rest_prefix', key: 'rest_prefix' },
            {
              title: 'Routes',
              key: 'routes',
              render: (_, svc) => (
                <Button type="link" onClick={() => showRoutes(svc)}>
                  View
                </Button>
              ),
            },
          ]}
        />
      </Space>

      <Modal
        title="Register Service"
        open={isRegisterModalOpen}
        onCancel={() => setIsRegisterModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={register}>
          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter service name' }]}
          >
            <Input placeholder="order-service" />
          </Form.Item>
          <Form.Item
            label="gRPC Host"
            name="grpc_host"
            rules={[{ required: true, message: 'Please enter gRPC host' }]}
          >
            <Input placeholder="order-service:8080" />
          </Form.Item>
          <Form.Item
            label="REST Prefix"
            name="rest_prefix"
            rules={[{ required: true, message: 'Please enter REST prefix' }]}
          >
            <Input placeholder="/api/order/" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={registering}>
                Register
              </Button>
              <Button onClick={() => setIsRegisterModalOpen(false)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`Routes for ${selectedServiceName}`}
        open={isRoutesModalOpen}
        onCancel={() => setIsRoutesModalOpen(false)}
        footer={null}
      >
        {selectedRoutes.map((route, idx) => (
          <Typography.Paragraph key={idx}>
            <Typography.Text code>{route.method}</Typography.Text>{' '}
            <Typography.Text>{route.path}</Typography.Text>
          </Typography.Paragraph>
        ))}
      </Modal>
    </Card>
  );
}
