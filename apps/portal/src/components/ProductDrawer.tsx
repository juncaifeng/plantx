import React, { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Drawer, Space, Spin, Typography } from 'antd';
import { AppstoreOutlined } from '@ant-design/icons';
import type { Application } from '@plantx/kit-sdk-api/registry';
import { useI18n } from '../i18n';

export interface ProductDrawerProps {
  applications: Application[];
  loading?: boolean;
  error?: Error | null;
  value?: Application;
  onChange?: (application: Application) => void;
  open?: boolean;
  onClose?: () => void;
}

export function ProductDrawer({
  applications,
  loading,
  error,
  value,
  onChange,
  open,
  onClose,
}: ProductDrawerProps) {
  const navigate = useNavigate();
  const { t } = useI18n();

  const allItem: Application = useMemo(
    () => ({
      id: '',
      key: '',
      name: t('product.all'),
      labelKey: 'product.all',
      status: 'APPLICATION_STATUS_ACTIVE',
      sortOrder: -1,
    }),
    [t]
  );

  const items = useMemo(
    () => [allItem, ...applications],
    [allItem, applications]
  );

  const handleSelect = (application: Application) => {
    onChange?.(application);
    onClose?.();
    navigate('/');
  };

  return (
    <Drawer
      title={t('product.switcher')}
      placement="left"
      width={320}
      open={open}
      onClose={onClose}
      bodyStyle={{ padding: 12 }}
    >
      {loading && (
        <Space style={{ justifyContent: 'center', width: '100%', padding: 16 }}>
          <Spin size="small" />
          <Typography.Text type="secondary">{t('product.loading')}</Typography.Text>
        </Space>
      )}
      {!loading && error && (
        <Typography.Text type="danger">{error.message}</Typography.Text>
      )}
      {!loading && !error && items.length === 0 && (
        <Typography.Text type="secondary">{t('product.empty')}</Typography.Text>
      )}
      {!loading && !error && (
        <Space direction="vertical" style={{ width: '100%' }}>
          {items.map((app) => (
            <Button
              key={app.id || 'all'}
              type={app.id === value?.id ? 'primary' : 'text'}
              block
              icon={<AppstoreOutlined />}
              style={{
                justifyContent: 'flex-start',
                height: 'auto',
                padding: '8px 12px',
              }}
              onClick={() => handleSelect(app)}
            >
              <div style={{ textAlign: 'left' }}>
                <div>{app.name}</div>
                {app.description && (
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {app.description}
                  </Typography.Text>
                )}
              </div>
            </Button>
          ))}
        </Space>
      )}
    </Drawer>
  );
}
