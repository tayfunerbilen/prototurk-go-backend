# ProtoTürk API

ProtoTürk, Türkçe soru-cevap platformudur. Bu repository, platformun Go ile yazılmış backend API'sini içerir.

## Teknolojiler

- Go 1.21
- PostgreSQL
- GORM (ORM)
- Gin (Web Framework)
- JWT (Authentication)
- Air (Hot Reload)

## Zaman Yönetimi

Tüm tarih ve zaman işlemleri UTC+0 olarak kaydedilir ve yönetilir:

- Veritabanında tüm timestamp'ler `TIMESTAMP WITH TIME ZONE` tipinde ve UTC+0 olarak saklanır
- Tüm zaman işlemleri için `pkg/utils/time.go` içindeki UTC-aware fonksiyonlar kullanılır
- GORM model hook'ları (`BeforeCreate`, `BeforeUpdate`) ile tüm timestamp'lerin UTC olması garanti edilir
- JWT token'ları için expiration time UTC olarak hesaplanır

## Proje Yapısı

```
.
├── cmd/
│   └── api/            # Ana uygulama giriş noktası
├── internal/
│   ├── database/       # Database bağlantısı ve konfigürasyonu
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Middleware'ler
│   ├── models/         # Database modelleri
│   └── validator/      # Validasyon kuralları
├── pkg/
│   ├── response/       # Global response yapısı
│   └── errors/         # Global error yapısı
├── migrations/         # Database migrasyonları
└── config/            # Konfigürasyon dosyaları
```

## Kurulum

1. PostgreSQL'i yükleyin ve bir veritabanı oluşturun:
```bash
createdb prototurk
```

2. Repository'yi klonlayın:
```bash
git clone <repo-url>
cd prototurk
```

3. Gerekli Go paketlerini yükleyin:
```bash
go mod download
```

4. `.env` dosyasını oluşturun:
```bash
cp .env.example .env
```

5. `.env` dosyasını düzenleyin:
```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=prototurk
JWT_SECRET=your-secret-key-here
```

6. Uygulamayı başlatın:
```bash
go run cmd/api/main.go
```

### Development Ortamı

Hot reload özelliği için Air kullanabilirsiniz:

1. Air'i yükleyin:
```bash
go install github.com/air-verse/air@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

2. Uygulamayı Air ile başlatın:
```bash
air
```

## API Endpoints

### Auth (User)

#### Register
- **POST** `/api/auth/register`
```json
{
    "username": "test",
    "email": "test@example.com",
    "password": "123456"
}
```

#### Login
- **POST** `/api/auth/login`
```json
{
    "identifier": "test", // username veya email
    "password": "123456"
}
```

#### Me (Authentication Required)
- **GET** `/api/auth/me`
- Headers:
  - Authorization: Bearer <token>

#### Update Profile (Authentication Required)
- **PUT** `/api/auth/profile`
- Headers:
  - Authorization: Bearer <token>
```json
{
    "username": "new_username",  // optional
    "email": "new@example.com"   // optional
}
```
Not: `username` ve `email` alanlarından en az birinin gönderilmesi gerekir. İki alan da opsiyoneldir.

#### Update Password (Authentication Required)
- **PUT** `/api/auth/password`
- Headers:
  - Authorization: Bearer <token>
```json
{
    "current_password": "mevcut123",
    "new_password": "yeni123"
}
```

### Admin

#### Login
- **POST** `/api/admin/login`
```json
{
    "email": "admin@example.com",
    "password": "123456"
}
```

#### Me (Admin Authentication Required)
- **GET** `/api/admin/me`
- Headers:
  - Authorization: Bearer <token>

#### Create Admin (Super Admin Only)
- **POST** `/api/admin`
- Headers:
  - Authorization: Bearer <token>
```json
{
    "email": "newadmin@example.com",
    "name": "New Admin",
    "password": "123456",
    "role": "admin",        // super_admin, admin, editor
    "status": "active"      // active, passive
}
```

#### List Admins (Admin Authentication Required)
- **GET** `/api/admin`
- Headers:
  - Authorization: Bearer <token>

#### Get Admin (Admin Authentication Required)
- **GET** `/api/admin/:id`
- Headers:
  - Authorization: Bearer <token>

#### Update Admin (Super Admin or Self)
- **PUT** `/api/admin/:id`
- Headers:
  - Authorization: Bearer <token>
```json
{
    "email": "updated@example.com",     // optional
    "name": "Updated Name",             // optional
    "password": "newpass",              // optional
    "role": "editor",                   // optional
    "status": "passive"                 // optional
}
```

#### Delete Admin (Super Admin Only)
- **DELETE** `/api/admin/:id`
- Headers:
  - Authorization: Bearer <token>

## Response Format

### Başarılı Response
```json
{
    "success": true,
    "data": {
        // response data
    }
}
```

### Hata Response
```json
{
    "success": false,
    "error": {
        "code": "ERROR_CODE",
        "message": "Error message",
        "details": {} // optional
    }
}
```

## Error Codes

- `VALIDATION_ERROR`: İstek validasyonu başarısız
- `SERVER_ERROR`: Sunucu hatası
- `UNAUTHORIZED`: Yetkilendirme hatası
- `FORBIDDEN`: Yetki yetersiz
- `USERNAME_EXISTS`: Kullanıcı adı zaten mevcut
- `EMAIL_EXISTS`: Email zaten mevcut
- `INVALID_CREDENTIALS`: Geçersiz kullanıcı adı/email veya şifre
- `USER_BANNED`: Kullanıcı yasaklanmış
- `USER_NOT_FOUND`: Kullanıcı bulunamadı
- `ACCOUNT_INACTIVE`: Hesap aktif değil
- `NOT_FOUND`: Kayıt bulunamadı
- `INVALID_ROLE`: Geçersiz rol
- `INVALID_STATUS`: Geçersiz durum

## User Status

- `active`: Aktif kullanıcı
- `passive`: Pasif kullanıcı
- `banned`: Yasaklanmış kullanıcı

## Development

1. Yeni bir özellik eklerken branch oluşturun:
```bash
git checkout -b feature/feature-name
```

2. Kodunuzu yazın ve commit edin:
```bash
git add .
git commit -m "feat: add new feature"
```

3. Pull request oluşturun

## Yapılacaklar

- [ ] Admin paneli için endpoints
- [ ] Soru-cevap endpoints
- [ ] Kullanıcı profili güncelleme
- [ ] Şifre sıfırlama
- [ ] Email doğrulama
- [ ] Rate limiting
- [ ] Cache mekanizması
- [ ] Test coverage
- [ ] API documentation (Swagger)
- [ ] Docker deployment
- [ ] CI/CD pipeline