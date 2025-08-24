
gateway {
  address               = "localhost:8080"
  log_file               = "gateway.log"
  db_file                = "limiter.db"
  user_cache_ttl_minutes = 10

}

api {
  address = "localhost:8081"
  key      = "topsecret"
}

routes {
  path        = "/foo"
  strategy    = "token_bucket"
  capacity    = 5
}

routes {
  path        = "/bar"
  strategy    = "fixed_window"
  window_size = 10 // seconds
  sql_table    = "request_count"
}