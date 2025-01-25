# ProtoTürk API

ProtoTürk, Türkçe soru-cevap platformudur. Bu repository, platformun Go ile yazılmış backend API'sini içerir.

## Gereksinimler

Projeyi çalıştırmadan önce aşağıdaki yazılımların kurulu olduğundan emin olun:

- [Go 1.21](https://golang.org/doc/install) veya üzeri
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Teknolojiler

- Go 1.21
- PostgreSQL
- GORM (ORM)
- Gin (Web Framework)
- JWT (Authentication)
- Air (Hot Reload)

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
DB_USER=prototurk
DB_PASSWORD=prototurk123
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

2. Uygulamayı Air ile başlatın:
```bash
air
```

## API Endpoints

### Auth

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
- `USERNAME_EXISTS`: Kullanıcı adı zaten mevcut
- `EMAIL_EXISTS`: Email zaten mevcut
- `INVALID_CREDENTIALS`: Geçersiz kullanıcı adı/email veya şifre
- `USER_BANNED`: Kullanıcı yasaklanmış
- `USER_NOT_FOUND`: Kullanıcı bulunamadı

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