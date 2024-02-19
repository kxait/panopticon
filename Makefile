all: clean program

main.css:
	touch ./web/page/tailwind-src/main.css

output.css: main.css 
	npm i
	npx tailwind -i ./web/page/tailwind-src/main.css -o ./web/page/static/output.css

program: output.css
	go build -o panopticon

clean:
	rm -f ./web/page/static/output.css
	rm -f panopticon
	rm -rf node_modules

# run with -j2
watch: watch_tailwind watch_program

watch_tailwind:
	npm i
	npx tailwind -i ./web/page/tailwind-src/main.css -o ./web/page/static/output.css --watch

watch_program:
	wgo run -file .html . $(PANOP)
