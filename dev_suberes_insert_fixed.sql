INSERT INTO users (id, complete_name, email, phone_number, country_code, password, user_type, user_level, color_code_level, user_rating, user_gender, user_profile_image, is_logged_in, user_status, is_active, is_busy, is_document_completed, is_mitra_invited, is_mitra_accepted, is_mitra_rejected, is_mitra_activated, is_suspended, suspended_reason, violation_count, violation_danger_count, is_auto_bid, order_id_running, sub_order_id_running, customer_id_running, service_id_running, sub_service_id_running, latitude, longitude, firebase_token, age, ktp_number, date_of_birth, place_of_birth, ktp_image, kk_image, address, domisili_address, rtrw, sub_district, district, province, city, postal_code, work_experience, user_tool, work_experience_cleaning, is_ex_golife, kind_of_mitra, work_experience_duration, emergency_contact_name, emergency_contact_relation, emergency_contact_country_code, emergency_contact_phone, cover_savings_book, time_invited, place_invited, note_invited, today_order, today_income, total_order, account_balance, pay_pin, disbursement_pin, total_bills, shared_prime, shared_base, shared_secret, private_key_pay_pin, public_key_pay_pin, private_key_disbursement_pin, public_key_disbursement_pin, rejection_count, socket_id, browser_name, is_in_call, nonactivate_reason, activate_reason, registered_from_mobile, created_at, updated_at) VALUES 
('013a8111-0671-46a2-bd79-5ba869dbe07f', 'naufal daffa', 'naufalrachmat91@gmail.com', '+6285712356777', '+62', NULL, 'superadmin', 'no level', '#CECECE', '0.0', 'male', NULL, '0', 'stay', 'no', 'no', '0', '0', '0', '0', '0', '0', NULL, '0', '0', 'no', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '0', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '0', '0', 'Tanpa Alat', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '0', '0', '0', '0', NULL, NULL, '0', '0', '0', '0', '...', '...', '...', '...', '0', NULL, NULL, '0', NULL, NULL, '0', '2026-01-24 21:26:11', '2026-01-24 21:26:11'),
('158c4da4-50fc-4022-808f-8d999e311cac', 'Naufal Rachmat', 'starknaufal123@gmail.com', '+6285712356420', '+62', '...', 'mitra', 'no level', '#CECECE', '4.0', 'male', '/mitra_candidate/...', '0', 'stay', 'yes', 'no', '0', '1', '1', '0', '1', '0', 'Mitra terlibat pencabulan', '0', '0', 'no', NULL, NULL, NULL, NULL, NULL, '-6.1924617', '107.015335', 'csYWyDf...', '22 Tahun', '3171030106011002', '2001-06-1', 'Bekasi', '/mitra_candidate/...', '/mitra_candidate/...', 'Jl. Cempaka Wangi', 'Jl Assalam III Gg. 18 RT 003 RW 015 No 137', '03/015', 'Bahagia', 'Babelan', 'Jawa Barat', 'Bekasi', '17610', NULL, NULL, '0', '0', 'Dengan Alat', NULL, 'Dahlia', 'Ibu Kandung ', '+62', '+6285712356430', NULL, NULL, NULL, NULL, '0', '0', '54', '1488900', NULL, NULL, '0', '0', '0', '4691', '...', '...', '...', '...', '0', NULL, NULL, '0', NULL, NULL, '0', '2024-01-21 15:14:36', '2024-01-21 15:14:36');

INSERT INTO bank_lists (id, bank_image, name, code, disbursement_code, method_type, can_topup, can_disbursement, min_topup, min_disbursement, topup_fee, disbursement_fee, created_at, updated_at) VALUES
('12', '/bank_images/logo_bank_mandiri_cropped.png', 'Bank Mandiri', 'mandiri', 'MANDIRI', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('13', '/bank_images/logo_bank_bri_cropped.png', 'Bank Rakyat Indonesia', 'bri', 'BRI', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('14', '/bank_images/logo_bank_bni_cropped.png', 'BNI (Bank Negara Indonesia)', 'bni', 'BNI', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('15', '/bank_images/logo_bank_bca_cropped.png', 'Bank Central Asia', 'bca', 'BCA', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('16', '/bank_images/logo_bank_bsi_cropped.png', 'BSI (Bank Syariah Indonesia)', 'bsm', 'BSI', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('17', '/bank_images/logo_bank_cimb_cropped.png', 'CIMB Niaga Syariah (UUS)', 'cimb', 'CIMB_UUS', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('18', '/bank_images/logo_bank_muamalat_cropped.png', 'Muamalat', 'muamalat', 'MUAMALAT', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('20', '/bank_images/logo_bank_permata_cropped.png', 'Bank Permata', 'permata', 'PERMATA', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('21', '/bank_images/logo_bank_maybank_cropped.png', 'Maybank Indonesia', 'bii', 'MAYBANK', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('22', '/bank_images/logo_bank_panin.png', 'Panin Bank', 'panin', 'PANIN', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('24', '/bank_images/logo_bank_ocbc_cropped.png', 'OCBC NISP', 'ocbc', 'OCBC', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06'),
('25', '/bank_images/logo_bank_citibank_cropped.png', 'Citibank', 'citibank', 'CITIBANK', 'bank', '0', '1', NULL, '10000', NULL, '2775.0', '2022-09-18 15:30:06', '2022-09-18 15:30:06');

-- Insert Layanan Service (ID 4)
INSERT INTO layanan_services (id, creator_id, layanan_title, layanan_description, layanan_image, layanan_image_size, layanan_image_dimension, is_active, created_at, updated_at) VALUES 
(4, '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Suberes', 'Layanan jasa kebersihan', '/layanan_image/ic_layanan.png', '3184', '71px and 71px', '1', '2024-03-03 13:53:26', '2024-03-03 13:53:26');

-- Insert Category Service (ID 5 - Butuh ID 4)
INSERT INTO category_services (id, layanan_id, category_service, creator_id) VALUES 
(5, 4, 'Suberes Cleaning', '013a8111-0671-46a2-bd79-5ba869dbe07f');

INSERT INTO services (id, parent_id, service_name, service_description, service_image_thumbnail, service_count, service_type, service_category, is_active, max_order_count, payment_timeout, created_at, updated_at, is_residental, service_status) VALUES 
(37, 5, 'Home Cleaning', 'Jam operasional...', 'cuci_karpet.png', NULL, 'Durasi', 'Cleaning', '1', NULL, NULL, '2024-03-03 22:47:01', '2024-03-03 22:47:01', 'true', 'Regular'),
(39, 5, 'Deep Cleaning House', 'Layanan Deep Cleaning...', 'ServiceImageMock.png', NULL, 'Durasi', 'Cleaning', '1', NULL, NULL, '2024-08-11 08:18:16', '2024-08-11 08:18:16', 'true', 'Regular');
-- Banner
INSERT INTO banner_lists (id, creator_id, creator_name, banner_title, banner_body, banner_image, banner_image_size, banner_image_dimension, banner_type, is_revision, is_broadcast, created_at, updated_at) VALUES 
(30, '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Admin 1', 'Lorem ipsum dolor sit amet', '...', '/banners/BNR_IMG_sample.png', '143765', '1280px and 720px', 'promo', '1', '1', '2023-09-23 05:36:31', '2023-09-23 05:36:31');

-- News
INSERT INTO news_lists (id, creator_id, creator_name, news_title, news_body, news_type, news_image, news_image_size, news_image_dimension, is_revision, read_count, like_count, comment_count, share_count, narasumber, is_broadcast, timezone_code, created_at, updated_at) VALUES 
('14', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Admin 1', 'Sanitasi Indonesia', '...', 'News', '/news/NEWS_IMG_sample.png', '362343', '1280px and 720', '1', 0, 0, 0, 0, 'Daffa Naufal', '1', 'Asia/Jakarta', '2024-01-20 05:48:56', '2024-01-20 05:48:56'),
('15', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Admin 1', 'Sanitasi Lorem Ipsum', '...', 'News', '/news/NEWS_IMG_2.png', '11106', '1280px and 720', '1', 0, 0, 0, 0, 'Daffa Naufal Rachmat', '1', 'Asia/Jakarta', '2025-08-18 04:04:27', '2025-08-18 04:04:27'),
('16', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Admin 1', 'dddd', '<p>...</p>', 'News', '/news/NEWS_IMG_3.png', '4637', '1280px and 720', '1', 0, 0, 0, 0, 'daffa', '1', 'Asia/Jakarta', '2025-08-18 05:59:24', '2025-08-18 05:59:24');

-- Syarat Ketentuan
INSERT INTO syarat_ketentuans (id, creator_id, title, body, is_pendaftaran_mitra, is_active, created_at, updated_at) VALUES 
('2', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Persyaratan Mitra', '...', '1', '1', '2021-05-30 18:09:05', '2021-05-30 18:09:05');

-- Terms & Privacy
INSERT INTO terms_conditions (id, creator_id, title, body, is_active, can_select, toc_type, toc_user_type, created_at, updated_at) VALUES 
('41', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Terms of Service Customer', '...', '1', '1', 'terms_of_service', 'customer', '2024-01-04 05:59:33', '2024-01-04 05:59:33'),
('42', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Privacy Policy Customer', '...', '1', '1', 'privacy_policy', 'customer', '2024-01-04 06:00:18', '2024-01-04 06:00:18'),
('43', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Terms of Condition Mitra', '...', '1', '1', 'terms_of_condition', 'mitra', '2024-01-04 06:00:59', '2024-01-04 06:00:59'),
('44', '013a8111-0671-46a2-bd79-5ba869dbe07f', 'Privacy Policy Mitra', '...', '1', '1', 'privacy_policy', 'mitra', '2024-01-04 06:01:48', '2024-01-04 06:01:48');

-- Payments
INSERT INTO payments (id, icon, is_active, title, type, "desc", created_at, updated_at) VALUES 
('6', '/payments/ic_card.png', '1', 'Virtual Account', 'virtual account', 'Gampang bayarnya', '2022-09-14 20:55:00', '2022-09-14 20:55:00'),
('8', '/payments/ic_cash.png', '1', 'Tunai', 'tunai', 'Uang pas ya', '2022-09-14 20:55:00', '2022-09-14 20:55:00'),
('9', '/payments/ic_cash_2.png', '0', 'Saldo', 'balance', 'Saldo terpotong', '2022-09-14 20:55:00', '2022-09-14 20:55:00'),
('10', '/payments/wallet.png', '1', 'E-Wallet', 'ewallet', 'Pakai DANA, OVO, dll', '2023-01-19 21:04:21', '2023-01-19 21:04:21');

-- Sub Payments
INSERT INTO sub_payments (id, payment_id, icon, title, title_payment, enabled, "desc", created_at, updated_at) VALUES 
('5', '6', '/sub_payments/mandiri.png', 'Mandiri Virtual Account', 'MANDIRI', '1', 'Mandiri VA', NULL, NULL),
('6', '6', '/sub_payments/bca.png', 'BCA Virtual Account', 'BCA', '1', 'BCA VA', NULL, NULL),
('7', '8', '/sub_payments/tunai.png', 'Tunai', 'TUNAI', '1', 'TUNAI', NULL, NULL),
('10', '9', '/sub_payments/saldo.png', 'Saldo Suberes', 'BALANCE', '1', 'Pakai saldo', '2022-05-08 15:30:36', '2022-05-08 15:30:36'),
('11', '10', '/sub_payments/spay.png', 'ShopeePay', 'ID_SHOPEEPAY', '0', 'ShopeePay', '2023-01-19 21:06:56', '2023-01-19 21:06:56'),
('12', '10', '/sub_payments/dana.png', 'DANA', 'ID_DANA', '1', 'DANA', '2023-01-19 21:07:19', '2023-01-19 21:07:19');

-- Tutorials
INSERT INTO sub_payment_tutorials (id, payment_id, sub_payment_id, title, description, created_at, updated_at) VALUES 
('1', '6', 5, NULL, '1. Buka Livin...', '2022-09-02 06:40:00', '2022-09-02 06:40:00'),
('2', '6', 6, NULL, '1. Buka m-BCA...', '2022-09-02 06:40:00', '2022-09-02 06:40:00'),
('3', '8', 7, NULL, NULL, '2022-09-02 06:40:00', '2022-09-02 06:40:00'),
('4', '9', 10, NULL, NULL, '2022-05-08 15:30:36', '2022-05-08 15:30:36');

