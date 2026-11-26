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
    "lấy 2 cái nhé", "2 cái áo size L", "mua 3 đôi giá 1.250.000đ", "đơn DH0912345678 giao chưa", "912.345.678đ",
    "giá 350000000đ", "chuyển 350000000 đồng", "SP0912345678", "tổng 912345678 vnđ", "1234567890", "mã đơn #ORD-2026-912345678",
    # Order codes whose length or prefix cannot be a Vietnamese phone.
    "mã đơn 1234567890", "đơn hàng 091234567890", "order 0012345678", "mã đơn 12345678",
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


@pytest.mark.parametrize("text,expected", [
    # Standalone hamlet, quarter and residential group with a number.
    ("ấp 3 giúp em", "[ADDRESS] giúp em"),
    ("nhà ở kp 5 nha", "nhà ở [ADDRESS] nha"),
    ("KP.2 Linh Trung", "[ADDRESS]"),
    ("khu phố 4 phường Linh Trung", "[ADDRESS]"),
    ("tổ 7 ạ", "[ADDRESS] ạ"),
    ("thôn 2 Đông Anh", "[ADDRESS]"),
    ("tổ 12 khu phố 3 phường Linh Trung", "[ADDRESS]"),
    ("giao tới ấp 3, xã Tân Phú, huyện Củ Chi nha", "giao tới [ADDRESS] nha"),
    # House + street before a known district or city, without any unit keyword.
    ("15B Phan Đình Phùng, Ba Đình, Hà Nội", "[ADDRESS]"),
    ("22 Hai Bà Trưng Hoàn Kiếm", "[ADDRESS]"),
    ("ship 22 hai ba trung hoan kiem nhé", "ship [ADDRESS] nhé"),
    ("giao số 10 Trần Phú Hà Đông giúp", "giao [ADDRESS] giúp"),
    ("12 Hàng Bạc, Hoàn Kiếm", "[ADDRESS]"),
    ("45 Lê Văn Sỹ Phú Nhuận, HCM", "[ADDRESS]"),
    # A city ends an address after a comma; capitalized serving words are street names.
    ("15B Phan Đình Phùng, Hà Nội", "[ADDRESS]"),
    ("giao 15b phan dinh phung, ha noi nhé", "giao [ADDRESS] nhé"),
    ("12 Tô Hiến Thành, Hai Bà Trưng", "[ADDRESS]"),
    ("3 Hàng Mắm Hoàn Kiếm", "[ADDRESS]"),
    ("5 hàng bún, ba đình", "[ADDRESS]"),
    # "kp" without its dot after an address word or a comma.
    ("giao kp3 Linh Trung", "giao [ADDRESS]"),
    ("địa chỉ kp 4 Linh Trung", "địa chỉ [ADDRESS]"),
])
def test_local_units_and_addresses_before_known_places(text, expected):
    assert redact_pii(text) == expected


@pytest.mark.parametrize("text", [
    "tổ chức 5 sự kiện", "tổng 5 cái", "thông tin 3 sản phẩm", "tổ 5 người đặt chung", "ấp trứng",
    "có ship Hà Nội không", "shop ở Hà Nội à", "ship 2 ngày tới Hà Nội", "mua 2 cái về Hà Nội",
    "giá 299.000 ship Hà Nội", "còn 3 màu Hà Nội", "có 2 chi nhánh Hà Nội",
    # Dishes and products named after a city, amounts and product models are not addresses.
    "cho 3 tô bún bò Huế", "2 phở Hà Nội", "1 cà phê Sài Gòn", "10 mè xửng Huế", "đặt 5 bánh cốm Hà Nội",
    "iPhone 15 Pro Max Hà Nội", "299k freeship Hà Nội", "2 phở Ba Đình", "cho 3 tô bún bò, Huế",
    "iPhone 15 Pro Max, Hà Nội", "Loa KP2", "loa KP 2 còn không", "KP2 còn hàng không",
])
def test_local_unit_and_place_lookalikes_are_kept(text):
    assert redact_pii(text) == text


@pytest.mark.parametrize("text", ["tôi muốn 912345678 cái", "lấy 912345678 sp"])
def test_quantities_shaped_like_phones_are_kept(text):
    assert redact_pii(text) == text


@pytest.mark.parametrize("text,expected", [
    # A customer often answers "which order?" with their phone: an order word never keeps a valid phone shape.
    ("kiểm tra đơn hàng 0912345678 giúp em", "kiểm tra đơn hàng [PHONE] giúp em"),
    ("cho em order 0987654321 Lan nhé", "cho em order [PHONE] Lan nhé"),
    ("Đơn hàng: 0987654321, Nguyễn Văn A", "Đơn hàng: [PHONE], Nguyễn Văn A"),
    ("đơn hàng 0912345678 sđt 0987654321", "đơn hàng [PHONE] sđt [PHONE]"),
    ("order 0912345678", "order [PHONE]"),
    ("order #0912345678", "order #[PHONE]"),
    ("mã đơn: 0912345678", "mã đơn: [PHONE]"),
    ("order id 0912345678 giao chưa", "order id [PHONE] giao chưa"),
    ("mã vận đơn 0912 345 678", "mã vận đơn [PHONE]"),
    ("đơn hàng +84 912 345 678", "đơn hàng [PHONE]"),
    ("order 84912345678", "order [PHONE]"),
    ("đơn hàng 0912.345.678", "đơn hàng [PHONE]"),
    ("mã đơn 091-234-5678", "mã đơn [PHONE]"),
    ("số đơn 024 3826 1234", "số đơn [PHONE]"),
    # Without the leading 0 too: the 9-digit form is a valid mobile whatever word precedes it.
    ("mã đơn 912345678", "mã đơn [PHONE]"),
    ("mã: 912345678", "mã: [PHONE]"),
    ("order #912345678", "order #[PHONE]"),
    ("đơn 512345678 sao rồi", "đơn [PHONE] sao rồi"),
])
def test_phones_after_order_words_are_redacted(text, expected):
    assert redact_pii(text) == expected
    # contact_values sees the same phone, so a tenant's own number written after an order word can still be allowed.
    assert contact_values(text) and "[PHONE]" not in redact_pii(text, keep=contact_values(text))


@pytest.mark.parametrize("text,expected", [
    ("gọi 0912345678 về đơn hàng", "gọi [PHONE] về đơn hàng"),
    ("sđt 0912345678 order giúp", "sđt [PHONE] order giúp"),
    ("order xong gọi 0912345678", "order xong gọi [PHONE]"),
    ("đơn giao 0912345678", "đơn giao [PHONE]"),
    ("912345678 cho mình", "[PHONE] cho mình"),
    ("lấy 2 cái, sđt 912345678", "lấy 2 cái, sđt [PHONE]"),
])
def test_phones_near_order_or_quantity_words_are_still_redacted(text, expected):
    assert redact_pii(text) == expected
