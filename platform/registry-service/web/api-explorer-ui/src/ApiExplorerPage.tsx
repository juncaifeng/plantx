import React, { useEffect, useState } from 'react';
import { Card, Select, Spin, Typography, message } from 'antd';
import SwaggerUI from 'swagger-ui-react';
import 'swagger-ui-react/swagger-ui.css';
import YAML from 'yaml';

const SPEC_FILES = [
  'openapi/audit.yaml',
  'openapi/gateway.yaml',
  'openapi/iam.yaml',
  'openapi/order.yaml',
  'openapi/registry.yaml',
  'openapi/tenant.yaml',
  'openapi/test.yaml',
];

export function ApiExplorerPage() {
  const [selected, setSelected] = useState<string>(SPEC_FILES[0]);
  const [spec, setSpec] = useState<unknown | null>(null);
  const [loading, setLoading] = useState(false);

  const loadSpec = async (path: string) => {
    setLoading(true);
    try {
      const res = await fetch(`/${path}`);
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
      }
      const text = await res.text();
      const parsed = YAML.parse(text);
      setSpec(parsed);
    } catch (e) {
      message.error('Failed to load OpenAPI spec');
      // eslint-disable-next-line no-console
      console.error('failed to load spec', e);
      setSpec(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadSpec(selected);
  }, [selected]);

  return (
    <Card
      title="API Explorer"
      extra={
        <Select
          value={selected}
          onChange={setSelected}
          options={SPEC_FILES.map((f) => ({
            label: f.replace('openapi/', '').replace('.yaml', ''),
            value: f,
          }))}
          style={{ minWidth: 200 }}
        />
      }
    >
      {loading ? (
        <Spin tip="Loading spec..." />
      ) : spec ? (
        <SwaggerUI spec={spec} />
      ) : (
        <Typography.Text type="secondary">Select a spec to explore.</Typography.Text>
      )}
    </Card>
  );
}
