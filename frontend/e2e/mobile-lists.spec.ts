import { expect, test } from '@playwright/test'

const apiBaseURL = process.env.PLAYWRIGHT_API_BASE_URL ?? 'http://127.0.0.1:8080/api/v1'
const tokenStorageKey = 'construction-hrms.access-token'

test('shows the compact talent record list instead of the desktop table on mobile', async ({ page, request }) => {
  const loginResponse = await request.post(`${apiBaseURL}/auth/login`, {
    data: { username: 'admin', password: '123456' },
  })
  expect(loginResponse.ok()).toBeTruthy()
  const loginPayload = await loginResponse.json() as { data: { access_token: string } }

  await page.addInitScript(({ key, token }) => {
    sessionStorage.setItem(key, token)
  }, { key: tokenStorageKey, token: loginPayload.data.access_token })

  await page.goto('/#/talents')

  await expect(page.getByRole('heading', { name: '人才证书' })).toBeVisible()
  await expect(page.getByTestId('mobile-talent-list')).toBeVisible()
  await expect(page.getByTestId('desktop-talent-table')).toBeHidden()
  await page.screenshot({ path: 'test-results/mobile-talents.png', fullPage: true })
})

test('shows dashboard reminders and recent talents as compact records on mobile', async ({ page, request }) => {
  const loginResponse = await request.post(`${apiBaseURL}/auth/login`, {
    data: { username: 'admin', password: '123456' },
  })
  expect(loginResponse.ok()).toBeTruthy()
  const loginPayload = await loginResponse.json() as { data: { access_token: string } }

  await page.addInitScript(({ key, token }) => {
    sessionStorage.setItem(key, token)
  }, { key: tokenStorageKey, token: loginPayload.data.access_token })

  await page.goto('/#/dashboard')

  await expect(page.getByTestId('mobile-dashboard-reminders')).toBeVisible()
  await expect(page.getByTestId('mobile-dashboard-talents')).toBeVisible()
  await expect(page.getByTestId('desktop-dashboard-reminders')).toBeHidden()
  await expect(page.getByTestId('desktop-dashboard-talents')).toBeHidden()
})

test('shows companies and delivery orders as compact records on mobile', async ({ page, request }) => {
  const loginResponse = await request.post(`${apiBaseURL}/auth/login`, {
    data: { username: 'admin', password: '123456' },
  })
  expect(loginResponse.ok()).toBeTruthy()
  const loginPayload = await loginResponse.json() as { data: { access_token: string } }

  await page.addInitScript(({ key, token }) => {
    sessionStorage.setItem(key, token)
  }, { key: tokenStorageKey, token: loginPayload.data.access_token })

  await page.goto('/#/companies')
  await expect(page.getByTestId('mobile-company-list')).toBeVisible()
  await expect(page.getByTestId('desktop-company-table')).toBeHidden()
  await page.screenshot({ path: 'test-results/mobile-companies.png', fullPage: true })

  await page.goto('/#/delivery-orders')
  await expect(page.getByTestId('mobile-delivery-order-list')).toBeVisible()
  await expect(page.getByTestId('desktop-delivery-order-table')).toBeHidden()
  await page.screenshot({ path: 'test-results/mobile-delivery-orders.png', fullPage: true })
})
