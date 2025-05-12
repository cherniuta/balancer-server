# Balancer server ⚖️

**Балансировщик**

[![Go](https://img.shields.io/badge/Go-%2300ADD8.svg?&logo=go&logoColor=white)](#)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](#)

## 🌟 О проекте

- ⚙️ Алгоритм балансировки [Round-robin]([https://xkcd.com](https://en.wikipedia.org/wiki/Round-robin_scheduling))
- 🔀 Concurrency limiter ограничивает количество одновременных запросов
- 🚦 Rate limiting ограничивает количество запросов от одного пользователя

## 🛠 Технологический стек

- **Go (Golang)** - основной язык разработки
- **Docker + Compose** - средство запуска

## 🚀 Команды
### Сборка
```bash
docker compose up --build -d
```
### Запуск
```bash
docker compose up
```
