module ecksbee.com/telefacts-sec

go 1.16

replace ecksbee.com/telefacts => github.com/ecksbee/telefacts v0.0.0-20240128

replace ecksbee.com/kushim => github.com/ecksbee/kushim v0.0.0-20231122

require (
	ecksbee.com/kushim v0.0.0
	ecksbee.com/telefacts v0.0.0
	github.com/gorilla/mux v1.8.0
	github.com/joshuanario/r8lmt v0.0.0-20190907165225-782e183364f7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	golang.org/x/net v0.0.0-20211005001312-d4b1ae081e3b
)
