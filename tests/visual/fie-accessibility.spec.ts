import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

const routes = [
  ["/fie", "Akwaaba home, Ama."],
  ["/fie/abonten", "Step outside. Stay yourself."],
  ["/fie/adiwo", "Familiar people. Shared purpose."],
  ["/fie/epono-ano", "Pause before you open."],
  ["/fie/dan-mu", "Private means private."],
  ["/fie/okyeame", "Help should know its place."],
] as const;

for (const [path, heading] of routes) {
  test(`${path} keeps the Fie shell accessible`, async ({ page }) => {
    await page.goto(path);
    await expect(page.getByRole("heading", { name: heading })).toBeVisible();

    const navigation = page.getByRole("navigation", {
      name: /compound navigation/i,
    });
    await expect(navigation).toBeVisible();

    const accessibility = await new AxeBuilder({ page }).analyze();
    expect(accessibility.violations).toEqual([]);

    const horizontalOverflow = await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth,
    );
    expect(horizontalOverflow).toBe(false);
  });
}
