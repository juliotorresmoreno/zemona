# zemona

## Requerimientos
Es necesario instalar [mongodb](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/) y [redis](https://redis.io/docs/getting-started/installation/install-redis-on-linux/) en su computadora para probar el proyecto.

Es necesario instalar nginx en su computadora.
```bash
sudo apt install nginx
```

Deberá reemplazar el contenido del archivo /etc/nginx/sites-available/default con lo siguiente.

```nginx
server {
  listen      80;
  server_name _;

  proxy_set_header Upgrade           $http_upgrade;
  proxy_set_header Connection        "upgrade";
  proxy_set_header Host              $host;
  proxy_set_header X-Real-IP         $remote_addr;
  proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
  proxy_set_header X-Forwarded-Proto $scheme;
  proxy_set_header X-Forwarded-Host  $host;
  proxy_set_header X-Forwarded-Port  $server_port;

  # Proxy timeouts
  proxy_connect_timeout              60s;
  proxy_send_timeout                 60s;
  proxy_read_timeout                 60s;

  location /api/v1 { 
    proxy_pass http://localhost:8080;
  }

  location / {    
    proxy_pass http://localhost:3000/;
  }
}

```

## Instalación
Clonamos o copiamos el proyecto y entramos en el desde una terminal.
```bash
go get -v
```

## Configuración
debemos crear o editar el archivo .env en la raíz del proyecto. A continuación dejo una copia de como lo tengo pero se entiende que esto no se almacena en el repositorio ya que contiene llaves de twitter.

```

ADDR=:8080

TWITTER_REDIRECT_URL=https://onnasoft.com/twitter/callback
TWITTER_API_KEY=8ERSmrAXOI5e7UAOEILTi9RNu
TWITTER_API_KEY_SECRET=vv0jUD0YLLB2kxbXoAIBUDiaXvJcKnbKM7wzo7GQuZrgkdr2sD

TWITTER_CLIENT_ID=
TWITTER_CLIENT_SECRET=


MONGODB_URI=mongodb://127.0.0.1:27017/?retryWrites=true&w=majority&compressors=disabled&gssapiServiceName=mongodb
MONGODB_DATABASE=zemona

REDIS_URI=redis://localhost:6379/1

PORTAL_USERNAME=TwitterDev
PORTAL_PASSWORD=TwitterDev

```
## Ejecución
Solo debemos lanzar el siguiente comando sobre la raíz del proyecto.
```bash
go run main.go
```

## API Rest

**Iniciar session**
```http
POST /api/v1/session/ HTTP/1.0
Connection: upgrade
Host: localhost
Accept: */*
Referer: http://localhost/account
Content-Type: application/json
Origin: http://localhost
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: same-origin

{"username":"TwitterDev","password":"TwitterDev"}
```

**Consultar session activa.**
```http
GET /api/v1/session/ HTTP/1.0
Host: localhost
Authorization: {{authorization}}
```

**Consultar perfil del usuario**
```http
GET /profile/{{username}}
Host: localhost
```
**Consultar Tweets**
```http
GET /twitter/{{username}}/tweets
Host: localhost
```
**Actualizar perfil**
```http
PATCH /profile/juliotorresmor4
Host: localhost
Content-Type: application/json

{
    "name": "...",
    "description": "...",
    "image_src": "..."
}
```
**Consultar visitas al perfil**
```http
GET /twitter/{{username}}/requests
Host: localhost
```

