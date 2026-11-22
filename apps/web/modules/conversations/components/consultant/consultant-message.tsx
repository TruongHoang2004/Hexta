"use client";

import ReactMarkdown from "react-markdown";
import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ConsultDraft, ConsultProduct } from "../../services/consultant";
import type { ConsultTurn } from "../../types";
import { DraftOrderCard } from "./draft-order-card";
import { RecommendedProductCard } from "./recommended-product-card";

/** One chat message with any recommended products, draft order and quick replies. */
export function ConsultantMessage({
  turn,
  latest,
  busy,
  onReply,
  onProductDetails,
  onBuyProduct,
  onCheckout,
}: {
  turn: ConsultTurn;
  /** Quick replies are only offered on the newest message. */
  latest: boolean;
  busy: boolean;
  onReply: (text: string) => void;
  onProductDetails: (product: ConsultProduct) => void;
  onBuyProduct: (product: ConsultProduct) => void;
  onCheckout: (draft: ConsultDraft) => void;
}) {
  const t = useT();
  return (
    <article className={`space-y-3 rounded-lg p-3 text-sm ${turn.role === "user" ? "bg-primary/10" : "bg-surface"}`}>
      <p className="text-xs font-semibold">{turn.role === "user" ? t("consultant.you") : t("consultant.assistant")}</p>
      <div className="space-y-2 break-words [&_ul]:list-inside [&_ul]:list-disc [&_strong]:font-bold">
        {/* Model output: no raw HTML, links or images. */}
        <ReactMarkdown skipHtml components={{ a: ({ children }) => <span>{children}</span>, img: () => null }}>
          {turn.content}
        </ReactMarkdown>
      </div>
      {turn.needsHuman && <p role="status">{t("consultant.needsHuman")}</p>}
      {!!turn.products?.length && (
        <div className="flex snap-x gap-3 overflow-x-auto pb-2" aria-label={t("consultant.recommended")}>
          {turn.products.map((product) => (
            <RecommendedProductCard
              key={product.variant_id || product.product_id}
              product={product}
              onDetails={() => onProductDetails(product)}
              onBuy={() => onBuyProduct(product)}
            />
          ))}
        </div>
      )}
      {turn.draft && <DraftOrderCard draft={turn.draft} onConfirm={() => onCheckout(turn.draft!)} />}
      {latest && (
        <div className="flex flex-wrap gap-2">
          {turn.replies?.map((reply) => (
            <button disabled={busy} key={reply} className={secondaryClass} onClick={() => onReply(reply)}>
              {reply}
            </button>
          ))}
        </div>
      )}
    </article>
  );
}
