export function sanitizeOtp(value: string, length: number): string {
  return value.replace(/\D/g, "").slice(0, length);
}

export type OtpCells = ReadonlyArray<string>;

export function createOtpCells(value: string, length: number): OtpCells {
  const cleanValue = sanitizeOtp(value, length);
  return Array.from({ length }, (_, index) => cleanValue[index] ?? "");
}

export function serializeOtpCells(cells: OtpCells): string {
  return cells.join("");
}

export function replaceOtpDigits(
  cells: OtpCells,
  insertion: string,
  start: number,
  length: number,
): OtpCells {
  const digits = Array.from({ length }, (_, index) => cells[index] ?? "");
  const cleanInsertion = sanitizeOtp(insertion, length);
  const insertionStart = cleanInsertion.length === length ? 0 : start;

  for (const [offset, digit] of [...cleanInsertion].entries()) {
    const target = insertionStart + offset;
    if (target >= length) break;
    digits[target] = digit;
  }

  return digits;
}

export function clearOtpDigit(cells: OtpCells, index: number): OtpCells {
  const digits = [...cells];
  digits[index] = "";
  return digits;
}

export function removeOtpDigit(
  cells: OtpCells,
  index: number,
): Readonly<{ cells: OtpCells; focusIndex: number }> {
  const target = cells[index] ? index : Math.max(0, index - 1);
  const digits = clearOtpDigit(cells, target);

  return { cells: digits, focusIndex: target };
}
