import assert from "node:assert/strict";
import test from "node:test";

import { Linter } from "eslint";

import { productQualityConfig } from "../product-quality.mjs";

const linter = new Linter();

function verify(source) {
  return linter.verify(
    source,
    [
      {
        files: ["**/*.jsx"],
        languageOptions: {
          ecmaVersion: "latest",
          parserOptions: { ecmaFeatures: { jsx: true }, sourceType: "module" },
        },
        ...productQualityConfig({ enforcement: "error" }),
      },
    ],
    { filename: "fixture.jsx" },
  );
}

test("reports hardcoded product copy and missing accessible names", () => {
  const messages = verify(`
    export const Broken = () => (
      <section>
        <img src="/person.png" />
        <button><span /></button>
        <p>Save your profile</p>
      </section>
    );
  `);

  assert.ok(
    messages.some(
      ({ ruleId }) => ruleId === "obiara-quality/no-hardcoded-product-copy",
    ),
  );
  assert.ok(
    messages.some(({ ruleId }) => ruleId === "obiara-quality/accessible-names"),
  );
});

test("reports pressure copy and mixed register", () => {
  const messages = verify(`export const message = "Kindly don't miss out!!";`);
  assert.ok(
    messages.some(({ ruleId }) => ruleId === "obiara-quality/no-pressure-copy"),
  );
  assert.ok(
    messages.some(
      ({ ruleId }) => ruleId === "obiara-quality/no-mixed-register",
    ),
  );
});

test("accepts localized, accessible controls", () => {
  assert.deepEqual(
    verify(`
      export const Good = ({ t }) => (
        <section>
          <img src="/shape.png" alt="" />
          <button aria-label={t("action.save")}><Icon /></button>
          <p>{t("profile.summary")}</p>
        </section>
      );
    `),
    [],
  );
});
