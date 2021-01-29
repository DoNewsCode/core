package core

var defaultConfig = `
name: skeleton
version: 0.1.0
env: local
http:
  addr: :8080
grpc:
  addr: :9090
log:
  level: debug
redis:
  addrs:
    - 127.0.0.1:6379
  database: 0
gorm:
  database: mysql
  dsn: root@tcp(127.0.0.1:3306)/skeleton?charset=utf8mb4&parseTime=True&loc=Local
`
