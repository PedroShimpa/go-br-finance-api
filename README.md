# Go BR Finance API

API opensource em Go que fornece informações financeiras para aplicativo da Play Store.

## 🚀 Como executar

### Pré-requisitos
- Docker
- Docker Compose

### Executar com Docker Compose

1. Clone o repositório:
```bash
git clone https://github.com/PedroShimpa/go-br-finance-api.git
cd go-br-finance-api
```

2. Execute os serviços:
```bash
docker-compose up --build
```

3. A API estará disponível em: `http://localhost:8080`

## 📋 Endpoints

### Recomendações Financeiras
- `GET /recomendacoes` - Lista todas as recomendações
- `POST /recomendacoes` - Cria uma nova recomendação (requer autenticação)

### Autenticação
- `POST /login` - Login de usuário
- `POST /register` - Registro de novo usuário

### Cálculos Financeiros
- `POST /calculations/compound-interest` - Cálculo de juros compostos
- `POST /calculations/simple-interest` - Cálculo de juros simples

## 🗄️ Banco de Dados

O projeto utiliza PostgreSQL com as seguintes tabelas:

- `recomendacoes_financeiras` - Armazena recomendações financeiras
- `users` - Gerenciamento de usuários

## 🔧 Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `DATABASE_URL` | URL de conexão com PostgreSQL | `postgres://postgres:postgres@db:5432/gofinance?sslmode=disable` |
| `REDIS_HOST` | Host do Redis | `redis` |
| `PORT` | Porta do servidor | `8080` |

## 🛠️ Desenvolvimento

### Estrutura do Projeto
```
.
├── cache/          # Cache Redis
├── config/         # Configurações de banco
├── db/            # Scripts SQL
├── docs/          # Documentação
├── handlers/      # Handlers HTTP
├── models/        # Modelos de dados
└── main.go        # Ponto de entrada
```

### Comandos Úteis

```bash
# Construir imagem Docker
docker-compose build

# Executar apenas o banco
docker-compose up db redis

# Ver logs
docker-compose logs -f api

# Parar serviços
docker-compose down

# Limpar volumes
docker-compose down -v
```

## 📝 Licença

Este projeto é open source e está disponível sob a [Licença MIT](LICENSE).
