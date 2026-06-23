import React, { useCallback, useEffect, useState } from 'react';
import { Button, Input, List, Spin, Typography } from 'antd';
import { useKitContext, useKitPermission, useKitTenant, useKitUser } from '@plantx/kit-sdk-kit';

interface Item {
  id: string;
  tenant_id: string;
  title: string;
  created_at: number;
}

export function DemoHome() {
  const ctx = useKitContext();
  const user = useKitUser();
  const tenant = useKitTenant();
  const canCreate = useKitPermission('item:create');
  const canList = useKitPermission('item:list');

  const [items, setItems] = useState<Item[]>([]);
  const [loading, setLoading] = useState(false);
  const [title, setTitle] = useState('');
  const [creating, setCreating] = useState(false);

  const fetchItems = useCallback(async () => {
    if (!canList || !ctx.apiClient) return;
    setLoading(true);
    try {
      const data = await ctx.apiClient.get<{ items: Item[] }>('/demo/v1/items');
      setItems(data.items ?? []);
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error('failed to load items:', err);
    } finally {
      setLoading(false);
    }
  }, [canList, ctx.apiClient]);

  useEffect(() => {
    fetchItems();
  }, [fetchItems]);

  const handleCreate = async () => {
    if (!title.trim() || !ctx.apiClient) return;
    setCreating(true);
    try {
      await ctx.apiClient.post<Item>('/demo/v1/items', { title: title.trim() });
      setTitle('');
      await fetchItems();
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error('failed to create item:', err);
    } finally {
      setCreating(false);
    }
  };

  return (
    <div>
      <Typography.Title level={4}>Demo Home</Typography.Title>
      <Typography.Text>
        Tenant: {tenant?.name ?? 'unknown'} ({tenant?.id ?? '-'})
      </Typography.Text>
      <br />
      <Typography.Text type="secondary">
        User: {user?.displayName ?? user?.username ?? 'unknown'}
      </Typography.Text>

      {canCreate && (
        <div style={{ marginTop: 16, display: 'flex', gap: 8 }}>
          <Input
            placeholder="New item title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onPressEnter={handleCreate}
            style={{ width: 300 }}
          />
          <Button type="primary" onClick={handleCreate} loading={creating}>
            Create
          </Button>
        </div>
      )}

      <Typography.Title level={5} style={{ marginTop: 24 }}>
        Items
      </Typography.Title>
      {loading ? (
        <Spin />
      ) : (
        <List
          bordered
          dataSource={items}
          renderItem={(item) => (
            <List.Item>
              <Typography.Text strong>{item.title}</Typography.Text>
              <Typography.Text type="secondary" style={{ marginLeft: 16 }}>
                {new Date(item.created_at * 1000).toLocaleString()}
              </Typography.Text>
            </List.Item>
          )}
        />
      )}
    </div>
  );
}
