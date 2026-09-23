# API e exemplos

## Autenticação

Rotas protegidas usam:

```http
Authorization: Bearer <JWT>
```

### Criar usuário

```bash
curl -X POST http://localhost:8081/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Maria Silva","email":"maria@example.com","password":"secret123"}'
```

### Login

```bash
curl -X POST http://localhost:8081/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"maria@example.com","password":"secret123"}'
```

Use `data.token` nas próximas requisições e `data.user.id` como `<USER_ID>`.

## Processamento de vídeo

### Criar lote

```bash
curl -X POST http://localhost:8082/v1/uploads \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<USER_ID>"}'
```

Use `data.id` como `<BATCH_ID>`.

### Enviar vídeo

```bash
curl -X POST http://localhost:8082/v1/uploads/<BATCH_ID>/processings \
  -H "Authorization: Bearer <JWT>" \
  -F "file=@./video.mp4"
```

Extensões aceitas: `.mp4`, `.avi`, `.mov`, `.mkv`, `.wmv`, `.flv` e `.webm`.

Status possíveis: `PENDING`, `PROCESSING`, `COMPLETED` e `FAILED`.

### Consultar processamento

```bash
curl http://localhost:8082/v1/uploads/<BATCH_ID>/processings \
  -H "Authorization: Bearer <JWT>"
```

### Baixar imagens do lote

```bash
curl -L -o imagens.zip \
  http://localhost:8082/v1/uploads/<BATCH_ID>/download \
  -H "Authorization: Bearer <JWT>"
```

Para um processamento específico:

```text
GET /v1/uploads/<BATCH_ID>/processings/<PROCESSING_ID>/download
```

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

## Collection do Insomnia

A collection [collection.yaml](../collection.yaml) contém requests para usuários, login, lotes, upload de vídeos, consulta de processamentos e downloads.

Para importar:

1. Abra o Insomnia e selecione **Import**.
2. Escolha `collection.yaml`.
3. Execute `Users > Login` primeiro.
4. Crie um lote e atualize os IDs nas URLs conforme as respostas da API.

As requisições usam `localhost:8081` e `localhost:8082`.
