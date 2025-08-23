
gateway {
  hostname               = "localhost"
  port                   = 8080
  log_file               = "gateway.log"
  db_file                = "limiter.db"
  user_cache_ttl_minutes = 10

}

api {
  hostname = "localhost"
  key      = "topsecret"
  port     = 8081
}

routes {
  path        = "/foo"
  strategy    = "token_bucket"
  capacity    = 5
}

routes {
  path        = "/bar"
  strategy    = "fixed_window"
  limit       = 10
  window_size = 10 // seconds
}