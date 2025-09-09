# Geolocation Tracker

Sistema de geolocalização em tempo real para eventos. Permite que usuários compartilhem suas posições, encontrem amigos próximos e recebam notificações automáticas via Redis Streams.

**Desenvolvido por Victor Augusto**

## Tecnologias

- **Go 1.25** - Backend performático
- **PostgreSQL + PostGIS** - Consultas geoespaciais  
- **Redis Streams** - Eventos em tempo real
- **Docker** - Deploy simplificado
- **Clean Architecture** - Código organizando e escalável

## Início rápido

```bash
git clone https://github.com/VictorXdAugusto/engineer-test
cd engineer-test
docker-compose up --build
```

Aplicação rodando em: **http://localhost:8080**

Teste se está funcionando:
```bash
curl http://localhost:8080/health
```

## Funcionalidades

| Endpoint | Descrição |
|----------|-----------|
| `POST /api/v1/users` | Criar usuário |
| `POST /api/v1/positions` | Salvar posição (gera evento) |
| `GET /api/v1/users/{id}/position` | Posição atual |
| `GET /api/v1/users/{id}/positions/history` | Histórico de posições |
| `GET /api/v1/positions/nearby` | Usuários próximos |
| `GET /api/v1/positions/sector` | Usuários no setor |

## Sistema de Eventos (Redis Streams)

### Como funciona:
1. Usuário salva nova posição → Evento é publicado no Redis Stream
2. **3 consumers** processam o evento automaticamente:
   - **notifications**: Notificações push, emails
   - **analytics**: Métricas e análises  
   - **realtime**: WebSocket para tempo real

### Monitoramento:
```bash
# Ver quantos eventos foram processados
docker exec geolocation-redis redis-cli XLEN geolocation:position-events

# Ver últimos eventos
docker exec geolocation-redis redis-cli XREVRANGE geolocation:position-events + - COUNT 3

# Status dos consumers
curl http://localhost:8080/api/v1/events/stats
```

## Desenvolvimento

Executar localmente (sem Docker):

```bash
go mod download
go generate ./internal/wire
go build -o bin/server ./cmd/server
./bin/server
```

## Troubleshooting

**Problema com portas:**
```bash
lsof -i :8080 :5432 :6379
```

**Reiniciar tudo:**
```bash
docker-compose down
docker-compose up --build
```

**Ver logs:**
```bash
docker-compose logs app | tail -20
```

# Swagger:

**URL de acesso para a documentação:**
[http://localhost:8080/swagger/index.html]


# Testes unitarios

```bash
go test ./...

---

*Desenvolvido por **Victor Augusto***
