import React from 'react';

export interface LayoutProps {
  title?: string;
  user?: { displayName?: string };
  children?: React.ReactNode;
}

export function KitLayout({ title, user, children }: LayoutProps) {
  return React.createElement(
    'div',
    { style: { fontFamily: 'sans-serif' } },
    React.createElement(
      'header',
      { style: { padding: 16, borderBottom: '1px solid #eee' } },
      React.createElement('h1', null, title ?? 'PlantX'),
      user && React.createElement('span', null, user.displayName ?? 'User')
    ),
    React.createElement('main', { style: { padding: 16 } }, children)
  );
}

export function UserMenu(props: { onLogout?: () => void }) {
  return React.createElement(
    'button',
    { onClick: props.onLogout },
    'Logout'
  );
}
