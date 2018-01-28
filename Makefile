PROJECTS=tracker subscriber
QA_PROJECTS= $(addprefix qa/, $(PROJECTS))

#builds devbox
devbox/build:
	cd docker; \
	$(MAKE) build; \
	sudo docker-compose build 

#builds devbox and runs it
devbox/run: devbox/build
	sudo docker-compose up -d

#stops devbox
devbox/stop:
	sudo docker-compose stop

#stops devbox and destroys images
devbox/down:
	sudo docker-compose down

#show devbox logs
devbox/logs:
	sudo docker-compose logs -f

#run qa for projects using it
qa: $(QA_PROJECTS)

qa/%:
	$(MAKE) -C $* qa

help:
	@echo Commands for running and dealing with project
	@echo "\"devbox/build\" - builds devbox images and docker-compose enviroment"
	@echo "\"devbox/run\" - builds devbox and starts devbox"
	@echo "\"devbox/stop\" - stops devbox"
	@echo "\"devbox/down\" - stops devbox and removes images" 
	@echo "\"devbox/logs\" - prints a logs of a devbox"
	@echo "\"qa\" - runs a test for project"