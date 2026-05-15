INSERT INTO topics (name, description) VALUES
    ('Matematika Wajib',           'Kalkulus, analisis matematika, dan konsep dasar'),
    ('Matematika Lanjutan',        'Kurva, integrasi, dan analisis matematika'),
    ('Fisika',                     'Mekanika, termodinamika, gelombang, listrik-magnet, dan fisika modern'),
    ('Kimia',                      'Struktur atom, ikatan kimia, stoikiometri, larutan, dan reaksi kimia'),
    ('Biologi',                    'Sel, genetika, ekologi, anatomi, dan evolusi'),
    ('Bahasa Indonesia',           'Teks, ejaan, tata bahasa, sastra, dan keterampilan berbahasa'),
    ('Bahasa Inggris',             'Grammar, vocabulary, reading comprehension, dan writing'),
    ('Sejarah',                    'Sejarah Indonesia dan dunia, periodisasi, dan analisis peristiwa'),
    ('Geografi',                   'Peta, lingkungan hidup, sumber daya alam, dan dinamika kependudukan'),
    ('Ekonomi',                    'Mikro-makroekonomi, manajemen, akuntansi, dan koperasi'),
    ('Sosiologi',                  'Struktur sosial, budaya, interaksi sosial, dan perubahan masyarakat'),
    ('TIK',                        'Teknologi informasi, pemrograman dasar, dan literasi digital'),
    ('Pendidikan Kewarganegaraan', 'Pancasila, UUD 1945, sistem pemerintahan, dan hak-kewajiban')
ON CONFLICT (name) DO NOTHING;
