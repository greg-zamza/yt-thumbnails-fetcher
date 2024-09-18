# YouTube thumbnails fetcher

Собирает превью с активных русскоязычных каналов на youtube.com. Использует youtube data API v3, интерфейс реализован через telegram bot API (NicoNex/echotron). Написан на чистом Go!

## Getting Started

### Prerequisites

```
git clone https://github.com/greg-zamza/yt-thumbnails-fetcher
cd youtube_thumbnails_fetcher
```

### Running

Токен telegram бота, полученный у @BotFather, необходимо записать в bot_token,
youtube data API v3 key записать в yt_api_keys,
пароль, который бот будет требовать для получения доступа, хранится в bot_password.

Когда все эти данные записаны, можно запустить проект с помощью

```
docker compose up -d
```
Эта команда скачает необходимые docker images из hub.docker.com
Если же нужно внести какие-то изменения в проект, можно собрать образы вручную

```
docker build -t latexenjoyer/yt_fetcher_tgbot:1.0 BotService/
docker build -t latexenjoyer/yt_fetcher_filter:1.0 FilterService/
```

```
docker compose up -d
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
