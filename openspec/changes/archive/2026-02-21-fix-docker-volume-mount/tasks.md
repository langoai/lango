## 1. Docker Compose

- [x] 1.1 Change volume mount from `lango-data:/data` to `lango-data:/home/lango/.lango` in docker-compose.yml

## 2. Dockerfile

- [x] 2.1 Replace `/data` directory creation with `/home/lango/.lango` pre-creation (`mkdir -p && chown lango:lango`) for correct volume ownership
