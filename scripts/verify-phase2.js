const { chromium } = require('playwright');
const assert = require('assert');

const BASE_URL = process.env.BASE_URL || 'http://localhost/';
const USERNAME = process.env.USERNAME || 'admin';
const PASSWORD = process.env.PASSWORD || 'admin';

(async () => {
  const headless = process.env.CI === 'true' || process.env.HEADLESS === '1';
  const browser = await chromium.launch({ headless });
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    console.log(`Navigating to ${BASE_URL}`);
    await page.goto(BASE_URL);

    console.log('Logging in...');
    await page.getByLabel('用户名').fill(USERNAME);
    await page.getByLabel('密码').fill(PASSWORD);
    await page.getByRole('button', { name: '登录' }).click();

    console.log('Waiting for top menu...');
    await page.getByRole('menu').waitFor({ timeout: 10000 });

    console.log('Asserting menu items...');
    await page.getByRole('menuitem', { name: '订单' }).waitFor({ timeout: 5000 });
    await page.getByRole('menuitem', { name: '测试服务' }).waitFor({ timeout: 5000 });

    const menuText = await page.getByRole('menu').textContent();
    assert(menuText.includes('订单'), 'Menu should contain 订单');
    assert(menuText.includes('测试服务'), 'Menu should contain 测试服务');

    console.log('Clicking 测试服务...');
    await page.getByRole('menuitem', { name: '测试服务' }).click();

    console.log('Waiting for test-ui micro-app to load...');
    await page.waitForURL(/\/test$/, { timeout: 10000 });
    await page.getByText('Test Service').waitFor({ timeout: 10000 });

    const heading = await page.getByText('Test Service').textContent();
    assert(heading.includes('Test Service'), 'test-ui heading should be visible');

    console.log('Phase 2 verification passed.');
  } catch (err) {
    console.error('Phase 2 verification failed:', err);
    process.exitCode = 1;
    throw err;
  } finally {
    await browser.close();
  }
})();
