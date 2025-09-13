# 一个简单的 gin 认证 API 例子

所有 API 均使用 JSON 格式请求和响应
使用 SQLite 作为数据库
密码使用 bcrypt 哈希后存储

## build tags
`production` - 启用生产环境设置（如严格的 CORS）

## 环境变量
- `ADDR`: 监听地址（包括端口），默认 `:8080`
- `JWT_SECRET`: 用于签发 JWT 的密钥，默认 `default-jwt-secret`
- `DB_PATH`: SQLite 数据库文件路径，默认 `./app.db`

## API 列表

### POST /register
请求：
```json
{
  "username": "<string>",
  "password": "<string>"
}
```

响应：
201：
```json
{
  "id": <int>,
  "username": "<string>"
}
```

其他：
```json
{
  "error": "<string>"
}
```

### POST /login

请求：
```json
{
  "username": "<string>",
  "password": "<string>"
}
```

响应：
200：
```json
{
  "token": "<string>"
}
```

其他：
```json
{
  "error": "<string>"
}
```

### GET /profile
需要在请求头中包含 `Authorization: Bearer <token>`

响应：
200：
```json
{
  "id": <int>,
  "username": "<string>"
}
```

其他：
```json
{
  "error": "<string>"
}
```
