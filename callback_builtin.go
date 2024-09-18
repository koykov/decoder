package decoder

import "fmt"

func cbPrint(_ *Ctx, args []any) error {
	fmt.Print(args...)
	return nil
}

func cbPrintln(_ *Ctx, args []any) error {
	fmt.Println(args...)
	return nil
}
