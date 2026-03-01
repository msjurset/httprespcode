build:
	go build -o hrc .

run:
	go run .

deploy: build install-man install-completion
	cp hrc ~/.local/bin/

install-man:
	install -d /usr/local/share/man/man1
	install -m 644 hrc.1 /usr/local/share/man/man1/hrc.1

install-completion:
	install -d ~/.oh-my-zsh/custom/completions
	install -m 644 _hrc ~/.oh-my-zsh/custom/completions/_hrc
