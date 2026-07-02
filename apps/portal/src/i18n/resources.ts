export type Locale = 'zh-CN' | 'en-US';

export const DEFAULT_LOCALE: Locale = 'zh-CN';

export const LOCALE_STORAGE_KEY = 'plantx-locale';

export const LOCALE_LABELS: Record<Locale, string> = {
  'zh-CN': '简体中文',
  'en-US': 'English',
};

export type TranslationKey =
  | 'nav.home'
  | 'nav.orders'
  | 'nav.test'
  | 'nav.admin'
  | 'nav.tenants'
  | 'nav.iam'
  | 'nav.gateway'
  | 'nav.audit'
  | 'nav.registry'
  | 'nav.registry.applications'
  | 'nav.registry.menus'
  | 'nav.registry.routes'
  | 'nav.registry.permissions'
  | 'nav.registry.attributes'
  | 'nav.registry.conditions'
  | 'nav.registry.policies'
  | 'nav.apiExplorer'
  | 'header.logout'
  | 'header.language'
  | 'login.title'
  | 'login.username'
  | 'login.usernamePlaceholder'
  | 'login.password'
  | 'login.passwordPlaceholder'
  | 'login.submit'
  | 'login.usernameRequired'
  | 'login.passwordRequired'
  | 'home.welcome'
  | 'home.userId'
  | 'home.tenant'
  | 'home.roles'
  | 'home.permissions'
  | 'orders.loading'
  | 'microapp.loading'
  | 'product.switcher'
  | 'product.all'
  | 'product.loading'
  | 'product.empty';

type Resources = Record<Locale, Record<TranslationKey, string>>;

export const resources: Resources = {
  'zh-CN': {
    'nav.home': '首页',
    'nav.orders': '订单',
    'nav.test': '测试服务',
    'nav.admin': '管理',
    'nav.tenants': '租户',
    'nav.iam': 'IAM',
    'nav.gateway': '网关',
    'nav.audit': '审计',
    'nav.registry': '注册中心',
    'nav.registry.applications': '应用管理',
    'nav.registry.menus': '菜单管理',
    'nav.registry.routes': '路由管理',
    'nav.registry.permissions': '权限管理',
    'nav.registry.attributes': '属性管理',
    'nav.registry.conditions': '条件管理',
    'nav.registry.policies': '策略管理',
    'nav.apiExplorer': 'API 浏览器',
    'header.logout': '退出登录',
    'header.language': '语言',
    'login.title': 'PlantX 门户',
    'login.username': '用户名',
    'login.usernamePlaceholder': 'demo-a',
    'login.password': '密码',
    'login.passwordPlaceholder': 'demo-a',
    'login.submit': '登录',
    'login.usernameRequired': '请输入用户名！',
    'login.passwordRequired': '请输入密码！',
    'home.welcome': '欢迎，{name}',
    'home.userId': '用户 ID',
    'home.tenant': '租户',
    'home.roles': '角色',
    'home.permissions': '权限',
    'orders.loading': '正在加载 order-ui...',
    'microapp.loading': '正在加载 {name}...',
    'product.switcher': '产品',
    'product.all': '全部产品',
    'product.loading': '加载中...',
    'product.empty': '无产品',
  },
  'en-US': {
    'nav.home': 'Home',
    'nav.orders': 'Orders',
    'nav.test': 'Test Service',
    'nav.admin': 'Admin',
    'nav.tenants': 'Tenants',
    'nav.iam': 'IAM',
    'nav.gateway': 'Gateway',
    'nav.audit': 'Audit',
    'nav.registry': 'Registry',
    'nav.registry.applications': 'Applications',
    'nav.registry.menus': 'Menus',
    'nav.registry.routes': 'Routes',
    'nav.registry.permissions': 'Permissions',
    'nav.registry.attributes': 'Attributes',
    'nav.registry.conditions': 'Conditions',
    'nav.registry.policies': 'Policies',
    'nav.apiExplorer': 'API Explorer',
    'header.logout': 'Logout',
    'header.language': 'Language',
    'login.title': 'PlantX Portal',
    'login.username': 'Username',
    'login.usernamePlaceholder': 'demo-a',
    'login.password': 'Password',
    'login.passwordPlaceholder': 'demo-a',
    'login.submit': 'Login',
    'login.usernameRequired': 'Please input your username!',
    'login.passwordRequired': 'Please input your password!',
    'home.welcome': 'Welcome, {name}',
    'home.userId': 'User ID',
    'home.tenant': 'Tenant',
    'home.roles': 'Roles',
    'home.permissions': 'Permissions',
    'orders.loading': 'Loading order-ui...',
    'microapp.loading': 'Loading {name}...',
    'product.switcher': 'Products',
    'product.all': 'All Products',
    'product.loading': 'Loading...',
    'product.empty': 'No products',
  },
};

export function getInitialLocale(): Locale {
  const stored = typeof window !== 'undefined' ? localStorage.getItem(LOCALE_STORAGE_KEY) : null;
  if (stored === 'zh-CN' || stored === 'en-US') {
    return stored;
  }
  const browser = typeof navigator !== 'undefined' ? navigator.language : '';
  if (browser.startsWith('zh')) {
    return 'zh-CN';
  }
  return DEFAULT_LOCALE;
}
