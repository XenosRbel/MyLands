# Общий harness: план реализации

> **For agentic workers:** Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Настроить общий harness Codex и Claude Code в текущей ветке.

**Architecture:** Общая реализация на Python обслуживает команды и JSON-события обоих инструментов.
Отдельные конфигурации подключают одинаковые проверки без изменения глобальной среды.

**Tech Stack:** Python ≥3.11, Git, Make, Codex CLI, Claude Code.

**Spec:** `docs/superpowers/specs/2026-11-01-agent-harness-design.md`.

## Global Constraints

- Изменения остаются в ветке `improve/game-experience-audit-20261031`.
- Сохранить `AGENT.md` и ссылку `AGENTS.md` как источник инженерных правил.
- Не менять глобальные модели, плагины, память и разрешения.
- Не выдавать инструкцию агенту за техническую блокировку.
- Не объявлять разбор shell-команды полноценной песочницей.

## Review Focus

- Пути с пробелами и вложенный рабочий каталог должны сохранять правильный корень проекта.
- Codex apply_patch должен проверять исходный путь и назначение Move to.
- Новая миграция разрешена; миграция из HEAD защищена даже после переименования.
- Ошибка или таймаут проверки блокирует первый Stop без бесконечного повторения.
- Отсутствующая зависимость должна давать явную ошибку, а не успешный результат.

## Task 1: Общие команды и политики

**Files:** `scripts/harness/harness.py`, `scripts/harness/policy.py`, `scripts/harness/test_harness.py`.

**Interfaces:** CLI `context|doctor|check|hook`; `hook` читает JSON из stdin.
Политика `check_tool(root, cwd, tool, args) -> str | None` возвращает причину запрета.

- [x] Написать тесты контекста, защищённых путей, shell-команд, apply_patch и Stop.
- [x] Запустить `python3 -m unittest discover -s scripts/harness -p 'test_*.py'`.
- [x] Подтвердить красный результат до реализации.
- [x] Реализовать политики, контекст, диагностику и ограниченный по времени запуск проверок.
- [x] Подтвердить зелёный результат той же командой.

## Task 2: Конфигурация обоих инструментов

**Files:** `.codex/config.toml`, `.codex/hooks.json`, `.claude/settings.json`, `.agents/HARNESS.md`, `Makefile`.

**Interfaces:** SessionStart, PreToolUse, PostToolUse, Stop вызывают общий CLI `hook`.

- [x] Добавить общие команды Make с абсолютным корнем через путь Makefile.
- [x] Подключить общий процесс из `AGENT.md` и `.claude/CLAUDE.md`.
- [x] Добавить относительные ссылки для переиспользования существующих skills.
- [x] Проверить конфигурации тестами и штатной командой Codex.
- [x] Выполнить адаптеры с синтетическими событиями через реальные строки команд конфигурации.

## Task 3: Проверка и документация

**Files:** `docs/AGENT_HARNESS.md`, `docs/IMPLEMENTATION_STATUS.md`, `docs/KNOWN_LIMITATIONS.md`, `docs/DECISIONS.md`, `tasks/T-032.md`.

**Interfaces:** `make harness-check` запускает harness-test, lint, test, packs.

- [x] Документировать запуск, доверие хукам, текущую задачу, MCP и ограничения защиты.
- [x] Запустить `make harness-doctor` и `make harness-check`.
- [x] Проверить diff и провести отдельный обзор реализации.
- [x] Зафиксировать результаты и непроверенные сценарии в T-032.

## Журнал

- План исполняется в текущей ветке согласно прямому запросу пользователя.
- Коммиты и публикация не входят в объём запроса.
- Пользователь одобрил дизайн; реализация продолжается в этой сессии.


## Результат выполнения

- Task 1: complete; 30 автономных тестов проходят после исправления CRLF-регрессии.
- Task 2: complete; обе команды конфигурации проверены реальными JSON-событиями.
- Task 3: verification partial; документация и независимый обзор выполнены.
- Обязательный `make test` прошёл с изолированной PostgreSQL 16 и Go race detector.
- Ruling: сохранить существующие Go-файлы без правок согласно согласованному объёму.
  Цена решения: `make lint` и общий completion gate остаются красными до отдельного исправления форматирования.
- Новые клиентские сессии не проверены; Codex требует ручной проверки доверия, Claude Code требует авторизации.
- Полный отчёт: `tasks/T-032.md`.
