import type { Product } from "./sales";

export function checkoutProductName(name: string, variantName: string) {
  const label = variantName.trim();
  return label ? `${name} (${label})` : name;
}

// Candidate storage can contain old product-only IDs or the legacy variant-as-product ID.
// An explicit variant must belong to the named product; never guess among several variants.
export function resolveCheckoutVariant(
  products: Product[],
  productID: string,
  variantID: unknown,
) {
  const explicitVariant = typeof variantID === "string" ? variantID.trim() : "";
  const product = products.find(
    (p) => p.id === productID || p.variants.some((v) => v.id === productID),
  );
  if (!product) return undefined;
  const variant = explicitVariant
    ? product.variants.find((v) => v.id === explicitVariant)
    : product.variants.find((v) => v.id === productID) ||
      (product.variants.length === 1 ? product.variants[0] : undefined);
  return variant ? { product, variant } : undefined;
}
