"""Typed PII redaction applied before customer text leaves for a third-party router.

Heuristic defense in depth for Vietnamese chat text, not a completeness guarantee: phone numbers
(+84/84/0 prefixes, spaced, dotted or dashed, glued to a phone keyword or a trailing word, and 9-digit
mobiles written without the leading 0), emails and best-effort street addresses (including ``ấp 3``,
``kp 5``, ``tổ 7``, ``thôn 2`` and house + street before a known district, or before a city after a comma)
become typed placeholders. Dishes and products named after a place (``3 tô bún bò Huế``, ``iPhone 15 Pro Max
Hà Nội``) and product models such as ``Loa KP2`` are not addresses.

Any valid phone shape is redacted whatever word precedes it, so ``đơn hàng 0912345678`` and ``order
912345678`` lose the number: PII is redacted before text leaves, and a customer often answers "which order?"
with their phone. Only values that cannot be a phone stay intact: prices such as ``1.250.000đ``, quantities
such as ``2 cái`` or ``912345678 cái``, codes glued to letters such as ``DH0912345678`` or ``#ORD-2026-…``,
and digit runs whose length or prefix is not a Vietnamese phone (``mã đơn 1234567890``).
"""

import re
import unicodedata

PHONE = "[PHONE]"
EMAIL = "[EMAIL]"
ADDRESS = "[ADDRESS]"

_EMAIL = re.compile(r"(?<![\w.+-])[\w.+-]+@[\w-]+(?:\.[\w-]+)*\.[A-Za-z]{2,}(?![\w-])")
# A phone may be glued to a preceding phone keyword ("sđt0912345678"); any other letter before it (an order code
# such as "DH0912345678") keeps it. Letters may follow ("0912345678nha"), digits may not.
_PHONE_KEYWORDS = ("sdt", "sđt", "dt", "đt", "zalo", "tel", "phone", "hotline", "lh", "goi", "gọi", "call", "số", "so")
_PHONE_START = "(?:(?<![\\w+])|" + "|".join(f"(?i:(?<={k}))" for k in _PHONE_KEYWORDS) + ")"
# A candidate run of digits with optional single separators; validated by digit count below.
_PHONE = re.compile(_PHONE_START + r"(?:\(?\+84\)?|\(?84\)?|0)[\s.\-]?(?:\d[\s.\-]?){7,10}\d(?!\d)")
# Mobile written without its leading 0 ("912345678", "912 345 678"): 9 digits starting 3/5/7/8/9, never dotted (a
# dotted run is a price), never inside a dashed or slashed code and never an amount or a quantity (checked by
# _is_bare_mobile). An order word before it does not keep it: "mã đơn 912345678" may be the customer's phone.
_PHONE9 = re.compile(_PHONE_START.replace("(?<![\\w+])", "(?<![\\w+.,/-])") + r"[35789]\d{2} ?\d{3} ?\d{3}(?![\d.,/-])")
_MONEY_AFTER = re.compile(r"(?i:\s*(?:đ|₫|vnđ|vnd|đồng|dong|k|tr|triệu|trieu|nghìn|nghin|ngàn|ngan)(?!\w))")
# A count unit right after a bare 9-digit number makes it a quantity ("tôi muốn 912345678 cái").
_QUANTITY_AFTER = re.compile(r"(?i:\s*(?:cái|cai|chiếc|chiec|đôi|doi|bộ|bo|hộp|hop|gói|goi|sp|sản phẩm|san pham|"
                             r"món|mon|cuốn|cuon|ly|kg|ký|cặp|cap)(?!\w))")

# Uppercase Vietnamese letters only (the À-Ỹ range also contains lowercase letters).
_UPPER = "A-Z" + "".join(sorted({c for c in map(chr, range(0xC0, 0x1EFA)) if c.isupper() and c.isalpha()}))
_WORD = r"[\wÀ-ỹđĐ]"
_STREET_KEYWORDS = r"(?:đường|đ\.|phố|ngõ|ngách|hẻm|kiệt|ấp|thôn|khu phố|kp\.?|duong|pho|hem|ngo|ngach)"
# Small curated list of major cities and Hà Nội / Hồ Chí Minh City districts, so "15B Phan Đình Phùng, Ba Đình"
# is an address without any unit keyword. Unaccented spellings are derived below. A city alone also names
# dishes and products ("bún bò Huế", "phở Hà Nội"), so only a district, or a city after a comma, ends an address.
_CITIES = "Hà Nội|Hồ Chí Minh|HCM|Sài Gòn|Đà Nẵng|Hải Phòng|Cần Thơ|Biên Hòa|Nha Trang|Huế"
_DISTRICTS = """Ba Đình|Hoàn Kiếm|Hai Bà Trưng|Đống Đa|Tây Hồ|Cầu Giấy|Thanh Xuân|Hoàng Mai|Long Biên|Hà Đông|
Nam Từ Liêm|Bắc Từ Liêm|Gia Lâm|Đông Anh|Thanh Trì|Bình Thạnh|Phú Nhuận|Gò Vấp|Tân Bình|Tân Phú|Bình Tân|Thủ Đức|
Nhà Bè|Bình Chánh|Hóc Môn|Củ Chi"""
_ADMIN_ACCENTED = r"(?:phường|quận|huyện|xã|thị xã|thị trấn|tỉnh|thành phố)"
_ADMIN_SHORT = r"(?:tp\.?|p\.|q\.|phuong|quan|huyen|xa)"
_NAME = rf"[{_UPPER}\d]{_WORD}*(?:\s+[{_UPPER}\d]{_WORD}*){{0,4}}"
_HOUSE = r"(?:(?i:số)\s*(?:(?i:nhà)\s*)?)?\d{1,5}[A-Za-z]?(?:[/\-]\d{1,4}[A-Za-z]?)*"

# An administrative unit right after a street: "Quận 5", "phường Bến Nghé", "quan 5", "q.1", "Q1".
_ADMIN_AHEAD = rf"(?:(?i:{_ADMIN_ACCENTED})\s+\S|(?i:{_ADMIN_SHORT})\s*\d|[qpQP]\d{{1,2}}(?!\w))"
_LETTERS = rf"[^\W\d_]{_WORD}*"
# Words that make "<number> <words> quận 7" a quantity or an instruction ("2 cái về quận 7"), not a house number.
_NOT_STREET = frozenset("""cái cai chiếc chiec đôi bộ hộp hop gói sp sản phẩm size áo ao quần cặp đơn don về ve tới
đến ở tại tai giao ship gửi gui cho nha nhé nhe mua lấy lay món mon người nguoi kg ký cuốn ly chi nhánh nhanh
shop kho ngày ngay tiếng tieng giờ gio phút phut km màu mau mẫu loại loai tháng thang""".split())
# Product words that make "<number> <words> <place>" a product line, not a house + street ("iPhone 15 Pro Max").
_NOT_STREET_PLACE = _NOT_STREET | frozenset("freeship combo pro max plus ultra mini gb tb".split())
# Servings and dishes ("3 tô bún bò Huế", "5 bánh cốm", "1 cà phê"). They block only as written in lowercase,
# since capitalized they are street names: Tô Hiến Thành, Bát Đàn, Chả Cá, Hàng Bún, Hàng Mắm, Hàng Thùng.
_SERVINGS = frozenset("""tô bát chén phần suất đĩa dĩa bịch túi hũ chai lon thùng bún phở cơm bánh cà phê cafe café
coffee trà sữa mè xửng cốm chả nem giò mắm kẹo mứt đặc""".split())


def _fold(text: str) -> str:
    return "".join(c for c in unicodedata.normalize("NFD", text) if not unicodedata.combining(c)).replace(
        "đ", "d").replace("Đ", "D")


def _place_pattern(places: str) -> str:
    names = sorted({name for place in places.split("|") for name in (place.strip(), _fold(place.strip()))},
                   key=len, reverse=True)
    return r"(?i:" + "|".join(re.escape(name).replace(r"\ ", r"\s+") for name in names) + r")(?!\w)"


_CITY = _place_pattern(_CITIES)
_DISTRICT = _place_pattern(_DISTRICTS)
_PLACE = _place_pattern(_CITIES + "|" + _DISTRICTS)


def _street_or_keep(match):
    words = match.group("words").casefold().split()
    return match.group(0) if any(w in _NOT_STREET for w in words) else ADDRESS


def _street_place_or_keep(match):
    words = match.group("words").split()
    street = words[0].casefold() in ("hàng", "hang")  # Hà Nội "Hàng <goods>" streets: Hàng Bún, Hàng Mắm
    if any(w.casefold() in _NOT_STREET_PLACE for w in words) or (not street and any(w in _SERVINGS for w in words)):
        return match.group(0)
    return ADDRESS


# House number + 1-4 street words (any case, with or without diacritics) followed by an administrative unit:
# 123 Nguyễn Trãi, Quận 5 / 123 nguyễn trãi quận 5 / 88 Đồng Khởi Quận 1 / 12 Nguyễn Huệ, Q1 / 25 le loi q1
_HOUSE_STREET = re.compile(rf"{_HOUSE}\s+(?P<words>(?:{_LETTERS}\s+){{0,3}}{_LETTERS})(?=\s*,?\s*{_ADMIN_AHEAD})")

# House number + 1-4 street words + a known district, or a city after a comma, without any unit keyword:
# 15B Phan Đình Phùng, Ba Đình, Hà Nội / 22 Hai Bà Trưng Hoàn Kiếm / 15B Phan Đình Phùng, Hà Nội. Never inside a
# price, an amount ("299k freeship Hà Nội") or a code, and never after a product name ("iPhone 15 Pro Max").
_HOUSE_PLACE = re.compile(rf"(?<![\w.,/])(?!\d+(?i:k|đ|tr)(?!\w)){_HOUSE}\s+"
                          rf"(?P<words>(?:{_LETTERS}\s+){{0,3}}{_LETTERS})(?:\s*,?\s*{_DISTRICT}|\s*,\s*{_CITY})"
                          rf"(?:\s*,?\s*{_PLACE})*")
# Standalone hamlet, quarter or residential group with a number: ấp 3 / kp 5 / KP.2 / khu phố 4 / tổ 7 / thôn 2.
# A count unit after it ("tổ 5 người") keeps it. "kp" is also a product-model prefix ("Loa KP2"), so without its dot
# it is a quarter only right after an address word or a comma ("nhà ở kp 5", "12/3, kp 5").
_KP_CONTEXT = "|".join(f"(?i:(?<={w}\\s))" for w in ("ở", "tại", "tai", "giao", "về", "tới", "toi", "đến", "den",
                                                     "nhà", "chỉ", "đc", "dc", ","))
_LOCAL_UNIT = re.compile(rf"(?<!\w)(?:(?i:ấp|thôn|khu phố|khu pho|tổ dân phố|tổ)\s+|(?i:kp)\.\s*|"
                         rf"(?:{_KP_CONTEXT}|(?<=,))(?i:kp)\s*)\d{{1,3}}[A-Za-z]?"
                         rf"(?!\w)(?!\s+(?i:người|nguoi|cái|cai|chiếc|đôi|bộ|sp|món|bạn|ban)(?!\w))"
                         rf"(?:\s+[{_UPPER}]{_WORD}*(?:\s+[{_UPPER}]{_WORD}*){{0,2}})?")

_ADDRESS_PATTERNS = [
    # First, so "tổ 12 khu phố 3" is not split by the house-number pattern below.
    _LOCAL_UNIT,
    # 12 đường Lê Lợi / 45/6 hẻm 12 / số 5 ngõ 34 Láng Hạ / số 5, đường số 7
    re.compile(rf"{_HOUSE}\s*,?\s*(?i:{_STREET_KEYWORDS})\s+(?:(?:(?i:số)\s*)?\d+[\w/]*(?:\s+{_NAME})?|{_NAME})"),
    # ngõ 34 / hẻm 45/6 / kiệt 12
    re.compile(rf"(?i:ngõ|ngách|hẻm|kiệt|hem|ngo|ngach)\s+\d+[\w/]*(?:\s+{_NAME})?"),
    (_HOUSE_PLACE, _street_place_or_keep),
    (_HOUSE_STREET, _street_or_keep),
    # phường Bến Nghé / Quận 7 / TP.HCM / thành phố Hồ Chí Minh
    re.compile(rf"(?i:{_ADMIN_ACCENTED})\s+(?:\d+|{_WORD}+(?:\s+[{_UPPER}]{_WORD}*){{0,3}})"),
    # unaccented or abbreviated units only with a number (quan 7, p.2, q7) or TP.HCM / tp HCM
    re.compile(rf"(?<!\w)(?i:{_ADMIN_SHORT})\s*\d{{1,2}}(?!\w)"),
    re.compile(r"(?<!\w)[qpQP]\d{1,2}(?!\w)"),
    re.compile(rf"(?<!\w)(?i:tp\.?)\s*(?:[{_UPPER}]{_WORD}+(?:\s+[{_UPPER}]{_WORD}*){{0,3}}|[A-Z]{{2,}})"),
]
_MERGE = re.compile(rf"{re.escape(ADDRESS)}(?:[\s,.\-]+{re.escape(ADDRESS)})+")


def _is_bare_mobile(text: str, match) -> bool:
    """A 9-digit match is a phone unless an amount of money or a count unit follows it."""
    return not (_MONEY_AFTER.match(text, match.end()) or _QUANTITY_AFTER.match(text, match.end()))


def _phone_digits(value: str) -> str | None:
    digits = re.sub(r"\D", "", value)
    if len(digits) == 9 and digits[0] in "35789":
        return "0" + digits
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
    values |= {d for m in _PHONE9.finditer(text) if _is_bare_mobile(text, m) and (d := _phone_digits(m.group(0)))}
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

    def bare_mobile(match):
        return phone(match) if _is_bare_mobile(match.string, match) else match.group(0)

    text = _EMAIL.sub(email, text)
    text = _PHONE.sub(phone, text)
    text = _PHONE9.sub(bare_mobile, text)
    if addresses:
        for pattern in _ADDRESS_PATTERNS:
            pattern, replacement = pattern if isinstance(pattern, tuple) else (pattern, ADDRESS)
            text = pattern.sub(replacement, text)
        text = _MERGE.sub(ADDRESS, text)
    return text
