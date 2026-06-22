import React, { useState } from 'react';
import { Button, Card, Form, Input, message, Space, Typography } from 'antd';
import { useI18n } from './i18n';
import { LocaleSwitch } from './components/LocaleSwitch';

interface LoginPageProps {
  onLogin: (token: string) => void;
}

export function LoginPage({ onLogin }: LoginPageProps) {
  const [loading, setLoading] = useState(false);
  const { t } = useI18n();

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const body = new URLSearchParams();
      body.set('grant_type', 'password');
      body.set('client_id', 'plantx-portal');
      body.set('username', values.username);
      body.set('password', values.password);

      const response = await fetch('/oauth/token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: body.toString(),
      });

      const data = await response.json();
      if (!response.ok || !data.access_token) {
        message.error(data.error_description || data.error || `Login failed (${response.status})`);
        return;
      }

      onLogin(data.access_token as string);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Network error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <div style={{ position: 'absolute', top: 16, right: 16 }}>
        <LocaleSwitch />
      </div>
      <Space direction="vertical" align="center" style={{ width: '100%', paddingTop: 80 }}>
        <Typography.Title level={3}>{t('login.title')}</Typography.Title>
        <Card style={{ width: 360 }}>
          <Form layout="vertical" onFinish={handleSubmit} autoComplete="off">
            <Form.Item
              label={t('login.username')}
              name="username"
              rules={[{ required: true, message: t('login.usernameRequired') }]}
            >
              <Input placeholder={t('login.usernamePlaceholder')} />
            </Form.Item>
            <Form.Item
              label={t('login.password')}
              name="password"
              rules={[{ required: true, message: t('login.passwordRequired') }]}
            >
              <Input.Password placeholder={t('login.passwordPlaceholder')} />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} block>
                {t('login.submit')}
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </Space>
    </>
  );
}
