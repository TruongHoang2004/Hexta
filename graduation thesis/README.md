# Khung LaTeX khóa luận tốt nghiệp

Khung tham khảo bố cục PDF mẫu được cung cấp, không sao chép nội dung nghiên cứu. Thông tin trường trên bìa lấy theo mẫu; thông tin sinh viên, giảng viên, tên đề tài và nội dung đều cần điền lại.

## Bắt đầu

1. Mở `config/metadata.tex`, sửa nhóm **THÔNG TIN KHÓA LUẬN**.
2. Điền các đoạn `[ ... ]` và đổi tên chương/mục cho phù hợp đề tài.
3. Đặt logo chính thức tại `assets/logo.png`; nếu chưa có, bìa hiển thị ô giữ chỗ.
4. Thay hình minh họa, công thức và tài liệu tham khảo mẫu bằng nội dung thật.

## Biên dịch

Khung hỗ trợ cả pdfLaTeX (font Latin Modern với mã hóa T5 cho tiếng Việt), XeLaTeX và LuaLaTeX (font TeX Gyre). Nếu trình biên dịch mặc định là pdfLaTeX, không cần đổi engine. Hình thức chữ có thể khác nhẹ giữa các engine.

Trên Overleaf: tải cả thư mục lên, chọn **Main document: main.tex**, **Compiler: XeLaTeX**, rồi Recompile. Có thể cần biên dịch lại để mục lục và tham chiếu cập nhật.

Trên máy đã có TeX Live hoặc MacTeX:

```sh
cd 'graduation thesis'
xelatex -interaction=nonstopmode -halt-on-error main.tex
xelatex -interaction=nonstopmode -halt-on-error main.tex
```

Hoặc: `latexmk -xelatex main.tex`.

Font TeX Gyre Termes hỗ trợ tiếng Việt và có sẵn trong các bộ TeX đầy đủ. Mẫu dùng A4, cỡ chữ 12pt, giãn dòng 1.3, lề trái 3 cm, phải 2 cm, trên/dưới 2.5 cm. Đây là cấu hình gần với PDF tham khảo, không phải xác nhận quy chuẩn của trường; chỉnh theo hướng dẫn chính thức khi có.

## Bố cục

Bìa không hiện số trang. Từ Tóm tắt bắt đầu đánh số Ả Rập liên tục, giống mẫu. Tiếp theo là lời cảm ơn, lời cam đoan, mục lục, danh sách hình/bảng, năm chương, tài liệu tham khảo và phụ lục. Hình, bảng và công thức đánh số theo chương; mục lục và liên kết tham chiếu được tạo tự động.

Mã nguồn được chia thành nhiều file và ghép bằng `\input{...}` trong `main.tex`:

```text
main.tex                         # File chính: thứ tự các phần
config/preamble.tex              # Gói LaTeX, font, lề và định dạng
config/metadata.tex              # Tên đề tài, sinh viên, giảng viên, năm
frontmatter/cover.tex            # Bìa
frontmatter/abstract.tex         # Tóm tắt
frontmatter/acknowledgements.tex # Lời cảm ơn
frontmatter/declaration.tex      # Lời cam đoan
frontmatter/contents.tex         # Mục lục và danh sách hình/bảng
chapters/01-introduction.tex
chapters/02-background.tex
chapters/03-system-design.tex
chapters/04-implementation-evaluation.tex
chapters/05-conclusion.tex
backmatter/references.tex        # Tài liệu tham khảo
appendices/database.tex
appendices/installation.tex
assets/                         # Logo và hình ảnh
```

Luôn biên dịch `main.tex` từ thư mục `graduation thesis`, không biên dịch riêng từng chương. Các file con có dòng `% !TeX root = ../main.tex` để trình soạn thảo nhận biết tài liệu chính. Đường dẫn hình ảnh trong các file con vẫn tính từ thư mục chứa `main.tex`, ví dụ `assets/kien-truc.png`.

Muốn chia một chương thành các mục riêng, tạo file trong `chapters/` rồi dùng `\input{chapters/ten-muc}` ngay tại vị trí cần chèn trong file chương. File con chỉ chứa nội dung; không thêm `\documentclass` hay `\begin{document}`.

Trình biên dịch tích hợp Codex hiện chỉ hỗ trợ tài liệu một file, nên có thể không xem trước được cấu trúc nhiều file này. Dùng MacTeX/LaTeX Workshop hoặc tải **toàn bộ thư mục** lên Overleaf để biên dịch.

## Chèn hình và tham chiếu

```tex
\begin{figure}[htbp]
  \centering
  \includegraphics[width=0.85\linewidth]{assets/kien-truc.png}
  \caption{Kiến trúc hệ thống}
  \label{fig:system}
\end{figure}
```

Dùng `Hình~\ref{fig:system}`, `Bảng~\ref{tab:environment}`, `\eqref{eq:objective}` và `\cite{reference-placeholder}`. Thêm nguồn thực vào môi trường `thebibliography`, mỗi nguồn có một khóa `\bibitem{...}` duy nhất.
