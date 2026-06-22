import React, { useEffect, useState } from 'react';
import { Button, Card, Input, Space, Table, Typography, message } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { AuditServiceClient, type AuditLog } from '@plantx/kit-sdk-api/audit';

export function AuditAdminPage() {
  const { apiClient } = useKitContext();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [tenantFilter, setTenantFilter] = useState('');

  const query = async () => {
    if (!apiClient) return;
    setLoading(true);
    try {
      const client = new AuditServiceClient(apiClient);
      const res = await client.listAuditLogs({
        tenant_id: tenantFilter.trim() || undefined,
      });
      setLogs(res.logs);
    } catch (e) {
      message.error('Failed to load audit logs');
      // eslint-disable-next-line no-console
      console.error('failed to query audit logs', e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    query();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiClient]);

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      ellipsis: true,
      render: (id: string) => <Typography.Text copyable={{ text: id }}>{id.slice(0, 8)}</Typography.Text>,
    },
    {
      title: 'Tenant',
      dataIndex: 'tenant_id',
      key: 'tenant_id',
    },
    {
      title: 'User',
      dataIndex: 'user_id',
      key: 'user_id',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: 'Resource',
      dataIndex: 'resource',
      key: 'resource',
    },
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (ts: number | string) => {
        const n = typeof ts === 'number' ? ts : parseInt(ts, 10);
        return new Date(n * 1000).toLocaleString();
      },
    },
    {
      title: 'Detail',
      dataIndex: 'detail',
      key: 'detail',
      ellipsis: true,
    },
  ];

  return (
    <Card title="Audit Log Explorer">
      <Space.Compact style={{ width: '100%', maxWidth: 480, marginBottom: 24 }}>
        <Input
          placeholder="Filter by tenant id"
          value={tenantFilter}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTenantFilter(e.target.value)}
          onPressEnter={query}
        />
        <Button type="primary" onClick={query}>
          Query
        </Button>
      </Space.Compact>

      <Table
        rowKey="id"
        loading={loading}
        dataSource={logs}
        columns={columns}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  );
}
