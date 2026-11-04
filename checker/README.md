# Wahrwelt checker

Checker использует стандартный ForcAD CLI protocol и десять flag places.
Успешный код возврата - `101`. Для public object numbers используется PFR.

Параметры задачи ForcAD:

```yaml
checker: wahrwelt/checker.py
checker_type: pfr
checker_timeout: 20
places: 10
puts: 10
gets: 1
```

Каждый раунд ForcAD запускает десять PUT на команду, по одному для каждого
места от 1 до 10. `gets: 1` задает одну попытку проверки ранее поставленного
флага.
