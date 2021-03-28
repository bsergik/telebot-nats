## telebot-nats

The bot consumes data from channels (NATS) and push them to telegram.

```sh
docker run --network deployments_internal --rm -it synadia/nats-box
# then
stan-pub -s stan -c my-cluster-id -id abc "telebot.v1.errors" '{"subsystem": "test-subssystem", "message": "test message"}'
```