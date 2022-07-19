# tix

Some scaffolds using TiDB


# Config Template

```
[common]
dsn = mysql://service.tidb.endpoint.com
txn_batch_size = 1000


[jobqueue]
table_name = "tix_job_queue_tbl"


[pubsub]
table_name = "tix_pubsub_tbl"

[alive]
table_name = "tix_alive_tbl"
```
