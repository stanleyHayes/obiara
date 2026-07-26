const placeholderPattern = /\{([A-Za-z][A-Za-z0-9_]*)\}/gu;
const pressurePattern =
  /\b(?:hurry|don't miss out|do not miss out|last chance|act now)\b|!{2,}/iu;
const formalPattern = /\b(?:kindly|please be advised|therefore|hereby)\b/iu;
const informalPattern = /\b(?:gonna|wanna|hey|y'all|can't|won't|don't)\b/iu;

export function placeholders(message) {
  return [
    ...new Set([...message.matchAll(placeholderPattern)].map((m) => m[1])),
  ].sort();
}

export function validateCatalog(source, candidate) {
  const findings = [];
  for (const [key, message] of Object.entries(source)) {
    if (!Object.hasOwn(candidate, key)) {
      findings.push({ rule: "missing-registry-key", key });
      continue;
    }
    const expected = placeholders(message);
    const actual = placeholders(candidate[key]);
    if (
      expected.length !== actual.length ||
      expected.some((value, index) => value !== actual[index])
    ) {
      findings.push({
        rule: "invalid-placeholders",
        key,
        expected,
        actual,
      });
    }
  }
  for (const key of Object.keys(candidate)) {
    if (!Object.hasOwn(source, key)) {
      findings.push({ rule: "unknown-registry-key", key });
    }
  }
  return findings;
}

export function validateCopy(message, context = {}) {
  const findings = [];
  if (pressurePattern.test(message)) {
    findings.push({ rule: "banned-pressure-copy" });
  }
  if (formalPattern.test(message) && informalPattern.test(message)) {
    findings.push({ rule: "mixed-register" });
  }
  if (context.registeredKeys && !context.registeredKeys.has(context.key)) {
    findings.push({ rule: "missing-registry-key", key: context.key });
  }
  return findings;
}

export function parseEscape(comment) {
  const match = comment.match(
    /^\s*(?:\/\/|\/\*)?\s*quality-ignore-next-line\s+non-product-data:\s*(.{8,}?)(?:\s*\*\/)?\s*$/u,
  );
  return match
    ? { classification: "non-product-data", reason: match[1] }
    : null;
}

export function inspectProductSource(source, filename = "<source>") {
  const findings = [];
  const lines = source.split(/\r?\n/u);
  lines.forEach((line, index) => {
    const previous = lines[index - 1] ?? "";
    const escaped = parseEscape(previous);
    const jsxText = [...line.matchAll(/>([^<>{]*[A-Za-z][^<>{]*)</gu)];
    const attributes = [
      ...line.matchAll(
        /\b(?:aria-label|alt|label|placeholder|title)=["']([^"']*[A-Za-z][^"']*)["']/gu,
      ),
    ];
    for (const match of [...jsxText, ...attributes]) {
      if (!escaped) {
        findings.push({
          rule: "hardcoded-product-copy",
          filename,
          line: index + 1,
          text: match[1].trim(),
        });
      }
      for (const finding of validateCopy(match[1])) {
        findings.push({ ...finding, filename, line: index + 1 });
      }
    }
    for (const match of line.matchAll(/["'`]([^"'`]*[A-Za-z][^"'`]*)["'`]/gu)) {
      for (const finding of validateCopy(match[1])) {
        findings.push({ ...finding, filename, line: index + 1 });
      }
    }
  });
  return findings;
}
