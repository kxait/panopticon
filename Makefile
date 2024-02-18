all: clean program

main.css:
	touch ./web/page/tailwind-src/main.css

output.css: main.css 
	npx tailwind -i ./web/page/tailwind-src/main.css -o ./web/page/static/output.css

program: output.css
	go build -o panopticon

clean:
	rm ./web/page/static/output.css
	rm panopticon

# run with -j2
watch: watch_tailwind watch_program

watch_tailwind:
	npx tailwind -i ./web/page/tailwind-src/main.css -o ./web/page/static/output.css --watch

watch_program:
	wgo run -file .html .
