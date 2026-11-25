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
])
def test_phones_and_emails_become_typed_placeholders(text, expected):
    assert redact_pii(text) == expected


@pytest.mark.parametrize("text", [
    "giá 1.250.000đ còn không", "giá 299.000 ạ", "mã đơn 1234567890", "size 42 còn 5 đôi",
    "mua quan size 30", "2 Áo Polo đen", "có đường không", "84.000đ",
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
])
def test_street_addresses_are_redacted_best_effort(text, expected):
    assert redact_pii(text) == expected


def test_addresses_can_be_kept_and_known_contacts_allowed():
    text = "giao 12 đường Lê Lợi, gọi 0912 345 678 hoặc shop@hexta.vn"
    assert redact_pii(text, addresses=False) == "giao 12 đường Lê Lợi, gọi [PHONE] hoặc [EMAIL]"
    allowed = contact_values("Hotline 0912.345.678, shop@hexta.vn")
    assert redact_pii(text, addresses=False, keep=allowed) == text
    assert redact_pii("") == ""
