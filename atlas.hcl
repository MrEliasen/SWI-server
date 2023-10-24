variable "token" {
  type    = string
  default = getenv("TURSO_TOKEN")
}

env "turso" {
  url     = "libsql+wss://<your-server-id>.turso.io?authToken=${var.token}"
  exclude = ["_litestream*"]
}
