# nginx vhost — party.kayushkin.com

`party.kayushkin.com.conf` is the vhost that puts inber-party on the internet:
TLS, the HTTP→HTTPS redirect, and the WebSocket upgrade that the camp view's
real-time updates ride on. It was captured byte-for-byte from what nginx was
actually serving.

Until it was committed here it existed only as a hand-edit in `/etc/nginx`,
tracked by no repo — so rebuilding this box would have restored the service but
not the config that makes its live updates work.

## Installing it

```sh
./deploy/nginx/install.sh
```

Safe to run on its own: it touches nothing but nginx, and a vhost already
matching this repo is skipped without a reload. It preflights `nginx -t`, takes
a timestamped backup, rolls back if the new config fails to parse **or** fails
to serve, and only then leaves the reload in place.

**This repo has no `deploy.sh`.** Every other repo that ships a vhost calls
`install.sh` from the end of its deploy script; here there is no deploy script
to call it from, so run it by hand after changing the vhost. `install.sh` is
identical in all seven vhost-owning repos (dash, llmux, kayushkin.com,
argraphments, inber-party, multichat, forge) — copy it whole, do not fork it.
