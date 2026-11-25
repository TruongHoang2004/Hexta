"""Typed redaction before customer text reaches the third-party router."""

import pytest

from hexta_ai.redaction import contact_values, redact_pii


@pytest.mark.parametrize("text,expected", [
    ("sdt mình 0912 345 678 nhé", "sdt mình [PHONE] nhé"),
    ("gọi 0912.345.678", "gọi [PHONE]"),
    ("số 091-234-5678", "số [PHONE]"),
    ("+84 912 345 678", "[PHONE]"),
    ("(+84) 912345678", "[PHONE]"),
    ("84912345678 ạ", "[PHONE] ạ"),
    ("bàn 024 3826 1234", "bàn [PHONE]"),
    ("mail abc.def+shop@gmail.com giúp", "mail [EMAIL] giúp"),
    # Glued to a phone keyword or to a trailing word, and mobiles written without the leading 0.
    ("sđt0912345678", "sđt[PHONE]"),
    ("SDT0912345678 nhé", "SDT[PHONE] nhé"),
    ("zalo0912345678", "zalo[PHONE]"),
    ("0912345678nha", "[PHONE]nha"),
    ("gọi mình 0912345678ạ", "gọi mình [PHONE]ạ"),
    ("912345678", "[PHONE]"),
    ("sdt 912345678 nhé", "sdt [PHONE] nhé"),
    ("số mình 912 345 678", "số mình [PHONE]"),
])
def test_phones_and_emails_become_typed_placeholders(text, expected):
    assert redact_pii(text) == expected


@pytest.mark.parametrize("text", [
    "giá 1.250.000đ còn không", "giá 299.000 ạ", "mã đơn 1234567890", "size 42 còn 5 đôi",
    "mua quan size 30", "2 Áo Polo đen", "có đường không", "84.000đ",
    # Prices, quantities and order codes that look like phones.
    "lấy 2 cái nhé", "2 cái áo size L", "mua 3 đôi giá 1.250.000đ", "đơn DH0912345678 giao chưa",
    "mã đơn 912345678", "mã: 912345678", "order #912345678", "đơn 512345678 sao rồi", "912.345.678đ",
    "giá 350000000đ", "chuyển 350000000 đồng", "SP0912345678", "tổng 912345678 vnđ", "1234567890", "mã đơn #ORD-2026-912345678",
])
def test_prices_sizes_and_ordinary_words_are_kept(text):
    assert redact_pii(text) == text


@pytest.mark.parametrize("text,expected", [
    ("giao 123 Nguyễn Trãi, Quận 5, TP.HCM giúp", "giao [ADDRESS] giúp"),
    ("số 12 ngõ 34 Láng Hạ", "[ADDRESS]"),
    ("45/6 hẻm 12 Lê Lợi", "[ADDRESS]"),
    ("12 đường Lê Lợi, phường Bến Thành", "[ADDRESS]"),
    ("ship về quận 7 bao nhiêu", "ship về [ADDRESS] bao nhiêu"),
    ("q7 ship bao lâu", "[ADDRESS] ship bao lâu"),
    ("tp Hồ Chí Minh", "[ADDRESS]"),
    ("12 Nguyễn Huệ, Q1", "[ADDRESS]"),
    ("giao 12 Nguyễn Huệ, Q1 nha", "giao [ADDRESS] nha"),
    ("123 nguyễn trãi quận 5", "[ADDRESS]"),
    ("88 Đồng Khởi Quận 1", "[ADDRESS]"),
    ("25 le loi q1", "[ADDRESS]"),
    ("ship tới 25 le loi q1 giúp", "ship tới [ADDRESS] giúp"),
    ("số 5, đường số 7", "[ADDRESS]"),
    ("số 5, đường số 7, phường An Phú", "[ADDRESS]"),
])
def test_street_addresses_are_redacted_best_effort(text, expected):
    assert redact_pii(text) == expected


@pytest.mark.parametrize("text,expected", [
    # A quantity before a district stays; only the district is redacted.
    ("mua 2 cái về quận 7", "mua 2 cái về [ADDRESS]"),
    ("lấy 3 áo giao q1", "lấy 3 áo giao [ADDRESS]"),
    ("2 cái ship quan 5 bao nhiêu", "2 cái ship [ADDRESS] bao nhiêu"),
])
def test_quantities_next_to_a_district_are_not_house_numbers(text, expected):
    assert redact_pii(text) == expected


def test_addresses_can_be_kept_and_known_contacts_allowed():
    text = "giao 12 đường Lê Lợi, gọi 0912 345 678 hoặc shop@hexta.vn"
    assert redact_pii(text, addresses=False) == "giao 12 đường Lê Lợi, gọi [PHONE] hoặc [EMAIL]"
    allowed = contact_values("Hotline 0912.345.678, shop@hexta.vn")
    assert redact_pii(text, addresses=False, keep=allowed) == text
    # A shop hotline written without the leading 0 is allowed the same way.
    assert redact_pii("hotline 912 345 678", keep=contact_values("Hotline 0912.345.678")) == "hotline 912 345 678"
    assert redact_pii("") == ""
