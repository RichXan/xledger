const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 } });
  
  console.log('正在访问前端页面...');
  await page.goto('http://localhost:4175');
  await page.waitForTimeout(3000);
  
  // 截图
  await page.screenshot({ path: '/tmp/frontend-login.png', fullPage: true });
  console.log('登录页面截图已保存: /tmp/frontend-login.png');
  
  // 获取页面内容
  const title = await page.title();
  const url = page.url();
  console.log('页面标题:', title);
  console.log('当前URL:', url);
  
  await browser.close();
})();
