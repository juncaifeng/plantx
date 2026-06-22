import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import enUS from 'antd/locale/en_US';
import { App } from './App';
import { I18nProvider, useI18n, type Locale } from './i18n';
import 'antd/dist/reset.css';

const antdLocales: Record<Locale, typeof zhCN> = {
  'zh-CN': zhCN,
  'en-US': enUS,
};

function AppRoot() {
  const { locale } = useI18n();
  return (
    <ConfigProvider locale={antdLocales[locale]}>
      <App />
    </ConfigProvider>
  );
}

const root = document.getElementById('root');
if (root) {
  ReactDOM.createRoot(root).render(
    <I18nProvider>
      <AppRoot />
    </I18nProvider>
  );
}
