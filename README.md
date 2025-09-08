# Sistema de Geolocalização para Eventos

Compartilha localização entre pessoas em eventos. Você vê os amigos, eles te veem, tudo em tempo real.

## Rodando

Precisa de docker e docker compose. Se não tem: [baixa aqui](https://www.docker.com/get-started/).

git clone 
cd engineer-test
docker-compose up --build

API rodando em: http://localhost:8080

Testa se tá ok:  
curl http://localhost:8080/health
# {"status":"healthy"}

## Tech stack

- Go 1.25  
- PostgreSQL + PostGIS  
- Redis  
- Docker  

## Usando a API

Criar usuário:
curl -X POST http://localhost:8080/api/v1/users \
-H "Content-Type: application/json" \
-d '{"id":"123e4567-e89b-12d3-a456-426614174000","name":"João Silva","email":"joao@exemplo.com","event_id":"meu-evento-2024"}'

Salvar localização:
curl -X POST http://localhost:8080/api/v1/positions \
-H "Content-Type: application/json" \
-d '{"user_id":"123e4567-e89b-12d3-a456-426614174000","latitude":-23.550520,"longitude":-46.633308,"event_id":"meu-evento-2024"}'

Ver sua posição:
curl http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000/position

Pessoas próximas (1km):
curl "http://localhost:8080/api/v1/positions/nearby?user_id=123e4567-e89b-12d3-a456-426614174000&latitude=-23.550520&longitude=-46.633308&radius_meters=1000"

Se travar, tenta:  
1. Checar portas 8080, 5432 e 6379  
2. docker-compose down && docker-compose up --build  
3. Logs: docker-compose logs app
