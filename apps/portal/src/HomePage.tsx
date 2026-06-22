import React from 'react';
import { Card, Descriptions, Typography } from 'antd';
import { useKitContext } from '@plantx/kit-sdk-kit';
import { useI18n } from './i18n';

export function HomePage() {
  const ctx = useKitContext();
  const { t } = useI18n();
  const displayName = ctx.user?.displayName ?? ctx.user?.username ?? 'User';

  return (
    <Card>
      <Typography.Title level={4}>{t('home.welcome', { name: displayName })}</Typography.Title>
      <Descriptions bordered column={1}>
        <Descriptions.Item label={t('home.userId')}>{ctx.user?.id}</Descriptions.Item>
        <Descriptions.Item label={t('home.tenant')}>{ctx.tenant?.name ?? ctx.tenant?.id}</Descriptions.Item>
        <Descriptions.Item label={t('home.roles')}>{ctx.user?.roles?.join(', ') ?? '-'}</Descriptions.Item>
        <Descriptions.Item label={t('home.permissions')}>
          {ctx.permissions?.join(', ') ?? '-'}
        </Descriptions.Item>
      </Descriptions>
    </Card>
  );
}
