install:
	@go build
	@cp -rf noProxy /usr/local/bin/noProxy
	@[ -f /usr/local/etc/noProxy.yaml ] || cp -rf noProxy.yaml /usr/local/etc/noProxy.yaml
	@cp -rf noproxy.service /etc/systemd/system/noproxy.service
	@systemctl daemon-reload
	@systemctl restart noproxy.service
	@systemctl enable noproxy.service