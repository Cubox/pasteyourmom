# Pasteyourmom

Simple paste service.

Demo here: [http://paste.cubox.dev]

## How to install?

You need [Go](http://golang.org) installed.

```
mkdir paste
cd paste
go get github.com/Cubox-/pasteyourmom
wget https://raw.githubusercontent.com/Cubox-/pasteyourmom/master/index.html
wget https://raw.githubusercontent.com/Cubox-/pasteyourmom/master/style.css
wget https://raw.githubusercontent.com/Cubox-/pasteyourmom/master/config.json
vim config.json
pasteyourmom -bind :(your port here) -conf config.json
```

Those steps are here for the example, modify following your needs.
You can put a nginx or apache in front of this, use the XRealIP configuration variable.
