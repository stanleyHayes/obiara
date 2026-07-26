import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

test("design lab remains accessible and visually stable", async ({ page }) => {
  await page.goto("/design-lab");
  await expect(
    page.getByRole("heading", {
      name: "Four gestures. Every one has another way.",
    }),
  ).toBeVisible();

  const accessibility = await new AxeBuilder({ page }).analyze();
  expect(accessibility.violations).toEqual([]);
  await expect(page).toHaveScreenshot("design-lab.png", { fullPage: true });
});
