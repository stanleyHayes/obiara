const productCopyAttributes = new Set([
  "aria-label",
  "alt",
  "label",
  "placeholder",
  "title",
]);

const allowedLiteralValues = new Set(["", "·", "—"]);
const pressurePattern =
  /\b(?:hurry|don't miss out|do not miss out|last chance|act now)\b|!{2,}/iu;
const formalPattern = /\b(?:kindly|please be advised|therefore|hereby)\b/iu;
const informalPattern = /\b(?:gonna|wanna|hey|y'all|can't|won't|don't)\b/iu;

function literalText(node) {
  if (!node) return null;
  if (node.type === "Literal" && typeof node.value === "string") {
    return node.value;
  }
  if (
    node.type === "JSXExpressionContainer" &&
    node.expression?.type === "Literal" &&
    typeof node.expression.value === "string"
  ) {
    return node.expression.value;
  }
  if (node.type === "JSXText") return node.value;
  return null;
}

function meaningful(value) {
  return value?.replace(/\s+/gu, " ").trim() ?? "";
}

function openingElementName(node) {
  return node.name?.type === "JSXIdentifier" ? node.name.name : null;
}

function hasAttribute(node, names) {
  return node.attributes.some(
    (attribute) =>
      attribute.type === "JSXAttribute" &&
      attribute.name.type === "JSXIdentifier" &&
      names.has(attribute.name.name),
  );
}

function reportCopy(context, node, value) {
  const text = meaningful(value);
  if (!text || allowedLiteralValues.has(text) || !/[A-Za-z]/u.test(text)) {
    return;
  }
  context.report({
    node,
    messageId: "hardcoded",
    data: { text: text.slice(0, 48) },
  });
}

const noHardcodedProductCopy = {
  meta: {
    type: "problem",
    docs: {
      description: "Require product-facing copy to use the i18n registry.",
    },
    messages: {
      hardcoded:
        'Product copy "{{text}}" must use a registered localization key.',
    },
    schema: [],
  },
  create(context) {
    return {
      JSXText(node) {
        reportCopy(context, node, node.value);
      },
      JSXAttribute(node) {
        if (
          node.name.type === "JSXIdentifier" &&
          productCopyAttributes.has(node.name.name)
        ) {
          reportCopy(context, node, literalText(node.value));
        }
      },
    };
  },
};

const noPressureCopy = {
  meta: {
    type: "problem",
    docs: { description: "Reject coercive or artificial-urgency copy." },
    messages: {
      pressure: "Avoid coercive or artificial-urgency product copy.",
    },
    schema: [],
  },
  create(context) {
    return {
      Literal(node) {
        if (
          typeof node.value === "string" &&
          pressurePattern.test(node.value)
        ) {
          context.report({ node, messageId: "pressure" });
        }
      },
      JSXText(node) {
        if (pressurePattern.test(node.value)) {
          context.report({ node, messageId: "pressure" });
        }
      },
    };
  },
};

const noMixedRegister = {
  meta: {
    type: "problem",
    docs: { description: "Keep a consistent conversational register." },
    messages: {
      mixed: "Do not mix formal and conversational registers in one message.",
    },
    schema: [],
  },
  create(context) {
    return {
      Literal(node) {
        if (
          typeof node.value === "string" &&
          formalPattern.test(node.value) &&
          informalPattern.test(node.value)
        ) {
          context.report({ node, messageId: "mixed" });
        }
      },
    };
  },
};

const accessibleNames = {
  meta: {
    type: "problem",
    docs: { description: "Require accessible names for common JSX controls." },
    messages: {
      image:
        'Images require non-empty alt text; use alt="" only when decorative.',
      control:
        "Interactive controls require visible text or an accessible label.",
    },
    schema: [],
  },
  create(context) {
    return {
      JSXOpeningElement(node) {
        const name = openingElementName(node);
        if (name === "img" && !hasAttribute(node, new Set(["alt"]))) {
          context.report({ node, messageId: "image" });
        }
        if (
          (name === "button" || name === "a") &&
          node.parent?.type === "JSXElement"
        ) {
          const labelled = hasAttribute(
            node,
            new Set(["aria-label", "aria-labelledby"]),
          );
          const visible = node.parent.children.some(
            (child) =>
              child.type === "JSXExpressionContainer" ||
              meaningful(literalText(child)),
          );
          if (!labelled && !visible) {
            context.report({ node, messageId: "control" });
          }
        }
      },
    };
  },
};

export const productQualityPlugin = {
  meta: { name: "@obiara/product-quality", version: "0.0.0" },
  rules: {
    "accessible-names": accessibleNames,
    "no-hardcoded-product-copy": noHardcodedProductCopy,
    "no-mixed-register": noMixedRegister,
    "no-pressure-copy": noPressureCopy,
  },
};

export function productQualityConfig({ enforcement = "warn" } = {}) {
  if (!["off", "warn", "error"].includes(enforcement)) {
    throw new TypeError("enforcement must be off, warn, or error");
  }
  return {
    plugins: { "obiara-quality": productQualityPlugin },
    rules: Object.fromEntries(
      Object.keys(productQualityPlugin.rules).map((rule) => [
        `obiara-quality/${rule}`,
        enforcement,
      ]),
    ),
  };
}
