module testMQ

replace (
	golang.org/x/crypto => C:/workspace/GOLOCAL/github.com/crypto
	golang.org/x/net => C:/workspace/GOLOCAL/github.com/net
	golang.org/x/sync => C:/workspace/GOLOCAL/github.com/sync
	golang.org/x/sys => C:/workspace/GOLOCAL/github.com/sys
	golang.org/x/text => C:/workspace/GOLOCAL/github.com/text
	golang.org/x/tools => C:/workspace/GOLOCAL/github.com/tools
	golang.org/x/xerrors => C:/workspace/GOLOCAL/github.com/xerrors
)

require github.com/Shopify/sarama v1.24.1 // indirect
