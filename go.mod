module ecksbee.com/telefacts-sec

go 1.16

replace ecksbee.com/telefacts => github.com/ecksbee/telefacts v0.0.0-20230102

replace ecksbee.com/telefacts-taxonomy-package => github.com/ecksbee/telefacts-taxonomy-package v0.1.4

require (
	ecksbee.com/telefacts v0.0.0
	ecksbee.com/telefacts-taxonomy-package v0.0.0
	github.com/gorilla/mux v1.8.0
	github.com/joshuanario/r8lmt v0.0.0-20190907165225-782e183364f7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	golang.org/x/net v0.0.0-20211005001312-d4b1ae081e3b
)
