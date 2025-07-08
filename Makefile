BINARY=Enty


default: development

dev: development

development:
	wails dev


build-app:
	wails build
	cp -r ./config ./build/bin

clean:
	rm -rf ./build/bin