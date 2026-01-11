-- Bangladesh Stations Seed Data
-- Generated from bangladesh_administrative_divisions.json
-- Total Stations: 64
-- ============================================================================

\c travio_catalog

-- Insert all Bangladesh district headquarters as stations
INSERT INTO stations (id, code, name, city, state, country, latitude, longitude, status) VALUES
  ('80449607-2125-58cd-bf86-1732871be373', 'COM', 'Comilla', 'Comilla', 'Chattogram', 'Bangladesh', 23.4682747, 91.1788135, 'active'),  -- কুমিল্লা
  ('7edcb3f2-a522-5b98-b65a-b0f343e32a58', 'FEN', 'Feni', 'Feni', 'Chattogram', 'Bangladesh', 23.023231, 91.3840844, 'active'),  -- ফেনী
  ('2ce76550-1d07-506c-acd3-f072bd904323', 'BRA', 'Brahmanbaria', 'Brahmanbaria', 'Chattogram', 'Bangladesh', 23.9570904, 91.1119286, 'active'),  -- ব্রাহ্মণবাড়িয়া
  ('adc950bc-e3fa-52d1-b162-c78fe4f5b0c2', 'RAN', 'Rangamati', 'Rangamati', 'Chattogram', 'Bangladesh', 22.65561018, 92.17541121, 'active'),  -- রাঙ্গামাটি
  ('01103775-0d01-5ee1-9cf7-d3b96e21e588', 'NOA', 'Noakhali', 'Noakhali', 'Chattogram', 'Bangladesh', 22.869563, 91.099398, 'active'),  -- নোয়াখালী
  ('d2f92e18-fa3a-5267-91fd-58f38af52339', 'CHA', 'Chandpur', 'Chandpur', 'Chattogram', 'Bangladesh', 23.2332585, 90.6712912, 'active'),  -- চাঁদপুর
  ('854267f1-1da3-5c83-b501-bc77eb30005f', 'LAK', 'Lakshmipur', 'Lakshmipur', 'Chattogram', 'Bangladesh', 22.942477, 90.841184, 'active'),  -- লক্ষ্মীপুর
  ('5fa58708-8aa2-53ff-bb5c-8e97d608a093', 'CHT', 'Chattogram', 'Chattogram', 'Chattogram', 'Bangladesh', 22.335109, 91.834073, 'active'),  -- চট্টগ্রাম
  ('4229dc7c-f9c8-5b76-b8dd-99f4ef72fbb8', 'COX', 'Coxsbazar', 'Coxsbazar', 'Chattogram', 'Bangladesh', 21.44315751, 91.97381741, 'active'),  -- কক্সবাজার
  ('77b27f5f-90d0-5fdf-aaaa-27832d3b33dc', 'KHA', 'Khagrachhari', 'Khagrachhari', 'Chattogram', 'Bangladesh', 23.119285, 91.984663, 'active'),  -- খাগড়াছড়ি
  ('f8640da8-e5da-5d6b-8e3d-db70475b3bce', 'BAN', 'Bandarban', 'Bandarban', 'Chattogram', 'Bangladesh', 22.1953275, 92.2183773, 'active'),  -- বান্দরবান
  ('d8d79ac5-b996-537b-8afb-0d45fb53f741', 'SIR', 'Sirajganj', 'Sirajganj', 'Rajshahi', 'Bangladesh', 24.4533978, 89.7006815, 'active'),  -- সিরাজগঞ্জ
  ('2ffa3bb0-56a7-564c-a290-604be9efd23b', 'PAB', 'Pabna', 'Pabna', 'Rajshahi', 'Bangladesh', 23.998524, 89.233645, 'active'),  -- পাবনা
  ('87c4cba4-5a8c-52f6-b09a-f9a3f7794ae3', 'BOG', 'Bogura', 'Bogura', 'Rajshahi', 'Bangladesh', 24.8465228, 89.377755, 'active'),  -- বগুড়া
  ('12b59d63-f0a3-5772-86a4-10fccaa3e72b', 'RAJ', 'Rajshahi', 'Rajshahi', 'Rajshahi', 'Bangladesh', 24.37230298, 88.56307623, 'active'),  -- রাজশাহী
  ('d41f56c3-594c-55f2-ba1c-eb3199da5b44', 'NAT', 'Natore', 'Natore', 'Rajshahi', 'Bangladesh', 24.420556, 89.000282, 'active'),  -- নাটোর
  ('c67c70bd-9517-50c9-b062-c08ffc13a7e0', 'JOY', 'Joypurhat', 'Joypurhat', 'Rajshahi', 'Bangladesh', 25.09636876, 89.04004280, 'active'),  -- জয়পুরহাট
  ('66d380c3-2ee0-5b32-ac0b-f38bb17bedf6', 'CHP', 'Chapainawabganj', 'Chapainawabganj', 'Rajshahi', 'Bangladesh', 24.5965034, 88.2775122, 'active'),  -- চাঁপাইনবাবগঞ্জ
  ('a926c5a6-7ff1-5052-8ce4-a151b5f4d1ea', 'NAO', 'Naogaon', 'Naogaon', 'Rajshahi', 'Bangladesh', 24.83256191, 88.92485205, 'active'),  -- নওগাঁ
  ('4241104c-d0e7-564c-8b39-46c83561708d', 'JAS', 'Jashore', 'Jashore', 'Khulna', 'Bangladesh', 23.16643, 89.2081126, 'active'),  -- যশোর
  ('7cedd528-8206-5be3-af9a-1c894596d714', 'SAT', 'Satkhira', 'Satkhira', 'Khulna', 'Bangladesh', 22.7180905, 89.0687033, 'active'),  -- সাতক্ষীরা
  ('da241951-75e2-52fc-a8b3-e84c47f42ef5', 'MEH', 'Meherpur', 'Meherpur', 'Khulna', 'Bangladesh', 23.762213, 88.631821, 'active'),  -- মেহেরপুর
  ('7e1460a9-8cce-5c0b-a86d-59da4c8192e7', 'NAR', 'Narail', 'Narail', 'Khulna', 'Bangladesh', 23.172534, 89.512672, 'active'),  -- নড়াইল
  ('19e9a12e-5e8a-5927-9d00-a78955c2788d', 'CHU', 'Chuadanga', 'Chuadanga', 'Khulna', 'Bangladesh', 23.6401961, 88.841841, 'active'),  -- চুয়াডাঙ্গা
  ('a6c3eeed-7e91-5e18-8fe5-5998a682eac8', 'KUS', 'Kushtia', 'Kushtia', 'Khulna', 'Bangladesh', 23.901258, 89.120482, 'active'),  -- কুষ্টিয়া
  ('2228d58c-c0a1-53e5-846b-e8a502daeb61', 'MAG', 'Magura', 'Magura', 'Khulna', 'Bangladesh', 23.487337, 89.419956, 'active'),  -- মাগুরা
  ('9cf71956-150b-5364-8091-5847eb3420ab', 'KHU', 'Khulna', 'Khulna', 'Khulna', 'Bangladesh', 22.815774, 89.568679, 'active'),  -- খুলনা
  ('3469cb32-997e-5ee7-b8ef-c76c754e0a03', 'BAG', 'Bagerhat', 'Bagerhat', 'Khulna', 'Bangladesh', 22.651568, 89.785938, 'active'),  -- বাগেরহাট
  ('09d1b01d-782d-500b-8574-df20efe9d8fc', 'JHE', 'Jhenaidah', 'Jhenaidah', 'Khulna', 'Bangladesh', 23.5448176, 89.1539213, 'active'),  -- ঝিনাইদহ
  ('1d3d6d5a-3a28-5112-969d-3f4822b585aa', 'JHA', 'Jhalakathi', 'Jhalakathi', 'Rangpur', 'Bangladesh', 22.6422689, 90.2003932, 'active'),  -- ঝালকাঠি
  ('0dd549fe-6e15-5b7b-9cf3-6069214ede60', 'PAT', 'Patuakhali', 'Patuakhali', 'Rangpur', 'Bangladesh', 22.3596316, 90.3298712, 'active'),  -- পটুয়াখালী
  ('937a2ce8-a817-52e7-86ad-2fd9418d7de4', 'PIR', 'Pirojpur', 'Pirojpur', 'Rangpur', 'Bangladesh', 22.5781398, 89.9983909, 'active'),  -- পিরোজপুর
  ('c05e6b9c-e1c5-561b-af15-ac7b400fe20f', 'BAR', 'Barisal', 'Barisal', 'Rangpur', 'Bangladesh', 22.7004179, 90.3731568, 'active'),  -- বরিশাল
  ('39dd6ead-26e7-52d4-8f1c-2ab8010bcfd9', 'BHO', 'Bhola', 'Bhola', 'Rangpur', 'Bangladesh', 22.685923, 90.648179, 'active'),  -- ভোলা
  ('ad42d307-7043-56ad-b357-c8237a98c7c7', 'BRG', 'Barguna', 'Barguna', 'Rangpur', 'Bangladesh', 22.159182, 90.125581, 'active'),  -- বরগুনা
  ('3ac3674c-b75c-50c8-93ec-22988a656d53', 'SYL', 'Sylhet', 'Sylhet', 'Barishal', 'Bangladesh', 24.8897956, 91.8697894, 'active'),  -- সিলেট
  ('9320630f-caee-52e1-a50b-88dbe30f1c0f', 'MOU', 'Moulvibazar', 'Moulvibazar', 'Barishal', 'Bangladesh', 24.482934, 91.777417, 'active'),  -- মৌলভীবাজার
  ('521607d8-d601-5ed9-91e2-a63222cc467a', 'HAB', 'Habiganj', 'Habiganj', 'Barishal', 'Bangladesh', 24.374945, 91.41553, 'active'),  -- হবিগঞ্জ
  ('f0aff3f9-3146-574a-82c8-86b125088598', 'SUN', 'Sunamganj', 'Sunamganj', 'Barishal', 'Bangladesh', 25.0658042, 91.3950115, 'active'),  -- সুনামগঞ্জ
  ('49201457-af3f-59d1-9443-ad72e8e14e82', 'NRS', 'Narsingdi', 'Narsingdi', 'Dhaka', 'Bangladesh', 23.932233, 90.71541, 'active'),  -- নরসিংদী
  ('77163c76-6521-550a-8983-561659a8889f', 'GAZ', 'Gazipur', 'Gazipur', 'Dhaka', 'Bangladesh', 24.0022858, 90.4264283, 'active'),  -- গাজীপুর
  ('b00c5a8b-67a6-56ed-9879-f50abcd1ddc0', 'SHA', 'Shariatpur', 'Shariatpur', 'Dhaka', 'Bangladesh', 23.2060195, 90.3477725, 'active'),  -- শরীয়তপুর
  ('7238f1f5-f59d-57dc-8c51-a25eefb46827', 'NRY', 'Narayanganj', 'Narayanganj', 'Dhaka', 'Bangladesh', 23.63366, 90.496482, 'active'),  -- নারায়ণগঞ্জ
  ('3f83ab7a-c1fd-5c82-b24e-48405a103303', 'TAN', 'Tangail', 'Tangail', 'Dhaka', 'Bangladesh', 24.264145, 89.918029, 'active'),  -- টাঙ্গাইল
  ('cf46a493-ac20-5fc5-9a7f-8b1e3c46457c', 'KIS', 'Kishoreganj', 'Kishoreganj', 'Dhaka', 'Bangladesh', 24.444937, 90.776575, 'active'),  -- কিশোরগঞ্জ
  ('08be5cec-a2fa-5e7b-a660-3141f554e5b8', 'MAN', 'Manikganj', 'Manikganj', 'Dhaka', 'Bangladesh', 23.8602262, 90.0018293, 'active'),  -- মানিকগঞ্জ
  ('0a987eb7-ecdb-5c37-9802-16bfbbbb75b7', 'DHA', 'Dhaka', 'Dhaka', 'Dhaka', 'Bangladesh', 23.7115253, 90.4111451, 'active'),  -- ঢাকা
  ('223c2790-e839-54e7-ad03-d50900183c61', 'MUN', 'Munshiganj', 'Munshiganj', 'Dhaka', 'Bangladesh', 23.5435742, 90.5354327, 'active'),  -- মুন্সিগঞ্জ
  ('1a8a68a8-7102-5fdd-91a0-d153ca7cb220', 'RJB', 'Rajbari', 'Rajbari', 'Dhaka', 'Bangladesh', 23.7574305, 89.6444665, 'active'),  -- রাজবাড়ী
  ('6770de9c-fd86-5dea-9e08-5573f657b097', 'MAD', 'Madaripur', 'Madaripur', 'Dhaka', 'Bangladesh', 23.164102, 90.1896805, 'active'),  -- মাদারীপুর
  ('bcd6e067-a670-5da3-a1de-2b51f5ebb705', 'GOP', 'Gopalganj', 'Gopalganj', 'Dhaka', 'Bangladesh', 23.0050857, 89.8266059, 'active'),  -- গোপালগঞ্জ
  ('9120709d-8110-5d39-8b13-54f032424740', 'FAR', 'Faridpur', 'Faridpur', 'Dhaka', 'Bangladesh', 23.6070822, 89.8429406, 'active'),  -- ফরিদপুর
  ('b7362293-9eba-53af-b833-4b4b7d97ddbf', 'PAN', 'Panchagarh', 'Panchagarh', 'Sylhet', 'Bangladesh', 26.3411, 88.5541606, 'active'),  -- পঞ্চগড়
  ('cbf17559-a4fd-52d2-ae22-7ef072fdee30', 'DIN', 'Dinajpur', 'Dinajpur', 'Sylhet', 'Bangladesh', 25.6217061, 88.6354504, 'active'),  -- দিনাজপুর
  ('dcf78920-9244-53d7-91a4-766382c844ef', 'LAL', 'Lalmonirhat', 'Lalmonirhat', 'Sylhet', 'Bangladesh', 25.9165451, 89.4532409, 'active'),  -- লালমনিরহাট
  ('b458c3ee-afcf-5e61-bf35-0ba4c7c070dc', 'NIL', 'Nilphamari', 'Nilphamari', 'Sylhet', 'Bangladesh', 25.931794, 88.856006, 'active'),  -- নীলফামারী
  ('20d0f671-2e03-5522-89fb-e093038b0668', 'GAI', 'Gaibandha', 'Gaibandha', 'Sylhet', 'Bangladesh', 25.328751, 89.528088, 'active'),  -- গাইবান্ধা
  ('86d5a3e9-e654-5ddc-b93f-f5d1baeda590', 'THA', 'Thakurgaon', 'Thakurgaon', 'Sylhet', 'Bangladesh', 26.0336945, 88.4616834, 'active'),  -- ঠাকুরগাঁও
  ('9f08dc53-d540-57b7-b41b-586855da848b', 'RNG', 'Rangpur', 'Rangpur', 'Sylhet', 'Bangladesh', 25.7558096, 89.244462, 'active'),  -- রংপুর
  ('5885da69-fff3-5efe-93dd-9e922fc01b0d', 'KUR', 'Kurigram', 'Kurigram', 'Sylhet', 'Bangladesh', 25.805445, 89.636174, 'active'),  -- কুড়িগ্রাম
  ('98aa6113-1570-5549-a859-a1706df0b949', 'SHE', 'Sherpur', 'Sherpur', 'Mymensingh', 'Bangladesh', 25.0204933, 90.0152966, 'active'),  -- শেরপুর
  ('0abccc46-98fb-5b1a-9b5c-5b46cb0e6a85', 'MYM', 'Mymensingh', 'Mymensingh', 'Mymensingh', 'Bangladesh', 24.7465670, 90.4072093, 'active'),  -- ময়মনসিংহ
  ('3b8f36d9-275e-5a9e-86fd-7f7409d73536', 'JAM', 'Jamalpur', 'Jamalpur', 'Mymensingh', 'Bangladesh', 24.937533, 89.937775, 'active'),  -- জামালপুর
  ('b6c6f929-9500-51f1-8de5-9c51c476ed45', 'NET', 'Netrokona', 'Netrokona', 'Mymensingh', 'Bangladesh', 24.870955, 90.727887, 'active');  -- নেত্রকোণা

ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  latitude = EXCLUDED.latitude,
  longitude = EXCLUDED.longitude,
  updated_at = NOW();

-- Verification query
-- SELECT state, COUNT(*) as station_count FROM stations WHERE country = 'Bangladesh' GROUP BY state ORDER BY state;
