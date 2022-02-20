package random

import "github.com/ShiraazMoollatjie/goluhn"

func OrderID() string {
	return goluhn.Generate(16)
}
