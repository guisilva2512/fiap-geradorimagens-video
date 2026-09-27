# FIAP Gerador de Imagens a partir de Vídeo

Backend em Go para receber vídeos, extrair imagens com FFmpeg e disponibilizar os resultados em ZIP. O processamento é assíncrono, usando RabbitMQ, PostgreSQL e MinIO/S3.

## Documentação

| Contexto | Documento |
| --- | --- |
| Arquitetura, componentes e diagramas C4 | [docs/architecture.md](docs/architecture.md) |
| Instalação, Docker Compose, execução manual e testes | [docs/local-development.md](docs/local-development.md) |
| Fluxo da API, endpoints e collection do Insomnia | [docs/api.md](docs/api.md) |
| Métricas, logs, traces e consultas no Grafana | [docs/observability.md](docs/observability.md) |
| Pipeline do GitHub Actions | [docs/ci-cd.md](docs/ci-cd.md) |

## Início rápido

```bash
docker compose up -d --build
```

Serviços principais:

- Frontend: `http://localhost:5173`
- Auth API: `http://localhost:8081`
- Upload API: `http://localhost:8082`
- Grafana: `http://localhost:3000`
- Prometheus: `http://localhost:9090`

Consulte [docs/local-development.md](docs/local-development.md) para pré-requisitos, credenciais e encerramento do ambiente.

## Fluxo funcional

1. Crie um usuário em `POST /v1/users`.
2. Faça login em `POST /v1/login` e obtenha o JWT.
3. Crie um lote em `POST /v1/uploads`.
4. Envie um vídeo em `POST /v1/uploads/:id/processings`.
5. Consulte o status e baixe as imagens.

O frontend React cobre esse fluxo em uma interface web: cadastre-se, faça login,
selecione ou crie um lote, envie vídeos e acompanhe o processamento. O token JWT
fica salvo apenas no `localStorage` do navegador para autenticar as requisições.

Os exemplos completos estão em [docs/api.md](docs/api.md). A collection para importação no Insomnia está em [collection.yaml](collection.yaml).

## Observabilidade

A dashboard local está em [http://localhost:3000/d/fiap-observability/fiap-observabilidade](http://localhost:3000/d/fiap-observability/fiap-observabilidade).

Exemplos de consultas para logs, métricas e traces estão em [docs/observability.md](docs/observability.md).

## Tecnologias

Go, Gin, GORM, PostgreSQL, RabbitMQ, MinIO/S3, FFmpeg, Docker Compose, JWT, OpenTelemetry, Prometheus, Grafana, Loki, Tempo e Grafana Alloy.
