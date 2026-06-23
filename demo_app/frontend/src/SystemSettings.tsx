import React from 'react';
import { Alert, Typography } from 'antd';
import { useKitPermission, useKitUser } from '@plantx/kit-sdk-kit';

export function SystemSettings() {
  const user = useKitUser();
  const isAdmin = useKitPermission('setting:admin');

  return (
    <div>
      <Typography.Title level={4}>System Settings</Typography.Title>
      <Typography.Text type="secondary">
        Current user: {user?.displayName ?? user?.username ?? 'unknown'}
      </Typography.Text>

      {isAdmin ? (
        <Alert
          message="Platform Admin Area"
          description="This page is only visible to users with the setting:admin permission. It represents an ABAC-protected resource where platform-level attributes (role, tenant scope) control access."
          type="success"
          showIcon
          style={{ marginTop: 16 }}
        />
      ) : (
        <Alert
          message="Restricted"
          description="You do not have setting:admin permission. This demonstrates RBAC denial."
          type="warning"
          showIcon
          style={{ marginTop: 16 }}
        />
      )}
    </div>
  );
}
