import { test } from "node:test";
import assert from "node:assert/strict";
import {
  resolveCheckoutVariant,
  checkoutProductName,
} from "./consultant-checkout.ts";
const variant = (id, label) => ({
  id,
  sku: id,
  variant_name: label,
  price: "100000",
  price_provided: true,
});
const products = [
  {
    id: "single",
    name: "Milk",
    description: "",
    category: "",
    variants: [variant("bottle", "500ml")],
  },
  {
    id: "multiple",
    name: "Shirt",
    description: "",
    category: "",
    variants: [variant("small", "Size S"), variant("large", "Size L")],
  },
];
test("checkout resolves canonical, missing, empty and legacy variant IDs", () => {
  for (const id of [undefined, "", "  ", "bottle"])
    assert.equal(
      resolveCheckoutVariant(products, "single", id)?.variant.id,
      "bottle",
    );
  assert.equal(
    resolveCheckoutVariant(products, "bottle", undefined)?.product.id,
    "single",
  );
  assert.equal(
    resolveCheckoutVariant(products, "multiple", "large")?.variant.variant_name,
    "Size L",
  );
});
test("checkout rejects ambiguity, stale IDs and variants from another product", () => {
  assert.equal(
    resolveCheckoutVariant(products, "multiple", undefined),
    undefined,
  );
  assert.equal(resolveCheckoutVariant(products, "single", "large"), undefined);
  assert.equal(
    resolveCheckoutVariant(products, "single", "removed"),
    undefined,
  );
  assert.equal(
    resolveCheckoutVariant(products, "missing", "bottle"),
    undefined,
  );
});
test("checkout labels include current catalog variant names", () => {
  assert.equal(checkoutProductName("Shirt", " Size L "), "Shirt (Size L)");
  assert.equal(checkoutProductName("Shirt", ""), "Shirt");
});
