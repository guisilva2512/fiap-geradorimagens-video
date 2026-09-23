# Arquitetura

## Objetivo

O sistema recebe vídeos, processa-os de forma assíncrona com FFmpeg e disponibiliza as imagens extraídas em arquivos ZIP.

O fluxo principal é:

1. O usuário cria uma conta e obtém um JWT.
2. A `upload-api` cria um lote e recebe o vídeo.
3. O vídeo original é salvo no MinIO/S3.
4. A API registra o processamento como `PENDING` e publica uma tarefa no RabbitMQ.
5. O `video-worker` consome a tarefa, baixa o vídeo, executa o FFmpeg e envia as imagens para o storage.
6. O processamento termina como `COMPLETED` ou `FAILED`.

## Componentes

| Componente | Responsabilidade | Porta local |
| --- | --- | --- |
| `auth-api` | Cadastro, login e gerenciamento de usuários | `8081` |
| `upload-api` | Lotes, upload, status e downloads | `8082` |
| `video-worker` | Consumo da fila e extração de imagens | métricas `8083` |
| PostgreSQL | Usuários, lotes e processamentos | `5432` |
| RabbitMQ | Fila `video_processing_queue` | `5672` / painel `15672` |
| MinIO | Storage S3 para vídeos e imagens | `9000` / console `9001` |
| Grafana | Dashboards de métricas, logs e traces | `3000` |
| Prometheus | Armazenamento de métricas | `9090` |
| Loki | Armazenamento de logs | `3100` |
| Tempo | Armazenamento de traces | `3200` |
| OpenTelemetry Collector | Recebimento e encaminhamento de traces | `4319` / `4320` |
| Grafana Alloy | Coleta de logs dos containers | - |

Os serviços de aplicação seguem uma separação inspirada em arquitetura hexagonal: handlers HTTP adaptam entradas e saídas, serviços concentram regras e PostgreSQL, S3, RabbitMQ e FFmpeg ficam nos adaptadores.

## Desenho da arquitetura

```mermaid
flowchart LR
    Client[Cliente HTTP]
    Auth[auth-api<br/>:8081]
    Upload[upload-api<br/>:8082]
    Worker[video-worker]
    DB[(PostgreSQL)]
    S3[(MinIO / S3<br/>video-bucket)]
    MQ[[RabbitMQ<br/>video_processing_queue]]
    FFmpeg[FFmpeg]
    Prom[Prometheus]
    Graf[Grafana]
    Loki[Loki]
    Tempo[Tempo]

    Client --> Auth
    Client --> Upload
    Auth --> DB
    Upload --> DB
    Upload --> S3
    Upload --> MQ
    MQ --> Worker
    Worker --> DB
    Worker --> S3
    Worker --> FFmpeg
    Auth --> Prom
    Upload --> Prom
    Worker --> Prom
    Prom --> Graf
    Loki --> Graf
    Tempo --> Graf
```

## C4 - Contexto

```mermaid
C4Context
    title Contexto do sistema - Gerador de imagens a partir de vídeo
    Person(usuario, "Usuário", "Envia vídeos e consulta resultados")
    System(sistema, "Gerador de imagens a partir de vídeo", "Recebe vídeos, extrai imagens e disponibiliza ZIPs")
    System_Ext(ffmpeg, "FFmpeg", "Extrai frames dos vídeos")
    System_Ext(storage, "MinIO/S3", "Armazena vídeos e imagens")
    System_Ext(messaging, "RabbitMQ", "Transporta tarefas assíncronas")
    System_Ext(database, "PostgreSQL", "Persiste usuários e processamentos")

    Rel(usuario, sistema, "Cadastra, autentica, envia vídeos e baixa imagens", "HTTP")
    Rel(sistema, database, "Lê e grava dados", "SQL")
    Rel(sistema, storage, "Armazena e recupera arquivos", "S3 API")
    Rel(sistema, messaging, "Publica e consome tarefas", "AMQP")
    Rel(sistema, ffmpeg, "Extrai imagens", "Processo local")
```

## C4 - Contêineres

```mermaid
C4Container
    title Contêineres - Gerador de imagens a partir de vídeo
    Person(usuario, "Usuário")
    Container(auth, "Auth API", "Go + Gin", "Autentica usuários e emite JWT")
    Container(upload, "Upload API", "Go + Gin", "Gerencia lotes, uploads e downloads")
    Container(worker, "Video Worker", "Go + FFmpeg", "Processa vídeos e extrai imagens")
    ContainerDb(db, "PostgreSQL", "PostgreSQL", "Usuários, lotes e processamentos")
    ContainerQueue(queue, "RabbitMQ", "AMQP", "Fila de processamento")
    ContainerDb(storage, "MinIO", "S3 API", "Vídeos e imagens")

    Rel(usuario, auth, "Cadastra e faz login", "HTTP")
    Rel(usuario, upload, "Envia e consulta vídeos", "HTTP + JWT")
    Rel(auth, db, "Persiste usuários", "GORM/SQL")
    Rel(upload, db, "Persiste metadados", "GORM/SQL")
    Rel(upload, storage, "Salva vídeos e lê imagens", "S3")
    Rel(upload, queue, "Publica tarefas", "AMQP")
    Rel(queue, worker, "Entrega tarefas", "AMQP")
    Rel(worker, db, "Atualiza status", "GORM/SQL")
    Rel(worker, storage, "Baixa vídeos e envia imagens", "S3")
```

## Estrutura

```text
cmd/
  auth-api/       Entrada do serviço de autenticação
  upload-api/     Entrada da API de upload
  video-worker/   Entrada do consumidor assíncrono
internal/
  auth/           Domínio e persistência de usuários
  upload/         Domínio e persistência de uploads
  video_processing/
                  Worker, RabbitMQ, aplicação e FFmpeg
  pkg/            Banco, storage, mensageria, middleware e observabilidade
observability/    Configuração local de Prometheus, Grafana, Loki, Tempo e Alloy
```
