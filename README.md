# Candidate Application Process Analysis

CAPA, açık kaynak bir aday başvuru ve davranış analitiği projesidir. Amaç; adayın başvuru formunu doldururken hangi alanlarda zaman harcadığını, nerede zorlandığını ve hangi davranış sinyallerinin incelemeye değer olduğunu yönetici panelinde görünür kılmaktır.

## Mevcut durum

Faz 1 tamamlandı:

- Candidate Next.js uygulaması
- Admin Next.js uygulaması
- Ortak `@capa/ui` component paketi
- Go API
- PostgreSQL compose ortamı
- `/health/live` ve `/health/ready` endpointleri
- Yerel çalıştırma yönergesi
- Kök kalite komutu

## Yerel geliştirme

Ayrıntılı çalıştırma notları: [`agents/docs/runbooks/local-development.md`](agents/docs/runbooks/local-development.md)

Temel komutlar:

```powershell
npm run db:up
npm run dev:api
npm run dev:candidate
npm run dev:admin
```

Toplu kontrol:

```powershell
npm run check
```

## Mimari

Proje tek repo içinde iki Next.js yüzeyi, tek Go modüler monolit API ve PostgreSQL kullanır. Ayrıntılı proje hafızası `agents/` klasöründedir.
