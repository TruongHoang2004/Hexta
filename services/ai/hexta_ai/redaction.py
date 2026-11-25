"""Typed PII redaction applied before customer text leaves for a third-party router.

Heuristic defense in depth for Vietnamese chat text, not a completeness guarantee: phone numbers
(+84/84/0 prefixes, spaced, dotted or dashed), emails and best-effort street addresses become typed
placeholders. Prices such as ``1.250.000đ`` are not phone numbers and stay intact.
"""

import re
import unicodedata

PHONE = "[PHONE]"
EMAIL = "[EMAIL]"
ADDRESS = "[ADDRESS]"

_EMAIL = re.compile(r"(?<![\w.+-])[\w.+-]+@[\w-]+(?:\.[\w-]+)*\.[A-Za-z]{2,}(?![\w-])")
# A candidate run of digits with optional single separators; validated by digit count below.
_PHONE = re.compile(r"(?<![\w+])(?:\(?\+84\)?|\(?84\)?|0)[\s.\-]?(?:\d[\s.\-]?){7,10}\d(?!\w)")

# Uppercase Vietnamese letters only (the À-Ỹ range also contains lowercase letters).
_UPPER = "A-Z" + "".join(sorted({c for c in map(chr, range(0xC0, 0x1EFA)) if c.isupper() and c.isalpha()}))
_WORD = r"[\wÀ-ỹđĐ]"
_STREET_KEYWORDS = r"(?:đường|đ\.|phố|ngõ|ngách|hẻm|kiệt|ấp|thôn|khu phố|kp\.?|duong|pho|hem|ngo|ngach)"
_ADMIN_ACCENTED = r"(?:phường|quận|huyện|xã|thị xã|thị trấn|tỉnh|thành phố)"
_ADMIN_SHORT = r"(?:tp\.?|p\.|q\.|phuong|quan|huyen|xa)"
_NAME = rf"[{_UPPER}\d]{_WORD}*(?:\s+[{_UPPER}\d]{_WORD}*){{0,4}}"
_HOUSE = r"(?:(?i:số)\s*(?:(?i:nhà)\s*)?)?\d{1,5}[A-Za-z]?(?:[/\-]\d{1,4}[A-Za-z]?)*"

_ADDRESS_PATTERNS = [
    # 12 đường Lê Lợi / 45/6 hẻm 12 / số 5 ngõ 34 Láng Hạ
    re.compile(rf"{_HOUSE}\s*,?\s*(?i:{_STREET_KEYWORDS})\s+(?:\d+[\w/]*(?:\s+{_NAME})?|{_NAME})"),
    # ngõ 34 / hẻm 45/6 / kiệt 12
    re.compile(rf"(?i:ngõ|ngách|hẻm|kiệt|hem|ngo|ngach)\s+\d+[\w/]*(?:\s+{_NAME})?"),
    # 123 Nguyễn Trãi, Quận 5 (house number + capitalized street followed by an administrative unit)
    re.compile(rf"{_HOUSE}\s+{_NAME}(?=\s*,\s*(?i:{_ADMIN_ACCENTED}|{_ADMIN_SHORT}))"),
    # phường Bến Nghé / Quận 7 / TP.HCM / thành phố Hồ Chí Minh
    re.compile(rf"(?i:{_ADMIN_ACCENTED})\s+(?:\d+|{_WORD}+(?:\s+[{_UPPER}]{_WORD}*){{0,3}})"),
    # unaccented or abbreviated units only with a number (quan 7, p.2, q7) or TP.HCM / tp HCM
    re.compile(rf"(?<!\w)(?i:{_ADMIN_SHORT})\s*\d{{1,2}}(?!\w)"),
    re.compile(r"(?<!\w)[qpQP]\d{1,2}(?!\w)"),
    re.compile(rf"(?<!\w)(?i:tp\.?)\s*(?:[{_UPPER}]{_WORD}+(?:\s+[{_UPPER}]{_WORD}*){{0,3}}|[A-Z]{{2,}})"),
]
_MERGE = re.compile(rf"{re.escape(ADDRESS)}(?:[\s,.\-]+{re.escape(ADDRESS)})+")


def _phone_digits(value: str) -> str | None:
    digits = re.sub(r"\D", "", value)
    if digits.startswith("84"):
        digits = "0" + digits[2:]
    # Vietnamese mobiles have 10 digits and landlines 11 (0 + 2x area code + subscriber).
    if digits.startswith("0") and digits[1:2] not in ("", "0") and len(digits) in (10, 11):
        return digits
    return None


def contact_values(text: str) -> set[str]:
    """Normalized phones/emails present in text (used to allow tenant-provided contacts in answers)."""
    values = {m.group(0).casefold() for m in _EMAIL.finditer(text)}
    values |= {d for m in _PHONE.finditer(text) if (d := _phone_digits(m.group(0)))}
    return values


def redact_pii(text: str, *, addresses: bool = True, keep: set[str] | None = None) -> str:
    """Replace phones, emails and (optionally) addresses with typed placeholders.

    ``keep`` holds normalized contact values (see ``contact_values``) that may remain, such as a shop
    hotline that appears in the tenant's own knowledge.
    """
    if not text:
        return text
    keep = keep or set()
    text = unicodedata.normalize("NFC", text)

    def email(match):
        return match.group(0) if match.group(0).casefold() in keep else EMAIL

    def phone(match):
        digits = _phone_digits(match.group(0))
        if digits is None or digits in keep:
            return match.group(0)
        return PHONE

    text = _EMAIL.sub(email, text)
    text = _PHONE.sub(phone, text)
    if addresses:
        for pattern in _ADDRESS_PATTERNS:
            text = pattern.sub(ADDRESS, text)
        text = _MERGE.sub(ADDRESS, text)
    return text
