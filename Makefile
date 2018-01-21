devbox/build:
	cd docker; \
	$(MAKE) build; \
	sudo docker-compose build 

devbox/run: devbox/build
	sudo docker-compose up -d