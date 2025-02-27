install:
	@go build
	@cp -rf noProxy /usr/local/bin/noProxy
	@cp -rf noProxy.service /etc/systemd/system/noProxy.service
	@systemctl daemon-reload
	@systemctl restart noProxy.service
	@systemctl enable noProxy.service