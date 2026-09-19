# CAPA Postman testleri

## Postman Desktop'a ekleme

1. Postman'ı aç.
2. **Import** seçeneğine tıkla.
3. `tests/postman/capa.postman_collection.json` dosyasını seç.
4. Tekrar **Import** ile `tests/postman/local.postman_environment.json` dosyasını ekle.
5. Sağ üstte environment olarak **CAPA Local** seç.

Collection artık Postman içinde görünür. Endpoint veya request değişirse CAPA içindeki JSON dosyası güncellenir. Postman Desktop dosyayı otomatik izlemeyebileceği için güncel dosyayı tekrar import etmek gerekebilir; aynı collection'ı silmeden yeni import ile güncelleyebilirsin.

## Terminalden çalıştırma

API ve PostgreSQL açıkken kök dizinde:

```bash
npm run postman:test
```

Collection sentetik bir e-posta üretir ve şu akışları test eder:

- liveness `200`
- readiness `200`
- candidate application `201`
- invalid application `400`
- duplicate application `409`
- interaction session `201`
- interaction event batch `202`
- tekrar event batch'i idempotent `202`
- yetkisiz admin listesi `401`

Gerçek kişisel veri kullanılmaz.
