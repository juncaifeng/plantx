import React, { useState } from 'react';
import { Button, Card, Descriptions, Input, Space, Typography, message } from 'antd';
import { useKitContext, useKitTenant, useKitUser } from '@plantx/kit-sdk-kit';
import { TestServiceClient, type EchoResponse } from '@plantx/test-sdk-api';

export function TestPage() {
  const { apiClient } = useKitContext();
  const user = useKitUser();
  const tenant = useKitTenant();
  const [input, setInput] = useState('');
  const [result, setResult] = useState<EchoResponse | null>(null);
  const [loading, setLoading] = useState(false);

  const echo = async () => {
    if (!apiClient || !input.trim()) return;
    setLoading(true);
    try {
      const client = new TestServiceClient(apiClient);
      const res = await client.echo({ message: input.trim() });
      setResult(res);
      message.success('Echo received');
    } catch (e) {
      message.error('Echo failed');
      // eslint-disable-next-line no-console
      console.error('failed to echo', e);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Space direction="vertical" size="middle" style={{ display: 'flex' }}>
      <Card title="Identity">
        <Descriptions size="small" column={2}>
          <Descriptions.Item label="User">{user?.displayName ?? user?.username ?? user?.id ?? '-'}</Descriptions.Item>
          <Descriptions.Item label="Tenant">{tenant?.name ?? tenant?.id ?? '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="Echo">
        <Space.Compact style={{ width: '100%', maxWidth: 480, marginBottom: 24 }}>
          <Input
            placeholder="Type a message"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onPressEnter={echo}
          />
          <Button type="primary" onClick={echo} loading={loading}>
            Echo
          </Button>
        </Space.Compact>

        {result && (
          <Descriptions size="small" column={1} bordered>
            <Descriptions.Item label="Message">{result.message}</Descriptions.Item>
            <Descriptions.Item label="User ID">{result.userId || '-'}</Descriptions.Item>
            <Descriptions.Item label="Tenant ID">{result.tenantId || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
    </Space>
  );
}
