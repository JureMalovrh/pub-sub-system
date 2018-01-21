PROJECTS = go mongo node
IMAGES = $(addprefix image/, $(PROJECTS))

image/%:
	@cd $*; sudo docker build -t tracker/$* .

build: $(IMAGES)