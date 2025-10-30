const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    // Login
    await page.goto('http://localhost:8080/login.html');
    await page.fill('input[type="email"]', 'sarah.cto@acme.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard.html');
    console.log('✅ Login successful');

    // Go to agents page
    await page.goto('http://localhost:8080/agents.html');
    await page.waitForLoadState('networkidle');
    console.log('✅ Agents page loaded');

    // Wait for agents to load
    await page.waitForSelector('text=Claude Code', { timeout: 5000 });
    console.log('✅ Claude Code agent found');

    // Check button text
    const claudeCodeCard = page.locator('text=Claude Code').locator('..').locator('..').locator('..');
    const buttonText = await claudeCodeCard.locator('button').textContent();
    console.log(`📌 Button text: "${buttonText.trim()}"`);

    // Click the button
    console.log('🖱️  Clicking button...');
    await claudeCodeCard.locator('button').click();
    await page.waitForTimeout(1000);

    // Check if modal appeared or if tab switched
    const modalVisible = await page.locator('text=Configure Claude Code').isVisible().catch(() => false);
    const configuredTabActive = await page.locator('button:has-text("Organization Configs")').evaluate(el =>
      el.className.includes('border-blue-500')
    ).catch(() => false);

    console.log(`📊 Modal visible: ${modalVisible}`);
    console.log(`📊 Organization Configs tab active: ${configuredTabActive}`);

    if (modalVisible) {
      console.log('✅ Modal opened - checking for API key field...');
      const apiKeyField = await page.locator('input[type="password"]').isVisible();
      console.log(`📊 API key field visible: ${apiKeyField}`);
    }

    if (configuredTabActive) {
      console.log('⚠️  Switched to Organization Configs tab instead of opening modal');
      console.log('📊 Checking edit mode...');
      const editingConfig = await page.locator('textarea.font-mono').isVisible();
      console.log(`📊 JSON editor visible: ${editingConfig}`);
    }

    await page.waitForTimeout(3000);

  } catch (error) {
    console.error('❌ Error:', error.message);
  } finally {
    await browser.close();
  }
})();
