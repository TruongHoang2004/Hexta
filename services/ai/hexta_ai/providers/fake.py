"""Deterministic offline routing and answers, without network clients or paid inference."""

import re

from hexta_ai.providers.base import BaseAIProvider, Candidate, Classification, GroundedAnswer, Usage
from hexta_ai.routing.decision import exact_script_match

# Upper bound for the quoted part of the best record; whole sentences only, so no number is ever cut in half.
EXCERPT_CHARS = 300


def excerpt(title: str, text: str) -> str:
    """Leading whole sentences of a record (without its repeated title), at most EXCERPT_CHARS long."""
    body = text.strip()
    if title and body.startswith(title + "\n"):
        # Stored chunks repeat the title on their first line.
        body = body[len(title) + 1:]
    body = re.sub(r"\s+", " ", body).strip()
    sentences = [s for s in re.split(r"(?<=[.!?])\s+", body) if s]
    chosen = ""
    for sentence in sentences:
        if len(chosen) + len(sentence) + 1 > EXCERPT_CHARS:
            break
        chosen = f"{chosen} {sentence}".strip()
    if not chosen and sentences:
        # One long sentence: keep whole words, dropping any trailing word with a digit (no truncated numbers).
        words = sentences[0][:EXCERPT_CHARS].split(" ")[:-1]
        while words and re.search(r"\d", words[-1]):
            words.pop()
        chosen = " ".join(words) + " …" if words else ""
    if chosen and chosen[-1] not in ".!?…":
        chosen += "."
    return chosen.replace("{", "(").replace("}", ")")


def compose_answer(item: dict) -> str:
    """Vietnamese shop reply quoting the record: title, excerpt, then price and stock from its attributes."""
    title = item.get("title", "").strip()
    attributes = item.get("attributes") or {}
    parts = [f"Dạ, về {title} ạ:" if title else "Dạ, theo thông tin của shop ạ:"]
    quoted = excerpt(title, item.get("text", ""))
    if quoted:
        parts.append(quoted)
    if attributes.get("price"):
        price = attributes["price"].strip()
        currency = "đ" if re.fullmatch(r"[\d.,]+", price) else ""
        parts.append(f"Giá: {price}{currency}.")
    stock = attributes.get("stock", "").strip()
    if stock == "0":
        parts.append("Hiện sản phẩm đang hết hàng.")
    elif stock:
        parts.append(f"Tồn kho: {stock} sản phẩm.")
    parts.append("Anh/chị cần shop hỗ trợ thêm gì không ạ?")
    return " ".join(parts).replace("{", "(").replace("}", ")")[:2000]


class FakeProvider(BaseAIProvider):
    async def classify(self, context):
        script_id = exact_script_match(context["message"], context["scripts"])
        return Classification(route="script" if script_id else "human", script_id=script_id,
                              confidence=1 if script_id else 0), Usage()

    async def generate_candidate(self, context):
        return Candidate(text="Please ask a team member for assistance.", fact_ids=[]), Usage()

    async def answer(self, context):
        # Extractive stand-in for a model: a reply built only from the best retrieved record, citing it.
        knowledge = context.get("knowledge") or []
        if not knowledge:
            return GroundedAnswer(kind="answer", text="", cited_source_ids=[], confidence=0, needs_human=True), Usage()
        best = knowledge[0]
        return GroundedAnswer(kind="answer", text=compose_answer(best), cited_source_ids=[best["id"]],
                              confidence=0.9, needs_human=False), Usage()
