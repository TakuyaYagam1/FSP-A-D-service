# Wahrwelt - A/D сервис на финал ЧРПР РФ

## Логистика и доставка

Wahrwelt - логистическая система контроля экспорта/импорта товаров и услуг.
Действия происходят в научном городе Орле, где строят первый российский узел "Рассвет".
Логистическая служба доставляет для него материалы и оборудование, а Wahrwelt
ведет сведения о поставках и пропусках, лицензии, акты досмотра, договоры,
счета и платежи.

## Состав проекта

- [`wahrwelt/`](wahrwelt/) - уязвимая версия сервиса
- [`wahrwelt-patch/`](wahrwelt-patch/) - исправленная версия
- [`checker/`](checker/) - чекер для 10 типов флагов
- [`wahrwelt-patch/writeup/`](wahrwelt-patch/writeup/) - описание уязвимостей, цепочка и эксплойты

Для запуска сервиса:

```sh
cd wahrwelt
make prepare
```

Инструкции по работе во время раунда находятся в [`wahrwelt/README.md`](wahrwelt/README.md).

## Уязвимости

| ID | Уязвимость | Целевое поле |
| --- | --- | --- |
| V1 | IDOR редакции манифеста | `sealed_route` |
| V2 | SSRF проверки сертификата | `origin_phrase` |
| V3 | SQLI поиска счетов | `payment_ref` |
| V4 | Mass assignment акта досмотра | `inspector_note` |
| V5 | Токен лицензии без привязки к объекту | `release_pin` |
| V6 | SSTI в шаблоне договора | `crew_phrase` |
| V7 | Повторное согласование слота | `slot_code` |
| V8 | Обход проверки поля квоты сменой регистра | `quota_code` |
| V9 | Ghostscript CVE-2024-29510 в рендерере | `archive_code` |
| V10 | Доверие клиентскому `X-Forwarded-Role` | `gate_word` |

Подробности и команды запуска эксплойтов - в [`wahrwelt-patch/writeup/README.md`](wahrwelt-patch/writeup/README.md).
Связь уязвимостей с общим сценарием описана в [`wahrwelt-patch/writeup/CHAINS.md`](wahrwelt-patch/writeup/CHAINS.md).

Автор: [TakuyaYagam1](https://t.me/TakuyaYagam1)
