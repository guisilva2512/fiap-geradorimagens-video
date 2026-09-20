# FIAP Gerador de Imagens a partir de Vídeo

Serviço backend para receber vídeos, processá-los de forma assíncrona e disponibilizar as imagens extraídas em arquivos ZIP. O projeto foi desenvolvido em Go e organiza cada etapa em serviços independentes: autenticação, upload/API de processamento e worker de vídeo.

## Objetivo

O sistema permite:

- cadastrar usuários e autenticar chamadas com JWT;
- criar lotes de processamento;
- enviar vídeos para armazenamento compatível com S3;
- processar os vídeos em segundo plano com FFmpeg, extraindo imagens na taxa configurada;
- consultar o status do processamento;
- baixar as imagens de um processamento ou de um lote completo em ZIP.

O processamento é assíncrono: a API salva o vídeo, registra o processamento como `PENDING` e publica uma mensagem no RabbitMQ. O worker consome essa mensagem, baixa o vídeo do MinIO, executa o FFmpeg, envia as imagens de volta ao storage e atualiza o status para `COMPLETED` ou `FAILED`.

## Arquitetura

### Componentes

| Componente | Responsabilidade | Porta local |
| --- | --- | --- |
| `auth-api` | Cadastro, login e gerenciamento de usuários | `8081` |
| `upload-api` | Lotes, upload de vídeos, consulta e download das imagens | `8082` |
| `video-worker` | Consumo da fila e extração de imagens com FFmpeg | - |
| PostgreSQL | Persistência de usuários, lotes e processamentos | `5432` |
| RabbitMQ | Fila de processamento assíncrono | `5672` / painel `15672` |
| MinIO | Storage S3 local para vídeos e imagens | API `9000` / console `9001` |
| pgAdmin | Interface opcional para administrar o PostgreSQL | `54321` |

Os serviços de aplicação seguem uma separação inspirada em arquitetura hexagonal: handlers HTTP fazem a adaptação de entrada e saída, os serviços de domínio/aplicação concentram as regras e as implementações de PostgreSQL, S3, RabbitMQ e FFmpeg ficam nos adaptadores.

### Desenho da arquitetura

```mermaid
flowchart LR
		Client[Cliente HTTP]
		Auth[auth-api<br/>:8081]
		Upload[upload-api<br/>:8082]
		Worker[video-worker]
		DB[(PostgreSQL)]
		S3[(MinIO / S3<br/>bucket video-bucket)]
		MQ[[RabbitMQ<br/>video_processing_queue]]
		FFmpeg[FFmpeg]

		Client -->|cadastro e login| Auth
		Client -->|lotes, upload e downloads| Upload
		Auth --> DB
		Upload --> DB
		Upload -->|salva vídeo| S3
		Upload -->|publica processamento| MQ
		MQ -->|consome mensagem| Worker
		Worker -->|consulta e atualiza status| DB
		Worker -->|baixa vídeo e envia imagens| S3
		Worker --> FFmpeg
		Upload -->|lê imagens e gera ZIP| S3
```

### C4 - Contexto

```mermaid
C4Context
		title Contexto do sistema - Gerador de imagens a partir de vídeo

		Person(usuario, "Usuário", "Envia vídeos e consulta os resultados")
		System(sistema, "Gerador de imagens a partir de vídeo", "Recebe vídeos, extrai imagens e disponibiliza ZIPs")
		System_Ext(ffmpeg, "FFmpeg", "Extrai frames dos vídeos")
		System_Ext(storage, "MinIO/S3", "Armazena vídeos e imagens")
		System_Ext(messaging, "RabbitMQ", "Transporta tarefas assíncronas")
		System_Ext(database, "PostgreSQL", "Persiste usuários e processamentos")

		Rel(usuario, sistema, "Cadastra, autentica, envia vídeos e baixa imagens", "HTTP/JSON e multipart")
		Rel(sistema, database, "Lê e grava dados", "SQL")
		Rel(sistema, storage, "Armazena e recupera arquivos", "S3 API")
		Rel(sistema, messaging, "Publica e consome tarefas", "AMQP")
		Rel(sistema, ffmpeg, "Extrai imagens", "Processo local")
```

### C4 - Contêineres

```mermaid
C4Container
		title Contêineres - Gerador de imagens a partir de vídeo

		Person(usuario, "Usuário")
		Container(auth, "Auth API", "Go + Gin", "Autentica usuários e emite JWT")
		Container(upload, "Upload API", "Go + Gin", "Gerencia lotes, uploads, status e downloads")
		Container(worker, "Video Worker", "Go + FFmpeg", "Processa vídeos e extrai imagens")
		ContainerDb(db, "PostgreSQL", "PostgreSQL", "Usuários, lotes e processamentos")
		ContainerQueue(queue, "RabbitMQ", "AMQP", "Fila video_processing_queue")
		ContainerDb(storage, "MinIO", "S3 API", "Vídeos originais e imagens geradas")

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

## Pré-requisitos

- Docker e Docker Compose;
- Go `1.26.6` para executar os serviços fora dos containers;
- `curl` para testar a API;
- FFmpeg apenas para execução manual do worker. A imagem Docker já o instala.

## Como rodar localmente com Docker

Na raiz do projeto:

```bash
docker compose up -d --build
```

Verifique os serviços e acompanhe os logs:

```bash
docker compose ps
docker compose logs -f auth-api upload-api video-worker
```

Interfaces úteis:

- Auth API: `http://localhost:8081`
- Upload API: `http://localhost:8082`
- Console do MinIO: `http://localhost:9001` (`root` / `password`)
- Dashboard do RabbitMQ: `http://localhost:15672` (`root` / `root`)
- pgAdmin: `http://localhost:54321` (`teste@email.com` / `123`)

O Compose cria automaticamente o banco `root`, o bucket `video-bucket`, a fila `video_processing_queue` e executa as migrations na inicialização do PostgreSQL. Para encerrar os serviços:

```bash
docker compose down
```

Para remover também os dados persistidos localmente, use `docker compose down -v`.

## Como rodar manualmente

Suba somente a infraestrutura:

```bash
docker compose up -d postgres minio createbuckets rabbitmq
go mod download
```

Em três terminais, execute:

```bash
go run ./cmd/auth-api
```

```bash
go run ./cmd/upload-api
```

```bash
go run ./cmd/video-worker
```

As variáveis usadas pelo Compose devem estar disponíveis no ambiente. No Windows PowerShell, por exemplo:

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
```

## Fluxo de uso da API

Todas as rotas protegidas usam o token retornado no login:

```http
Authorization: Bearer <JWT>
```

### 1. Criar usuário

```bash
curl -X POST http://localhost:8081/v1/users \
	-H "Content-Type: application/json" \
	-d '{"name":"Maria Silva","email":"maria@example.com","password":"secret123"}'
```

### 2. Fazer login

```bash
curl -X POST http://localhost:8081/v1/login \
	-H "Content-Type: application/json" \
	-d '{"email":"maria@example.com","password":"secret123"}'
```

Copie o valor de `data.token` para as requisições seguintes.

### 3. Criar lote

```bash
curl -X POST http://localhost:8082/v1/uploads \
	-H "Authorization: Bearer <JWT>" \
	-H "Content-Type: application/json" \
	-d '{"user_id":"<USER_ID>"}'
```

Copie o `data.id` retornado como `<BATCH_ID>`.

### 4. Enviar vídeo para processamento

```bash
curl -X POST http://localhost:8082/v1/uploads/<BATCH_ID>/processings \
	-H "Authorization: Bearer <JWT>" \
	-F "file=@./video.mp4"
```

Extensões aceitas: `.mp4`, `.avi`, `.mov`, `.mkv`, `.wmv`, `.flv` e `.webm`. O processamento inicia em `PENDING`, passa por `PROCESSING` e termina em `COMPLETED` ou `FAILED`.

### 5. Consultar e baixar resultados

```bash
curl http://localhost:8082/v1/uploads/<BATCH_ID>/processings \
	-H "Authorization: Bearer <JWT>"

curl -L -o imagens.zip \
	http://localhost:8082/v1/uploads/<BATCH_ID>/download \
	-H "Authorization: Bearer <JWT>"
```

Para baixar apenas um processamento, use `/v1/uploads/<BATCH_ID>/processings/<PROCESSING_ID>/download`.

## Collection do Insomnia

O projeto inclui a collection [collection.yaml](collection.yaml) para importar no Insomnia. Ela contém requests prontos para autenticação, usuários, lotes de upload, envio de vídeos, consulta de processamentos e downloads.

Para usar:

1. Abra o Insomnia e selecione **Import**.
2. Escolha o arquivo `collection.yaml` deste projeto.
3. Execute os requests na pasta `gerador-imagens-video`.
4. Faça `Users > Login` primeiro. As requisições protegidas reutilizam automaticamente o JWT retornado por esse login.
5. Crie um usuário e faça login antes de criar um lote. Depois, atualize os IDs de lote, usuário e processamento nas URLs dos requests conforme os valores retornados pela API.

As requisições da collection usam `http://localhost:8081` para autenticação e `http://localhost:8082` para uploads, portanto a aplicação precisa estar em execução conforme as instruções acima.

## Endpoints principais

| Serviço | Método | Rota | Autenticação |
| --- | --- | --- | --- |
| Auth | `POST` | `/v1/users` | Não |
| Auth | `POST` | `/v1/login` | Não |
| Auth | `GET/PUT/DELETE` | `/v1/users[/:id]` | JWT |
| Upload | `GET/POST/DELETE` | `/v1/uploads[/:id]` | JWT |
| Upload | `GET/POST` | `/v1/uploads/:id/processings` | JWT |
| Upload | `PUT/DELETE` | `/v1/uploads/:id/processings/:video_id` | JWT |
| Upload | `GET` | `/v1/uploads/:id/download` | JWT |
| Upload | `GET` | `/v1/uploads/:id/processings/:video_id/download` | JWT |

## Como testar

Execute todos os testes unitários:

```bash
go test ./...
```

Para uma execução mais detalhada:

```bash
go test -v ./...
```

Os testes atuais cobrem handlers, middleware de autenticação, regras de domínio, processamento com FFmpeg e o serviço de aplicação do worker. Não é necessário subir o Docker Compose para executar a suíte unitária.

## Pipeline do GitHub Actions

O workflow [`.github/workflows/ci-cd.yml`](.github/workflows/ci-cd.yml), chamado `CI/CD`, é executado nos eventos abaixo:

- `push` nas branches `main`, `master` e `develop`;
- `pull_request` direcionado para `main`, `master` ou `develop`.

O pipeline usa `ubuntu-latest` e Go `1.26.6`, definido pela variável global `GO_VERSION`.

### Fluxo do pipeline

```mermaid
flowchart LR
	Trigger[Push ou Pull Request<br/>main / master / develop]
	Build[build-test<br/>build Go + testes + cobertura]
	Sonar[sonar<br/>análise + quality gate]
	Docker[docker<br/>build das 3 imagens]
	Deploy[deploy<br/>placeholder]
	Artifact[(Artifact<br/>go-coverage)]

	Trigger --> Build
	Build --> Artifact
	Build --> Sonar
	Artifact --> Sonar
	Sonar --> Docker
	Docker --> Deploy
```

### Jobs e dependências

#### `build-test`

É o primeiro estágio e executa:

1. checkout do repositório;
2. configuração do Go com cache;
3. `go mod download`;
4. compilação de `auth-api`, `upload-api` e `video-worker`;
5. testes unitários com `go test ./... -covermode=atomic -coverprofile=coverage.out`;
6. upload do arquivo `coverage.out` como artifact `go-coverage`.

#### `sonar`

Depende de `build-test`, baixa o artifact de cobertura e executa a análise usando o arquivo [sonar-project.properties](sonar-project.properties). Depois executa o quality gate do SonarQube. Para esse job funcionar, o repositório precisa ter:

- secret `SONAR_TOKEN`;
- secret `SONAR_HOST_URL`;
- variável `SONAR_ORGANIZATION`;
- variável opcional `SONAR_PROJECT_KEY`. Quando ausente, o workflow usa `github.repository` como chave do projeto.

#### `docker`

Depende de `build-test` e `sonar`. Usa uma matriz com os serviços `auth-api`, `upload-api` e `video-worker` e, para cada um:

- configura Docker Buildx;
- usa o [Dockerfile](Dockerfile) multi-stage;
- passa `SERVICE=<nome-do-servico>` como build argument;
- gera a tag `fiap-geradorimagens-video/<servico>:<github.sha>`;
- utiliza cache do GitHub Actions.

As imagens são construídas com `push: false`; portanto, o workflow atual não publica imagens em Docker Hub, GHCR, ECR ou outro registry.

#### `deploy`

Depende do job `docker`, mas atualmente não realiza uma implantação. O passo apenas imprime `Deploy será configurado em outro momento.`. Para habilitar deploy real, esse job deverá ser substituído por etapas específicas do ambiente de destino, como autenticação no registry, publicação das imagens e atualização dos serviços.

## Estrutura do projeto

```text
cmd/
	auth-api/       Entrada do serviço de autenticação
	upload-api/     Entrada da API de upload e processamento
	video-worker/   Entrada do consumidor assíncrono
internal/
	auth/           Domínio, handlers, persistência e migrations de usuários
	upload/         Domínio, handlers, persistência e migrations de uploads
	video_processing/
									Worker, consumidor RabbitMQ, aplicação e processador FFmpeg
	pkg/             Banco, storage S3, mensageria e middleware compartilhados
```

## Configuração principal

| Variável | Uso |
| --- | --- |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Conexão com PostgreSQL |
| `JWT_SECRET`, `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_EXPIRATION_HOURS` | Emissão e validação de tokens |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` | Credenciais S3/MinIO |
| `AWS_S3_ENDPOINT`, `AWS_S3_BUCKET` | Endpoint e bucket de arquivos |
| `RABBITMQ_URL`, `RABBITMQ_QUEUE` | Conexão e nome da fila |
| `FFMPEG_BINARY` | Executável do FFmpeg, padrão `ffmpeg` |
| `VIDEO_FRAMES_PER_SECOND` | Quantidade de frames extraídos por segundo, padrão `1` |
| `WORKER_TEMP_DIR` | Diretório temporário do worker |

## Tecnologias

- Go, Gin e GORM;
- PostgreSQL;
- RabbitMQ;
- MinIO usando a API compatível com Amazon S3;
- FFmpeg;
- Docker Compose;
- JWT para autenticação.
