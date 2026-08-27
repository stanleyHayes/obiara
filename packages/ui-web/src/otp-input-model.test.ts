import assert from "node:assert/strict";
import test from "node:test";

import {
  clearOtpDigit,
  createOtpCells,
  removeOtpDigit,
  replaceOtpDigits,
  sanitizeOtp,
  serializeOtpCells,
} from "./otp-input-model.ts";

test("sanitizes pasted and autofilled codes", () => {
  assert.equal(sanitizeOtp("12 3-45a67", 6), "123456");
});

test("inserts partial codes from the active digit", () => {
  assert.deepEqual(replaceOtpDigits(createOtpCells("123456", 6), "90", 2, 6), [
    "1",
    "2",
    "9",
    "0",
    "5",
    "6",
  ]);
  assert.deepEqual(replaceOtpDigits(createOtpCells("123456", 6), "90", 4, 6), [
    "1",
    "2",
    "3",
    "4",
    "9",
    "0",
  ]);
});

test("a full code pasted into any cell replaces all cells", () => {
  assert.deepEqual(
    replaceOtpDigits(createOtpCells("123456", 6), "908172", 4, 6),
    ["9", "0", "8", "1", "7", "2"],
  );
});

test("backspace preserves positional cells without compacting later digits", () => {
  const cleared = removeOtpDigit(createOtpCells("123456", 6), 2);
  assert.deepEqual(cleared, {
    cells: ["1", "2", "", "4", "5", "6"],
    focusIndex: 2,
  });

  assert.equal(serializeOtpCells(cleared.cells), "12456");

  assert.deepEqual(removeOtpDigit(cleared.cells, 2), {
    cells: ["1", "", "", "4", "5", "6"],
    focusIndex: 1,
  });
});

test("delete clears only the active cell and a replacement restores the code", () => {
  const cleared = clearOtpDigit(createOtpCells("123456", 6), 2);
  assert.deepEqual(cleared, ["1", "2", "", "4", "5", "6"]);
  assert.equal(serializeOtpCells(cleared), "12456");

  const restored = replaceOtpDigits(cleared, "3", 2, 6);
  assert.deepEqual(restored, ["1", "2", "3", "4", "5", "6"]);
  assert.equal(serializeOtpCells(restored), "123456");
});
