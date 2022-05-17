module GoSchool_Assignment/main

go 1.18

replace github.com/shiweii/binarysearchtree => ../binarysearchtree

replace github.com/shiweii/appointment => ../appointment

require (
	github.com/gorilla/mux v1.8.0
	github.com/satori/go.uuid v1.2.0
	github.com/shiweii/appointment v0.0.0-00010101000000-000000000000
	github.com/shiweii/binarysearchtree v0.0.0-00010101000000-000000000000
	github.com/shiweii/cryptography v0.0.0-00010101000000-000000000000
	github.com/shiweii/doublylinkedlist v0.0.0-00010101000000-000000000000
	github.com/shiweii/logger v0.0.0-00010101000000-000000000000
	github.com/shiweii/user v0.0.0-00010101000000-000000000000
	github.com/shiweii/utility v0.0.0-00010101000000-000000000000
	github.com/shiweii/validator v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20220517005047-85d78b3ac167
)

require (
	github.com/joho/godotenv v1.4.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/shiweii/utility => ../utility

replace github.com/shiweii/logger => ../logger

replace github.com/shiweii/cryptography => ../cryptography

replace github.com/shiweii/doublylinkedlist => ../doublylinkedlist

replace github.com/shiweii/user => ../user

replace github.com/shiweii/validator => ../validator
