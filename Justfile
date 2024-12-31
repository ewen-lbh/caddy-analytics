build:
	go mod tidy
	xcaddy build --with github.com/ewen-lbh/caddy-analytics=. --output ~/.local/bin/caddy-analytics

dev:
	just build
	caddy-analytics adapt --config example/Caddyfile | jq .
	caddy-analytics run --config example/Caddyfile

example:
	just build
	caddy-analytics start --config example/Caddyfile
	wget -O example/responses/custom.html http://localhost:8081/custom
	wget -O example/responses/default.html http://localhost:8081/default
	caddy-analytics stop --config example/Caddyfile
