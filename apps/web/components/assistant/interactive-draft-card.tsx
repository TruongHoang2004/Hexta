"use client";

import React, { useState } from "react";
import {
  CheckCircle2,
  Clock,
  Minus,
  Plus,
  ShoppingBag,
  Trash2,
  User,
  Phone,
  MapPin,
  CreditCard,
  Sparkles,
  AlertCircle,
  ExternalLink,
} from "lucide-react";
import { Button } from "@hexta/ui";
import { OrderDraft } from "@/lib/services/ai-agent-service";

export interface InteractiveDraftCardProps {
  draft: OrderDraft;
  onUpdateItem?: (index: number, quantity: number, variant?: string) => void;
  onConfirm?: (draft: OrderDraft) => Promise<void>;
  onDiscard?: (draftId: string) => void;
  isConfirming?: boolean;
}

export function InteractiveDraftCard({
  draft,
  onUpdateItem,
  onConfirm,
  onDiscard,
  isConfirming = false,
}: InteractiveDraftCardProps) {
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat("vi-VN", {
      style: "currency",
      currency: "VND",
    }).format(val);
  };

  const handleConfirm = async () => {
    if (!onConfirm) return;
    setConfirmError(null);
    try {
      await onConfirm(draft);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Xác nhận đơn thất bại";
      setConfirmError(msg);
    }
  };

  const isProposed = draft.status === "proposed";
  const isConfirmed = draft.status === "confirmed";
  const isDiscarded = draft.status === "discarded";

  return (
    <div
      className={`my-3 overflow-hidden rounded-2xl border transition-all duration-200 ${
        isConfirmed
          ? "border-emerald-500/40 bg-emerald-500/5 shadow-md"
          : isDiscarded
          ? "border-muted/30 bg-muted/10 opacity-70"
          : "border-primary/40 bg-card shadow-lg ring-1 ring-primary/10"
      }`}
    >
      {/* Header Banner */}
      <div className="flex items-center justify-between border-b border-border/60 bg-muted/30 px-4 py-2.5">
        <div className="flex items-center gap-2">
          <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Sparkles className="h-3.5 w-3.5" />
          </div>
          <span className="text-xs font-semibold uppercase tracking-wider text-primary">
            Đề Xuất Đơn Hàng (AI Draft)
          </span>
        </div>

        <div>
          {isProposed && (
            <span className="inline-flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5 text-[11px] font-medium text-amber-600 dark:text-amber-400">
              <Clock className="h-3 w-3" /> Chờ xác nhận
            </span>
          )}
          {isConfirmed && (
            <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
              <CheckCircle2 className="h-3 w-3" /> Đã tạo đơn
            </span>
          )}
          {isDiscarded && (
            <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">
              Đã hủy
            </span>
          )}
        </div>
      </div>

      <div className="space-y-4 p-4 text-sm">
        {/* Customer Information if available */}
        {(draft.customer_name || draft.customer_phone || draft.shipping_address) && (
          <div className="grid grid-cols-1 gap-2 rounded-xl bg-muted/20 p-3 text-xs sm:grid-cols-2">
            {draft.customer_name && (
              <div className="flex items-center gap-1.5 text-foreground/80">
                <User className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="font-medium">{draft.customer_name}</span>
              </div>
            )}
            {draft.customer_phone && (
              <div className="flex items-center gap-1.5 text-foreground/80">
                <Phone className="h-3.5 w-3.5 text-muted-foreground" />
                <span>{draft.customer_phone}</span>
              </div>
            )}
            {draft.shipping_address && (
              <div className="col-span-full flex items-center gap-1.5 text-foreground/70">
                <MapPin className="h-3.5 w-3.5 text-muted-foreground" />
                <span>{draft.shipping_address}</span>
              </div>
            )}
          </div>
        )}

        {/* Line Items List */}
        <div className="space-y-2">
          <div className="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
            <ShoppingBag className="h-3.5 w-3.5" />
            <span>Danh sách mặt hàng ({draft.items.length})</span>
          </div>

          <div className="divide-y divide-border/40 rounded-xl border border-border/40 bg-background/50">
            {draft.items.map((item, idx) => (
              <div
                key={`${item.product_name}-${idx}`}
                className="flex items-center justify-between p-3 gap-2"
              >
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium text-foreground">
                    {item.product_name}
                  </p>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    {item.variant && (
                      <span className="rounded bg-muted px-1.5 py-0.2 text-[10px]">
                        {item.variant}
                      </span>
                    )}
                    <span>{formatCurrency(item.unit_price)}</span>
                  </div>
                </div>

                {/* Inline Quantity Controls */}
                <div className="flex items-center gap-3">
                  {isProposed ? (
                    <div className="flex items-center rounded-lg border border-border/70 bg-background">
                      <button
                        type="button"
                        onClick={() => onUpdateItem?.(idx, item.quantity - 1)}
                        disabled={item.quantity <= 1 || isConfirming}
                        className="flex h-7 w-7 items-center justify-center text-muted-foreground transition hover:text-foreground disabled:opacity-30"
                        title="Giảm số lượng"
                      >
                        <Minus className="h-3 w-3" />
                      </button>
                      <span className="w-6 text-center text-xs font-semibold">
                        {item.quantity}
                      </span>
                      <button
                        type="button"
                        onClick={() => onUpdateItem?.(idx, item.quantity + 1)}
                        disabled={isConfirming}
                        className="flex h-7 w-7 items-center justify-center text-muted-foreground transition hover:text-foreground"
                        title="Tăng số lượng"
                      >
                        <Plus className="h-3 w-3" />
                      </button>
                    </div>
                  ) : (
                    <span className="text-xs font-semibold text-muted-foreground">
                      x{item.quantity}
                    </span>
                  )}

                  <div className="w-20 text-right font-medium text-foreground">
                    {formatCurrency(item.subtotal)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Pricing Summary */}
        <div className="flex items-baseline justify-between border-t border-border/40 pt-3">
          <span className="text-xs font-medium text-muted-foreground">Tổng cộng</span>
          <span className="text-lg font-bold text-primary">
            {formatCurrency(draft.total_amount)}
          </span>
        </div>

        {/* Payment & Notes Meta */}
        <div className="flex flex-wrap items-center justify-between gap-2 text-[11px] text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <CreditCard className="h-3 w-3" />
            <span>Thanh toán: <b className="uppercase">{draft.payment_method || "COD"}</b></span>
          </div>
          {draft.notes && <span className="italic">Ghi chú: {draft.notes}</span>}
        </div>

        {confirmError && (
          <div className="flex items-center gap-2 rounded-lg bg-destructive/10 p-2.5 text-xs text-destructive">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{confirmError}</span>
          </div>
        )}

        {/* Interactive Action Controls */}
        {isProposed && (
          <div className="flex items-center gap-2 pt-2">
            <Button
              type="button"
              onClick={handleConfirm}
              disabled={isConfirming}
              className="flex-1 gap-1.5 rounded-xl font-medium shadow-md shadow-primary/20"
            >
              {isConfirming ? (
                <>
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                  <span>Đang xử lý đơn...</span>
                </>
              ) : (
                <>
                  <CheckCircle2 className="h-4 w-4" />
                  <span>Xác nhận đặt hàng</span>
                </>
              )}
            </Button>

            {onDiscard && (
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={() => onDiscard(draft.draft_id)}
                disabled={isConfirming}
                title="Hủy bản nháp"
                className="h-10 w-10 rounded-xl text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        )}

        {/* Confirmed State Notification */}
        {isConfirmed && (
          <div className="rounded-xl bg-emerald-500/10 p-3 text-xs text-emerald-700 dark:text-emerald-300">
            <div className="flex items-center justify-between">
              <span className="font-semibold">Mã đơn hàng:</span>
              <span className="font-mono font-bold">
                {draft.order_number || draft.order_id || "ORD-SUCCESS"}
              </span>
            </div>
            <p className="mt-1 text-emerald-600/90 dark:text-emerald-400/90">
              Đơn hàng đã được lưu vào hệ thống và giữ tồn kho thành công.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
