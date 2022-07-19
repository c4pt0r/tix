# tix

Some scaffolds using TiDB


# Config Template

```
dsn = "root:@tcp(localhost:4000)/test?charset=utf8&parseTime=True&loc=Local"
max_transaction_size = 1000

[job_queue]
  table_prefix = "tix_job_queue"
  poll_interval = "1s"
  enable_gc = true
  gc_keep_items = 10000
  gc_interval = "1m"

[pubsub]
  table_prefix = "tix_pubsub"
  poll_interval = "1s"
  enable_gc = true
  gc_keep_items = 10000
  gc_interval = "1m"

[election]
  table_name = "tix_election"
  poll_interval = "1s"
  term_timeout = "1m"
```
