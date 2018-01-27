devbox/build:
	cd docker; \
	$(MAKE) build; \
	sudo docker-compose build 

devbox/run: devbox/build
	sudo docker-compose up -d

devbox/stop:
	sudo docker-compose down

devbox/logs:
	sudo docker-compose logs -f