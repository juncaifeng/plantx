import React, { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, Dropdown, Space, Spin, Typography } from 'antd';
import { AppstoreOutlined, DownOutlined } from '@ant-design/icons';
import type { Application } from '@plantx/kit-sdk-api/registry';
import { useApplications } from '@plantx/kit-sdk-kit';
import { useI18n } from '../i18n';
import { storeApplicationId } from '../applicationStorage';

export interface ProductSwitcherProps {
  value?: Application;
  onChange?: (application: Application) => void;
  variant?: 'header' | 'sider';
}

export function ProductSwitcher({ value, onChange, variant = 'header' }: ProductSwitcherProps) {
  const navigate = useNavigate();
  const { t } = useI18n();
  const { activeApplications, loading, error } = useApplications();

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

  const items = useMemo(() => [allItem, ...activeApplications], [allItem, activeApplications]);

  const handleSelect = (application: Application) => {
    storeApplicationId(application.id || null);
    onChange?.(application);
    navigate('/');
  };

  const label = value?.name ?? t('product.switcher');

  const dropdownRender = () => (
    <Card
      size="small"
      style={{
        width: 320,
        maxHeight: 480,
        overflow: 'auto',
        boxShadow: '0 6px 16px rgba(0,0,0,0.15)',
      }}
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
    </Card>
  );

  const isSider = variant === 'sider';

  return (
    <Dropdown dropdownRender={dropdownRender} placement="bottomLeft" trigger={['click']}>
      <Button
        type="text"
        block={isSider}
        style={{
          color: '#fff',
          justifyContent: isSider ? 'space-between' : undefined,
          height: isSider ? 48 : undefined,
          padding: isSider ? '0 16px' : undefined,
          borderRadius: 0,
        }}
      >
        <Space>
          <AppstoreOutlined />
          <span>{label}</span>
        </Space>
        {!isSider && <DownOutlined />}
      </Button>
    </Dropdown>
  );
}
