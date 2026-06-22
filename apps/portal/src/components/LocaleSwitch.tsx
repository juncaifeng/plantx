import React from 'react';
import { Button, Dropdown } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';
import { useI18n, LOCALE_LABELS, type Locale } from '../i18n';

export function LocaleSwitch() {
  const { locale, setLocale } = useI18n();

  const items = Object.entries(LOCALE_LABELS).map(([key, label]) => ({
    key,
    label,
    onClick: () => setLocale(key as Locale),
  }));

  return (
    <Dropdown menu={{ items, selectedKeys: [locale] }} placement="bottomRight">
      <Button icon={<GlobalOutlined />}>{LOCALE_LABELS[locale]}</Button>
    </Dropdown>
  );
}
