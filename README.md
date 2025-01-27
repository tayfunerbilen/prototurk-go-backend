# ProtoTürk API

ProtoTürk, Türkçe soru-cevap platformudur. Bu repository, platformun Go ile yazılmış backend API'sini içerir.

## Gereksinimler

Projeyi çalıştırmadan önce aşağıdaki yazılımların kurulu olduğundan emin olun:

- [Go 1.23](https://golang.org/doc/install) veya üzeri
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Teknolojiler

- Go 1.23
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
└── migrations/         # Database migrasyonları
```

## Kurulum

1. Repository'yi klonlayın:
```bash
git clone https://github.com/tayfunerbilen/prototurk-go-backend
cd prototurk-go-backend
```

2. Docker Compose ile PostgreSQL'i başlatın (veritabanı ve kullanıcı otomatik oluşturulacaktır):
```bash
docker compose up -d
```

3. Gerekli Go paketlerini yükleyin:
```bash
go mod download
go mod tidy
```

4. Örnek çevre değişkenleri dosyasını kopyalayın:
```bash
cp .env.example .env
```

5. `.env` dosyasını düzenleyin (Docker Compose ile uyumlu değerler):
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

7. API http://localhost:8080 adresinde çalışmaya başlayacaktır.

### Olası Sorunlar ve Çözümleri

1. PostgreSQL bağlantı hatası alırsanız:
   - Docker konteynerinin çalıştığından emin olun: `docker ps`
   - Veritabanı bağlantı bilgilerinin doğru olduğunu kontrol edin
   - Docker loglarını kontrol edin: `docker logs prototurk_db`

2. Go paketleri ile ilgili sorun yaşarsanız:
   - Go modüllerini temizleyip tekrar yükleyin:
     ```bash
     go clean -modcache
     go mod download
     ```

3. Go versiyon hatası alırsanız:
   - Projenin Go 1.23 veya üstünü gerektirdiğinden emin olun
   - Sisteminizde yüklü Go versiyonunu kontrol edin: `go version`
   - Eğer eski bir versiyon yüklüyse, Go'yu güncelleyin:
     ```bash
     # macOS için:
     brew install go
     # Yeni bir terminal açın ve versiyon kontrolü yapın:
     go version  # Go 1.23 veya üstü olmalı
     ```
   - go.mod dosyasında Go versiyonunun 1.23 olduğundan emin olun:
     ```go
     module prototurk
     
     go 1.23
     ```

4. "invalid go version" hatası alırsanız:
   - Bu hata genellikle go.mod dosyasındaki versiyon formatının yanlış olmasından kaynaklanır
   - go.mod dosyasındaki versiyon formatının doğru olduğundan emin olun (örn: "1.23" şeklinde olmalı, "1.23.5" değil)

5. Docker Compose hataları:
   - Docker'ın yüklü ve çalışır durumda olduğundan emin olun
   - Docker Compose versiyonunuzun güncel olduğundan emin olun
   - PostgreSQL portunu (5432) başka bir uygulama kullanıyorsa, docker-compose.yml dosyasında portu değiştirin

### Development Ortamı

Hot reload özelliği için Air kullanabilirsiniz:

1. Air'i yükleyin:
```bash
go install github.com/air-verse/air@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

2. `.air.toml` dosyası projede hazır olarak gelir. Air'i başlatın:
```bash
air
```

### API Test Etme

Projeyle birlikte gelen Postman collection'ını kullanabilirsiniz:

1. Postman'i açın
2. Import > Upload Files
3. `prototurk.postman_collection.json` dosyasını seçin
4. Collection içindeki istekleri kullanmaya başlayın

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