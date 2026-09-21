# Desenvolvimento local

## Pré-requisitos

- Docker e Docker Compose;
- Go `1.26.6` para executar fora dos containers;
- `curl`;
- FFmpeg apenas na execução manual do worker.

## Docker Compose

Na raiz do projeto:

```bash
docker compose up -d --build
```

Verifique os serviços:

```bash
docker compose ps
docker compose logs -f auth-api upload-api video-worker
```

Interfaces úteis:

| Serviço | URL | Credenciais |
| --- | --- | --- |
| Auth API | `http://localhost:8081` | - |
| Upload API | `http://localhost:8082` | - |
| Grafana | `http://localhost:3000` | `admin` / `admin` |
| Prometheus | `http://localhost:9090` | - |
| Tempo | `http://localhost:3200` | - |
| Loki | `http://localhost:3100` | - |
| MinIO Console | `http://localhost:9001` | `root` / `password` |
| RabbitMQ Dashboard | `http://localhost:15672` | `root` / `root` |
| pgAdmin | `http://localhost:54321` | `teste@email.com` / `123` |

O Compose cria o banco, executa as migrations, cria o bucket `video-bucket` e declara a fila `video_processing_queue`.

Para encerrar:

```bash
docker compose down
```

Para remover todos os dados locais, incluindo volumes:

```bash
docker compose down -v
```

## Execução manual

Suba a infraestrutura:

```bash
docker compose up -d postgres minio createbuckets rabbitmq

go mod download
```

Configure as variáveis no PowerShell:

```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="root"
$env:DB_PASSWORD="root"
$env:DB_NAME="root"
$env:JWT_SECRET="your_super_secret_key_here"
$env:JWT_ISSUER="your_issuer_here"
$env:JWT_AUDIENCE="your_audience_here"
$env:JWT_EXPIRATION_HOURS="2"
$env:AWS_ACCESS_KEY_ID="root"
$env:AWS_SECRET_ACCESS_KEY="password"
$env:AWS_REGION="us-east-1"
$env:AWS_S3_ENDPOINT="http://localhost:9000"
$env:AWS_S3_BUCKET="video-bucket"
$env:RABBITMQ_URL="amqp://root:root@localhost:5672"
$env:RABBITMQ_QUEUE="video_processing_queue"
$env:FFMPEG_BINARY="ffmpeg"
$env:VIDEO_FRAMES_PER_SECOND="1"
$env:WORKER_TEMP_DIR="$env:TEMP\video-worker"
$env:OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4320"
```

Execute em três terminais:

```bash
go run ./cmd/auth-api
go run ./cmd/upload-api
go run ./cmd/video-worker
```

## Testes

```bash
go test ./...
go test -v ./...
```

A suíte cobre handlers, middleware JWT, regras de domínio, FFmpeg e o serviço do worker.

## Configuração principal

| Variável | Uso |
| --- | --- |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | PostgreSQL |
| `JWT_SECRET`, `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_EXPIRATION_HOURS` | JWT |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` | S3/MinIO |
| `AWS_S3_ENDPOINT`, `AWS_S3_BUCKET` | Storage |
| `RABBITMQ_URL`, `RABBITMQ_QUEUE` | RabbitMQ |
| `FFMPEG_BINARY` | Executável FFmpeg |
| `VIDEO_FRAMES_PER_SECOND` | Frames extraídos por segundo |
| `WORKER_TEMP_DIR` | Diretório temporário |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Endpoint HTTP do Collector OTel |
