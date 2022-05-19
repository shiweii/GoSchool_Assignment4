module github.com/shiweii/appointment

go 1.18

require (
	github.com/shiweii/binarysearchtree v0.0.0-00010101000000-000000000000
	github.com/shiweii/doublylinkedlist v0.0.0-00010101000000-000000000000
	github.com/shiweii/logger v0.0.0-00010101000000-000000000000
	github.com/shiweii/user v0.0.0-00010101000000-000000000000
	github.com/shiweii/utility v0.0.0-00010101000000-000000000000
)

require (
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/shiweii/cryptography v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/shiweii/binarysearchtree => ../binarysearchtree

replace github.com/shiweii/utility => ../utility

replace github.com/shiweii/logger => ../logger

replace github.com/shiweii/user => ../user

replace github.com/shiweii/cryptography => ../cryptography

replace github.com/shiweii/doublylinkedlist => ../doublylinkedlist
