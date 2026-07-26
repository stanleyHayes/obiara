import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

const routes = [
  ["/fie", "Akwaaba home, Ama."],
  ["/fie/abonten", "Step outside. Stay yourself."],
  ["/fie/adiwo", "Familiar people. Shared purpose."],
  ["/fie/epono-ano", "Pause before you open."],
  ["/fie/dan-mu", "Private means private."],
  ["/fie/garden", "Sow with intention."],
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

test("private room detail stays accessible and bounded", async ({ page }) => {
  await page.goto("/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa");
  await expect(
    page.getByRole("heading", { name: "Make room for honesty." }),
  ).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth,
    ),
  ).toBe(false);
});

test("Abusua Gate requires mutual consent and stays bounded", async ({
  page,
}) => {
  await page.goto("/fie/abusua-gate");
  await expect(
    page.getByRole("heading", { name: "Open one careful window." }),
  ).toBeVisible();
  const issue = page.getByRole("button", {
    name: "Create private reviewer passage",
  });
  await expect(issue).toBeDisabled();
  await page.getByRole("button", { name: "Preview mutual consent" }).click();
  await expect(issue).toBeEnabled();
  await issue.click();
  await expect(
    page.getByRole("heading", {
      name: "The gate is open for one visit.",
    }),
  ).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth,
    ),
  ).toBe(false);
});

test("Fire degradation keeps captions, safety and leave available", async ({
  page,
}) => {
  await page.goto("/fie/fires/fire_7Qp9kL2xV4mN8zTa");
  await expect(
    page.getByRole("heading", { name: "Stories we inherited." }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Use audio only" }).click();
  await expect(page.getByText("Audio only · using less data.")).toBeVisible();
  await page.getByRole("button", { name: "Safety" }).click();
  await expect(
    page.getByRole("heading", {
      name: "Help stays available in every mode.",
    }),
  ).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth,
    ),
  ).toBe(false);
});
