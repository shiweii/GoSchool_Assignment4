module github.com/shiweii/user

go 1.18

replace github.com/shiweii/encryptdecrypt => ../encryptdecrypt

replace github.com/shiweii/utility => ../utility

replace github.com/shiweii/logger => ../logger

require (
	github.com/shiweii/encryptdecrypt v0.0.0-00010101000000-000000000000
	github.com/shiweii/utility v0.0.0-00010101000000-000000000000
)

require (
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/shiweii/logger v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/text v0.3.7 // indirect
)
