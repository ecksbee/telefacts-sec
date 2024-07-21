module ecksbee.com/telefacts-sec

go 1.22

toolchain go1.22.0

replace ecksbee.com/telefacts => github.com/ecksbee/telefacts v0.0.0-20240713

replace ecksbee.com/telefacts-taxonomy-package => github.com/ecksbee/telefacts-taxonomy-package v0.1.6

require (
	ecksbee.com/telefacts v0.0.0
	ecksbee.com/telefacts-taxonomy-package v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/joshuanario/r8lmt v0.0.0-20190907165225-782e183364f7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	golang.org/x/net v0.21.0
)

require (
	github.com/antchfx/xmlquery v1.3.18 // indirect
	github.com/antchfx/xpath v1.2.4 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joshuanario/arcs v0.0.0-20221030000450-bf6ace5a19ba // indirect
	github.com/joshuanario/digits v0.5.2 // indirect
	github.com/klauspost/lctime v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
