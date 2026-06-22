import React, { useEffect, useState } from 'react';
import { Button, Card, Input, List, Space, Tag, Typography, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import type { Order, CreateOrderRequest } from '@plantx/order-sdk-api';

export function OrderPage() {
  const { apiClient } = useKitContext();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [name, setName] = useState('');

  const list = async () => {
    if (!apiClient) return;
    setLoading(true);
    try {
      const res = await apiClient.get<{ orders: Order[] }>('/order/v1/orders');
      setOrders(res.orders);
    } catch (e) {
      message.error('Failed to load orders');
      // eslint-disable-next-line no-console
      console.error('failed to list orders', e);
    } finally {
      setLoading(false);
    }
  };

  const create = async () => {
    if (!apiClient || !name.trim()) return;
    setCreating(true);
    try {
      const req: CreateOrderRequest = { customer_name: name.trim() };
      const created = await apiClient.post<Order>('/order/v1/orders', req);
      setOrders((prev) => [created, ...prev]);
      setName('');
      message.success('Order created');
    } catch (e) {
      message.error('Failed to create order');
      // eslint-disable-next-line no-console
      console.error('failed to create order', e);
    } finally {
      setCreating(false);
    }
  };

  useEffect(() => {
    list();
  }, [apiClient]);

  return (
    <Card title="Orders">
      <Space.Compact style={{ width: '100%', maxWidth: 480, marginBottom: 24 }}>
        <Input
          placeholder="Customer name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onPressEnter={create}
        />
        <Button type="primary" onClick={create} loading={creating}>
          Create Order
        </Button>
      </Space.Compact>

      <List
        loading={loading}
        bordered
        dataSource={orders}
        renderItem={(order) => (
          <List.Item>
            <Space>
              <Typography.Text strong>{order.customer_name}</Typography.Text>
              <Tag color="blue">{order.status}</Tag>
              <Typography.Text type="secondary">{order.id}</Typography.Text>
            </Space>
          </List.Item>
        )}
      />
    </Card>
  );
}
