-- Mevcut unique constraint'i kaldır
ALTER TABLE admins DROP CONSTRAINT IF EXISTS admins_email_key;

-- Partial unique index oluştur
DROP INDEX IF EXISTS idx_admins_email;
CREATE UNIQUE INDEX idx_admins_email ON admins(email) WHERE deleted_at IS NULL; 