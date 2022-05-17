module github.com/shiweii/user

go 1.18

replace github.com/shiweii/utility => ../utility

replace github.com/shiweii/cryptography => ../cryptography

replace github.com/shiweii/logger => ../logger

require (
	github.com/shiweii/cryptography v0.0.0-00010101000000-000000000000
	github.com/shiweii/doublylinkedlist v0.0.0-00010101000000-000000000000
	github.com/shiweii/utility v0.0.0-00010101000000-000000000000
)

require (
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/shiweii/logger v0.0.0-00010101000000-000000000000 // indirect
	github.com/yuin/goldmark v1.4.12 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220516155154-20f960328961 // indirect
	golang.org/x/sys v0.0.0-20220513210249-45d2b4557a2a // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
)

replace github.com/shiweii/doublylinkedlist => ../doublylinkedlist
